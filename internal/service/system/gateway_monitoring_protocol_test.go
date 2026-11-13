// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestGatewayMonitoringProtocolUpdateAndLookup(t *testing.T) {
	ctx := context.Background()
	const typeName = "opnsense_system_gateway"
	saved := map[string]string{"name": "fixture_gateway", "descr": "fixture", "disabled": "0", "interface": "wan", "ipprotocol": "inet6", "gateway": "", "defaultgw": "0", "monitor_disable": "0", "weight": "1", "priority": "255", "monitor": "2001:db8::53", "force_down": "1", "latencylow": "200", "latencyhigh": "500", "losslow": "10", "losshigh": "20", "interval": "1", "time_period": "60", "loss_interval": "4", "data_length": "1", "fargw": "1", "monitor_noroute": "0", "monitor_killstates": "1", "monitor_killstates_priority": "0", "nosync": "0"}
	saves, activations := 0, 0
	malformed := false
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/core/menu/search":
			_, _ = fmt.Fprint(w, `{}`)
		case "/api/routing/settings/get_gateway/fixture-id":
			response := map[string]string{}
			for key, value := range saved {
				response[key] = value
			}
			if malformed {
				response["force_down"] = "invalid"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"gateway_item": response})
		case "/api/routing/settings/set_gateway/fixture-id":
			saves++
			var body struct {
				Gateway map[string]string `json:"gateway_item"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			for key, want := range map[string]string{"gateway": "", "defaultgw": "0", "monitor": "", "force_down": "0", "latencylow": "", "data_length": "0", "fargw": "1", "interval": "1"} {
				if got, ok := body.Gateway[key]; !ok || got != want {
					t.Errorf("write %s=%q present=%v want=%q", key, got, ok, want)
				}
			}
			for key, value := range body.Gateway {
				saved[key] = value
			}
			_, _ = fmt.Fprint(w, `{"result":"saved"}`)
		case "/api/routing/settings/reconfigure":
			activations++
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		default:
			t.Errorf("unexpected endpoint %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer api.Close()
	server, schemas := groupServer(t, api.URL)
	schema := schemas.ResourceSchemas[typeName]
	imported, err := server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: typeName, ID: "fixture-id"})
	if err != nil || groupErrors(imported.Diagnostics) {
		t.Fatal(err, imported.Diagnostics)
	}
	read, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: typeName, CurrentState: imported.ImportedResources[0].State})
	if err != nil || groupErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	attrs := groupAttrs(t, schema, read.NewState)
	for key, value := range map[string]string{"monitor": "", "latency_low": "", "data_length": "0"} {
		attrs[key] = tftypes.NewValue(tftypes.String, value)
	}
	attrs["force_down"] = tftypes.NewValue(tftypes.Bool, false)
	proposed, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
	if err != nil {
		t.Fatal(err)
	}
	config := groupValue(t, schema, map[string]any{"name": "fixture_gateway", "description": "fixture", "interface": "wan", "ip_protocol": "inet6", "gateway": "", "monitor_disable": false, "monitor": "", "force_down": false, "latency_low": "", "data_length": "0"})
	plan, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{TypeName: typeName, PriorState: read.NewState, Config: config, ProposedNewState: &proposed})
	if err != nil || groupErrors(plan.Diagnostics) {
		t.Fatal(err, plan.Diagnostics)
	}
	applied, err := server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{TypeName: typeName, PriorState: read.NewState, Config: config, PlannedState: plan.PlannedState})
	if err != nil || groupErrors(applied.Diagnostics) {
		t.Fatal(err, applied.Diagnostics)
	}
	next, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{TypeName: typeName, PriorState: applied.NewState, Config: config, ProposedNewState: applied.NewState})
	if err != nil || groupErrors(next.Diagnostics) {
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
		t.Fatalf("unstable update or unexpected calls: saves=%d activations=%d", saves, activations)
	}
	lookup, err := server.ReadDataSource(ctx, &tfprotov6.ReadDataSourceRequest{TypeName: typeName, Config: groupValue(t, schemas.DataSourceSchemas[typeName], map[string]any{"id": "fixture-id"})})
	if err != nil || groupErrors(lookup.Diagnostics) {
		t.Fatal(err, lookup.Diagnostics)
	}
	lookupAttrs := groupAttrs(t, schemas.DataSourceSchemas[typeName], lookup.State)
	if !lookupAttrs["monitor"].Equal(tftypes.NewValue(tftypes.String, "")) || !lookupAttrs["data_length"].Equal(tftypes.NewValue(tftypes.String, "0")) {
		t.Fatal("lookup lost clear/zero values")
	}
	for _, tc := range []struct {
		low, high, length string
		invalid           bool
	}{
		{"500", "200", "0", true}, {"10", "20", "-1", true}, {"", "200", "0", false}, {"00200", "500", "0", true}, {"200", "500", "00", true},
	} {
		validate, err := server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{TypeName: typeName, Config: groupValue(t, schema, map[string]any{"name": "fixture_gateway", "interface": "wan", "gateway": "", "latency_low": tc.low, "latency_high": tc.high, "data_length": tc.length})})
		if err != nil || groupErrors(validate.Diagnostics) != tc.invalid {
			t.Fatal("threshold validation mismatch", err, validate.Diagnostics, tc)
		}
	}
	malformed = true
	invalid, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: typeName, CurrentState: applied.NewState})
	if err != nil || !groupErrors(invalid.Diagnostics) || saves != 1 {
		t.Fatal("malformed boolean did not stop refresh", err, invalid.Diagnostics)
	}
}
