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

func TestRIPDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newRIPDataSource())
	if len(schema.Attributes) != len(ripLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(ripLookupFields))
	}
	for _, name := range ripLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{"password", "secret", "key", "passive_interfaces", "passiveinterfaces", "ignoredExtra"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestRIPDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != ripReqOpts.GetEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(ripLookupFixture()))
	}))
	t.Cleanup(server.Close)

	state := readRIP(t, server.URL, ripID)
	if !state.Enabled.ValueBool() || state.Version.ValueInt64() != 2 || state.DefaultMetric.ValueInt64() != 5 || state.ID.ValueString() != ripID {
		t.Fatalf("identity fields changed: %#v", state)
	}
	if got := setMembers(t, state.Networks); len(got) != 2 || got[0] != "192.0.2.0/24" || got[1] != "198.51.100.0/25" {
		t.Fatalf("networks = %v", got)
	}
	if got := setMembers(t, state.Redistribute); len(got) != 2 || got[0] != "connected" || got[1] != "static" {
		t.Fatalf("redistribute = %v", got)
	}
	if len(calls) != 1 || calls[0] != "GET "+ripReqOpts.GetEndpoint {
		t.Fatalf("calls = %v", calls)
	}
}

func TestRIPDataSource_preservesDisabledAndEmptySelections(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{"enabled":"0","version":"1","networks":{"":{"value":"","selected":1}},"redistribute":{"bgp":{"value":"Border Gateway Protocol (BGP)","selected":0},"connected":{"value":"Connected","selected":0},"kernel":{"value":"Kernel","selected":0},"ospf":{"value":"OSPF","selected":0},"static":{"value":"Static","selected":0}},"defaultmetric":"","passiveinterfaces":{"wan":{"selected":1}},"ignoredExtra":"not-lookup-state"}}`))
	}))
	t.Cleanup(server.Close)

	state := readRIP(t, server.URL, ripID)
	if state.Enabled.IsNull() || state.Enabled.ValueBool() || state.Version.ValueInt64() != 1 || !state.DefaultMetric.IsNull() {
		t.Fatalf("disabled flag, version 1, or blank metric changed: %#v", state)
	}
	if got := setMembers(t, state.Networks); len(got) != 0 {
		t.Fatalf("empty stored networks = %v", got)
	}
	if got := setMembers(t, state.Redistribute); len(got) != 0 {
		t.Fatalf("unselected redistribute = %v", got)
	}
}

func TestRIPDataSource_omittedOptionalsStayEmpty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{"enabled":"1","version":"2"}}`))
	}))
	t.Cleanup(server.Close)

	state := readRIP(t, server.URL, ripID)
	if state.Version.ValueInt64() != 2 || !state.DefaultMetric.IsNull() || len(setMembers(t, state.Networks)) != 0 || len(setMembers(t, state.Redistribute)) != 0 {
		t.Fatalf("omitted optional fields were synthesized: %#v", state)
	}
}

func TestRIPDataSource_literalZeroFailsClosed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		body  string
		field string
	}{
		{`{"rip":{"enabled":"1","version":"0","networks":[],"redistribute":[],"defaultmetric":"1"}}`, "version"},
		{`{"rip":{"enabled":"1","version":"2","networks":[],"redistribute":[],"defaultmetric":"0"}}`, "default_metric"},
	}
	for _, tc := range cases {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(tc.body))
		}))
		resp := readRIPResponse(t, server.URL, ripID)
		server.Close()
		if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, tc.field) || bgpDiagnosticContains(resp, "not found") {
			t.Fatalf("literal 0 for %s must fail closed, got %#v", tc.field, resp.Diagnostics)
		}
	}
}

func TestRIPDataSource_versionAndMetricRanges(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{"enabled":"1","version":"1","networks":[],"redistribute":[],"defaultmetric":"16"}}`))
	}))
	t.Cleanup(server.Close)
	state := readRIP(t, server.URL, ripID)
	if state.Version.ValueInt64() != 1 || state.DefaultMetric.ValueInt64() != 16 {
		t.Fatalf("boundary values changed: %#v", state)
	}

	above := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{"enabled":"1","version":"3","networks":[],"redistribute":[],"defaultmetric":"17"}}`))
	}))
	t.Cleanup(above.Close)
	resp := readRIPResponse(t, above.URL, ripID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "version") {
		t.Fatalf("version above maximum must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestRIPDataSource_metricAboveMaximumFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{"enabled":"1","version":"2","networks":[],"redistribute":[],"defaultmetric":"17"}}`))
	}))
	t.Cleanup(server.Close)
	resp := readRIPResponse(t, server.URL, ripID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "default_metric") {
		t.Fatalf("metric above maximum must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestRIPDataSource_trimsTokensOnce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{"enabled":" 0 ","version":" 2 ","networks":[],"redistribute":[],"defaultmetric":" 8 "}}`))
	}))
	t.Cleanup(server.Close)
	state := readRIP(t, server.URL, ripID)
	if state.Enabled.ValueBool() || state.Version.ValueInt64() != 2 || state.DefaultMetric.ValueInt64() != 8 {
		t.Fatalf("padded tokens changed: %#v", state)
	}
}

