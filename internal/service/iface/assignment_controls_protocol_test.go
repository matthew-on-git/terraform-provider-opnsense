// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

func controlValue(t *testing.T, schema *tfprotov6.Schema, values map[string]any) *tfprotov6.DynamicValue {
	t.Helper()
	fields := map[string]any{}
	for _, attribute := range schema.Block.Attributes {
		fields[attribute.Name] = nil
	}
	for key, value := range values {
		fields[key] = value
	}
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return &tfprotov6.DynamicValue{JSON: raw}
}

func controlErrors(ds []*tfprotov6.Diagnostic) bool {
	for _, d := range ds {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func TestAssignmentControlsProtocolUpdatePreservesIdentity(t *testing.T) {
	ctx := context.Background()
	const typeName = "opnsense_interface_assignment"
	saved := map[string]any{"if": map[string]any{"vtnet2": map[string]any{"selected": 1}, "vtnet3": map[string]any{"selected": 0}}, "enable": "1", "type4": "dhcp", "ipaddr": "", "type6": "dhcp6", "ipaddrv6": "", "gateway": "", "gatewayv6": "", "descr": "FIXTURE", "mtu": "1500", "mss": "", "blockpriv": "1", "blockbogons": "1", "hw_settings_overwrite": "1", "disablechecksumoffloading": "1", "disablesegmentationoffloading": "1", "disablelargereceiveoffloading": "1", "disablevlanhwfilter": "1", "dhcp6-ia-pd-len": "8", "dhcp6prefixonly": "1", "dhcp6-ia-pd-send-hint": "1", "dhcp6-prefix-id": "0x1", "dhcp6_ifid": "0x1", "dhcp6_assoc_pd": "", "identifier": "opt3", "pending_action": "", "icon": "ignored", "optgroup": "ignored", "future_field": "preserve"}
	for _, key := range []string{"dhcphostname", "dhcprejectfrom", "dhcphonourmtu", "dhcpvlanprio", "dhcp_send_options", "dhcp_request_options", "dhcp6vlanprio", "dhcp6_request_dns", "dhcp6-information-only", "dhcp6_rapid_commit", "dhcp6_send_options", "dhcp6_request_options"} {
		saved[key] = ""
	}
	saves, activations := 0, 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/core/menu/search":
			_, _ = fmt.Fprint(w, `{}`)
		case "/api/core/firmware/info":
			_, _ = fmt.Fprint(w, `{"product_version":"26.7.8"}`)
		case "/api/interfaces/assignment/get_item/opt3":
			_ = json.NewEncoder(w).Encode(map[string]any{"interface": saved})
		case "/api/interfaces/assignment/set_item/opt3":
			saves++
			var body struct {
				Interface map[string]any `json:"interface"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			for key, want := range map[string]string{"if": "vtnet3", "descr": "UPDATED", "mtu": "", "mss": "1460", "blockpriv": "0", "blockbogons": "1", "disablevlanhwfilter": "2", "dhcp6-ia-pd-len": "0", "dhcp6prefixonly": "0", "dhcp6-ia-pd-send-hint": "0", "dhcp6-prefix-id": "", "dhcp6_ifid": "0x", "dhcp6_assoc_pd": "0", "type4": "dhcp", "type6": "dhcp6", "future_field": "preserve"} {
				if got := body.Interface[key]; got != want {
					t.Errorf("write %s=%v want %q", key, got, want)
				}
			}
			for _, key := range []string{"identifier", "pending_action", "icon", "optgroup"} {
				if _, ok := body.Interface[key]; ok {
					t.Errorf("wrote response-only metadata %s", key)
				}
			}
			for key, value := range body.Interface {
				saved[key] = value
			}
			// Tagged toLegacy uses intval(value,16); fromLegacy formats 0 as 0x0.
			saved["dhcp6_ifid"] = "0x0"
			_, _ = fmt.Fprint(w, `{"result":"saved"}`)
		case "/api/interfaces/assignment/reconfigure":
			activations++
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer api.Close()
	server := providerserver.NewProtocol6(provider.New("fixture")())()
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil || controlErrors(schemas.Diagnostics) {
		t.Fatal(err, schemas.Diagnostics)
	}
	configured, err := server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{Config: controlValue(t, schemas.Provider, map[string]any{"uri": api.URL, "api_key": "fixture-key", "api_secret": "fixture-secret"}), TerraformVersion: "1.14.0"})
	if err != nil || controlErrors(configured.Diagnostics) {
		t.Fatal(err, configured.Diagnostics)
	}
	schema := schemas.ResourceSchemas[typeName]
	imported, err := server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: typeName, ID: "opt3"})
	if err != nil || controlErrors(imported.Diagnostics) {
		t.Fatal(err, imported.Diagnostics)
	}
	read, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: typeName, CurrentState: imported.ImportedResources[0].State})
	if err != nil || controlErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	prior, err := read.NewState.Unmarshal(schema.ValueType())
	if err != nil {
		t.Fatal(err)
	}
	var attrs map[string]tftypes.Value
	if err := prior.As(&attrs); err != nil {
		t.Fatal(err)
	}
	if !attrs["dhcp6_prefix_length"].Equal(tftypes.NewValue(tftypes.String, "56")) {
		t.Fatal("encoded prefix length exposed instead of displayed length")
	}
	changes := map[string]any{"name": "opt3", "device": "vtnet3", "description": "UPDATED", "mtu": 0, "mss": 1460, "block_private": false, "vlan_hardware_filtering": "default", "dhcp6_prefix_length": "64", "dhcp6_prefix_only": false, "dhcp6_send_prefix_hint": false, "dhcp6_prefix_id": "", "dhcp6_interface_id": "0x", "dhcp6_identity_association_id": "0"}
	for key, value := range changes {
		attrs[key] = tftypes.NewValue(attrs[key].Type(), value)
	}
	proposed, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
	if err != nil {
		t.Fatal(err)
	}
	config := controlValue(t, schema, changes)
	plan, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{TypeName: typeName, PriorState: read.NewState, Config: config, ProposedNewState: &proposed})
	if err != nil || controlErrors(plan.Diagnostics) {
		t.Fatal(err, plan.Diagnostics)
	}
	applied, err := server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{TypeName: typeName, PriorState: read.NewState, Config: config, PlannedState: plan.PlannedState})
	if err != nil || controlErrors(applied.Diagnostics) {
		t.Fatal(err, applied.Diagnostics)
	}
	next, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{TypeName: typeName, PriorState: applied.NewState, Config: config, ProposedNewState: applied.NewState})
	if err != nil || controlErrors(next.Diagnostics) {
		t.Fatal(err, next.Diagnostics)
	}
	actual, err := applied.NewState.Unmarshal(schema.ValueType())
	if err != nil {
		t.Fatal(err)
	}
	planned, err := next.PlannedState.Unmarshal(schema.ValueType())
	if err != nil {
		t.Fatal(err)
	}
	if !actual.Equal(planned) || saves != 1 || activations != 1 {
		t.Fatal("update did not converge", saves, activations)
	}
	for _, value := range []string{"00", "9223372036854775808", "-1"} {
		invalidConfig := controlValue(t, schema, map[string]any{"name": "opt3", "dhcp6_identity_association_id": value})
		validation, err := server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{TypeName: typeName, Config: invalidConfig})
		if err != nil || !controlErrors(validation.Diagnostics) {
			t.Fatal("invalid association ID accepted", value, err, validation.Diagnostics)
		}
	}
	for _, tc := range []struct {
		value   string
		invalid bool
	}{{"0x7fffffffffffffff", false}, {"0x8000000000000000", true}} {
		validation, err := server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{TypeName: typeName, Config: controlValue(t, schema, map[string]any{"name": "opt3", "dhcp6_interface_id": tc.value})})
		if err != nil || controlErrors(validation.Diagnostics) != tc.invalid {
			t.Fatal("hexadecimal boundary validation mismatch", tc, err, validation.Diagnostics)
		}
	}
	saved["mtu"] = nil
	invalid, err := server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{TypeName: typeName, PriorState: applied.NewState, Config: config, PlannedState: next.PlannedState})
	if err != nil || !controlErrors(invalid.Diagnostics) || saves != 1 {
		t.Fatal("null API capability was written", err, invalid.Diagnostics, saves)
	}
}
