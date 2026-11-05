// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestProtocolAdvancedReadFailsClosed(t *testing.T) {
	for _, tc := range []struct {
		field string
		value any
	}{
		{"weight", nil}, {"weight", "65536"}, {"localas", "0"}, {"connecttimer", "65001"}, {"removeprivateas", "removeallas"}, {"allowas_in", "val1"}, {"attributeunchanged", "as-path,med"}, {"capabilities", "unsupported"}, {"bfd", "2"},
	} {
		t.Run(tc.field, func(t *testing.T) {
			writes := 0
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					writes++
				}
				if frrFixture(w, r) {
					return
				}
				response := finalNeighbor()
				if tc.value == nil {
					delete(response, tc.field)
				} else {
					response[tc.field] = tc.value
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"neighbor": response})
			}))
			defer api.Close()
			server, schemas := frrProtocolServer(t, api.URL)
			name := "opnsense_quagga_bgp_neighbor"
			schema := schemas.ResourceSchemas[name]
			prior := frrProtocolValue(t, schema, frrNeighborProtocolValues())
			read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: prior})
			if err != nil || !frrProtocolErrors(read.Diagnostics) || writes != 0 {
				t.Fatal("malformed advanced read did not fail closed", err, writes)
			}
			before, _ := prior.Unmarshal(schema.ValueType())
			after, _ := read.NewState.Unmarshal(schema.ValueType())
			if !before.Equal(after) {
				t.Fatal("failed read replaced prior state")
			}
		})
	}
}

func TestProtocolAdvancedInvalidIntentHasNoWrites(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
	}{
		{"bfd_strict", true}, {"next_hop_self_all", true}, {"password", "configured-test-input"}, {"capabilities", []string{"unsupported"}}, {"capabilities", []string{"dynamic,fqdn"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					writes++
				}
				if !frrFixture(w, r) {
					t.Error("unexpected API", r.URL.Path)
				}
			}))
			defer api.Close()
			server, schemas := frrProtocolServer(t, api.URL)
			name := "opnsense_quagga_bgp_neighbor"
			schema := schemas.ResourceSchemas[name]
			values := frrNeighborProtocolValues()
			delete(values, "id")
			values[tc.name] = tc.value
			config := frrProtocolValue(t, schema, values)
			prior := &tfprotov6.DynamicValue{JSON: []byte("null")}
			planned, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: prior, ProposedNewState: config, Config: config})
			if err != nil {
				t.Fatal(err)
			}
			if !frrProtocolErrors(planned.Diagnostics) {
				applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: planned.PlannedState, Config: config})
				if err != nil || !frrProtocolErrors(applied.Diagnostics) {
					t.Fatal("invalid advanced intent accepted", err)
				}
			}
			if writes != 0 {
				t.Fatal("invalid intent mutated appliance")
			}
		})
	}
	// A configured unknown must remain intent, not an omission/default.
	api := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/core/menu/search" {
			t.Error("unknown configured intent reached appliance")
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	config := frrProtocolValue(t, schema, frrNeighborProtocolValues())
	attrs := frrProtocolAttrs(t, schema, config)
	attrs["weight"] = tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)
	value, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
	if err != nil {
		t.Fatal(err)
	}
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: &tfprotov6.DynamicValue{JSON: []byte("null")}, PlannedState: &value, Config: &value})
	if err != nil || !frrProtocolErrors(applied.Diagnostics) {
		t.Fatal("configured unknown was accepted", err)
	}
}

func TestProtocolInvalidListenRangesHaveNoWrites(t *testing.T) {
	for _, tc := range []struct{ name, value string }{
		{"CSV member", "192.0.2.0/24,2001:db8::/64"},
		{"malformed IPv4", "192.0.2.0/33"},
		{"malformed IPv6", "2001:db8::/129"},
		{"IPv4 host bits", "192.0.2.1/24"},
		{"IPv6 host bits", "2001:db8::1/64"},
		{"IPv4 host only", "192.0.2.1"},
		{"IPv6 host only", "2001:db8::1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					writes++
				}
				if !frrFixture(w, r) {
					t.Error("invalid set reached API", r.URL.Path)
				}
			}))
			defer api.Close()
			server, schemas := frrProtocolServer(t, api.URL)
			name := "opnsense_quagga_bgp_peergroup"
			schema := schemas.ResourceSchemas[name]
			config := frrProtocolValue(t, schema, map[string]any{"name": "fixture", "listen_ranges": []string{tc.value}})
			null := &tfprotov6.DynamicValue{JSON: []byte("null")}
			plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: config, Config: config})
			if err != nil {
				t.Fatal(err)
			}
			applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: plan.PlannedState, Config: config})
			if err != nil || !frrProtocolErrors(applied.Diagnostics) || writes != 0 {
				t.Fatal("invalid listen range mutated appliance")
			}
		})
	}
}

func TestProtocolOldAbsentSchemaRejectsAdvancedBeforeWrites(t *testing.T) {
	writes := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writes++
		}
		switch r.URL.Path {
		case "/api/core/firmware/info":
			_, _ = w.Write([]byte(`{"package":[{"name":"os-frr","version":"1.46","installed":"1"}]}`))
		case "/api/quagga/bgp/get_neighbor":
			_, _ = w.Write([]byte(`{"neighbor":{"multiprotocol":"0"}}`))
		default:
			if !frrFixture(w, r) {
				t.Error("unsupported intent reached mutation endpoint")
			}
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	values := frrNeighborProtocolValues()
	delete(values, "id")
	values["weight"] = 0
	config := frrProtocolValue(t, schema, values)
	null := &tfprotov6.DynamicValue{JSON: []byte("null")}
	plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: config, Config: config})
	if err != nil {
		t.Fatal(err)
	}
	result, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: plan.PlannedState, Config: config})
	if err != nil || !frrProtocolErrors(result.Diagnostics) || writes != 0 {
		t.Fatal("old absent schema accepted advanced configured intent")
	}
}
