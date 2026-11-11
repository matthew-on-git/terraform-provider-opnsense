// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dnsmasq_test

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestProtocolDNSMasqSettingsLookupReadOnly(t *testing.T) {
	ctx := context.Background()
	calls := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/core/menu/search" {
			_, _ = fmt.Fprint(w, `{}`)
			return
		}
		calls++
		if r.Method != http.MethodGet || r.URL.Path != "/api/dnsmasq/settings/get" {
			t.Errorf("lookup mutated appliance: %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprintf(w, "%s", `{"dnsmasq":{"enable":"0","port":"5353","expand_hosts":"1","dhcp":`+protocolDHCP+`}}`)
	}))
	t.Cleanup(api.Close)

	server := lookupProviderServer(t, api.URL)
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil || dnsmasqErrors(schemas.Diagnostics) {
		t.Fatal(err, schemas.Diagnostics)
	}
	schema := schemas.DataSourceSchemas["opnsense_dnsmasq_settings"]
	if schema == nil {
		t.Fatal("dnsmasq settings lookup is not registered")
	}
	read, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_dnsmasq_settings",
		Config:   dnsmasqValue(t, schema, map[string]any{"id": "dnsmasq"}),
	})
	if err != nil || dnsmasqErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	attrs := lookupAttrs(t, schema, read.State)
	var id, port string
	var enable bool
	if err := attrs["id"].As(&id); err != nil || id != "dnsmasq" {
		t.Fatalf("id = %q %v", id, err)
	}
	if err := attrs["enable"].As(&enable); err != nil || enable {
		t.Fatalf("enable = %v %v", enable, err)
	}
	if err := attrs["port"].As(&port); err != nil || port != "5353" {
		t.Fatalf("port = %q %v", port, err)
	}
	if _, ok := attrs["expand_hosts"]; ok {
		t.Fatal("expand_hosts is lookup state")
	}
	dhcp, ok := attrs["dhcp"]
	if !ok || dhcp.IsNull() {
		t.Fatal("dhcp object missing")
	}
	var nested map[string]tftypes.Value
	if err := dhcp.As(&nested); err != nil {
		t.Fatal(err)
	}
	var leaseMax big.Float
	if err := nested["lease_max"].As(&leaseMax); err != nil || leaseMax.Cmp(big.NewFloat(128)) != 0 {
		t.Fatalf("lease_max = %v %v", nested["lease_max"], err)
	}
	if _, ok := nested["%domain"]; ok {
		t.Fatal("response metadata published")
	}
	if calls != 1 {
		t.Fatalf("settings calls = %d, want 1 GET", calls)
	}

	rejected, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{
		TypeName: "opnsense_dnsmasq_settings",
		Config:   dnsmasqValue(t, schema, map[string]any{"id": "other"}),
	})
	if err != nil || !dnsmasqErrors(rejected.Diagnostics) || calls != 1 {
		t.Fatalf("wrong id err=%v diagnostics=%v calls=%d", err, rejected.Diagnostics, calls)
	}
}

func lookupProviderServer(t *testing.T, uri string) tfprotov6.ProviderServer {
	t.Helper()
	server, _ := dnsmasqServer(t, uri)
	return server
}

func lookupAttrs(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue) map[string]tftypes.Value {
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
