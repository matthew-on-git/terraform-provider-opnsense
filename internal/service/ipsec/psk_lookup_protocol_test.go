// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

func TestProtocolPSKLookupReadOnly(t *testing.T) {
	ctx := context.Background()
	const id = "12345678-1234-1234-1234-123456789012"
	var calls []string
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls = append(calls, req.Method+" "+req.URL.Path)
		if req.Method != http.MethodGet {
			t.Errorf("lookup mutated appliance: %s %s", req.Method, req.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		switch req.URL.Path {
		case "/api/core/menu/search":
			_, _ = io.WriteString(w, `{}`)
		case "/api/ipsec/pre_shared_keys/get_item/" + id:
			_, _ = io.WriteString(w, `{"preSharedKey":{"ident":"local@example.com","remote_ident":"remote@example.com","keyType":{"PSK":{"value":"PSK","selected":1}},"Key":"not-lookup-state","description":"site tunnel","ignoredExtra":"not-lookup-state"}}`)
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(api.Close)

	server, schemas := pskLookupServer(t, api.URL)
	schema := schemas.DataSourceSchemas["opnsense_ipsec_psk"]
	if schema == nil {
		t.Fatal("ipsec psk lookup is not registered")
	}
	for _, name := range []string{"key", "Key", "pre_shared_key", "psk"} {
		if pskSchemaHas(schema, name) {
			t.Fatalf("excluded field %s is lookup state", name)
		}
	}
	read, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_ipsec_psk",
		Config:   pskLookupValue(t, schema, map[string]any{"id": id}),
	})
	if err != nil || pskLookupErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	attrs := pskLookupAttrs(t, schema, read.State)
	var got string
	if err := attrs["id"].As(&got); err != nil || got != id {
		t.Fatalf("id = %q %v", got, err)
	}
	if err := attrs["identity"].As(&got); err != nil || got != "local@example.com" {
		t.Fatalf("identity = %q %v", got, err)
	}
	if err := attrs["key_type"].As(&got); err != nil || got != "PSK" {
		t.Fatalf("key_type = %q %v", got, err)
	}
	if _, ok := attrs["key"]; ok {
		t.Fatal("key is lookup state")
	}
	before := len(calls)
	rejected, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_ipsec_psk",
		Config:   pskLookupValue(t, schema, map[string]any{"id": "not-a-uuid"}),
	})
	if err != nil || !pskLookupErrors(rejected.Diagnostics) || len(calls) != before {
		t.Fatalf("wrong id err=%v diagnostics=%v calls=%d/%d", err, rejected.Diagnostics, len(calls), before)
	}
	for _, call := range calls {
		if !strings.HasPrefix(call, "GET ") {
			t.Fatalf("mutation call %s", call)
		}
	}
}

func pskLookupServer(t *testing.T, uri string) (tfprotov6.ProviderServer, *tfprotov6.GetProviderSchemaResponse) {
	t.Helper()
	server := providerserver.NewProtocol6(provider.New("test")())()
	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil || pskLookupErrors(schemas.Diagnostics) {
		t.Fatal(err, schemas.Diagnostics)
	}
	config := pskLookupValue(t, schemas.Provider, map[string]any{
		"uri": uri, "api_key": "fixture-key", "api_secret": "fixture-secret", "insecure": true,
	})
	result, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: config, TerraformVersion: "1.14.0"})
	if err != nil || pskLookupErrors(result.Diagnostics) {
		t.Fatalf("configure: %v %+v", err, result.Diagnostics)
	}
	return server, schemas
}

func pskLookupValue(t *testing.T, schema *tfprotov6.Schema, fields map[string]any) *tfprotov6.DynamicValue {
	t.Helper()
	values := map[string]any{}
	for _, attr := range schema.Block.Attributes {
		values[attr.Name] = nil
	}
	for key, value := range fields {
		values[key] = value
	}
	raw, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	return &tfprotov6.DynamicValue{JSON: raw}
}

func pskLookupAttrs(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue) map[string]tftypes.Value {
	t.Helper()
	value, err := state.Unmarshal(schema.ValueType())
	if err != nil {
		t.Fatal(err)
	}
	var attrs map[string]tftypes.Value
	if err = value.As(&attrs); err != nil {
		t.Fatal(err)
	}
	return attrs
}

func pskLookupErrors(diagnostics []*tfprotov6.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func pskSchemaHas(schema *tfprotov6.Schema, name string) bool {
	for _, attr := range schema.Block.Attributes {
		if attr.Name == name {
			return true
		}
	}
	return false
}
