// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDHCPv6SettingsDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newDHCPv6SettingsDataSource())
	if len(schema.Attributes) != len(dhcpv6SettingsLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(dhcpv6SettingsLookupFields))
	}
	for _, name := range dhcpv6SettingsLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{
		"password", "secret", "key", "dhcp_socket_type", "mac_sources",
		"service_sockets_max_retries", "service_sockets_retry_wait_time",
		"decline_probation_period", "compatibility",
	} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestDHCPv6SettingsDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != dhcpv6SettingsReqOpts.GetEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(dhcpv6SettingsLookupFixture(false)))
	}))
	t.Cleanup(server.Close)

	state := readDHCPv6Settings(t, server.URL, dhcpv6SettingsID)
	if !state.Enabled.ValueBool() || state.ManualConfig.IsNull() || !state.ManualConfig.ValueBool() {
		t.Fatalf("flags = enabled %#v manual %#v", state.Enabled, state.ManualConfig)
	}
	if state.ValidLifetime.ValueInt64() != 7200 || state.FirewallRules.ValueBool() {
		t.Fatalf("non-default values were replaced: %#v", state)
	}
	if state.ID.ValueString() != dhcpv6SettingsID || !stringSetEquals(t, state.Interfaces, "lan", "opt1") {
		t.Fatalf("identity or interfaces changed: %#v", state)
	}
	if len(calls) != 1 || calls[0] != "GET "+dhcpv6SettingsReqOpts.GetEndpoint {
		t.Fatalf("calls = %v", calls)
	}
}

func TestDHCPv6SettingsDataSource_preservesDisabledAndEmptyInterfaces(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv6":{"general":{"enabled":"0","manual_config":"0","interfaces":[],"valid_lifetime":"4000","fwrules":"1","mac_sources":"ipv6-link-local"},"lexpire":{"hold_reclaimed_time":"3600"},"ha":{"enabled":"0"}}}`))
	}))
	t.Cleanup(server.Close)

	state := readDHCPv6Settings(t, server.URL, dhcpv6SettingsID)
	if state.Enabled.IsNull() || state.Enabled.ValueBool() || state.ManualConfig.IsNull() || state.ManualConfig.ValueBool() {
		t.Fatalf("disabled flags changed: enabled %#v manual %#v", state.Enabled, state.ManualConfig)
	}
	if state.FirewallRules.IsNull() || !state.FirewallRules.ValueBool() || !stringSetEquals(t, state.Interfaces) {
		t.Fatalf("empty interfaces or firewall flag changed: %#v", state)
	}
}

func TestDHCPv6SettingsDataSource_trimsManualConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv6":{"general":{"enabled":"1","manual_config":" 1 ","interfaces":[],"valid_lifetime":"7200","fwrules":"0"}}}`))
	}))
	t.Cleanup(server.Close)

	state := readDHCPv6Settings(t, server.URL, dhcpv6SettingsID)
	if state.ManualConfig.IsNull() || !state.ManualConfig.ValueBool() {
		t.Fatalf("padded manual_config was not true: %#v", state.ManualConfig)
	}
}

func TestDHCPv6SettingsDataSource_blankOptionalStaysNull(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv6":{"general":{"enabled":"1","manual_config":"","interfaces":[],"valid_lifetime":"7200","fwrules":"0"}}}`))
	}))
	t.Cleanup(server.Close)

	state := readDHCPv6Settings(t, server.URL, dhcpv6SettingsID)
	if !state.ManualConfig.IsNull() {
		t.Fatalf("blank optional manual_config became %v", state.ManualConfig)
	}
}

func TestDHCPv6SettingsDataSource_missingFieldFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv6":{"general":{"enabled":"1","manual_config":"0","interfaces":[],"fwrules":"1"}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readDHCPv6SettingsResponse(t, server.URL, dhcpv6SettingsID)
	if !resp.Diagnostics.HasError() || !diagnosticContains(resp, "valid_lifetime") {
		t.Fatalf("missing required lifetime must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestDHCPv6SettingsDataSource_doesNotSynthesizeDefaults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv6":{"general":{"enabled":"","manual_config":"","interfaces":[],"valid_lifetime":"","fwrules":"","dhcp_socket_type":""}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readDHCPv6SettingsResponse(t, server.URL, dhcpv6SettingsID)
	if !resp.Diagnostics.HasError() {
		t.Fatal("empty required fields must not become model defaults")
	}
}

func TestDHCPv6SettingsDataSource_ignoresUnreadableSocketType(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv6":{"general":{"enabled":"1","manual_config":"0","interfaces":[],"valid_lifetime":"7200","fwrules":"0","dhcp_socket_type":"raw","mac_sources":"ipv6-link-local","ignoredExtra":"not-lookup-state"}}}`))
	}))
	t.Cleanup(server.Close)

	state := readDHCPv6Settings(t, server.URL, dhcpv6SettingsID)
	if state.ValidLifetime.ValueInt64() != 7200 || state.FirewallRules.ValueBool() {
		t.Fatalf("non-allowlisted fields changed readable state: %#v", state)
	}
}

func TestDHCPv6SettingsDataSource_notFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv6":{"general":{}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readDHCPv6SettingsResponse(t, server.URL, dhcpv6SettingsID)
	if !resp.Diagnostics.HasError() || !diagnosticContains(resp, "not found") {
		t.Fatalf("empty general must be not-found before field validation, got %#v", resp.Diagnostics)
	}
}

func TestDHCPv6SettingsDataSource_wrongIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	resp := readDHCPv6SettingsResponse(t, server.URL, "kea-dhcpv4")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func dhcpv6SettingsLookupFixture(blankManual bool) string {
	manual := "1"
	if blankManual {
		manual = ""
	}
	return `{"dhcpv6":{"general":{"enabled":"1","manual_config":"` + manual + `","interfaces":{"lan":{"value":"LAN","selected":1},"opt1":{"value":"OPT1","selected":1},"wan":{"value":"WAN","selected":0}},"valid_lifetime":"7200","fwrules":"0","mac_sources":{"ipv6-link-local":{"value":"ipv6-link-local","selected":1}},"service_sockets_max_retries":"5","service_sockets_retry_wait_time":"5000","decline_probation_period":"600","dhcp_socket_type":"raw","ignoredExtra":"not-lookup-state"},"lexpire":{"hold_reclaimed_time":"3600"},"ha":{"enabled":"0"}}}`
}

func readDHCPv6Settings(t *testing.T, baseURL, id string) dhcpv6SettingsLookupModel {
	t.Helper()
	resp := readDHCPv6SettingsResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state dhcpv6SettingsLookupModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readDHCPv6SettingsResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newDHCPv6SettingsDataSource()
	configureDataSource(t, ds, testClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := dhcpv6SettingsLookupModel{ID: types.StringValue(id), Interfaces: types.SetNull(types.StringType)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}
