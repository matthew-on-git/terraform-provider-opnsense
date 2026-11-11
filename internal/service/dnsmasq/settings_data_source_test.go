// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dnsmasq

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestSettingsDataSourceSchemaAllowlist(t *testing.T) {
	t.Parallel()

	var resp datasource.SchemaResponse
	newSettingsDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if len(resp.Schema.Attributes) != len(settingsLookupFields) {
		t.Fatalf("schema attributes = %d, allowlist = %d", len(resp.Schema.Attributes), len(settingsLookupFields))
	}
	for _, name := range settingsLookupFields {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("allowlist field %q missing from schema", name)
		}
		if attr.IsSensitive() || attr.IsWriteOnly() {
			t.Fatalf("%s is sensitive or write-only; this lookup has no secret field", name)
		}
		if name == "id" {
			id, ok := attr.(dsschema.StringAttribute)
			if !ok || !id.Required || id.Computed || id.Optional {
				t.Fatal("id must be the required singleton identity")
			}
			continue
		}
		if !attr.IsComputed() || attr.IsRequired() || attr.IsOptional() {
			t.Fatalf("%s must be computed-only", name)
		}
	}
	for name := range resp.Schema.Attributes {
		if !slices.Contains(settingsLookupFields, name) {
			t.Fatalf("schema field %q is outside the allowlist", name)
		}
	}
	nested, ok := resp.Schema.Attributes["dhcp"].(dsschema.SingleNestedAttribute)
	if !ok || !nested.Computed {
		t.Fatal("dhcp must be a computed nested object")
	}
	for _, excluded := range []string{"expand_hosts", "%domain", "hosts", "domain"} {
		if _, ok := resp.Schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %q is lookup state", excluded)
		}
	}
	if _, ok := nested.Attributes["%domain"]; ok {
		t.Fatal("response metadata %domain is lookup state")
	}
}

func TestSettingsDataSourceReadAllowlist(t *testing.T) {
	t.Parallel()

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != "/api/dnsmasq/settings/get" {
			t.Errorf("lookup mutated appliance: %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"dnsmasq":{"enable":"1","regdhcp":"0","regdhcpstatic":"1","expand_hosts":"1","port":"","dns_port":"53","interface":{"lan":{"value":"LAN label","selected":1},"wan":{"value":"WAN label","selected":0}},"add_mac":{"standard":{"value":"Standard","selected":1}},"hosts":{"item":{"host":"not-settings"}},"dhcp":` + dhcpFixture + `}}`))
	}))
	t.Cleanup(server.Close)

	ds := &settingsDataSource{client: lookupClient(t, server.URL)}
	schema := dataSourceSchema(t, ds)
	req := datasource.ReadRequest{Config: lookupConfig(t, schema, settingsID)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	var state SettingsResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatal(diags)
	}
	if state.ID.ValueString() != settingsID || !state.Enable.ValueBool() || state.RegisterDHCP.ValueBool() || !state.RegisterDHCPStatic.ValueBool() {
		t.Fatalf("identity or flags lost: %#v", state)
	}
	if state.Port.ValueString() != "" || state.DNSPort.ValueString() != "53" || state.AddMAC.ValueString() != "standard" {
		t.Fatalf("string allowlist lost: port=%q dns_port=%q add_mac=%q", state.Port.ValueString(), state.DNSPort.ValueString(), state.AddMAC.ValueString())
	}
	var interfaces []string
	if diags := state.Interface.ElementsAs(context.Background(), &interfaces, false); diags.HasError() || !slices.Equal(interfaces, []string{"lan"}) {
		t.Fatalf("interface selection lost: %v %v", interfaces, diags)
	}
	dhcp := state.DHCP.Attributes()
	if !dhcp["lease_max"].IsNull() || !dhcp["domain"].Equal(types.StringValue("")) || !dhcp["fqdn"].Equal(types.BoolValue(true)) {
		t.Fatalf("dhcp null/empty boundary lost: %v", dhcp)
	}
	if _, ok := dhcp["%domain"]; ok {
		t.Fatal("published response metadata")
	}
	if len(calls) != 1 || calls[0] != "GET /api/dnsmasq/settings/get" {
		t.Fatalf("calls = %v", calls)
	}
}

func TestSettingsDataSourceRejectsWrongIDWithoutRead(t *testing.T) {
	t.Parallel()

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	t.Cleanup(server.Close)
	ds := &settingsDataSource{client: lookupClient(t, server.URL)}
	schema := dataSourceSchema(t, ds)
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: lookupConfig(t, schema, "other")}, &resp)
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("diagnostics=%v calls=%d", resp.Diagnostics, calls)
	}
}

func TestSettingsDataSourceMissingDHCPFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected method %s", r.Method)
		}
		_, _ = w.Write([]byte(`{"dnsmasq":{"enable":"1"}}`))
	}))
	t.Cleanup(server.Close)
	ds := &settingsDataSource{client: lookupClient(t, server.URL)}
	schema := dataSourceSchema(t, ds)
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: lookupConfig(t, schema, settingsID)}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("missing dhcp published state")
	}
}

func lookupClient(t *testing.T, baseURL string) *opnsense.Client {
	t.Helper()
	client, err := opnsense.NewClient(opnsense.ClientConfig{
		BaseURL:      baseURL,
		APIKey:       "fixture-key",
		APISecret:    "fixture-secret",
		Insecure:     true,
		RetryMax:     1,
		RetryWaitMin: time.Nanosecond,
		RetryWaitMax: time.Nanosecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func dataSourceSchema(t *testing.T, ds datasource.DataSource) dsschema.Schema {
	t.Helper()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	return resp.Schema
}

func lookupConfig(t *testing.T, schema dsschema.Schema, id string) tfsdk.Config {
	t.Helper()
	state := tfsdk.State{Schema: schema}
	model := &SettingsResourceModel{
		ID:        types.StringValue(id),
		Interface: types.SetNull(types.StringType),
		DHCP:      types.ObjectNull(dhcpAttributeTypes()),
	}
	if diags := state.Set(context.Background(), model); diags.HasError() {
		t.Fatal(diags)
	}
	return tfsdk.Config{Raw: state.Raw, Schema: schema}
}
