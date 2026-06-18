// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient

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
		t.Fatalf("expected 2 ddclient data sources, got %d", len(constructors))
	}
	ds := constructors[0]()
	assertRequiredID(t, ds)
	if _, ok := dataSourceSchema(t, ds).Attributes["password"]; ok {
		t.Fatal("ddclient_account data source must not expose password")
	}
	settings := constructors[1]()
	assertRequiredID(t, settings)
}

func TestResources_includeSettings(t *testing.T) {
	t.Parallel()

	constructors := Resources()
	if len(constructors) != 2 {
		t.Fatalf("expected 2 ddclient resources, got %d", len(constructors))
	}
}

func TestAccountDataSource_read(t *testing.T) {
	t.Parallel()

	const id = "account-uuid"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != accountReqOpts.GetEndpoint+"/"+id {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"account":{"enabled":"1","service":{"cloudflare":{"value":"Cloudflare","selected":1}},"hostnames":"host.example.com","username":"user","password":"secret","description":"dyn"}}`))
	}))
	t.Cleanup(server.Close)

	client := testClient(t, server.URL)
	ds := newAccountDataSource()
	configureDataSource(t, ds, client)
	schema := dataSourceSchema(t, ds)
	req := datasource.ReadRequest{Config: modelConfig(t, schema, &accountDataSourceModel{ID: types.StringValue(id)})}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}

	ds.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state accountDataSourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	if state.Service.ValueString() != "cloudflare" || state.Username.ValueString() != "user" {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestSettingsDataSource_read(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != ddclientSettingsReqOpts.GetEndpoint {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"ddclient":{"general":{"enabled":"1","backend":{"ddclient":{"value":"ddclient","selected":1}},"daemon_delay":"300","verbose":"0","allowipv6":"1"}}}`))
	}))
	t.Cleanup(server.Close)

	client := testClient(t, server.URL)
	ds := newSettingsDataSource()
	configureDataSource(t, ds, client)
	schema := dataSourceSchema(t, ds)
	req := datasource.ReadRequest{Config: modelConfig(t, schema, &SettingsResourceModel{ID: types.StringValue(settingsID)})}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}

	ds.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state SettingsResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	if state.ID.ValueString() != settingsID || state.Backend.ValueString() != "ddclient" || state.Interval.ValueInt64() != 300 || !state.AllowIPv6.ValueBool() {
		t.Fatalf("unexpected state: %#v", state)
	}
}

func TestSettingsDataSource_rejectsInvalidID(t *testing.T) {
	t.Parallel()

	ds := newSettingsDataSource()
	schema := dataSourceSchema(t, ds)
	req := datasource.ReadRequest{Config: modelConfig(t, schema, &SettingsResourceModel{ID: types.StringValue("wrong")})}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}

	ds.Read(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected invalid singleton ID diagnostic")
	}
}

func TestSettingsReqOpts_useDdclientGeneralMonad(t *testing.T) {
	t.Parallel()

	if ddclientSettingsReqOpts.Monad != "ddclient.general" {
		t.Fatalf("unexpected settings monad %q", ddclientSettingsReqOpts.Monad)
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
	return client
}
