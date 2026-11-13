// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAssignmentLookupSchemaMatchesResourceAllowlist(t *testing.T) {
	t.Parallel()
	schema := assignmentLookupTestSchema(t)
	if len(schema.Attributes) != len(assignmentLookupFields) {
		t.Fatalf("published fields = %d, allowlist = %d", len(schema.Attributes), len(assignmentLookupFields))
	}
	for _, name := range assignmentLookupFields {
		attribute, ok := schema.Attributes[name]
		if !ok {
			t.Fatalf("allowlisted field %q is not published", name)
		}
		if attribute.IsSensitive() {
			t.Fatalf("%s is marked sensitive; the released model has no secret to publish", name)
		}
		if name == "id" {
			if !attribute.IsRequired() || attribute.IsComputed() {
				t.Fatal("id must be the required lookup identity")
			}
			continue
		}
		if !attribute.IsComputed() || attribute.IsRequired() {
			t.Fatalf("%s must be computed and not required", name)
		}
	}
}

func TestAssignmentLookupReadPublishesPublicFieldsOnly(t *testing.T) {
	t.Parallel()
	current := assignmentFixture("192.0.2.10/24", "2001:db8::10/64")
	current["type4"] = selected("staticv4")
	current["type6"] = selected("staticv6")
	current["gateway"] = "GW4"
	current["gatewayv6"] = "GW6"
	current["dhcprejectfrom"] = "192.0.2.1,192.0.2.2"
	current["spoofmac"] = "02:00:00:00:00:01"
	current["pending_action"] = "not-lookup-state"
	current["descr"] = "EXAMPLE"
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		methods = append(methods, req.Method+" "+req.URL.Path)
		if req.Method != http.MethodGet {
			t.Errorf("lookup mutated appliance: %s %s", req.Method, req.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		switch req.URL.Path {
		case "/api/core/firmware/info":
			_, _ = io.WriteString(w, `{"product_version":"26.7.9"}`)
		case "/api/interfaces/assignment/get_item/opt3":
			_ = json.NewEncoder(w).Encode(map[string]any{"interface": current})
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(server.Close)

	state, diags := readAssignmentLookup(t, server.URL, "opt3")
	if diags.HasError() {
		t.Fatalf("diagnostics: %v", diags)
	}
	if state.ID.ValueString() != "opt3" || state.Name.ValueString() != "opt3" {
		t.Fatalf("identity = %s/%s", state.ID.ValueString(), state.Name.ValueString())
	}
	if !state.Enabled.ValueBool() || state.IPv4Mode.ValueString() != "staticv4" || state.IPv4Address.ValueString() != "192.0.2.10" || state.IPv4PrefixLength.ValueInt64() != 24 {
		t.Fatalf("ipv4 = %#v", state)
	}
	if state.IPv6Address.ValueString() != "2001:db8::10" || state.IPv6PrefixLength.ValueInt64() != 64 || state.IPv4Gateway.ValueString() != "GW4" {
		t.Fatalf("gateways/ipv6 = %#v", state)
	}
	var rejected []string
	if diags := state.DHCPRejectFrom.ElementsAs(context.Background(), &rejected, false); diags.HasError() || len(rejected) != 2 {
		t.Fatalf("dhcp_reject_from = %#v %v", rejected, diags)
	}
	for _, call := range methods {
		if !strings.HasPrefix(call, "GET ") {
			t.Fatalf("mutation call %s", call)
		}
	}
}

func TestAssignmentLookupRejectsIdentityBeforeRead(t *testing.T) {
	t.Parallel()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls++
	}))
	t.Cleanup(server.Close)
	_, diags := readAssignmentLookup(t, server.URL, "em0")
	if !diags.HasError() || calls != 0 || !strings.Contains(diags[0].Detail(), "wan, lan, or optN") {
		t.Fatalf("diagnostics=%v calls=%d", diags, calls)
	}
}

func TestAssignmentLookupResponseIdentity(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name       string
		identifier any
		omit       bool
		wantError  bool
	}{
		{name: "matching", identifier: "opt3"},
		{name: "absent", omit: true},
		{name: "mismatched", identifier: "opt4", wantError: true},
		{name: "empty", identifier: "", wantError: true},
		{name: "null", identifier: nil, wantError: true},
		{name: "wrong type", identifier: 3, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			current := assignmentFixture("192.0.2.10/24", "")
			if test.omit {
				delete(current, "identifier")
			} else {
				current["identifier"] = test.identifier
			}
			state, diags := readAssignmentLookupFixture(t, current)
			if test.wantError {
				if !diags.HasError() || diags[0].Summary() != "Interface assignment identity changed" {
					t.Fatalf("diagnostics = %v", diags)
				}
				return
			}
			if diags.HasError() || state.ID.ValueString() != "opt3" || state.Name.ValueString() != "opt3" {
				t.Fatalf("identity = %s/%s, diagnostics = %v", state.ID.ValueString(), state.Name.ValueString(), diags)
			}
		})
	}
}

