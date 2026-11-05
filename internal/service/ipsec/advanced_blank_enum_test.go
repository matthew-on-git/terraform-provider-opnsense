// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Pinned core9697800 BaseListField::getNodeOptions emits a selected empty
// placeholder for optional single-choice fields, unlike multi-selection fields.
func TestIPsecBlankEnumSourceShapes(t *testing.T) {
	for _, tc := range []struct{ wire, attribute, choice string }{
		{"local_port", "local_port", "4500"},
		{"remote_port", "remote_port", "4500"},
		{"send_cert", "send_certificate", "never"},
	} {
		t.Run(tc.wire, func(t *testing.T) {
			api := &ipsecAPI{owner: "connection", fields: ipsecFixture("connection")}
			server, schema, lookupSchema := ipsecServer(t, api)
			initial, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{
				TypeName: "opnsense_ipsec_connection", CurrentState: ipsecValue(t, schema, map[string]any{"id": ipsecID}),
			})
			if err != nil || ipsecErrors(initial.Diagnostics) {
				t.Fatal("initial policy read failed")
			}
			for _, selected := range []string{"", tc.choice} {
				api.override = map[string]any{tc.wire: map[string]any{
					"":        map[string]any{"value": "Default", "selected": boolInt(selected == "")},
					tc.choice: map[string]any{"value": "Choice", "selected": boolInt(selected == tc.choice)},
				}}
				read, readErr := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{
					TypeName: "opnsense_ipsec_connection", CurrentState: initial.NewState,
				})
				lookup, lookupErr := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{
					TypeName: "opnsense_ipsec_connection", Config: ipsecValue(t, lookupSchema, map[string]any{"id": ipsecID}),
				})
				if readErr != nil || lookupErr != nil || ipsecErrors(read.Diagnostics) || ipsecErrors(lookup.Diagnostics) {
					t.Fatalf("legitimate selected %q rejected", selected)
				}
				if ipsecString(t, schema, read.NewState, tc.attribute) != selected || ipsecString(t, lookupSchema, lookup.State, tc.attribute) != selected {
					t.Fatal("selected enum value changed")
				}
			}
			for _, malformed := range []map[string]any{
				{"": map[string]any{"selected": 1}, tc.choice: map[string]any{"selected": 1}},
				{"not-a-valid-option": map[string]any{"selected": 1}},
				{" ": map[string]any{"selected": 1}},
			} {
				api.override = map[string]any{tc.wire: malformed}
				read, readErr := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{
					TypeName: "opnsense_ipsec_connection", CurrentState: initial.NewState,
				})
				lookup, lookupErr := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{
					TypeName: "opnsense_ipsec_connection", Config: ipsecValue(t, lookupSchema, map[string]any{"id": ipsecID}),
				})
				if readErr != nil || lookupErr != nil || !ipsecErrors(read.Diagnostics) || !ipsecErrors(lookup.Diagnostics) {
					t.Fatal("mixed or invalid enum selection accepted")
				}
				before, _ := initial.NewState.Unmarshal(schema.ValueType())
				after, _ := read.NewState.Unmarshal(schema.ValueType())
				if !before.Equal(after) {
					t.Fatal("malformed enum erased prior policy")
				}
			}
			if api.adds+api.saves+api.deletes+api.activations != 0 {
				t.Fatal("enum read or lookup mutated appliance")
			}
		})
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
