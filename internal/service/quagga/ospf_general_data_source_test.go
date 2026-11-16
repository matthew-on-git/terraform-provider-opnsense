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

func TestOSPFGeneralDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newOSPFGeneralDataSource())
	if len(schema.Attributes) != len(ospfGeneralLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(ospfGeneralLookupFields))
	}
	for _, name := range ospfGeneralLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{"password", "secret", "key", "passive_interfaces", "areas", "networks", "neighbors", "interfaces", "prefixlists", "routemaps", "redistributions", "ignoredExtra"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestOSPFGeneralDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != ospfGeneralReqOpts.GetEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(ospfGeneralLookupFixture()))
	}))
	t.Cleanup(server.Close)

	state := readOSPFGeneral(t, server.URL, ospfGeneralID)
	if !state.Enabled.ValueBool() || state.RouterID.ValueString() != "192.0.2.1" || state.CostReference.ValueInt64() != 100 || state.OriginateMetric.ValueInt64() != 100 {
		t.Fatalf("identity fields changed: %#v", state)
	}
	if state.Originate.ValueBool() || state.OriginateAlways.ValueBool() || state.LogAdjacencyChanges.ValueBool() || state.CARPDemote.ValueBool() || state.ID.ValueString() != ospfGeneralID {
		t.Fatalf("flags or identity changed: %#v", state)
	}
	if len(calls) != 1 || calls[0] != "GET "+ospfGeneralReqOpts.GetEndpoint {
		t.Fatalf("calls = %v", calls)
	}
}

func TestOSPFGeneralDataSource_preservesDisabledFlagsAndZeroMetric(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{"enabled":"0","routerid":"","costreference":"","logadjacencychanges":"0","originate":"0","originatealways":"0","originatemetric":"0","carp_demote":"0","areas":{"area":{}},"ignoredExtra":"not-lookup-state"}}`))
	}))
	t.Cleanup(server.Close)

	state := readOSPFGeneral(t, server.URL, ospfGeneralID)
	if state.Enabled.IsNull() || state.Enabled.ValueBool() || state.Originate.ValueBool() || state.LogAdjacencyChanges.ValueBool() || state.CARPDemote.ValueBool() {
		t.Fatalf("disabled flags were synthesized from model defaults: %#v", state)
	}
	if !state.RouterID.IsNull() || !state.CostReference.IsNull() || state.OriginateMetric.IsNull() || state.OriginateMetric.ValueInt64() != 0 {
		t.Fatalf("blank optional values or literal metric 0 changed: %#v", state)
	}
}

func TestOSPFGeneralDataSource_blankOptionalsStayNull(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{"enabled":"1","routerid":"198.51.100.1","costreference":"","logadjacencychanges":"1","originate":"1","originatealways":"0","originatemetric":"","carp_demote":"1"}}`))
	}))
	t.Cleanup(server.Close)

	state := readOSPFGeneral(t, server.URL, ospfGeneralID)
	if !state.CostReference.IsNull() || !state.OriginateMetric.IsNull() || state.RouterID.ValueString() != "198.51.100.1" || !state.CARPDemote.ValueBool() {
		t.Fatalf("blank cost or metric was synthesized: %#v", state)
	}
}

func TestOSPFGeneralDataSource_costZeroFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{"enabled":"1","routerid":"192.0.2.1","costreference":"0","logadjacencychanges":"0","originate":"0","originatealways":"0","originatemetric":"0","carp_demote":"0"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPFGeneralResponse(t, server.URL, ospfGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "cost_reference") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("literal cost 0 must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestOSPFGeneralDataSource_costAboveMaximumFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{"enabled":"1","routerid":"192.0.2.1","costreference":"4294968","logadjacencychanges":"0","originate":"0","originatealways":"0","originatemetric":"1","carp_demote":"0"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPFGeneralResponse(t, server.URL, ospfGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "cost_reference") {
		t.Fatalf("cost above maximum must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestOSPFGeneralDataSource_metricAboveMaximumFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{"enabled":"1","routerid":"192.0.2.1","costreference":"1","logadjacencychanges":"0","originate":"0","originatealways":"0","originatemetric":"16777215","carp_demote":"0"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPFGeneralResponse(t, server.URL, ospfGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "originate_default_metric") {
		t.Fatalf("metric above maximum must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestOSPFGeneralDataSource_trimsTokensOnce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{"enabled":" 1 ","routerid":" 192.0.2.10 ","costreference":" 20 ","logadjacencychanges":" 0 ","originate":" 0 ","originatealways":" 1 ","originatemetric":" 0 ","carp_demote":" 0 "}}`))
	}))
	t.Cleanup(server.Close)

	state := readOSPFGeneral(t, server.URL, ospfGeneralID)
	if !state.Enabled.ValueBool() || state.CARPDemote.ValueBool() || !state.OriginateAlways.ValueBool() || state.LogAdjacencyChanges.ValueBool() || state.RouterID.ValueString() != "192.0.2.10" || state.CostReference.ValueInt64() != 20 || state.OriginateMetric.ValueInt64() != 0 {
		t.Fatalf("padded tokens changed: %#v", state)
	}
}