func TestAssignmentLookupMissingFieldFailsClosed(t *testing.T) {
	t.Parallel()
	current := assignmentFixture("192.0.2.10/24", "")
	delete(current, "enable")
	_, diags := readAssignmentLookupFixture(t, current)
	if !diags.HasError() || !strings.Contains(diags[0].Detail(), "missing required fields") {
		t.Fatalf("diagnostics = %v", diags)
	}
}

func TestAssignmentLookupUnsupportedModeFailsClosed(t *testing.T) {
	t.Parallel()
	current := assignmentFixture("", "")
	current["type4"] = selected("pppoe")
	current["pending_action"] = "do-not-synthesize"
	_, diags := readAssignmentLookupFixture(t, current)
	if !diags.HasError() || !strings.Contains(diags[0].Detail(), "unsupported IPv4 mode") {
		t.Fatalf("diagnostics = %v", diags)
	}
}

func TestAssignmentLookupEmptyAddressKeepsNullPrefix(t *testing.T) {
	t.Parallel()
	state, diags := readAssignmentLookupFixture(t, assignmentFixture("", ""))
	if diags.HasError() {
		t.Fatalf("diagnostics = %v", diags)
	}
	if state.IPv4Address.ValueString() != "" || !state.IPv4PrefixLength.IsNull() || state.IPv6Mode.ValueString() != "none" {
		t.Fatalf("empty address normalized to %#v", state)
	}
}

func TestAssignmentLookupNotFound(t *testing.T) {
	t.Parallel()
	_, diags := readAssignmentLookupFixture(t, map[string]any{})
	if !diags.HasError() || !strings.Contains(strings.ToLower(diags[0].Detail()), "not found") {
		t.Fatalf("diagnostics = %v", diags)
	}
}

func readAssignmentLookup(t *testing.T, baseURL, id string) (AssignmentResourceModel, diag.Diagnostics) {
	t.Helper()
	source := &assignmentDataSource{client: assignmentTestClient(t, baseURL)}
	var response datasource.ReadResponse
	response.State = tfsdk.State{Schema: assignmentLookupTestSchema(t)}
	source.Read(context.Background(), datasource.ReadRequest{Config: assignmentLookupConfig(t, id)}, &response)
	var state AssignmentResourceModel
	if !response.Diagnostics.HasError() {
		response.State.Get(context.Background(), &state)
	}
	return state, response.Diagnostics
}

func readAssignmentLookupFixture(t *testing.T, assignment map[string]any) (AssignmentResourceModel, diag.Diagnostics) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			t.Errorf("lookup mutated appliance: %s %s", req.Method, req.URL.Path)
			http.Error(w, "mutation", http.StatusBadRequest)
			return
		}
		switch req.URL.Path {
		case "/api/core/firmware/info":
			_, _ = io.WriteString(w, `{"product_version":"26.7.9"}`)
		case "/api/interfaces/assignment/get_item/opt3":
			_ = json.NewEncoder(w).Encode(map[string]any{"interface": assignment})
		default:
			http.NotFound(w, req)
		}
	}))
	t.Cleanup(server.Close)
	return readAssignmentLookup(t, server.URL, "opt3")
}

func assignmentLookupTestSchema(t *testing.T) dsschema.Schema {
	t.Helper()
	var response datasource.SchemaResponse
	(&assignmentDataSource{}).Schema(context.Background(), datasource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", response.Diagnostics)
	}
	return response.Schema
}

func assignmentLookupConfig(t *testing.T, id string) tfsdk.Config {
	t.Helper()
	schema := assignmentLookupTestSchema(t)
	model := AssignmentResourceModel{ID: types.StringValue(id), DHCPRejectFrom: types.SetNull(types.StringType)}
	state := tfsdk.State{Schema: schema}
	if diags := state.Set(context.Background(), &model); diags.HasError() {
		t.Fatalf("config diagnostics: %v", diags)
	}
	return tfsdk.Config{Raw: state.Raw, Schema: schema}
}
