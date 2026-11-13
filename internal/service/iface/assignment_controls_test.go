// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAssignmentControlsPreserveOmittedAndClearExplicit(t *testing.T) {
	var current assignmentAPIResponse
	if err := json.Unmarshal([]byte(`{"if":"vtnet2","descr":"FIXTURE","mtu":"1500","mss":"","blockpriv":"1","blockbogons":"1","hw_settings_overwrite":"1","disablechecksumoffloading":"1","disablesegmentationoffloading":"1","disablelargereceiveoffloading":"1","dhcp6-ia-pd-len":"8","identifier":"opt3"}`), &current); err != nil {
		t.Fatal(err)
	}
	var adopted AssignmentResourceModel
	if err := adopted.fromAPI(context.Background(), &current, "opt3"); err != nil {
		t.Fatal(err)
	}
	if adopted.MTU.ValueInt64() != 1500 || adopted.MSS.ValueInt64() != 0 || !adopted.BlockPrivate.ValueBool() {
		t.Fatal("saved controls not adopted")
	}
	plan := AssignmentResourceModel{Description: types.StringValue("UPDATED"), MTU: types.Int64Value(0), BlockPrivate: types.BoolValue(false), DisableChecksumOffloading: types.BoolValue(false)}
	payload, err := plan.applyToAPI(context.Background(), &current)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"if": "vtnet2", "descr": "UPDATED", "mtu": "", "mss": "", "blockpriv": "0", "blockbogons": "1", "hw_settings_overwrite": "1", "disablechecksumoffloading": "0", "disablesegmentationoffloading": "1", "disablelargereceiveoffloading": "1", "dhcp6-ia-pd-len": "8"} {
		if got := (*payload)[key]; got != want {
			t.Errorf("%s: got %v want %q", key, got, want)
		}
	}
	if _, ok := (*payload)["identifier"]; ok {
		t.Fatal("wrote response metadata")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	var saved assignmentAPIResponse
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if err := adopted.fromAPI(context.Background(), &saved, "opt3"); err != nil {
		t.Fatal(err)
	}
	if adopted.MTU.ValueInt64() != 0 || adopted.BlockPrivate.ValueBool() || adopted.Description.ValueString() != "UPDATED" {
		t.Fatal("explicit changes not retained on refresh")
	}
}

func TestAssignmentControlsMissingAndMalformed(t *testing.T) {
	current := assignmentAPIResponse{Raw: map[string]json.RawMessage{}}
	var model AssignmentResourceModel
	if err := model.fromAPI(context.Background(), &current, "opt3"); err != nil {
		t.Fatal(err)
	}
	if !model.MTU.IsNull() || !model.BlockPrivate.IsNull() || !model.Description.IsNull() {
		t.Fatal("missing fields fabricated")
	}
	model.MTU = types.Int64Unknown()
	if _, err := model.applyToAPI(context.Background(), &current); err != nil {
		t.Fatal("unknown should preserve omission", err)
	}
	model.MTU = types.Int64Value(1500)
	if _, err := model.applyToAPI(context.Background(), &current); err == nil {
		t.Fatal("configured unsupported control accepted")
	}
	for _, raw := range []string{`{"mtu":"invalid"}`, `{"mss":"1"}`, `{"mtu":"65536"}`, `{"blockpriv":"invalid"}`} {
		var malformed assignmentAPIResponse
		if err := json.Unmarshal([]byte(raw), &malformed); err != nil {
			t.Fatal(err)
		}
		if err := model.fromAPI(context.Background(), &malformed, "opt3"); err == nil {
			t.Fatalf("malformed controls accepted: %s", raw)
		}
	}
}

func TestAssignmentControlNullRemainsUnmanaged(t *testing.T) {
	var api assignmentAPIResponse
	if err := json.Unmarshal([]byte(`{"descr":null,"mtu":null,"mss":null,"blockpriv":null,"blockbogons":null,"hw_settings_overwrite":null,"disablechecksumoffloading":null,"disablesegmentationoffloading":null,"disablelargereceiveoffloading":null}`), &api); err != nil {
		t.Fatal(err)
	}
	var model AssignmentResourceModel
	if err := model.readControls(&api); err != nil {
		t.Fatal(err)
	}
	if !model.Description.IsNull() || !model.MTU.IsNull() || !model.MSS.IsNull() || !model.BlockPrivate.IsNull() || !model.BlockBogons.IsNull() || !model.OverrideHardwareSettings.IsNull() || !model.DisableChecksumOffloading.IsNull() || !model.DisableSegmentationOffloading.IsNull() || !model.DisableLargeReceiveOffloading.IsNull() {
		t.Fatal("JSON null fabricated a writable default")
	}
	request, err := model.applyToAPI(context.Background(), &api)
	if err != nil {
		t.Fatal(err)
	}
	if len(*request) != 0 {
		t.Fatalf("null fields became writes: %v", *request)
	}
	model.MTU = types.Int64Value(1500)
	if _, err := model.applyToAPI(context.Background(), &api); err == nil {
		t.Fatal("null field accepted as supported write target")
	}
}

