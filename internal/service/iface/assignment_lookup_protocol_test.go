// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface_test

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

func TestProtocolAssignmentLookupReadOnly(t *testing.T) {
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
		case "/api/core/firmware/info":
			_, _ = io.WriteString(w, `{"product_version":"26.7.9"}`)
		case "/api/interfaces/assignment/get_item/opt3":
			_ = json.NewEncoder(w).Encode(map[string]any{"interface": protocolAssignment()})
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(api.Close)

	server, schemas := assignmentLookupServer(t, api.URL)
	schema := schemas.DataSourceSchemas["opnsense_interface_assignment"]
	if schema == nil {
		t.Fatal("interface assignment lookup is not registered")
	}
	for _, name := range []string{"password", "descr", "spoofmac", "if", "mtu"} {
		if assignmentSchemaHas(schema, name) {
			t.Fatalf("excluded field %s is lookup state", name)
		}
	}
	read, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_interface_assignment",
		Config:   assignmentLookupValue(t, schema, map[string]any{"id": "opt3"}),
	})
	if err != nil || assignmentLookupErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	attrs := assignmentLookupAttrs(t, schema, read.State)
	var id, address string
	if err := attrs["id"].As(&id); err != nil || id != "opt3" {
		t.Fatalf("id = %q %v", id, err)
	}
	if err := attrs["name"].As(&address); err != nil || address != "opt3" {
		t.Fatalf("name = %q %v", address, err)
	}
	if err := attrs["ipv4_address"].As(&address); err != nil || address != "192.0.2.10" {
		t.Fatalf("ipv4_address = %q %v", address, err)
	}
	if _, ok := attrs["password"]; ok {
		t.Fatal("password is lookup state")
	}
	before := len(calls)
	rejected, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_interface_assignment",
		Config:   assignmentLookupValue(t, schema, map[string]any{"id": "em0"}),
	})
	if err != nil || !assignmentLookupErrors(rejected.Diagnostics) || len(calls) != before {
		t.Fatalf("wrong id err=%v diagnostics=%v calls=%d/%d", err, rejected.Diagnostics, len(calls), before)
	}
	for _, call := range calls {
		if !strings.HasPrefix(call, "GET ") {
			t.Fatalf("mutation call %s", call)
		}
	}
}

func protocolAssignment() map[string]any {
	selected := func(value string) map[string]any {
		return map[string]any{value: map[string]any{"value": value, "selected": 1}}
	}
	return map[string]any{
		"identifier": "opt3", "if": selected("vtnet2"), "descr": "example", "lock": "1", "enable": "1",
		"spoofmac": "02:00:00:00:00:01", "pending_action": "not-lookup-state",
		"type4": selected("staticv4"), "ipaddr": "192.0.2.10/24", "gateway": "GW4",
		"type6": selected(""), "ipaddrv6": "", "gatewayv6": "",
		"dhcphostname": "", "dhcprejectfrom": "", "dhcphonourmtu": "0", "dhcpvlanprio": "",
		"dhcp_send_options": "", "dhcp_request_options": "", "dhcp6vlanprio": "",
		"dhcp6_request_dns": "0", "dhcp6-information-only": "0", "dhcp6_rapid_commit": "0",
		"dhcp6_send_options": "", "dhcp6_request_options": "",
	}
}

func assignmentLookupServer(t *testing.T, uri string) (tfprotov6.ProviderServer, *tfprotov6.GetProviderSchemaResponse) {
	t.Helper()
	server := providerserver.NewProtocol6(provider.New("test")())()
	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil || assignmentLookupErrors(schemas.Diagnostics) {
		t.Fatal(err, schemas.Diagnostics)
	}
	config := assignmentLookupValue(t, schemas.Provider, map[string]any{
		"uri": uri, "api_key": "fixture-key", "api_secret": "fixture-secret", "insecure": true,
	})
	result, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: config, TerraformVersion: "1.14.0"})
	if err != nil || assignmentLookupErrors(result.Diagnostics) {
		t.Fatalf("configure: %v %+v", err, result.Diagnostics)
	}
	return server, schemas
}

func assignmentLookupValue(t *testing.T, schema *tfprotov6.Schema, fields map[string]any) *tfprotov6.DynamicValue {
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

func assignmentLookupAttrs(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue) map[string]tftypes.Value {
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

func assignmentLookupErrors(diagnostics []*tfprotov6.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func assignmentSchemaHas(schema *tfprotov6.Schema, name string) bool {
	for _, attr := range schema.Block.Attributes {
		if attr.Name == name {
			return true
		}
	}
	return false
}
