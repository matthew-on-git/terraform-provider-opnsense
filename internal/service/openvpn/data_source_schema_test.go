// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestDataSources_schemaIDs(t *testing.T) {
	t.Parallel()

	constructors := DataSources()
	if len(constructors) != 2 {
		t.Fatalf("expected 2 OpenVPN data sources, got %d", len(constructors))
	}

	for _, constructor := range constructors {
		assertRequiredID(t, constructor())
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

func TestInstanceDataSource_read(t *testing.T) {
	t.Parallel()

	const id = "instance-uuid"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != instanceReqOpts.GetEndpoint+"/"+id {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"instance":{"vpnid":"srv1","enabled":"1","role":{"server":{"selected":1}},"description":"roadwarrior","dev_type":{"tun":{"selected":1}},"proto":{"udp":{"selected":1}},"port":"1194","local":"","remote":[],"server":"10.8.0.0/24","topology":{"subnet":{"selected":1}},"ca":{"ca":{"selected":1}},"cert":{"cert":{"selected":1}},"verify_client_cert":{"required":{"selected":1}},"tls_key":{"tls":{"selected":1}},"data-ciphers":{"AES-256-GCM":{"selected":1}},"auth":{"SHA256":{"selected":1}},"dns_servers":{"192.0.2.53":{"selected":1}},"push_route":{"192.0.2.0/24":{"selected":1}},"redirect_gateway":{"def1":{"selected":1}},"maxclients":"50","keepalive_interval":"10","keepalive_timeout":"60","verb":{"3":{"selected":1}}}}`))
	}))
	t.Cleanup(server.Close)

	ds := newInstanceDataSource()
	configureDataSource(t, ds, newTestClient(t, server.URL))
	schema := dataSourceSchema(t, ds)
	config := InstanceResourceModel{
		ID:              types.StringValue(id),
		DataCiphers:     types.SetNull(types.StringType),
		DNSServers:      types.SetNull(types.StringType),
		PushRoute:       types.SetNull(types.StringType),
		RedirectGateway: types.SetNull(types.StringType),
	}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}

	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}

	var state InstanceResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	if state.VPNID.ValueString() != "srv1" || state.Role.ValueString() != "server" || state.MaxClients.ValueInt64() != 50 {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func newTestClient(t *testing.T, baseURL string) *opnsense.Client {
	t.Helper()
	client, err := opnsense.NewClient(opnsense.ClientConfig{BaseURL: baseURL, APIKey: "key", APISecret: "secret", RetryMax: 1})
	if err != nil {
		t.Fatal(err)
	}
	return client
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
