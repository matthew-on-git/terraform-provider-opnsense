// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func delegationFixture(mode string) map[string]any {
	fixture := assignmentFixture("192.0.2.1/24", "")
	fixture["type4"] = selected("staticv4")
	fixture["type6"] = selected(mode)
	fixture["track6-interface"] = selected("wan")
	fixture["track6-prefix-id"] = "0x10"
	fixture["track6_ifid"] = "0x1"
	fixture["track6_prefix_range"] = ""
	fixture["track6_assoc_pd"] = "0"
	fixture["dhcpd6track6allowoverride"] = "1"
	// Unowned DHCPv6 and hardware fields must survive delegation updates.
	fixture["dhcp6-ia-pd-len"] = selected("8")
	fixture["dhcp6prefixonly"] = "1"
	fixture["mtu"] = ""
	return fixture
}

func decodeDelegationFixture(t *testing.T, fixture map[string]any) *assignmentAPIResponse {
	t.Helper()
	raw, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	var result assignmentAPIResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	return &result
}

func TestDelegationReadAndRoundTrip(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"idassoc6", "track6"} {
		t.Run(mode, func(t *testing.T) {
			current := decodeDelegationFixture(t, delegationFixture(mode))
			if err := validateSafeAssignmentModes(current); err != nil {
				t.Fatal(err)
			}
			var state AssignmentResourceModel
			if err := state.fromAPI(context.Background(), current, "opt3"); err != nil {
				t.Fatal(err)
			}
			if state.IPv6TrackInterface.ValueString() != "wan" || state.IPv6PrefixID.ValueString() != "0x10" || state.IPv6InterfaceID.ValueString() != "0x1" || state.IPv6IdentityAssociationID.ValueString() != "0" || state.IPv6PrefixRange.ValueString() != "" || !state.IPv6AllowManualDHCPv6.ValueBool() {
				t.Fatalf("delegation fields were not decoded: %+v", state)
			}
			if _, err := state.applyToAPI(context.Background(), current); err == nil {
				t.Fatal("inactive adopted prefix must be rejected before update")
			}
			// Explicitly select the default before updating an inactive parent.
			state.DHCP6PrefixLength = types.StringValue("64")
			payload, err := state.applyToAPI(context.Background(), current)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := current.toRequest()
			if err != nil {
				t.Fatal(err)
			}
			expected["dhcp6-ia-pd-len"] = "0"
			if !reflect.DeepEqual(*payload, expected) {
				t.Fatalf("no-change round trip altered the assignment: got %#v want %#v", *payload, expected)
			}
			// Exercise the framework Read path used after import, including schema/state conversion.
			server := assignmentTestServer(t, "26.7.7", delegationFixture(mode), http.StatusOK, "")
			defer server.Close()
			r := &assignmentResource{client: assignmentTestClient(t, server.URL)}
			imported := assignmentTestState(t, r, AssignmentResourceModel{ID: types.StringValue("opt3"), Name: types.StringValue("opt3")})
			resp := resource.ReadResponse{State: imported}
			r.Read(context.Background(), resource.ReadRequest{State: imported}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatal(resp.Diagnostics)
			}
			var got AssignmentResourceModel
			if d := resp.State.Get(context.Background(), &got); d.HasError() {
				t.Fatal(d)
			}
			if got.IPv6Mode.ValueString() != mode || got.IPv6IdentityAssociationID.ValueString() != "0" {
				t.Fatalf("import read state = %+v", got)
			}
		})
	}
}

func TestDelegationOmittedUnknownAndEmptyValues(t *testing.T) {
	t.Parallel()
	current := decodeDelegationFixture(t, delegationFixture("idassoc6"))
	plan := AssignmentResourceModel{Name: types.StringValue("opt3"), IPv6PrefixID: types.StringUnknown()}
	payload, err := plan.applyToAPI(context.Background(), current)
	if err != nil {
		t.Fatal(err)
	}
	if (*payload)["track6-prefix-id"] != "0x10" || (*payload)["track6_assoc_pd"] != "0" || (*payload)["dhcpd6track6allowoverride"] != "1" {
		t.Fatalf("omitted/unknown fields were changed: %#v", *payload)
	}
	plan.IPv6IdentityAssociationID = types.StringValue("")
	plan.IPv6PrefixRange = types.StringValue("4")
	plan.IPv6AllowManualDHCPv6 = types.BoolValue(false)
	payload, err = plan.applyToAPI(context.Background(), current)
	if err != nil {
		t.Fatal(err)
	}
	if (*payload)["track6_assoc_pd"] != "" || (*payload)["track6_prefix_range"] != "4" || (*payload)["dhcpd6track6allowoverride"] != "0" {
		t.Fatalf("explicit clear/false/number not preserved: %#v", *payload)
	}
	var state AssignmentResourceModel
	if err := state.fromAPI(context.Background(), decodeDelegationFixture(t, map[string]any(*payload)), "opt3"); err != nil {
		t.Fatal(err)
	}
	if state.IPv6IdentityAssociationID.IsNull() || state.IPv6IdentityAssociationID.ValueString() != "" {
		t.Fatal("blank association ID must round trip distinctly from zero")
	}
}

