// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestDataSources_schemaIDs(t *testing.T) {
	t.Parallel()

	constructors := DataSources()
	if len(constructors) != 11 {
		t.Fatalf("expected 11 Kea data sources, got %d", len(constructors))
	}
	for _, constructor := range constructors {
		var metadata datasource.MetadataResponse
		constructor().Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "opnsense"}, &metadata)
		if metadata.TypeName == "opnsense_kea_ddns_settings" {
			var singleton datasource.SchemaResponse
			constructor().Schema(context.Background(), datasource.SchemaRequest{}, &singleton)
			id, ok := singleton.Schema.Attributes["id"].(dsschema.StringAttribute)
			if !ok || !id.Computed || id.Required || id.Optional {
				t.Fatal("singleton lookup must expose only a computed fixed ID")
			}
			continue
		}
		assertRequiredID(t, constructor())
	}
}

func TestDHCPv4SubnetDataSource_read(t *testing.T) {
	t.Parallel()

	const id = "subnet4-uuid"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != dhcpv4SubnetReqOpts.GetEndpoint+"/"+id {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"subnet4":{"subnet":"192.168.40.0/24","allocator":{"random":{"value":"Random","selected":1}},"pools":"192.168.40.100-192.168.40.254","match-client-id":"1","ping_check":"0","option_data_autocollect":"0","option_data":{"domain_name_servers":{},"domain_search":{},"routers":{},"ntp_servers":{},"time_servers":{}},"description":"guest"}}`))
	}))
	t.Cleanup(server.Close)

	client := testClient(t, server.URL)
	ds := newDHCPv4SubnetDataSource()
	configureDataSource(t, ds, client)
	schema := dataSourceSchema(t, ds)
	req := datasource.ReadRequest{Config: modelConfig(t, schema, &DHCPv4SubnetResourceModel{
		ID: types.StringValue(id), DNSServers: types.SetNull(types.StringType), DomainSearch: types.SetNull(types.StringType),
		Routers: types.SetNull(types.StringType), StaticRoutes: types.SetNull(types.ObjectType{AttrTypes: staticRouteAttrTypes}),
		NTPServers: types.SetNull(types.StringType), TimeServers: types.SetNull(types.StringType),
	})}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}

	ds.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state DHCPv4SubnetResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	if state.Subnet.ValueString() != "192.168.40.0/24" || state.Pools.ValueString() != "192.168.40.100-192.168.40.254" {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestDHCPv4ReservationDataSource_read(t *testing.T) {
	t.Parallel()

	const id = "reservation4-uuid"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != dhcpv4ReservationReqOpts.GetEndpoint+"/"+id {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"reservation":{"subnet":{"subnet4-uuid":{"value":"192.168.40.0/24","selected":1}},"ip_address":"192.168.40.8","hw_address":"00:11:22:33:44:55","client_id":"","hostname":"ps5","description":"tfacc"}}`))
	}))
	t.Cleanup(server.Close)

	client := testClient(t, server.URL)
	ds := newDHCPv4ReservationDataSource()
	configureDataSource(t, ds, client)
	schema := dataSourceSchema(t, ds)
	req := datasource.ReadRequest{Config: modelConfig(t, schema, &DHCPv4ReservationResourceModel{ID: types.StringValue(id)})}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}

	ds.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state DHCPv4ReservationResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	if state.SubnetID.ValueString() != "subnet4-uuid" || state.HWAddress.ValueString() != "00:11:22:33:44:55" {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestDHCPv6SubnetDataSource_read(t *testing.T) {
	t.Parallel()

	const id = "subnet6-uuid"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != dhcpv6SubnetReqOpts.GetEndpoint+"/"+id {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"subnet6":{"subnet":"2001:db8::/64","interface":{"lan":{"value":"LAN","selected":1}},"allocator":{"random":{"value":"Random","selected":1}},"pd-allocator":{"random":{"value":"Random","selected":1}},"pools":"2001:db8::100-2001:db8::1ff","option_data":{"dns_servers":{},"domain_search":{}},"description":"lan6"}}`))
	}))
	t.Cleanup(server.Close)

	client := testClient(t, server.URL)
	ds := newDHCPv6SubnetDataSource()
	configureDataSource(t, ds, client)
	schema := dataSourceSchema(t, ds)
	req := datasource.ReadRequest{Config: modelConfig(t, schema, &DHCPv6SubnetResourceModel{ID: types.StringValue(id), DNSServers: types.SetNull(types.StringType), DomainSearch: types.SetNull(types.StringType)})}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}

	ds.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state DHCPv6SubnetResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	if state.Subnet.ValueString() != "2001:db8::/64" || state.Interface.ValueString() != "lan" {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestSubnetDataSourcesRejectMissingOptionListChildren(t *testing.T) {
	t.Run("DHCPv4", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"subnet4":{"subnet":"192.0.2.0/24","allocator":{},"match-client-id":"1","ping_check":"0","option_data_autocollect":"0","option_data":{"domain_name_servers":{},"domain_search":{},"routers":{},"ntp_servers":{}}}}`)
		}))
		defer server.Close()
		ds := newDHCPv4SubnetDataSource()
		configureDataSource(t, ds, testClient(t, server.URL))
		schema := dataSourceSchema(t, ds)
		request := datasource.ReadRequest{Config: modelConfig(t, schema, &DHCPv4SubnetResourceModel{
			ID: types.StringValue("id"), DNSServers: types.SetNull(types.StringType), DomainSearch: types.SetNull(types.StringType),
			Routers: types.SetNull(types.StringType), StaticRoutes: types.SetNull(types.ObjectType{AttrTypes: staticRouteAttrTypes}),
			NTPServers: types.SetNull(types.StringType), TimeServers: types.SetNull(types.StringType),
		})}
		response := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
		ds.Read(context.Background(), request, &response)
		if !response.Diagnostics.HasError() || !strings.Contains(response.Diagnostics.Errors()[0].Detail(), "time_servers") {
			t.Fatalf("missing option child diagnostics: %v", response.Diagnostics)
		}
	})

	t.Run("DHCPv6", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, `{"subnet6":{"subnet":"2001:db8::/64","interface":{"lan":{"selected":1}},"option_data":{"dns_servers":{}}}}`)
		}))
		defer server.Close()
		ds := newDHCPv6SubnetDataSource()
		configureDataSource(t, ds, testClient(t, server.URL))
		schema := dataSourceSchema(t, ds)
		request := datasource.ReadRequest{Config: modelConfig(t, schema, &DHCPv6SubnetResourceModel{
			ID: types.StringValue("id"), DNSServers: types.SetNull(types.StringType), DomainSearch: types.SetNull(types.StringType),
		})}
		response := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
		ds.Read(context.Background(), request, &response)
		if !response.Diagnostics.HasError() || !strings.Contains(response.Diagnostics.Errors()[0].Detail(), "domain_search") {
			t.Fatalf("missing option child diagnostics: %v", response.Diagnostics)
		}
	})
}