func TestOSPFGeneralDataSource_ignoresNonAllowlistedFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(ospfGeneralLookupFixture()))
	}))
	t.Cleanup(server.Close)

	state := readOSPFGeneral(t, server.URL, ospfGeneralID)
	if state.RouterID.ValueString() != "192.0.2.1" || !state.Enabled.ValueBool() || state.CostReference.ValueInt64() != 100 {
		t.Fatalf("non-allowlisted fields changed readable state: %#v", state)
	}
}

func TestOSPFGeneralDataSource_missingFieldFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{"routerid":"192.0.2.1","costreference":"1","logadjacencychanges":"0","originate":"0","originatealways":"0","originatemetric":"1","carp_demote":"0"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPFGeneralResponse(t, server.URL, ospfGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("missing required enabled must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestOSPFGeneralDataSource_doesNotSynthesizeDefaults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{"enabled":"","routerid":"","costreference":"","logadjacencychanges":"","originate":"","originatealways":"","originatemetric":"","carp_demote":""}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPFGeneralResponse(t, server.URL, ospfGeneralID)
	if !resp.Diagnostics.HasError() || bgpDiagnosticContains(resp, "not found") {
		t.Fatal("empty required fields must not become model defaults or not-found")
	}
}

func TestOSPFGeneralDataSource_invalidRouterIDFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{"enabled":"1","routerid":"192.0.2.1/32","costreference":"1","logadjacencychanges":"0","originate":"0","originatealways":"0","originatemetric":"1","carp_demote":"0"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPFGeneralResponse(t, server.URL, ospfGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "router_id") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("masked router ID must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestOSPFGeneralDataSource_notFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPFGeneralResponse(t, server.URL, ospfGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("empty OSPF object must be not-found before field validation, got %#v", resp.Diagnostics)
	}
}

func TestOSPFGeneralDataSource_paddedEmptyObjectIsMalformed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{ }}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPFGeneralResponse(t, server.URL, ospfGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("whitespace-padded empty object must be malformed, not not-found, got %#v", resp.Diagnostics)
	}
}

func TestOSPFGeneralDataSource_nullScalarIsMalformed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ospf":{"enabled":"1","routerid":null,"costreference":"1","logadjacencychanges":"0","originate":"0","originatealways":"0","originatemetric":"1","carp_demote":"0"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readOSPFGeneralResponse(t, server.URL, ospfGeneralID)
	if !resp.Diagnostics.HasError() || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("null router ID must be malformed, not not-found, got %#v", resp.Diagnostics)
	}
}

func TestOSPFGeneralDataSource_wrongIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	resp := readOSPFGeneralResponse(t, server.URL, "general")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func ospfGeneralLookupFixture() string {
	return `{"ospf":{"enabled":"1","routerid":"192.0.2.1","costreference":"100","logadjacencychanges":"0","originate":"0","originatealways":"0","originatemetric":"100","carp_demote":"0","areas":{"area":{"example":{"enabled":"1"}}},"networks":{"network":{}},"neighbors":{"neighbor":{}},"interfaces":{"interface":{}},"prefixlists":{"prefixlist":{}},"routemaps":{"routemap":{}},"redistributions":{"redistribution":{}},"ignoredExtra":"not-lookup-state"}}`
}

func readOSPFGeneral(t *testing.T, baseURL, id string) ospfGeneralLookupModel {
	t.Helper()
	resp := readOSPFGeneralResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state ospfGeneralLookupModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readOSPFGeneralResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newOSPFGeneralDataSource()
	configureDataSource(t, ds, newTestClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := ospfGeneralLookupModel{ID: types.StringValue(id)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}
