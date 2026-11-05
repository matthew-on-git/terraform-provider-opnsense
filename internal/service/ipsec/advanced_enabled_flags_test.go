// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestIPsecAdvancedEnabledFlags(t *testing.T) {
	for _, tc := range []struct{ owner, attribute, wire string }{
		{"connection", "aggressive", "aggressive"},
		{"child", "sha256_96", "sha256_96"},
	} {
		t.Run(tc.owner+"/"+tc.attribute, func(t *testing.T) {
			api := &ipsecAPI{owner: tc.owner, fields: ipsecFixture(tc.owner)}
			server, schema, lookupSchema := ipsecServer(t, api)
			assertFlag := func(state *tfprotov6.DynamicValue, want bool) {
				t.Helper()
				var got bool
				if err := ipsecAttrs(t, schema, state)[tc.attribute].As(&got); err != nil || got != want {
					t.Fatalf("%s readback = %t, want %t: %v", tc.attribute, got, want, err)
				}
			}
			read := func(state *tfprotov6.DynamicValue) *tfprotov6.DynamicValue {
				t.Helper()
				result, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + tc.owner, CurrentState: state})
				if err != nil || ipsecErrors(result.Diagnostics) {
					t.Fatalf("refresh: %v %v", err, result.Diagnostics)
				}
				return result.NewState
			}
			prior := read(ipsecValue(t, schema, map[string]any{"id": ipsecID}))
			assertFlag(prior, false)
			api.mu.Lock()
			api.fields[tc.wire] = "1"
			api.mu.Unlock()
			drift := read(prior)
			assertFlag(drift, true)

			imported, err := server.ImportResourceState(context.Background(), &tfprotov6.ImportResourceStateRequest{TypeName: "opnsense_ipsec_" + tc.owner, ID: ipsecID})
			if err != nil || ipsecErrors(imported.Diagnostics) || len(imported.ImportedResources) != 1 {
				t.Fatalf("import: %v %v", err, imported.Diagnostics)
			}
			assertFlag(read(imported.ImportedResources[0].State), true)
			lookup, err := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "opnsense_ipsec_" + tc.owner, Config: ipsecValue(t, lookupSchema, map[string]any{"id": ipsecID})})
			if err != nil || ipsecErrors(lookup.Diagnostics) {
				t.Fatalf("lookup: %v %v", err, lookup.Diagnostics)
			}
			var lookupFlag bool
			if err := ipsecAttrs(t, lookupSchema, lookup.State)[tc.attribute].As(&lookupFlag); err != nil || !lookupFlag {
				t.Fatalf("lookup lost enabled %s: %v", tc.attribute, err)
			}
			if api.adds+api.saves+api.deletes+api.activations != 0 {
				t.Fatal("refresh/import/lookup wrote or activated policy")
			}

			configured := ipsecConfig(tc.owner)
			config := ipsecValue(t, schema, configured)
			plan := ipsecPlan(t, server, tc.owner, drift, ipsecMerge(t, schema, drift, configured), config)
			assertFlag(plan.PlannedState, false)
			before, _ := drift.Unmarshal(schema.ValueType())
			planned, _ := plan.PlannedState.Unmarshal(schema.ValueType())
			if before.Equal(planned) {
				t.Fatal("enabled drift did not produce a repair plan")
			}
			repaired := ipsecApply(t, server, tc.owner, drift, config, plan)
			if ipsecErrors(repaired.Diagnostics) {
				t.Fatalf("repair: %v", repaired.Diagnostics)
			}
			assertFlag(repaired.NewState, false)
			if api.fields[tc.wire] != "0" || api.last[tc.wire] != "0" || api.saves != 1 || api.activations != 1 || api.adds+api.deletes != 0 {
				t.Fatal("repair did not persist and activate exactly one false write")
			}
			refreshed := read(repaired.NewState)
			assertFlag(refreshed, false)
			noChange := ipsecPlan(t, server, tc.owner, refreshed, refreshed, config)
			current, _ := refreshed.Unmarshal(schema.ValueType())
			next, _ := noChange.PlannedState.Unmarshal(schema.ValueType())
			if !current.Equal(next) {
				t.Fatal("repaired flag produced a perpetual diff")
			}
		})
	}
}
