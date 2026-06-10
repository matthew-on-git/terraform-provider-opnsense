// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

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
	if len(constructors) != 7 {
		t.Fatalf("expected 7 IPsec data sources, got %d", len(constructors))
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

func TestConnectionDataSource_read(t *testing.T) {
	t.Parallel()

	const id = "connection-uuid"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != connectionReqOpts.GetEndpoint+"/"+id {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"connection":{"enabled":"1","description":"site tunnel","remote_addrs":{"203.0.113.10":{"selected":1}},"version":[{"selected":0},{"selected":0},{"selected":1}],"proposals":{"aes256-sha256-modp2048":{"selected":1}},"unique":{"replace":{"selected":1}}}}`))
	}))
	t.Cleanup(server.Close)

	ds := newConnectionDataSource()
	configureDataSource(t, ds, newTestClient(t, server.URL))
	schema := dataSourceSchema(t, ds)
	config := ConnectionResourceModel{ID: types.StringValue(id)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}

	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}

	var state ConnectionResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	if state.RemoteAddrs.ValueString() != "203.0.113.10" || state.Version.ValueString() != "2" || state.Unique.ValueString() != "replace" {
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
