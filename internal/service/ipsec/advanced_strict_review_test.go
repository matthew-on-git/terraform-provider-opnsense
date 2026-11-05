// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestIPsecStrictDeleteAmbiguityRetainsOwnership(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		for _, reply := range []string{"", `{}`, `null`, `{"result":"unknown"}`} {
			t.Run(owner+"/"+reply, func(t *testing.T) {
				api := &ipsecAPI{owner: owner, fields: ipsecFixture(owner), deleteReply: &reply}
				server, schema, _ := ipsecServer(t, api)
				read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: ipsecValue(t, schema, map[string]any{"id": ipsecID})})
				if err != nil || ipsecErrors(read.Diagnostics) {
					t.Fatal("initial read failed")
				}
				empty := ipsecNull(t, schema)
				deleted := ipsecApply(t, server, owner, read.NewState, empty, &tfprotov6.PlanResourceChangeResponse{PlannedState: empty})
				if !ipsecErrors(deleted.Diagnostics) || ipsecString(t, schema, deleted.NewState, "id") != ipsecID || api.activations != 0 || api.deletes != 1 {
					t.Fatal("ambiguous delete lost ownership or activated")
				}
				api.deleteReply = nil
				success := ipsecApply(t, server, owner, deleted.NewState, empty, &tfprotov6.PlanResourceChangeResponse{PlannedState: empty})
				state, _ := success.NewState.Unmarshal(schema.ValueType())
				if ipsecErrors(success.Diagnostics) || !state.IsNull() || api.activations != 1 || api.deletes != 2 {
					t.Fatal("affirmative delete did not complete")
				}
			})
		}
	}
}

func TestIPsecStrictSelectionRejectsLossyNormalization(t *testing.T) {
	cases := []struct {
		field string
		value any
	}{
		{"local_addrs", "192.0.2.1,,192.0.2.2"},
		{"local_addrs", "192.0.2.1, 192.0.2.1"},
		{"local_addrs", map[string]any{"": map[string]any{"selected": 1}}},
		{"pools", "not-a-pool-uuid"},
		{"pools", map[string]any{"not-a-pool-uuid": map[string]any{"selected": 1}}},
	}
	for _, tc := range cases {
		t.Run(tc.field+"/"+t.Name(), func(t *testing.T) {
			api := &ipsecAPI{owner: "connection", fields: ipsecFixture("connection")}
			server, schema, lookupSchema := ipsecServer(t, api)
			initial, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_connection", CurrentState: ipsecValue(t, schema, map[string]any{"id": ipsecID})})
			if err != nil || ipsecErrors(initial.Diagnostics) {
				t.Fatal("valid control failed")
			}
			api.override = map[string]any{tc.field: tc.value}
			read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_connection", CurrentState: initial.NewState})
			if err != nil || !ipsecErrors(read.Diagnostics) {
				t.Fatal("malformed selection accepted")
			}
			before, _ := initial.NewState.Unmarshal(schema.ValueType())
			after, _ := read.NewState.Unmarshal(schema.ValueType())
			if !before.Equal(after) {
				t.Fatal("malformed selection erased prior state")
			}
			lookup, err := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "opnsense_ipsec_connection", Config: ipsecValue(t, lookupSchema, map[string]any{"id": ipsecID})})
			if err != nil || !ipsecErrors(lookup.Diagnostics) || api.adds+api.saves+api.deletes+api.activations != 0 {
				t.Fatal("malformed lookup accepted or wrote")
			}
		})
	}
}
