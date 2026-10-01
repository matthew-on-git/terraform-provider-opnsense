// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNatPortForwardRequestUsesNestedSourceDestination(t *testing.T) {
	m := NatPortForwardResourceModel{
		Enabled:         types.BoolValue(true),
		Sequence:        types.Int64Value(700),
		Interface:       types.StringValue("wan"),
		IPProtocol:      types.StringValue("inet"),
		Protocol:        types.StringValue("tcp"),
		SourceNet:       types.StringValue("any"),
		SourcePort:      types.StringValue(""),
		SourceNot:       types.BoolValue(false),
		DestinationNet:  types.StringValue("wanip"),
		DestinationPort: types.StringValue("3074"),
		DestinationNot:  types.BoolValue(false),
		Target:          types.StringValue("192.168.1.8"),
		LocalPort:       types.StringValue("3074"),
		Log:             types.BoolValue(false),
		Description:     types.StringValue("probe"),
		Reflection:      types.StringValue("disable"),
		Categories:      types.SetNull(types.StringType),
	}

	req := m.toAPI(context.Background())

	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	source, ok := payload["source"].(map[string]any)
	if !ok {
		t.Fatalf("payload has no nested \"source\" object; got %T: %s", payload["source"], raw)
	}
	if got := source["network"]; got != "any" {
		t.Errorf(`source.network = %v, want "any"`, got)
	}
	if got := source["port"]; got != "" {
		t.Errorf(`source.port = %v, want ""`, got)
	}
	if got := source["not"]; got != "0" {
		t.Errorf(`source.not = %v, want "0"`, got)
	}

	destination, ok := payload["destination"].(map[string]any)
	if !ok {
		t.Fatalf("payload has no nested \"destination\" object; got %T: %s", payload["destination"], raw)
	}
	if got := destination["network"]; got != "wanip" {
		t.Errorf(`destination.network = %v, want "wanip"`, got)
	}
	if got := destination["port"]; got != "3074" {
		t.Errorf(`destination.port = %v, want "3074"`, got)
	}
	if got := payload["sequence"]; got != "700" {
		t.Errorf(`sequence = %v, want "700"`, got)
	}
	if got := payload["natreflection"]; got != "disable" {
		t.Errorf(`natreflection = %v, want "disable"`, got)
	}

	if _, present := payload["source.network"]; present {
		t.Errorf(`flat "source.network" key still present in payload: %s`, raw)
	}
	if _, present := payload["destination.port"]; present {
		t.Errorf(`flat "destination.port" key still present in payload: %s`, raw)
	}
}

// Unset sequence must be OMITTED from the API request (not "0", not "1"):
// sending "0" is rejected by OPNsense (min 1), and hardcoding "1" would slot
// new rules to the top of the DNAT order. An omitted field lets OPNsense
// assign max+100 (append-at-end), which fromAPI then reads back.
func TestNatPortForwardRequestOmitsUnsetSequence(t *testing.T) {
	m := NatPortForwardResourceModel{
		Enabled:         types.BoolValue(true),
		Interface:       types.StringValue("wan"),
		IPProtocol:      types.StringValue("inet"),
		Protocol:        types.StringValue("tcp"),
		SourceNet:       types.StringValue("any"),
		DestinationNet:  types.StringValue("wanip"),
		DestinationPort: types.StringValue("3074"),
		Target:          types.StringValue("192.168.1.8"),
		LocalPort:       types.StringValue("3074"),
		Description:     types.StringValue("probe"),
		Reflection:      types.StringValue("disable"),
		Categories:      types.SetNull(types.StringType),
		// Sequence intentionally left null (never set in config).
	}

	req := m.toAPI(context.Background())

	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, present := payload["sequence"]; present {
		t.Errorf(`sequence present in payload when unset: %s`, raw)
	}
}

func TestNatPortForwardResponseNestedRoundTrip(t *testing.T) {
	const resp = `{
		"sequence": "700",
		"disabled": "0",
		"interface": {"wan": {"value": "WAN", "selected": 1}},
		"ipprotocol": {"inet": {"value": "IPv4", "selected": 1}},
		"protocol": {"tcp": {"value": "TCP", "selected": 1}},
		"source": {"network": "any", "port": "", "not": "0"},
		"destination": {"network": "wanip", "port": "3074", "not": "0"},
		"target": "192.168.1.8",
		"local-port": "3074",
		"log": "0",
		"descr": "probe",
		"natreflection": {"": {"value": "Use system default", "selected": 0}, "purenat": {"value": "Enable", "selected": 0}, "disable": {"value": "Disable", "selected": 1}},
		"categories": []
	}`

	var a natPortForwardAPIResponse
	if err := json.Unmarshal([]byte(resp), &a); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	var m NatPortForwardResourceModel
	m.fromAPI(context.Background(), &a, "test-uuid")

	if got := m.SourceNet.ValueString(); got != "any" {
		t.Errorf("SourceNet = %q, want %q", got, "any")
	}
	if got := m.DestinationPort.ValueString(); got != "3074" {
		t.Errorf("DestinationPort = %q, want %q", got, "3074")
	}
	if got := m.Target.ValueString(); got != "192.168.1.8" {
		t.Errorf("Target = %q, want %q", got, "192.168.1.8")
	}
	if got := m.Sequence.ValueInt64(); got != 700 {
		t.Errorf("Sequence = %v, want 700", got)
	}
	if got := m.Reflection.ValueString(); got != "disable" {
		t.Errorf("Reflection = %q, want %q", got, "disable")
	}
	if m.Enabled.ValueBool() != true {
		t.Errorf("Enabled = %v, want true (disabled=\"0\")", m.Enabled.ValueBool())
	}
}
