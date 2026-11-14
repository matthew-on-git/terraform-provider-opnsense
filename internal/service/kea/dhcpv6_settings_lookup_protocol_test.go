// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea_test

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

func TestProtocolDHCPv6SettingsLookupReadOnly(t *testing.T) {
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
		case "/api/kea/dhcpv6/get":
			_, _ = io.WriteString(w, `{"dhcpv6":{"general":{"enabled":"1","manual_config":"0","interfaces":{"lan":{"value":"LAN","selected":1}},"valid_lifetime":"7200","fwrules":"0","mac_sources":"ipv6-link-local","dhcp_socket_type":"raw","ignoredExtra":"not-lookup-state"},"ha":{"enabled":"0"}}}`)
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(api.Close)

	server, schemas := dhcpv6SettingsLookupServer(t, api.URL)
	schema := schemas.DataSourceSchemas["opnsense_kea_dhcpv6_settings"]
	if schema == nil {
		t.Fatal("kea dhcpv6 settings lookup is not registered")
	}
	for _, name := range []string{"password", "secret", "key", "dhcp_socket_type", "mac_sources", "compatibility", "ignoredExtra"} {
		if dhcpv6SettingsSchemaHas(schema, name) {
			t.Fatalf("excluded field %s is lookup state", name)
		}
	}
	read, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_kea_dhcpv6_settings",
		Config:   dhcpv6SettingsLookupValue(t, schema, map[string]any{"id": "kea-dhcpv6"}),
	})
	if err != nil || dhcpv6SettingsLookupErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	attrs := dhcpv6SettingsLookupAttrs(t, schema, read.State)
	var got string
	var enabled bool
	if err := attrs["id"].As(&got); err != nil || got != "kea-dhcpv6" {
		t.Fatalf("id = %q %v", got, err)
	}
	if err := attrs["enabled"].As(&enabled); err != nil || !enabled {
		t.Fatalf("enabled = %v %v", enabled, err)
	}
	var lifetime big.Float
	if err := attrs["valid_lifetime"].As(&lifetime); err != nil || lifetime.Cmp(big.NewFloat(7200)) != 0 {
		t.Fatalf("valid_lifetime = %s %v", lifetime.String(), err)
	}
	if _, ok := attrs["dhcp_socket_type"]; ok {
		t.Fatal("unreadable socket type is lookup state")
	}
	if _, ok := attrs["ignoredExtra"]; ok {
		t.Fatal("non-allowlisted field is lookup state")
	}
	before := len(calls)
	rejected, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_kea_dhcpv6_settings",
		Config:   dhcpv6SettingsLookupValue(t, schema, map[string]any{"id": "kea-dhcpv4"}),
	})
	if err != nil || !dhcpv6SettingsLookupErrors(rejected.Diagnostics) || len(calls) != before {
		t.Fatalf("wrong id err=%v diagnostics=%v calls=%d/%d", err, rejected.Diagnostics, len(calls), before)
	}
	for _, call := range calls {
		if !strings.HasPrefix(call, "GET ") {
			t.Fatalf("mutation call %s", call)
		}
	}
}

func dhcpv6SettingsLookupServer(t *testing.T, uri string) (tfprotov6.ProviderServer, *tfprotov6.GetProviderSchemaResponse) {
	t.Helper()
	server := providerserver.NewProtocol6(provider.New("test")())()
	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil || dhcpv6SettingsLookupErrors(schemas.Diagnostics) {
		t.Fatal(err, schemas.Diagnostics)
	}
	config := dhcpv6SettingsLookupValue(t, schemas.Provider, map[string]any{
		"uri": uri, "api_key": "fixture-key", "api_secret": "fixture-secret", "insecure": true,
	})
	result, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: config, TerraformVersion: "1.14.0"})
	if err != nil || dhcpv6SettingsLookupErrors(result.Diagnostics) {
		t.Fatalf("configure: %v %+v", err, result.Diagnostics)
	}
	return server, schemas
}

func dhcpv6SettingsLookupValue(t *testing.T, schema *tfprotov6.Schema, fields map[string]any) *tfprotov6.DynamicValue {
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

func dhcpv6SettingsLookupAttrs(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue) map[string]tftypes.Value {
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

func dhcpv6SettingsLookupErrors(diagnostics []*tfprotov6.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func dhcpv6SettingsSchemaHas(schema *tfprotov6.Schema, name string) bool {
	for _, attr := range schema.Block.Attributes {
		if attr.Name == name {
			return true
		}
	}
	return false
}