func TestRIPDataSource_invalidNetworkFailsClosed(t *testing.T) {
	t.Parallel()

	for _, network := range []string{"::ffff:192.0.2.1/128", "192.0.2.1", "192.0.2.0/33", "010.0.0.0/8", "192.0.2.0/024"} {
		body := `{"rip":{"enabled":"1","version":"2","networks":{"` + network + `":{"selected":1}},"redistribute":[],"defaultmetric":""}}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(body))
		}))
		resp := readRIPResponse(t, server.URL, ripID)
		server.Close()
		if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "networks") || bgpDiagnosticContains(resp, "not found") {
			t.Fatalf("network %s must fail closed, got %#v", network, resp.Diagnostics)
		}
	}
}

func TestRIPDataSource_unknownRedistributeFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{"enabled":"1","version":"2","networks":[],"redistribute":{" connected ":{"selected":1}},"defaultmetric":""}}`))
	}))
	t.Cleanup(server.Close)
	resp := readRIPResponse(t, server.URL, ripID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "redistribute") {
		t.Fatalf("padded redistribute token must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestRIPDataSource_missingFieldFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{"version":"2","networks":[],"redistribute":[],"defaultmetric":""}}`))
	}))
	t.Cleanup(server.Close)
	resp := readRIPResponse(t, server.URL, ripID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("missing required enabled must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestRIPDataSource_doesNotSynthesizeDefaults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{"enabled":"","version":"","networks":[],"redistribute":[],"defaultmetric":""}}`))
	}))
	t.Cleanup(server.Close)
	resp := readRIPResponse(t, server.URL, ripID)
	if !resp.Diagnostics.HasError() || bgpDiagnosticContains(resp, "not found") {
		t.Fatal("empty required fields must not become model defaults or not-found")
	}
}

func TestRIPDataSource_notFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{}}`))
	}))
	t.Cleanup(server.Close)
	resp := readRIPResponse(t, server.URL, ripID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("empty RIP object must be not-found before field validation, got %#v", resp.Diagnostics)
	}
}

func TestRIPDataSource_paddedEmptyObjectIsMalformed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{ }}`))
	}))
	t.Cleanup(server.Close)
	resp := readRIPResponse(t, server.URL, ripID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("whitespace-padded empty object must be malformed, not not-found, got %#v", resp.Diagnostics)
	}
}

func TestRIPDataSource_nullScalarIsMalformed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"rip":{"enabled":"1","version":null,"networks":[],"redistribute":[],"defaultmetric":""}}`))
	}))
	t.Cleanup(server.Close)
	resp := readRIPResponse(t, server.URL, ripID)
	if !resp.Diagnostics.HasError() || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("null version must be malformed, not not-found, got %#v", resp.Diagnostics)
	}
}

func TestRIPDataSource_wrongIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)
	resp := readRIPResponse(t, server.URL, "ospf")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func ripLookupFixture() string {
	return `{"rip":{"enabled":"1","version":"2","networks":{"198.51.100.0/25":{"value":"198.51.100.0/25","selected":1},"192.0.2.0/24":{"value":"192.0.2.0/24","selected":1},"203.0.113.0/24":{"value":"203.0.113.0/24","selected":0}},"redistribute":{"bgp":{"value":"Border Gateway Protocol (BGP)","selected":0},"connected":{"value":"Connected","selected":1},"kernel":{"value":"Kernel","selected":0},"ospf":{"value":"OSPF","selected":0},"static":{"value":"Static","selected":1}},"defaultmetric":"5","passiveinterfaces":{"wan":{"selected":1}},"ignoredExtra":"not-lookup-state"}}`
}

func readRIP(t *testing.T, baseURL, id string) ripLookupModel {
	t.Helper()
	resp := readRIPResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state ripLookupModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readRIPResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newRIPDataSource()
	configureDataSource(t, ds, newTestClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := ripLookupModel{
		ID:           types.StringValue(id),
		Networks:     types.SetNull(types.StringType),
		Redistribute: types.SetNull(types.StringType),
	}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}

func setMembers(t *testing.T, value types.Set) []string {
	t.Helper()
	if value.IsNull() || value.IsUnknown() {
		t.Fatalf("set is null or unknown: %#v", value)
	}
	var got []string
	if diags := value.ElementsAs(context.Background(), &got, false); diags.HasError() {
		t.Fatalf("set elements: %v", diags)
	}
	return got
}
