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

func TestStaticGeneralDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newStaticGeneralDataSource())
	if len(schema.Attributes) != len(staticLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(staticLookupFields))
	}
	for _, name := range staticLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{"password", "secret", "key", "routes", "route", "ignoredExtra"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestStaticGeneralDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != staticGeneralReqOpts.GetEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(staticLookupFixture()))
	}))
	t.Cleanup(server.Close)

	state := readStatic(t, server.URL, staticGeneralID)
	if !state.Enabled.ValueBool() || state.ID.ValueString() != staticGeneralID {
		t.Fatalf("identity fields changed: %#v", state)
	}
	if len(calls) != 1 || calls[0] != "GET "+staticGeneralReqOpts.GetEndpoint {
		t.Fatalf("calls = %v", calls)
	}
}

func TestStaticGeneralDataSource_preservesDisabled(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"staticd":{"enabled":"0","routes":{"route":{"11111111-1111-1111-1111-111111111111":{"enabled":"1","network":"192.0.2.0/24"}}},"ignoredExtra":"not-lookup-state"}}`))
	}))
	t.Cleanup(server.Close)

	state := readStatic(t, server.URL, staticGeneralID)
	if state.Enabled.IsNull() || state.Enabled.ValueBool() || state.ID.ValueString() != staticGeneralID {
		t.Fatalf("disabled flag changed: %#v", state)
	}
}

func TestStaticGeneralDataSource_trimsEnabledOnce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"staticd":{"enabled":" 1 "}}`))
	}))
	t.Cleanup(server.Close)

	state := readStatic(t, server.URL, staticGeneralID)
	if !state.Enabled.ValueBool() {
		t.Fatalf("padded enabled token changed: %#v", state)
	}
}

func TestStaticGeneralDataSource_blankEnabledFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"staticd":{"enabled":"","routes":{"route":{}}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readStaticResponse(t, server.URL, staticGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("blank enabled must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestStaticGeneralDataSource_missingEnabledFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"staticd":{"routes":{"route":{}}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readStaticResponse(t, server.URL, staticGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("missing enabled must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestStaticGeneralDataSource_nonBooleanFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"staticd":{"enabled":"true"}}`))
	}))
	t.Cleanup(server.Close)

	resp := readStaticResponse(t, server.URL, staticGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("non-boolean enabled must fail closed, got %#v", resp.Diagnostics)
	}
}

func TestStaticGeneralDataSource_notFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"staticd":{}}`))
	}))
	t.Cleanup(server.Close)

	resp := readStaticResponse(t, server.URL, staticGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("empty staticd object must be not-found before field validation, got %#v", resp.Diagnostics)
	}
}

func TestStaticGeneralDataSource_paddedEmptyObjectIsMalformed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"staticd":{ }}`))
	}))
	t.Cleanup(server.Close)

	resp := readStaticResponse(t, server.URL, staticGeneralID)
	if !resp.Diagnostics.HasError() || !bgpDiagnosticContains(resp, "enabled") || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("whitespace-padded empty object must be malformed, not not-found, got %#v", resp.Diagnostics)
	}
}

func TestStaticGeneralDataSource_nullScalarIsMalformed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"staticd":{"enabled":null}}`))
	}))
	t.Cleanup(server.Close)

	resp := readStaticResponse(t, server.URL, staticGeneralID)
	if !resp.Diagnostics.HasError() || bgpDiagnosticContains(resp, "not found") {
		t.Fatalf("null enabled must be malformed, not not-found, got %#v", resp.Diagnostics)
	}
}

func TestStaticGeneralDataSource_wrongIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	resp := readStaticResponse(t, server.URL, "rip")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func staticLookupFixture() string {
	return `{"staticd":{"enabled":"1","routes":{"route":{"11111111-1111-1111-1111-111111111111":{"enabled":"1","network":"192.0.2.0/24"}}},"ignoredExtra":"not-lookup-state"}}`
}

func readStatic(t *testing.T, baseURL, id string) staticLookupModel {
	t.Helper()
	resp := readStaticResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state staticLookupModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readStaticResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newStaticGeneralDataSource()
	configureDataSource(t, ds, newTestClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := staticLookupModel{ID: types.StringValue(id)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}
