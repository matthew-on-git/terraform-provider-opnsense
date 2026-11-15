// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestGeneralDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newGeneralDataSource())
	if len(schema.Attributes) != len(generalLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(generalLookupFields))
	}
	for _, name := range generalLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{"password", "secret", "key", "manual_config", "daemons", "ignoredExtra"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestGeneralDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != generalReqOpts.GetEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(generalLookupFixture()))
	}))
	t.Cleanup(server.Close)

	state := readGeneral(t, server.URL, generalID)
	if !state.Enabled.ValueBool() || state.Profile.ValueString() != "datacenter" {
		t.Fatalf("identity fields changed: %#v", state)
	}
	if state.EnableCARP.ValueBool() || !state.EnableSyslog.ValueBool() || state.EnableSNMP.ValueBool() || !state.FirewallRules.ValueBool() {
		t.Fatalf("flags changed: %#v", state)
	}
	if state.SyslogLevel.ValueString() != "debugging" || state.ID.ValueString() != generalID {
		t.Fatalf("syslog or identity changed: %#v", state)
	}
	if len(calls) != 1 || calls[0] != "GET "+generalReqOpts.GetEndpoint {
		t.Fatalf("calls = %v", calls)
	}
}

func TestGeneralDataSource_preservesDisabledFlags(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"general":{"enabled":"0","profile":"traditional","enablecarp":"0","enablesyslog":"0","enablesnmp":"0","sysloglevel":"alerts","fwrules":"0","manual_config":"1","daemons":{"bgp":{"selected":1}}}}`))
	}))
	t.Cleanup(server.Close)

	state := readGeneral(t, server.URL, generalID)
	if state.Enabled.IsNull() || state.Enabled.ValueBool() || state.EnableSyslog.IsNull() || state.EnableSyslog.ValueBool() || state.FirewallRules.ValueBool() {
		t.Fatalf("disabled flags were synthesized from model defaults: %#v", state)
	}
	if state.Profile.ValueString() != "traditional" || state.SyslogLevel.ValueString() != "alerts" {
		t.Fatalf("required options changed: %#v", state)
	}
}

func TestGeneralDataSource_trimsBoolTokensOnce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"general":{"enabled":" 1 ","profile":"datacenter","enablecarp":" 0 ","enablesyslog":" 1 ","enablesnmp":" 0 ","sysloglevel":"informational","fwrules":" 1 "}}`))
	}))
	t.Cleanup(server.Close)

	state := readGeneral(t, server.URL, generalID)
	if !state.Enabled.ValueBool() || state.EnableCARP.ValueBool() || !state.FirewallRules.ValueBool() || state.Profile.ValueString() != "datacenter" {
		t.Fatalf("padded bool tokens changed: %#v", state)
	}
}

func TestGeneralDataSource_paddedOptionFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"general":{"enabled":"1","profile":" datacenter ","enablecarp":"0","enablesyslog":"1","enablesnmp":"0","sysloglevel":"notifications","fwrules":"1"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readGeneralResponse(t, server.URL, generalID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "malformed") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("padded profile must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestGeneralDataSource_readsSelectedOptions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"general":{"enabled":"1","profile":{"traditional":{"value":"traditional","selected":0},"datacenter":{"value":"datacenter","selected":1}},"enablecarp":"1","enablesyslog":"1","enablesnmp":"0","sysloglevel":{"notifications":{"selected":0},"debugging":{"selected":"1"}},"fwrules":"1"}}`))
	}))
	t.Cleanup(server.Close)

	state := readGeneral(t, server.URL, generalID)
	if state.Profile.ValueString() != "datacenter" || state.SyslogLevel.ValueString() != "debugging" || !state.EnableCARP.ValueBool() {
		t.Fatalf("unselected option became lookup state: %#v", state)
	}
}

func TestGeneralDataSource_ignoresNonAllowlistedFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"general":{"enabled":"1","profile":"traditional","enablecarp":"0","enablesyslog":"1","enablesnmp":"0","sysloglevel":"notifications","fwrules":"1","manual_config":"1","daemons":{"ospf":{"selected":1}},"ignoredExtra":"not-lookup-state"}}`))
	}))
	t.Cleanup(server.Close)

	state := readGeneral(t, server.URL, generalID)
	if state.Profile.ValueString() != "traditional" || !state.EnableSyslog.ValueBool() {
		t.Fatalf("non-allowlisted fields changed readable state: %#v", state)
	}
}

func TestGeneralDataSource_missingFieldFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"general":{"enabled":"1","enablecarp":"0","enablesyslog":"1","enablesnmp":"0","sysloglevel":"notifications","fwrules":"1"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readGeneralResponse(t, server.URL, generalID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "profile") {
		t.Fatalf("missing required profile must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestGeneralDataSource_doesNotSynthesizeDefaults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"general":{"enabled":"","profile":"","enablecarp":"","enablesyslog":"","enablesnmp":"","sysloglevel":"","fwrules":""}}`))
	}))
	t.Cleanup(server.Close)

	resp := readGeneralResponse(t, server.URL, generalID)
	if !resp.Diagnostics.HasError() || bgpDiagnosticContains(resp, "not found") {
		t.Fatal("empty required fields must not become model defaults or not-found")
	}
}

func TestGeneralDataSource_notFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"general":{}}`))
	}))
	t.Cleanup(server.Close)

	resp := readGeneralResponse(t, server.URL, generalID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("empty general object must be not-found before field validation, got %#v", resp.Diagnostics)
	}
}

func TestGeneralDataSource_paddedEmptyObjectIsMalformed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"general":{ }}`))
	}))
	t.Cleanup(server.Close)

	resp := readGeneralResponse(t, server.URL, generalID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("whitespace-padded empty object must be malformed, not not-found, got %#v", resp.Diagnostics)
	}
}

func TestGeneralDataSource_wrongIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	resp := readGeneralResponse(t, server.URL, "bgp")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func generalLookupFixture() string {
	return `{"general":{"enabled":"1","profile":{"traditional":{"value":"traditional","selected":0},"datacenter":{"value":"datacenter","selected":1}},"enablecarp":"0","enablesyslog":"1","enablesnmp":"0","sysloglevel":{"notifications":{"selected":0},"debugging":{"selected":1}},"fwrules":"1","manual_config":"0","daemons":{"bgp":{"selected":1},"ospf":{"selected":0}},"ignoredExtra":"not-lookup-state"}}`
}

func readGeneral(t *testing.T, baseURL, id string) generalLookupModel {
	t.Helper()
	resp := readGeneralResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state generalLookupModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readGeneralResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newGeneralDataSource()
	configureDataSource(t, ds, newTestClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := generalLookupModel{ID: types.StringValue(id)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}