func TestAssignmentControlOptionsAndParentDelegation(t *testing.T) {
	var api assignmentAPIResponse
	raw := `{"if":{"vtnet2":{"selected":1},"vtnet3":{"selected":0}},"type6":"dhcp6","disablevlanhwfilter":[{"selected":0},{"selected":0},{"selected":1}],"dhcp6-ia-pd-len":{"0":{"selected":0},"8":{"selected":1},"none":{"selected":0}},"dhcp6prefixonly":"1","dhcp6-ia-pd-send-hint":"1","dhcp6-prefix-id":"0xa","dhcp6_ifid":"0x2","dhcp6_assoc_pd":"0"}`
	if err := json.Unmarshal([]byte(raw), &api); err != nil {
		t.Fatal(err)
	}
	model := AssignmentResourceModel{DHCP6PrefixID: types.StringValue("0X000A")}
	if err := model.readControls(&api); err != nil {
		t.Fatal(err)
	}
	if model.DHCP6PrefixLength.ValueString() != "56" || model.VLANHardwareFiltering.ValueString() != "default" || model.DHCP6PrefixID.ValueString() != "0X000A" {
		t.Fatal("option/hex normalization lost configured semantics")
	}
	for _, length := range []string{"48", "56", "64", "none"} {
		model.DHCP6PrefixLength = types.StringValue(length)
		payload, err := model.applyToAPI(context.Background(), &api)
		if err != nil {
			t.Fatal(err)
		}
		want := map[string]string{"48": "16", "56": "8", "64": "0", "none": "none"}[length]
		if (*payload)["dhcp6-ia-pd-len"] != want {
			t.Fatalf("length %s encoded %v", length, (*payload)["dhcp6-ia-pd-len"])
		}
	}
	model.Device = types.StringValue("vtnet3")
	if _, err := model.applyToAPI(context.Background(), &api); err != nil {
		t.Fatal(err)
	}
	model.Device = types.StringValue("unadvertised0")
	if _, err := model.applyToAPI(context.Background(), &api); err == nil {
		t.Fatal("unadvertised device accepted")
	}
	model.Device = types.StringNull()
	model.IPv6Mode = types.StringValue("none")
	model.DHCP6PrefixLength = types.StringValue("56")
	if _, err := model.applyToAPI(context.Background(), &api); err == nil {
		t.Fatal("inactive nondefault prefix request accepted")
	}
	for _, raw := range []string{`{"disablevlanhwfilter":"3"}`, `{"dhcp6-ia-pd-len":"17"}`, `{"dhcp6-prefix-id":"invalid"}`} {
		var bad assignmentAPIResponse
		if err := json.Unmarshal([]byte(raw), &bad); err != nil {
			t.Fatal(err)
		}
		if err := model.readControls(&bad); err == nil {
			t.Fatalf("malformed control accepted: %s", raw)
		}
	}
}

func TestAssignmentIdentifierIntegerBounds(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"", true}, {"0", true}, {"9223372036854775807", true}, {"9223372036854775808", false}, {"00", false}, {"-1", false},
	} {
		if valid := assignmentAssociationIDValue(tc.value) == nil; valid != tc.valid {
			t.Errorf("association ID %q valid=%v want=%v", tc.value, valid, tc.valid)
		}
	}
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"0x", true}, {"0x0", true}, {"0X000A", true}, {"0x7fffffffffffffff", true}, {"0x8000000000000000", false}, {"0xffffffffffffffffffffffffffffff", false},
	} {
		if _, valid := assignmentHexValue(tc.value); valid != tc.valid {
			t.Errorf("hexadecimal ID %q valid=%v want=%v", tc.value, valid, tc.valid)
		}
	}
}

func TestAssignmentControlsRejectUnchangedInactivePrefixOnMTUUpdate(t *testing.T) {
	for _, prefix := range []string{"56", "none"} {
		t.Run(prefix, func(t *testing.T) {
			raw := `{"if":"vtnet2","type6":"slaac","mtu":"1500","dhcp6-ia-pd-len":"` + assignmentPrefixLengthWrite(prefix) + `"}`
			var current assignmentAPIResponse
			if err := json.Unmarshal([]byte(raw), &current); err != nil {
				t.Fatal(err)
			}
			model := AssignmentResourceModel{IPv6Mode: types.StringValue("slaac"), DHCP6PrefixLength: types.StringValue(prefix), MTU: types.Int64Value(1400)}
			if _, err := model.applyToAPI(context.Background(), &current); err == nil {
				t.Fatal("unrelated MTU update accepted an unchanged inactive prefix that the appliance sanitizer discards")
			}
			model.IPv6Mode = types.StringValue("dhcp6")
			if _, err := model.applyToAPI(context.Background(), &current); err != nil {
				t.Fatal("active prefix rejected", err)
			}
			model.IPv6Mode = types.StringValue("slaac")
			model.DHCP6PrefixLength = types.StringValue("64")
			if _, err := model.applyToAPI(context.Background(), &current); err != nil {
				t.Fatal("default prefix rejected", err)
			}
		})
	}
}
