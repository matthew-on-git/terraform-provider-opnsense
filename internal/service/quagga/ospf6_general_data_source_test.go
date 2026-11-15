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

func TestOSPF6GeneralDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newOSPF6GeneralDataSource())
	if len(schema.Attributes) != len(ospf6GeneralLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(ospf6GeneralLookupFields))
	}
	for _, name := range ospf6GeneralLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{"password", "secret", "key", "areas", "interfaces", "prefixlists", "routemaps", "redistributions", "ignoredExtra"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestOSPF6GeneralDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != ospf6GeneralReqOpts.GetEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(ospf6GeneralLookupFixture()))
	}))
	t.Cleanup(server.Close)

	state := readOSPF6General(t, server.URL, ospf6GeneralID)
	if !state.Enabled.ValueBool() || state.RouterID.ValueString() != "192.0.2.1" || state.OriginateMetric.ValueInt64() != 100 {
		t.Fatalf("identity fields changed: %#v", state)
	}
	if state.Originate.ValueBool() || state.OriginateAlways.ValueBool() || state.CARPDemote.ValueBool() || state.ID.ValueString() != ospf6GeneralID {
		t.Fatalf("flags or identity changed: %#v", state)
	}
	if len(calls) != 1 || calls[0] != "GET "+ospf6GeneralReqOpts.GetEndpoint {
		t.Fatalf("calls = %v", calls)
	}
}

func TestOSPF6GeneralDataSource_preservesDisabledFlagsAndZeroMetric(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf6":{"enabled":"0","routerid":"","originate":"0","originatealways":"0","originatemetric":"0","carp_demote":"0","areas":{"area":{}},"ignoredExtra":"not-lookup-state"}}`))
	}))
	t.Cleanup(server.Close)

	state := readOSPF6General(t, server.URL, ospf6GeneralID)
	if state.Enabled.IsNull() || state.Enabled.ValueBool() || state.Originate.ValueBool() || state.CARPDemote.ValueBool() {
		t.Fatalf("disabled flags were synthesized from model defaults: %#v", state)
	}
	if !state.RouterID.IsNull() || state.OriginateMetric.IsNull() || state.OriginateMetric.ValueInt64() != 0 {
		t.Fatalf("blank router ID or literal metric 0 changed: %#v", state)
	}
}

func TestOSPF6GeneralDataSource_blankMetricStaysNull(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf6":{"enabled":"1","routerid":"198.51.100.1","originate":"1","originatealways":"0","originatemetric":"","carp_demote":"1"}}`))
	}))
	t.Cleanup(server.Close)

	state := readOSPF6General(t, server.URL, ospf6GeneralID)
	if !state.OriginateMetric.IsNull() || state.RouterID.ValueString() != "198.51.100.1" || !state.CARPDemote.ValueBool() {
		t.Fatalf("blank metric was synthesized as 0: %#v", state)
	}
}

func TestOSPF6GeneralDataSource_trimsBoolTokensOnce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf6":{"enabled":" 1 ","routerid":" 192.0.2.10 ","originate":" 0 ","originatealways":" 1 ","originatemetric":" 20 ","carp_demote":" 0 "}}`))
	}))
	t.Cleanup(server.Close)

	state := readOSPF6General(t, server.URL, ospf6GeneralID)
	if !state.Enabled.ValueBool() || state.CARPDemote.ValueBool() || !state.OriginateAlways.ValueBool() || state.RouterID.ValueString() != "192.0.2.10" || state.OriginateMetric.ValueInt64() != 20 {
		t.Fatalf("padded tokens changed: %#v", state)
	}
}

func TestOSPF6GeneralDataSource_ignoresNonAllowlistedFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(ospf6GeneralLookupFixture()))
	}))
	t.Cleanup(server.Close)

	state := readOSPF6General(t, server.URL, ospf6GeneralID)
	if state.RouterID.ValueString() != "192.0.2.1" || !state.Enabled.ValueBool() {
		t.Fatalf("non-allowlisted fields changed readable state: %#v", state)
	}
}

func TestOSPF6GeneralDataSource_missingFieldFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf6":{"routerid":"192.0.2.1","originate":"0","originatealways":"0","originatemetric":"1","carp_demote":"0"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPF6GeneralResponse(t, server.URL, ospf6GeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("missing required enabled must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestOSPF6GeneralDataSource_doesNotSynthesizeDefaults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf6":{"enabled":"","routerid":"","originate":"","originatealways":"","originatemetric":"","carp_demote":""}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPF6GeneralResponse(t, server.URL, ospf6GeneralID)
	if !resp.Diagnostics.HasError() || bgpDiagnosticContains(resp, "not found") {
		t.Fatal("empty required fields must not become model defaults or not-found")
	}
}

func TestOSPF6GeneralDataSource_invalidRouterIDFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf6":{"enabled":"1","routerid":"192.0.2.1/32","originate":"0","originatealways":"0","originatemetric":"1","carp_demote":"0"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPF6GeneralResponse(t, server.URL, ospf6GeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "router_id") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("masked router ID must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestOSPF6GeneralDataSource_metricAboveMaximumFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf6":{"enabled":"1","routerid":"192.0.2.1","originate":"0","originatealways":"0","originatemetric":"16777215","carp_demote":"0"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPF6GeneralResponse(t, server.URL, ospf6GeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "originate_default_metric") {
		t.Fatalf("metric above maximum must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestOSPF6GeneralDataSource_notFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf6":{}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPF6GeneralResponse(t, server.URL, ospf6GeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("empty OSPFv3 object must be not-found before field validation, got %#v", resp.Diagnostics)
	}
}

func TestOSPF6GeneralDataSource_paddedEmptyObjectIsMalformed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf6":{ }}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPF6GeneralResponse(t, server.URL, ospf6GeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("whitespace-padded empty object must be malformed, not not-found, got %#v", resp.Diagnostics)
	}
}

func TestOSPF6GeneralDataSource_nullScalarIsMalformed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf6":{"enabled":"1","routerid":null,"originate":"0","originatealways":"0","originatemetric":"1","carp_demote":"0"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPF6GeneralResponse(t, server.URL, ospf6GeneralID)
	if !resp.Diagnostics.HasError() || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("null router ID must be malformed, not not-found, got %#v", resp.Diagnostics)
	}
}

func TestOSPF6GeneralDataSource_wrongIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	resp := readOSPF6GeneralResponse(t, server.URL, "general")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func ospf6GeneralLookupFixture() string {
	return `{"ospf6":{"enabled":"1","routerid":"192.0.2.1","originate":"0","originatealways":"0","originatemetric":"100","carp_demote":"0","areas":{"area":{"example":{"enabled":"1"}}},"interfaces":{"interface":{}},"prefixlists":{"prefixlist":{}},"routemaps":{"routemap":{}},"redistributions":{"redistribution":{}},"ignoredExtra":"not-lookup-state"}}`
}

func readOSPF6General(t *testing.T, baseURL, id string) ospf6GeneralLookupModel {
	t.Helper()
	resp := readOSPF6GeneralResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state ospf6GeneralLookupModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readOSPF6GeneralResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newOSPF6GeneralDataSource()
	configureDataSource(t, ds, newTestClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := ospf6GeneralLookupModel{ID: types.StringValue(id)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}
