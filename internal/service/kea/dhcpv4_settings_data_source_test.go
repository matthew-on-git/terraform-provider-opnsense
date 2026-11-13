// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDHCPv4SettingsDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newDHCPv4SettingsDataSource())
	if len(schema.Attributes) != len(dhcpv4SettingsLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(dhcpv4SettingsLookupFields))
	}
	for _, name := range dhcpv4SettingsLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{
		"password", "secret", "key", "service_sockets_max_retries",
		"service_sockets_retry_wait_time", "decline_probation_period", "compatibility",
	} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestDHCPv4SettingsDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != dhcpv4SettingsReqOpts.GetEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(dhcpv4SettingsLookupFixture(false)))
	}))
	t.Cleanup(server.Close)

	state := readDHCPv4Settings(t, server.URL, dhcpv4SettingsID)
	if !state.Enabled.ValueBool() || state.ManualConfig.IsNull() || !state.ManualConfig.ValueBool() {
		t.Fatalf("flags = enabled %#v manual %#v", state.Enabled, state.ManualConfig)
	}
	if state.ValidLifetime.ValueInt64() != 7200 || state.FirewallRules.ValueBool() || state.DHCPSocketType.ValueString() != "udp" {
		t.Fatalf("non-default values were replaced: %#v", state)
	}
	if state.ID.ValueString() != dhcpv4SettingsID || !stringSetEquals(t, state.Interfaces, "lan", "opt1") {
		t.Fatalf("identity or interfaces changed: %#v", state)
	}
	if len(calls) != 1 || calls[0] != "GET "+dhcpv4SettingsReqOpts.GetEndpoint {
		t.Fatalf("calls = %v", calls)
	}
}

func TestDHCPv4SettingsDataSource_preservesDisabledAndEmptyInterfaces(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv4":{"general":{"enabled":"0","manual_config":"0","interfaces":[],"valid_lifetime":"4000","fwrules":"1","dhcp_socket_type":"raw"},"lexpire":{"hold_reclaimed_time":"3600"},"ha":{"enabled":"0"}}}`))
	}))
	t.Cleanup(server.Close)

	state := readDHCPv4Settings(t, server.URL, dhcpv4SettingsID)
	if state.Enabled.IsNull() || state.Enabled.ValueBool() || state.ManualConfig.IsNull() || state.ManualConfig.ValueBool() {
		t.Fatalf("disabled flags changed: enabled %#v manual %#v", state.Enabled, state.ManualConfig)
	}
	if state.FirewallRules.IsNull() || !state.FirewallRules.ValueBool() || !stringSetEquals(t, state.Interfaces) {
		t.Fatalf("empty interfaces or firewall flag changed: %#v", state)
	}
}

func TestDHCPv4SettingsDataSource_trimsManualConfig(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv4":{"general":{"enabled":"1","manual_config":" 1 ","interfaces":[],"valid_lifetime":"7200","fwrules":"0","dhcp_socket_type":"udp"}}}`))
	}))
	t.Cleanup(server.Close)

	state := readDHCPv4Settings(t, server.URL, dhcpv4SettingsID)
	if state.ManualConfig.IsNull() || !state.ManualConfig.ValueBool() {
		t.Fatalf("padded manual_config was not true: %#v", state.ManualConfig)
	}
}

func TestDHCPv4SettingsDataSource_blankOptionalStaysNull(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv4":{"general":{"enabled":"1","manual_config":"","interfaces":[],"valid_lifetime":"7200","fwrules":"0","dhcp_socket_type":"udp"}}}`))
	}))
	t.Cleanup(server.Close)

	state := readDHCPv4Settings(t, server.URL, dhcpv4SettingsID)
	if !state.ManualConfig.IsNull() {
		t.Fatalf("blank optional manual_config became %v", state.ManualConfig)
	}
}

func TestDHCPv4SettingsDataSource_missingFieldFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv4":{"general":{"enabled":"1","manual_config":"0","interfaces":[],"fwrules":"1","dhcp_socket_type":"raw"}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readDHCPv4SettingsResponse(t, server.URL, dhcpv4SettingsID)
	if !resp.Diagnostics.HasError() || !diagnosticContains(resp, "valid_lifetime") {
		t.Fatalf("missing required lifetime must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestDHCPv4SettingsDataSource_doesNotSynthesizeDefaults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv4":{"general":{"enabled":"","manual_config":"","interfaces":[],"valid_lifetime":"","fwrules":"","dhcp_socket_type":""}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readDHCPv4SettingsResponse(t, server.URL, dhcpv4SettingsID)
	if !resp.Diagnostics.HasError() {
		t.Fatal("empty required fields must not become model defaults")
	}
}

func TestDHCPv4SettingsDataSource_notFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dhcpv4":{"general":{}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readDHCPv4SettingsResponse(t, server.URL, dhcpv4SettingsID)
	if !resp.Diagnostics.HasError() || !diagnosticContains(resp, "not found") {
		t.Fatalf("empty general must be not-found before field validation, got %#v", resp.Diagnostics)
	}
}

func TestDHCPv4SettingsDataSource_wrongIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	resp := readDHCPv4SettingsResponse(t, server.URL, "kea-ctrl-agent")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func dhcpv4SettingsLookupFixture(blankManual bool) string {
	manual := "1"
	if blankManual {
		manual = ""
	}
	return `{"dhcpv4":{"general":{"enabled":"1","manual_config":"` + manual + `","interfaces":{"lan":{"value":"LAN","selected":1},"opt1":{"value":"OPT1","selected":1},"wan":{"value":"WAN","selected":0}},"valid_lifetime":"7200","fwrules":"0","dhcp_socket_type":{"udp":{"value":"UDP","selected":1},"raw":{"value":"RAW","selected":0}},"service_sockets_max_retries":"3","service_sockets_retry_wait_time":"2","decline_probation_period":"86400","compatibility":{"lenient-option-parsing":{"selected":1}},"ignoredExtra":"not-lookup-state"},"lexpire":{"hold_reclaimed_time":"3600"},"ha":{"enabled":"0","%this_server_name":"example"}}}`
}

func readDHCPv4Settings(t *testing.T, baseURL, id string) DHCPv4SettingsResourceModel {
	t.Helper()
	resp := readDHCPv4SettingsResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state DHCPv4SettingsResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readDHCPv4SettingsResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newDHCPv4SettingsDataSource()
	configureDataSource(t, ds, testClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := DHCPv4SettingsResourceModel{ID: types.StringValue(id), Interfaces: types.SetNull(types.StringType)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}

func diagnosticContains(resp datasource.ReadResponse, needle string) bool {
	for _, diagnostic := range resp.Diagnostics {
		if strings.Contains(strings.ToLower(diagnostic.Summary()), strings.ToLower(needle)) || strings.Contains(strings.ToLower(diagnostic.Detail()), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func stringSetEquals(t *testing.T, set types.Set, want ...string) bool {
	t.Helper()
	if set.IsNull() || set.IsUnknown() {
		return false
	}
	var got []string
	if diags := set.ElementsAs(context.Background(), &got, false); diags.HasError() {
		t.Fatal(diags)
	}
	if len(got) != len(want) {
		return false
	}
	seen := map[string]bool{}
	for _, value := range got {
		seen[value] = true
	}
	for _, value := range want {
		if !seen[value] {
			return false
		}
	}
	return true
}
