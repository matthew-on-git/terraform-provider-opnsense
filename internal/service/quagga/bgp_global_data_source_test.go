// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

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

func TestBGPGlobalDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newBGPGlobalDataSource())
	if len(schema.Attributes) != len(bgpGlobalLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(bgpGlobalLookupFields))
	}
	for _, name := range bgpGlobalLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{"password", "secret", "key", "bestpath", "neighbors", "compatibility"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestBGPGlobalDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != bgpGlobalReqOpts.GetEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(bgpGlobalLookupFixture()))
	}))
	t.Cleanup(server.Close)

	state := readBGPGlobal(t, server.URL, bgpGlobalID)
	if !state.Enabled.ValueBool() || state.ASNumber.ValueInt64() != 64496 {
		t.Fatalf("identity fields changed: %#v", state)
	}
	if state.RouterID.ValueString() != "192.0.2.1" || state.Distance.IsNull() || state.Distance.ValueInt64() != 20 {
		t.Fatalf("optional router or distance changed: %#v", state)
	}
	if state.GracefulRestart.ValueBool() || !state.NetworkImportCheck.ValueBool() || !state.EnforceFirstAS.ValueBool() || state.LogNeighborChanges.ValueBool() {
		t.Fatalf("flags changed: %#v", state)
	}
	if state.MaximumPaths.ValueInt64() != 4 || !state.MaximumPathsIBGP.IsNull() {
		t.Fatalf("maximum paths changed: %#v", state)
	}
	if state.ID.ValueString() != bgpGlobalID || !bgpStringSetEquals(t, state.Networks, "192.0.2.0/24", "203.0.113.0/24") {
		t.Fatalf("identity or networks changed: %#v", state)
	}
	if len(calls) != 1 || calls[0] != "GET "+bgpGlobalReqOpts.GetEndpoint {
		t.Fatalf("calls = %v", calls)
	}
}

func TestBGPGlobalDataSource_preservesDisabledAndEmptyNetworks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"bgp":{"enabled":"0","asnumber":"64496","routerid":"","distance":"","graceful":"0","networkimportcheck":"0","enforce_first_as":"0","logneighborchanges":"1","networks":[],"maximumpaths":"","maximumpathsibgp":"","bestpath":[]}}`))
	}))
	t.Cleanup(server.Close)

	state := readBGPGlobal(t, server.URL, bgpGlobalID)
	if state.Enabled.IsNull() || state.Enabled.ValueBool() || state.LogNeighborChanges.IsNull() || !state.LogNeighborChanges.ValueBool() {
		t.Fatalf("disabled flags changed: enabled %#v log %#v", state.Enabled, state.LogNeighborChanges)
	}
	if !state.RouterID.IsNull() || !state.Distance.IsNull() || !state.MaximumPaths.IsNull() || !state.MaximumPathsIBGP.IsNull() {
		t.Fatalf("blank optionals were synthesized: %#v", state)
	}
	if !bgpStringSetEquals(t, state.Networks) {
		t.Fatalf("empty networks changed: %#v", state.Networks)
	}
}

func TestBGPGlobalDataSource_trimsTokensOnce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"bgp":{"enabled":" 1 ","asnumber":" 64496 ","routerid":" 192.0.2.1 ","distance":" 12 ","graceful":" 0 ","networkimportcheck":" 1 ","enforce_first_as":" 1 ","logneighborchanges":" 0 ","networks":"192.0.2.0/24","maximumpaths":" 2 ","maximumpathsibgp":"  "}}`))
	}))
	t.Cleanup(server.Close)

	state := readBGPGlobal(t, server.URL, bgpGlobalID)
	if !state.Enabled.ValueBool() || state.GracefulRestart.ValueBool() || state.ASNumber.ValueInt64() != 64496 {
		t.Fatalf("padded required tokens changed: %#v", state)
	}
	if state.RouterID.ValueString() != "192.0.2.1" || state.Distance.ValueInt64() != 12 || state.MaximumPaths.ValueInt64() != 2 || !state.MaximumPathsIBGP.IsNull() {
		t.Fatalf("padded optionals changed: %#v", state)
	}
}