func TestDelegationMissingFieldsFailClosed(t *testing.T) {
	t.Parallel()
	for _, field := range assignmentDelegationAPIFields {
		t.Run(field, func(t *testing.T) {
			fixture := delegationFixture("idassoc6")
			delete(fixture, field)
			if err := validateSafeAssignmentModes(decodeDelegationFixture(t, fixture)); err == nil {
				t.Fatal("partial delegation API accepted")
			}
		})
	}
}

func TestDelegationParentValidationBeforeWrite(t *testing.T) {
	t.Parallel()
	current := decodeDelegationFixture(t, delegationFixture("idassoc6"))
	for _, parent := range []string{"", "opt3"} {
		plan := AssignmentResourceModel{Name: types.StringValue("opt3"), IPv6TrackInterface: types.StringValue(parent)}
		if _, err := plan.applyToAPI(context.Background(), current); err == nil {
			t.Fatalf("invalid parent %q accepted", parent)
		}
	}
}

func TestDelegationSchemaValidators(t *testing.T) {
	t.Parallel()
	tests := []struct {
		field, value string
		valid        bool
	}{
		{"ipv6_track_interface", "wan", true},
		{"ipv6_track_interface", "opt2", true},
		{"ipv6_track_interface", "vtnet0", false},
		{"ipv6_prefix_id", "0x0", true},
		{"ipv6_prefix_id", "", true},
		{"ipv6_prefix_id", "0xgg", false},
		{"ipv6_interface_id", "0x01", true},
		{"ipv6_interface_id", "2001:db8::1", false},
		{"ipv6_prefix_range", "1", true},
		{"ipv6_prefix_range", "", true},
		{"ipv6_prefix_range", "0", false},
		{"ipv6_identity_association_id", "0", true},
		{"ipv6_identity_association_id", "", true},
		{"ipv6_identity_association_id", "-1", false},
	}
	attributes := assignmentDelegationAttributes()
	for _, test := range tests {
		t.Run(test.field+"/"+test.value, func(t *testing.T) {
			attribute := attributes[test.field].(interface{ StringValidators() []validator.String })
			var response validator.StringResponse
			for _, v := range attribute.StringValidators() {
				v.ValidateString(context.Background(), validator.StringRequest{ConfigValue: types.StringValue(test.value)}, &response)
			}
			if response.Diagnostics.HasError() == test.valid {
				t.Fatalf("value %q valid=%v diagnostics=%v", test.value, test.valid, response.Diagnostics)
			}
		})
	}
}

func TestDelegationUpdateAndModeTransitions(t *testing.T) {
	t.Parallel()
	for _, transition := range []struct{ from, to string }{
		{"idassoc6", "idassoc6"},
		{"track6", "idassoc6"},
		{"none", "idassoc6"},
		{"idassoc6", "track6"},
		{"idassoc6", "none"},
	} {
		t.Run(transition.from+"/"+transition.to, func(t *testing.T) {
			mode := transition.to
			current := delegationFixture(transition.from)
			var posted map[string]map[string]any
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				switch req.URL.Path {
				case "/api/core/firmware/info":
					_, _ = fmt.Fprint(w, `{"product_version":"26.7.7"}`)
				case "/api/interfaces/assignment/get_item/opt3":
					_ = json.NewEncoder(w).Encode(map[string]any{"interface": current})
				case "/api/interfaces/assignment/set_item/opt3":
					writes++
					if err := json.NewDecoder(req.Body).Decode(&posted); err != nil {
						t.Error(err)
					}
					for key, value := range posted["interface"] {
						current[key] = value
					}
					_, _ = fmt.Fprint(w, `{"result":"saved"}`)
				case "/api/interfaces/assignment/reconfigure":
					_, _ = fmt.Fprint(w, `{"status":"ok"}`)
				default:
					http.NotFound(w, req)
				}
			}))
			defer server.Close()
			r := &assignmentResource{client: assignmentTestClient(t, server.URL)}
			plan := AssignmentResourceModel{Name: types.StringValue("opt3"), IPv6Mode: types.StringValue(mode), IPv6PrefixID: types.StringValue("0x20")}
			var diagnostics diag.Diagnostics
			var got AssignmentResourceModel
			r.apply(context.Background(), &plan, "opt3", "updating", &diagnostics, func(m *AssignmentResourceModel) { got = *m })
			if diagnostics.HasError() || writes != 1 {
				t.Fatalf("writes=%d diagnostics=%v", writes, diagnostics)
			}
			if got.IPv6Mode.ValueString() != mode || got.IPv6PrefixID.ValueString() != "0x20" {
				t.Fatalf("readback=%+v", got)
			}
			for key, want := range map[string]any{"if": "vtnet2", "descr": "disposable", "ipaddr": "192.0.2.1/24", "dhcp6-ia-pd-len": "8", "dhcp6prefixonly": "1", "mtu": "", "track6_assoc_pd": "0"} {
				if posted["interface"][key] != want {
					t.Errorf("unrelated field %s changed", key)
				}
			}
		})
	}
}

func TestDelegationTransitionRejectsPartialAPI(t *testing.T) {
	t.Parallel()
	current := decodeDelegationFixture(t, assignmentFixture("", ""))
	plan := AssignmentResourceModel{
		Name: types.StringValue("opt3"), IPv6Mode: types.StringValue("idassoc6"),
		IPv6TrackInterface: types.StringValue("wan"),
	}
	if _, err := plan.applyToAPI(context.Background(), current); err == nil {
		t.Fatal("transition to delegation accepted an API missing its dependent fields")
	}
}
