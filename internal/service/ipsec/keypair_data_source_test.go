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

func TestKeyPairDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newKeyPairDataSource())
	if len(schema.Attributes) != len(keyPairLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(keyPairLookupFields))
	}
	for _, name := range keyPairLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{"public_key", "private_key", "publicKey", "privateKey"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestKeyPairDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	const id = "12345678-1234-1234-1234-123456789012"
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != keyPairReqOpts.GetEndpoint+"/"+id {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"keyPair":{"name":"example-pair","keyType":{"rsa":{"value":"RSA","selected":1}},"publicKey":"not-lookup-state","privateKey":"not-lookup-state","keySize":"2048","keyFingerprint":"aa:bb","ignoredExtra":"not-lookup-state"}}`))
	}))
	t.Cleanup(server.Close)

	state := readKeyPair(t, server.URL, id)
	if state.Name.ValueString() != "example-pair" || state.KeyType.ValueString() != "rsa" || state.KeySize.ValueString() != "2048" || state.KeyFingerprint.ValueString() != "aa:bb" {
		t.Fatalf("unexpected state: %#v", state)
	}
	if len(calls) != 1 || calls[0] != "GET "+keyPairReqOpts.GetEndpoint+"/"+id {
		t.Fatalf("calls = %v", calls)
	}
}

func TestKeyPairDataSource_preservesEmptyDerivedFields(t *testing.T) {
	t.Parallel()

	const id = "12345678-1234-1234-1234-123456789012"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"keyPair":{"name":"peer-public","keyType":{"ecdsa":{"value":"ECDSA","selected":1}},"publicKey":"not-lookup-state","privateKey":"","keySize":"","keyFingerprint":""}}`))
	}))
	t.Cleanup(server.Close)

	state := readKeyPair(t, server.URL, id)
	if state.KeyType.ValueString() != "ecdsa" || state.KeySize.ValueString() != "" || state.KeyFingerprint.ValueString() != "" {
		t.Fatalf("empty derived fields changed: %#v", state)
	}
	if !state.KeySize.IsNull() && state.KeySize.ValueString() == "" && !state.KeyFingerprint.IsNull() {
		return
	}
	t.Fatal("empty derived fields must stay empty strings, not null or synthetic values")
}

func TestKeyPairDataSource_notFound(t *testing.T) {
	t.Parallel()

	const id = "12345678-1234-1234-1234-123456789012"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"keyPair":[]}`))
	}))
	t.Cleanup(server.Close)

	resp := readKeyPairResponse(t, server.URL, id)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected not-found error")
	}
}

func TestKeyPairDataSource_malformed(t *testing.T) {
	t.Parallel()

	const id = "12345678-1234-1234-1234-123456789012"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"keyPair":{"name":"","keyType":[],"keySize":"2048","keyFingerprint":"aa:bb"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readKeyPairResponse(t, server.URL, id)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected malformed response error")
	}
}

func TestKeyPairDataSource_invalidUUIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	resp := readKeyPairResponse(t, server.URL, "not-a-uuid")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func readKeyPair(t *testing.T, baseURL, id string) keyPairLookupModel {
	t.Helper()
	resp := readKeyPairResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state keyPairLookupModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readKeyPairResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newKeyPairDataSource()
	configureDataSource(t, ds, newTestClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := keyPairLookupModel{ID: types.StringValue(id)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}
