// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dnsmasq_test

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

func TestProtocolHostClearClientID(t *testing.T) {
	ctx := context.Background()
	clientID := "fixture-client"
	writes := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/core/menu/search":
			_, _ = fmt.Fprint(w, `{}`)
		case "/api/dnsmasq/settings/get_host/fixture-id":
			_ = json.NewEncoder(w).Encode(map[string]any{"host": map[string]any{"host": "fixture", "ip": "192.0.2.10", "client_id": clientID}})
		case "/api/dnsmasq/settings/set_host/fixture-id":
			writes++
			var body struct {
				Host map[string]string `json:"host"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			value, present := body.Host["client_id"]
			if !present || value != "" {
				t.Errorf("missing explicit clear: %v", body.Host)
			}
			if present {
				clientID = value
			}
			_, _ = fmt.Fprint(w, `{"result":"saved"}`)
		case "/api/dnsmasq/service/reconfigure":
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer api.Close()
	server, _ := dnsmasqServer(t, api.URL)
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil || dnsmasqErrors(schemas.Diagnostics) {
		t.Fatal(err, schemas.Diagnostics)
	}
	schema := schemas.ResourceSchemas["opnsense_dnsmasq_host"]
	imported, err := server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{TypeName: "opnsense_dnsmasq_host", ID: "fixture-id"})
	if err != nil || dnsmasqErrors(imported.Diagnostics) {
		t.Fatal(err, imported.Diagnostics)
	}
	read, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: "opnsense_dnsmasq_host", CurrentState: imported.ImportedResources[0].State})
	if err != nil || dnsmasqErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	prior, err := read.NewState.Unmarshal(schema.ValueType())
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]tftypes.Value
	if err := prior.As(&values); err != nil {
		t.Fatal(err)
	}
	values["client_id"] = tftypes.NewValue(tftypes.String, "")
	proposed, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), values))
	if err != nil {
		t.Fatal(err)
	}
	config := dnsmasqValue(t, schema, map[string]any{"host": "fixture", "ip": "192.0.2.10", "client_id": ""})
	plan, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{TypeName: "opnsense_dnsmasq_host", PriorState: read.NewState, Config: config, ProposedNewState: &proposed})
	if err != nil || dnsmasqErrors(plan.Diagnostics) {
		t.Fatal(err, plan.Diagnostics)
	}
	applied, err := server.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{TypeName: "opnsense_dnsmasq_host", PriorState: read.NewState, Config: config, PlannedState: plan.PlannedState})
	if err != nil || dnsmasqErrors(applied.Diagnostics) {
		t.Fatal(err, applied.Diagnostics)
	}
	next, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{TypeName: "opnsense_dnsmasq_host", PriorState: applied.NewState, Config: config, ProposedNewState: applied.NewState})
	if err != nil || dnsmasqErrors(next.Diagnostics) {
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
	if !actual.Equal(planned) || clientID != "" || writes != 1 {
		t.Fatalf("clear not stable: writes=%d clientID=%q", writes, clientID)
	}
}