func TestDHCPv6PDPoolPrefixDescriptionsDistinguishResourceAndDataSource(t *testing.T) {
	var resourceResponse resource.SchemaResponse
	(&dhcpv6PdPoolResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resourceResponse)
	resourcePrefix := resourceResponse.Schema.Attributes["prefix"].GetMarkdownDescription()
	if !strings.Contains(resourcePrefix, "address or CIDR") {
		t.Fatalf("resource prefix description = %q", resourcePrefix)
	}

	var dataSourceResponse datasource.SchemaResponse
	(&dhcpv6PdPoolDataSource{}).Schema(context.Background(), datasource.SchemaRequest{}, &dataSourceResponse)
	dataSourcePrefix := dataSourceResponse.Schema.Attributes["prefix"].GetMarkdownDescription()
	if !strings.Contains(dataSourcePrefix, "address-only") {
		t.Fatalf("data-source prefix description = %q", dataSourcePrefix)
	}
}

func assertRequiredID(t *testing.T, ds datasource.DataSource) {
	t.Helper()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	id, ok := resp.Schema.Attributes["id"].(dsschema.StringAttribute)
	if !ok {
		t.Fatalf("id attribute is %T, want StringAttribute", resp.Schema.Attributes["id"])
	}
	if !id.Required {
		t.Fatal("id attribute must be required")
	}
}

func configureDataSource(t *testing.T, ds datasource.DataSource, client *opnsense.Client) {
	t.Helper()
	configurable, ok := ds.(datasource.DataSourceWithConfigure)
	if !ok {
		t.Fatal("data source does not implement Configure")
	}
	var resp datasource.ConfigureResponse
	configurable.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: client}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("configure diagnostics: %v", resp.Diagnostics)
	}
}

func dataSourceSchema(t *testing.T, ds datasource.DataSource) dsschema.Schema {
	t.Helper()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

func modelConfig(t *testing.T, schema dsschema.Schema, model any) tfsdk.Config {
	t.Helper()
	state := tfsdk.State{Schema: schema}
	if diags := state.Set(context.Background(), model); diags.HasError() {
		t.Fatalf("config diagnostics: %v", diags)
	}
	return tfsdk.Config{Raw: state.Raw, Schema: schema}
}

func testClient(t *testing.T, baseURL string) *opnsense.Client {
	t.Helper()
	client, err := opnsense.NewClient(opnsense.ClientConfig{BaseURL: baseURL, APIKey: "key", APISecret: "secret", RetryMax: 1})
	if err != nil {
		t.Fatal(err)
	}
	client.HTTPClient().Transport = firmwareTestTransport{base: client.HTTPClient().Transport}
	return client
}

type firmwareTestTransport struct{ base http.RoundTripper }

func (t firmwareTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Path == keaFirmwareStatusEndpoint {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"product":{"product_version":"26.7.7"}}`)), Request: request}, nil
	}
	return t.base.RoundTrip(request)
}
