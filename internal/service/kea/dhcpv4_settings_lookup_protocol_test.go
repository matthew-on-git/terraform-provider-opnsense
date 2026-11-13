// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea_test

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

func TestProtocolDHCPv4SettingsLookupReadOnly(t *testing.T) {
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
		case "/api/kea/dhcpv4/get":
			_, _ = io.WriteString(w, `{"dhcpv4":{"general":{"enabled":"1","manual_config":"0","interfaces":{"lan":{"value":"LAN","selected":1}},"valid_lifetime":"7200","fwrules":"0","dhcp_socket_type":"udp","ignoredExtra":"not-lookup-state"},"ha":{"enabled":"0"}}}`)
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(api.Close)

	server, schemas := dhcpv4SettingsLookupServer(t, api.URL)
	schema := schemas.DataSourceSchemas["opnsense_kea_dhcpv4_settings"]
	if schema == nil {
		t.Fatal("kea dhcpv4 settings lookup is not registered")
	}
	for _, name := range []string{"password", "secret", "key", "service_sockets_max_retries", "compatibility", "ignoredExtra"} {
		if dhcpv4SettingsSchemaHas(schema, name) {
			t.Fatalf("excluded field %s is lookup state", name)
		}
	}
	read, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_kea_dhcpv4_settings",
		Config:   dhcpv4SettingsLookupValue(t, schema, map[string]any{"id": "kea-dhcpv4"}),
	})
	if err != nil || dhcpv4SettingsLookupErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	attrs := dhcpv4SettingsLookupAttrs(t, schema, read.State)
	var got string
	var enabled bool
	if err := attrs["id"].As(&got); err != nil || got != "kea-dhcpv4" {
		t.Fatalf("id = %q %v", got, err)
	}
	if err := attrs["enabled"].As(&enabled); err != nil || !enabled {
		t.Fatalf("enabled = %v %v", enabled, err)
	}
	if err := attrs["dhcp_socket_type"].As(&got); err != nil || got != "udp" {
		t.Fatalf("dhcp_socket_type = %q %v", got, err)
	}
	if _, ok := attrs["ignoredExtra"]; ok {
		t.Fatal("non-allowlisted field is lookup state")
	}
	before := len(calls)
	rejected, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_kea_dhcpv4_settings",
		Config:   dhcpv4SettingsLookupValue(t, schema, map[string]any{"id": "kea-ctrl-agent"}),
	})
	if err != nil || !dhcpv4SettingsLookupErrors(rejected.Diagnostics) || len(calls) != before {
		t.Fatalf("wrong id err=%v diagnostics=%v calls=%d/%d", err, rejected.Diagnostics, len(calls), before)
	}
	for _, call := range calls {
		if !strings.HasPrefix(call, "GET ") {
			t.Fatalf("mutation call %s", call)
		}
	}
}

func dhcpv4SettingsLookupServer(t *testing.T, uri string) (tfprotov6.ProviderServer, *tfprotov6.GetProviderSchemaResponse) {
	t.Helper()
	server := providerserver.NewProtocol6(provider.New("test")())()
	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil || dhcpv4SettingsLookupErrors(schemas.Diagnostics) {
		t.Fatal(err, schemas.Diagnostics)
	}
	config := dhcpv4SettingsLookupValue(t, schemas.Provider, map[string]any{
		"uri": uri, "api_key": "fixture-key", "api_secret": "fixture-secret", "insecure": true,
	})
	result, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: config, TerraformVersion: "1.14.0"})
	if err != nil || dhcpv4SettingsLookupErrors(result.Diagnostics) {
		t.Fatalf("configure: %v %+v", err, result.Diagnostics)
	}
	return server, schemas
}

func dhcpv4SettingsLookupValue(t *testing.T, schema *tfprotov6.Schema, fields map[string]any) *tfprotov6.DynamicValue {
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

func dhcpv4SettingsLookupAttrs(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue) map[string]tftypes.Value {
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

func dhcpv4SettingsLookupErrors(diagnostics []*tfprotov6.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func dhcpv4SettingsSchemaHas(schema *tfprotov6.Schema, name string) bool {
	for _, attr := range schema.Block.Attributes {
		if attr.Name == name {
			return true
		}
	}
	return false
}
