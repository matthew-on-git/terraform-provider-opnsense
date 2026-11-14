// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"context"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

func TestProtocolBGPGlobalLookupReadOnly(t *testing.T) {
	ctx := context.Background()
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
		case "/api/quagga/bgp/get":
			_, _ = io.WriteString(w, `{"bgp":{"enabled":"1","asnumber":"64496","routerid":"192.0.2.1","distance":"","graceful":"0","networkimportcheck":"1","enforce_first_as":"1","logneighborchanges":"0","networks":{"192.0.2.0/24":{"selected":1},"198.51.100.0/24":{"selected":0}},"maximumpaths":"4","maximumpathsibgp":"","bestpath":{"compare-routerid":{"selected":1}},"ignoredExtra":"not-lookup-state"}}`)
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(api.Close)

	server, schemas := bgpGlobalLookupServer(t, api.URL)
	schema := schemas.DataSourceSchemas["opnsense_quagga_bgp_global"]
	if schema == nil {
		t.Fatal("BGP global lookup is not registered")
	}
	for _, name := range []string{"password", "secret", "key", "bestpath", "neighbors", "compatibility", "ignoredExtra"} {
		if bgpGlobalSchemaHas(schema, name) {
			t.Fatalf("excluded field %s is lookup state", name)
		}
	}
	read, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_quagga_bgp_global",
		Config:   bgpGlobalLookupValue(t, schema, map[string]any{"id": "bgp"}),
	})
	if err != nil || bgpGlobalLookupErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	attrs := bgpGlobalLookupAttrs(t, schema, read.State)
	var got string
	var enabled bool
	if err := attrs["id"].As(&got); err != nil || got != "bgp" {
		t.Fatalf("id = %q %v", got, err)
	}
	if err := attrs["enabled"].As(&enabled); err != nil || !enabled {
		t.Fatalf("enabled = %v %v", enabled, err)
	}
	var asNumber big.Float
	if err := attrs["as_number"].As(&asNumber); err != nil || asNumber.Cmp(big.NewFloat(64496)) != 0 {
		t.Fatalf("as_number = %s %v", asNumber.String(), err)
	}
	if !attrs["distance"].IsNull() {
		t.Fatal("blank distance was synthesized")
	}
	var paths big.Float
	if err := attrs["maximum_paths"].As(&paths); err != nil || paths.Cmp(big.NewFloat(4)) != 0 {
		t.Fatalf("maximum_paths = %s %v", paths.String(), err)
	}
	if _, ok := attrs["bestpath"]; ok {
		t.Fatal("bestpath is lookup state")
	}
	if _, ok := attrs["ignoredExtra"]; ok {
		t.Fatal("non-allowlisted field is lookup state")
	}
	before := len(calls)
	rejected, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_quagga_bgp_global",
		Config:   bgpGlobalLookupValue(t, schema, map[string]any{"id": "not-bgp"}),
	})
	if err != nil || !bgpGlobalLookupErrors(rejected.Diagnostics) || len(calls) != before {
		t.Fatalf("wrong id err=%v diagnostics=%v calls=%d/%d", err, rejected.Diagnostics, len(calls), before)
	}
	for _, call := range calls {
		if !strings.HasPrefix(call, "GET ") {
			t.Fatalf("mutation call %s", call)
		}
	}
}

func bgpGlobalLookupServer(t *testing.T, uri string) (tfprotov6.ProviderServer, *tfprotov6.GetProviderSchemaResponse) {
	t.Helper()
	server := providerserver.NewProtocol6(provider.New("test")())()
	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil || bgpGlobalLookupErrors(schemas.Diagnostics) {
		t.Fatal(err, schemas.Diagnostics)
	}
	config := bgpGlobalLookupValue(t, schemas.Provider, map[string]any{
		"uri": uri, "api_key": "fixture-key", "api_secret": "fixture-secret", "insecure": true,
	})
	result, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: config, TerraformVersion: "1.14.0"})
	if err != nil || bgpGlobalLookupErrors(result.Diagnostics) {
		t.Fatalf("configure: %v %+v", err, result.Diagnostics)
	}
	return server, schemas
}

func bgpGlobalLookupValue(t *testing.T, schema *tfprotov6.Schema, fields map[string]any) *tfprotov6.DynamicValue {
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

func bgpGlobalLookupAttrs(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue) map[string]tftypes.Value {
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

func bgpGlobalLookupErrors(diagnostics []*tfprotov6.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func bgpGlobalSchemaHas(schema *tfprotov6.Schema, name string) bool {
	for _, attr := range schema.Block.Attributes {
		if attr.Name == name {
			return true
		}
	}
	return false
}