func TestBGPGlobalDataSource_readsSelectedNetworks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"bgp":{"enabled":"1","asnumber":"64496","routerid":"192.0.2.1","distance":"20","graceful":"0","networkimportcheck":"1","enforce_first_as":"1","logneighborchanges":"0","networks":{"192.0.2.0/24":{"selected":1},"198.51.100.0/24":{"selected":0},"203.0.113.0/24":{"selected":"1"}},"maximumpaths":"1","maximumpathsibgp":"1"}}`))
	}))
	t.Cleanup(server.Close)

	state := readBGPGlobal(t, server.URL, bgpGlobalID)
	if !bgpStringSetEquals(t, state.Networks, "192.0.2.0/24", "203.0.113.0/24") {
		t.Fatalf("unselected network became lookup state: %#v", state.Networks)
	}
}

func TestBGPGlobalDataSource_ignoresNonAllowlistedFields(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"bgp":{"enabled":"1","asnumber":"64496","routerid":"192.0.2.1","distance":"","graceful":"0","networkimportcheck":"1","enforce_first_as":"1","logneighborchanges":"0","networks":[],"maximumpaths":"","maximumpathsibgp":"","bestpath":{"compare-routerid":{"selected":1}},"ignoredExtra":"not-lookup-state","neighbors":{"neighbor":[]}}}`))
	}))
	t.Cleanup(server.Close)

	state := readBGPGlobal(t, server.URL, bgpGlobalID)
	if state.ASNumber.ValueInt64() != 64496 || !state.Distance.IsNull() {
		t.Fatalf("non-allowlisted fields changed readable state: %#v", state)
	}
}

func TestBGPGlobalDataSource_missingFieldFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"bgp":{"enabled":"1","routerid":"","distance":"","graceful":"0","networkimportcheck":"1","enforce_first_as":"1","logneighborchanges":"0","networks":[],"maximumpaths":"","maximumpathsibgp":""}}`))
	}))
	t.Cleanup(server.Close)

	resp := readBGPGlobalResponse(t, server.URL, bgpGlobalID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "as_number") {
		t.Fatalf("missing required AS number must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestBGPGlobalDataSource_doesNotSynthesizeDefaults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"bgp":{"enabled":"","asnumber":"","routerid":"","distance":"","graceful":"","networkimportcheck":"","enforce_first_as":"","logneighborchanges":"","networks":[],"maximumpaths":"","maximumpathsibgp":""}}`))
	}))
	t.Cleanup(server.Close)

	resp := readBGPGlobalResponse(t, server.URL, bgpGlobalID)
	if !resp.Diagnostics.HasError() {
		t.Fatal("empty required fields must not become model defaults")
	}
}

func TestBGPGlobalDataSource_zeroDistanceFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"bgp":{"enabled":"1","asnumber":"64496","routerid":"","distance":"0","graceful":"0","networkimportcheck":"1","enforce_first_as":"1","logneighborchanges":"0","networks":[],"maximumpaths":"","maximumpathsibgp":""}}`))
	}))
	t.Cleanup(server.Close)

	resp := readBGPGlobalResponse(t, server.URL, bgpGlobalID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "distance") {
		t.Fatalf("zero distance must not become the unset default, got %#v", resp.Diagnostics)
	}
}

func TestBGPGlobalDataSource_notFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"bgp":{}}`))
	}))
	t.Cleanup(server.Close)

	resp := readBGPGlobalResponse(t, server.URL, bgpGlobalID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("empty BGP object must be not-found before field validation, got %#v", resp.Diagnostics)
	}
}

func TestBGPGlobalDataSource_wrongIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	resp := readBGPGlobalResponse(t, server.URL, "kea-dhcpv6")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func bgpGlobalLookupFixture() string {
	return `{"bgp":{"enabled":"1","asnumber":"64496","routerid":"192.0.2.1","distance":"20","graceful":"0","networkimportcheck":"1","enforce_first_as":"1","logneighborchanges":"0","networks":"192.0.2.0/24,203.0.113.0/24","maximumpaths":"4","maximumpathsibgp":"","bestpath":{"compare-routerid":{"selected":1}},"ignoredExtra":"not-lookup-state"}}`
}

func readBGPGlobal(t *testing.T, baseURL, id string) bgpGlobalLookupModel {
	t.Helper()
	resp := readBGPGlobalResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state bgpGlobalLookupModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readBGPGlobalResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newBGPGlobalDataSource()
	configureDataSource(t, ds, newTestClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := bgpGlobalLookupModel{ID: types.StringValue(id), Networks: types.SetNull(types.StringType)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}

func bgpDiagnosticContains(resp datasource.ReadResponse, needle string) bool {
	for _, diagnostic := range resp.Diagnostics {
		if strings.Contains(strings.ToLower(diagnostic.Summary()), strings.ToLower(needle)) || strings.Contains(strings.ToLower(diagnostic.Detail()), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func bgpStringSetEquals(t *testing.T, set types.Set, want ...string) bool {
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
