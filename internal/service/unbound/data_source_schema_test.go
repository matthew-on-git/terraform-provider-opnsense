// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

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
	if len(constructors) != 4 {
		t.Fatalf("expected 4 Unbound data sources, got %d", len(constructors))
	}
	for _, constructor := range constructors {
		assertRequiredID(t, constructor())
	}
}

func TestHostOverrideDataSource_read(t *testing.T) {
	t.Parallel()

	const id = "host-uuid"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != hostOverrideReqOpts.GetEndpoint+"/"+id {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"host":{"enabled":"1","hostname":"www","domain":"example.com","rr":{"A":{"value":"A","selected":1}},"server":"192.0.2.10","description":"web"}}`))
	}))
	t.Cleanup(server.Close)

	client := testClient(t, server.URL)
	ds := newHostOverrideDataSource()
	configureDataSource(t, ds, client)
	schema := dataSourceSchema(t, ds)
	req := datasource.ReadRequest{Config: modelConfig(t, schema, &HostOverrideResourceModel{ID: types.StringValue(id)})}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}

	ds.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state HostOverrideResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	if state.Hostname.ValueString() != "www" || state.RR.ValueString() != "A" || !state.Enabled.ValueBool() {
		t.Fatalf("unexpected state: %#v", state)
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
