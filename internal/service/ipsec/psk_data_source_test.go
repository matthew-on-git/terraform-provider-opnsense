// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPSKDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newPSKDataSource())
	if len(schema.Attributes) != len(pskLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(pskLookupFields))
	}
	for _, name := range pskLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{"key", "Key", "pre_shared_key", "psk"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestPSKDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	const id = "12345678-1234-1234-1234-123456789012"
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != pskReqOpts.GetEndpoint+"/"+id {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"preSharedKey":{"ident":"local@example.com","remote_ident":"remote@example.com","keyType":{"PSK":{"value":"PSK","selected":1}},"Key":"not-lookup-state","description":"site tunnel","ignoredExtra":"not-lookup-state"}}`))
	}))
	t.Cleanup(server.Close)

	state := readPSK(t, server.URL, id)
	if state.Identity.ValueString() != "local@example.com" || state.RemoteIdentity.ValueString() != "remote@example.com" || state.KeyType.ValueString() != "PSK" || state.Description.ValueString() != "site tunnel" {
		t.Fatalf("unexpected state: %#v", state)
	}
	if len(calls) != 1 || calls[0] != "GET "+pskReqOpts.GetEndpoint+"/"+id {
		t.Fatalf("calls = %v", calls)
	}
}

func TestPSKDataSource_preservesEmptyOptionalFields(t *testing.T) {
	t.Parallel()

	const id = "12345678-1234-1234-1234-123456789012"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"preSharedKey":{"ident":"local@example.com","remote_ident":"","keyType":{"EAP":{"value":"EAP","selected":1}},"Key":"","description":""}}`))
	}))
	t.Cleanup(server.Close)

	state := readPSK(t, server.URL, id)
	if state.KeyType.ValueString() != "EAP" || state.RemoteIdentity.ValueString() != "" || state.Description.ValueString() != "" {
		t.Fatalf("empty optional fields changed: %#v", state)
	}
	if !state.RemoteIdentity.IsNull() && state.RemoteIdentity.ValueString() == "" && !state.Description.IsNull() {
		return
	}
	t.Fatal("empty optional fields must stay empty strings, not null or synthetic values")
}

func TestPSKDataSource_notFound(t *testing.T) {
	t.Parallel()

	const id = "12345678-1234-1234-1234-123456789012"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"preSharedKey":[]}`))
	}))
	t.Cleanup(server.Close)

	resp := readPSKResponse(t, server.URL, id)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected not-found error")
	}
}

func TestPSKDataSource_malformed(t *testing.T) {
	t.Parallel()

	const id = "12345678-1234-1234-1234-123456789012"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"preSharedKey":{"ident":"","keyType":[],"description":"site"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readPSKResponse(t, server.URL, id)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected malformed response error")
	}
}

func TestPSKDataSource_invalidUUIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	resp := readPSKResponse(t, server.URL, "not-a-uuid")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func readPSK(t *testing.T, baseURL, id string) pskLookupModel {
	t.Helper()
	resp := readPSKResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state pskLookupModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readPSKResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newPSKDataSource()
	configureDataSource(t, ds, newTestClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := pskLookupModel{ID: types.StringValue(id)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}
