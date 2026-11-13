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

func TestCtrlAgentDataSource_schemaAllowlist(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchema(t, newCtrlAgentDataSource())
	if len(schema.Attributes) != len(ctrlAgentLookupFields) {
		t.Fatalf("schema fields = %d, allowlist = %d", len(schema.Attributes), len(ctrlAgentLookupFields))
	}
	for _, name := range ctrlAgentLookupFields {
		if _, ok := schema.Attributes[name]; !ok {
			t.Fatalf("allowlisted field %s is missing", name)
		}
	}
	for _, excluded := range []string{"password", "secret", "key", "http_password"} {
		if _, ok := schema.Attributes[excluded]; ok {
			t.Fatalf("excluded field %s is lookup schema", excluded)
		}
	}
}

func TestCtrlAgentDataSource_readPublicMetadata(t *testing.T) {
	t.Parallel()

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		if r.Method != http.MethodGet || r.URL.Path != ctrlAgentReqOpts.GetEndpoint {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"ctrlagent":{"general":{"enabled":"1","http_host":"192.0.2.10","http_port":"8001","ignoredExtra":"not-lookup-state"}}}`))
	}))
	t.Cleanup(server.Close)

	state := readCtrlAgent(t, server.URL, ctrlAgentID)
	if !state.Enabled.ValueBool() || state.HTTPHost.ValueString() != "192.0.2.10" || state.HTTPPort.ValueString() != "8001" || state.ID.ValueString() != ctrlAgentID {
		t.Fatalf("unexpected state: %#v", state)
	}
	if len(calls) != 1 || calls[0] != "GET "+ctrlAgentReqOpts.GetEndpoint {
		t.Fatalf("calls = %v", calls)
	}
}

func TestCtrlAgentDataSource_preservesDisabled(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ctrlagent":{"general":{"enabled":"0","http_host":"192.0.2.10","http_port":"8000"}}}`))
	}))
	t.Cleanup(server.Close)

	state := readCtrlAgent(t, server.URL, ctrlAgentID)
	if state.Enabled.IsNull() || state.Enabled.ValueBool() {
		t.Fatalf("disabled flag changed: %#v", state.Enabled)
	}
}

func TestCtrlAgentDataSource_missingFieldFailsClosed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ctrlagent":{"general":{"enabled":"1","http_host":"192.0.2.10"}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readCtrlAgentResponse(t, server.URL, ctrlAgentID)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected malformed response error")
	}
}

func TestCtrlAgentDataSource_doesNotSynthesizeDefaults(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ctrlagent":{"general":{"enabled":"","http_host":"","http_port":""}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readCtrlAgentResponse(t, server.URL, ctrlAgentID)
	if !resp.Diagnostics.HasError() {
		t.Fatal("empty required fields must not become model defaults")
	}
}

func TestCtrlAgentDataSource_notFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ctrlagent":{"general":{}}}`))
	}))
	t.Cleanup(server.Close)

	resp := readCtrlAgentResponse(t, server.URL, ctrlAgentID)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected not-found error")
	}
}

func TestCtrlAgentDataSource_wrongIDMakesNoCall(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		http.Error(w, "called", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)

	resp := readCtrlAgentResponse(t, server.URL, "kea-dhcpv4")
	if !resp.Diagnostics.HasError() || calls != 0 {
		t.Fatalf("err=%v calls=%d", resp.Diagnostics, calls)
	}
}

func readCtrlAgent(t *testing.T, baseURL, id string) CtrlAgentResourceModel {
	t.Helper()
	resp := readCtrlAgentResponse(t, baseURL, id)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read diagnostics: %v", resp.Diagnostics)
	}
	var state CtrlAgentResourceModel
	if diags := resp.State.Get(context.Background(), &state); diags.HasError() {
		t.Fatalf("state diagnostics: %v", diags)
	}
	return state
}

func readCtrlAgentResponse(t *testing.T, baseURL, id string) datasource.ReadResponse {
	t.Helper()
	ds := newCtrlAgentDataSource()
	configureDataSource(t, ds, testClient(t, baseURL))
	schema := dataSourceSchema(t, ds)
	config := CtrlAgentResourceModel{ID: types.StringValue(id)}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: modelConfig(t, schema, &config)}, &resp)
	return resp
}
