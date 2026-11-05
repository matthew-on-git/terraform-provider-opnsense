// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func finalNeighbor() map[string]any {
	return map[string]any{"remote_as_mode": "", "localas": "", "weight": "", "localip": "", "linklocalinterface": "", "nexthopselfall": "0", "multihop": "0", "rrclient": "0", "soft_reconfiguration_inbound": "0", "bfd": "0", "bfd_strict": "0", "connecttimer": "", "defaultoriginate": "0", "asoverride": "0", "removeprivateas": "", "allowas_in": "", "disable_connected_check": "0", "attributeunchanged": "", "capabilities": []any{}, "enabled": "1", "description": "retained", "address": "192.0.2.1", "remoteas": "64512", "updatesource": "", "nexthopself": "0", "family": "ipv4", "peergroup": "", "keepalive": "", "holddown": "", "linkedPrefixlistIn": []any{}, "linkedPrefixlistOut": []any{}, "linkedRoutemapIn": []any{}, "linkedRoutemapOut": []any{}}
}

func TestFinalNeighborEmptyASRefreshImportDataSource(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if frrFixture(w, r) {
			return
		}
		response := finalNeighbor()
		response["remoteas"] = ""
		_ = json.NewEncoder(w).Encode(map[string]any{"neighbor": response})
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	prior := frrProtocolValue(t, schema, frrNeighborProtocolValues())
	read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: prior})
	if err != nil || frrProtocolErrors(read.Diagnostics) || !frrProtocolAttrs(t, schema, read.NewState)["remote_as"].IsNull() {
		t.Fatal("empty AS not read as null", err, read.Diagnostics)
	}
	imported, err := server.ImportResourceState(context.Background(), &tfprotov6.ImportResourceStateRequest{TypeName: name, ID: "11111111-1111-4111-8111-111111111111"})
	if err != nil || frrProtocolErrors(imported.Diagnostics) {
		t.Fatal(err, imported.Diagnostics)
	}
	read, err = server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: imported.ImportedResources[0].State})
	if err != nil || frrProtocolErrors(read.Diagnostics) || !frrProtocolAttrs(t, schema, read.NewState)["remote_as"].IsNull() {
		t.Fatal("import AS", err, read.Diagnostics)
	}
	ds := schemas.DataSourceSchemas[name]
	data, err := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: name, Config: frrProtocolValue(t, ds, map[string]any{"id": "11111111-1111-4111-8111-111111111111"})})
	if err != nil || frrProtocolErrors(data.Diagnostics) || !frrProtocolAttrs(t, ds, data.State)["remote_as"].IsNull() {
		t.Fatal("data source AS", err, data.Diagnostics)
	}
}

func TestFinalMalformedOwnedReadRetainsState(t *testing.T) {
	for _, tc := range []struct {
		kind, field string
		bad         any
	}{
		{"general", "profile", "bad"},
		{"general", "sysloglevel", "bad"},
		{"bgp", "routerid", "2001:db8::1"},
		{"bgp", "routerid", "192.0.2.1/32"},
		{"bgp", "networks", "198.18.31.1/24"},
		{"bgp", "networks", "198.18.31.0"},
		{"peergroup", "remote_as_mode", "bad"},
		{"redistribution", "redistribute", "bad"},
		{"neighbor", "remoteas", "4294967296"},
		{"neighbor", "linkedRoutemapOut", "a,b"},
		{"neighbor", "peergroup", "stale"},
	} {
		t.Run(tc.kind+"_"+tc.field+fmt.Sprint(tc.bad), func(t *testing.T) {
			monad := tc.kind
			name := "opnsense_quagga_bgp_" + tc.kind
			values := frrNeighborProtocolValues()
			response := finalNeighbor()
			switch tc.kind {
			case "general":
				name = "opnsense_quagga_general"
				values = map[string]any{"id": "general", "enabled": true, "profile": "traditional", "enable_carp": false, "enable_syslog": true, "enable_snmp": false, "syslog_level": "notifications", "firewall_rules": false}
				response = map[string]any{"enabled": "1", "profile": "traditional", "enablecarp": "0", "enablesyslog": "1", "enablesnmp": "0", "sysloglevel": "notifications", "fwrules": "0"}
			case "bgp":
				name = "opnsense_quagga_bgp_global"
				values = map[string]any{"id": "bgp", "enabled": true, "as_number": 65111, "router_id": "198.18.31.1", "distance": 0, "graceful_restart": false, "network_import_check": true, "enforce_first_as": true, "log_neighbor_changes": false, "networks": []string{}, "maximum_paths": 0, "maximum_paths_ibgp": 0}
				response = map[string]any{"enabled": "1", "asnumber": "65111", "routerid": "198.18.31.1", "distance": "", "graceful": "0", "networkimportcheck": "1", "enforce_first_as": "1", "logneighborchanges": "0", "networks": []any{}, "maximumpaths": "", "maximumpathsibgp": ""}
			case "peergroup":
				values = map[string]any{"id": "group", "enabled": true, "name": "group", "remote_as_mode": "", "remote_as": 64512, "family": []string{"ipv4"}, "update_source": "", "next_hop_self": false, "default_originate": false, "linked_prefixlist_in": []string{}, "linked_prefixlist_out": []string{}, "linked_routemap_in": []string{}, "linked_routemap_out": []string{}}
				response = map[string]any{"listenranges": []any{}, "enabled": "1", "name": "group", "remote_as_mode": "", "remoteas": "64512", "family": "ipv4", "updatesource": "", "nexthopself": "0", "defaultoriginate": "0", "linkedPrefixlistIn": []any{}, "linkedPrefixlistOut": []any{}, "linkedRoutemapIn": []any{}, "linkedRoutemapOut": []any{}}
			case "redistribution":
				values = map[string]any{"id": "redistribution", "enabled": true, "description": "retained", "redistribute": "connected", "route_map": ""}
				response = map[string]any{"remote_as_mode": "", "localas": "", "weight": "", "localip": "", "linklocalinterface": "", "nexthopselfall": "0", "multihop": "0", "rrclient": "0", "soft_reconfiguration_inbound": "0", "bfd": "0", "bfd_strict": "0", "connecttimer": "", "defaultoriginate": "0", "asoverride": "0", "removeprivateas": "", "allowas_in": "", "disable_connected_check": "0", "attributeunchanged": "", "capabilities": []any{}, "enabled": "1", "description": "retained", "redistribute": "connected", "linkedRoutemap": ""}
			}
			malformed := false
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if frrFixture(w, r) {
					return
				}
				if strings.Contains(r.URL.Path, "/search_routemap") {
					_, _ = fmt.Fprint(w, `{"rows":[{"uuid":"a"},{"uuid":"b"}],"total":2,"current":1,"rowCount":2}`)
					return
				}
				if strings.Contains(r.URL.Path, "/get_routemap/") {
					_ = json.NewEncoder(w).Encode(map[string]any{"routemap": map[string]any{"name": r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]}})
					return
				}
				if strings.Contains(r.URL.Path, "/get_peergroup/stale") {
					_, _ = fmt.Fprint(w, `{"peergroup":{}}`)
					return
				}
				if malformed {
					response[tc.field] = tc.bad
				}
				_ = json.NewEncoder(w).Encode(map[string]any{monad: response})
			}))
			defer api.Close()
			server, schemas := frrProtocolServer(t, api.URL)
			schema := schemas.ResourceSchemas[name]
			prior := frrProtocolValue(t, schema, values)
			valid, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: prior})
			if err != nil || frrProtocolErrors(valid.Diagnostics) {
				t.Fatal("complete control invalid", err, valid.Diagnostics)
			}
			malformed = true
			bad, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: prior})
			if err != nil || !frrProtocolErrors(bad.Diagnostics) {
				t.Fatal("malformed accepted", err, bad.Diagnostics)
			}
			before, _ := prior.Unmarshal(schema.ValueType())
			after, _ := bad.NewState.Unmarshal(schema.ValueType())
			if !before.Equal(after) {
				t.Fatal("malformed changed prior state")
			}
		})
	}
}

func TestFinalLegacyCreateUpdateWire(t *testing.T) {
	for _, version := range []string{"1.46", "1.55_1"} {
		for _, tc := range []struct {
			address string
			family  []string
			multi   string
		}{{"192.0.2.1", []string{"ipv4"}, "0"}, {"2001:db8::1", []string{"ipv6"}, "0"}, {"192.0.2.1", []string{"ipv4", "ipv6"}, "1"}} {
			t.Run(version+tc.address+tc.multi, func(t *testing.T) {
				writes := 0
				id := "11111111-1111-4111-8111-111111111111"
				api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/api/core/firmware/info":
						_, _ = fmt.Fprintf(w, `{"product_version":"26.7.7","package":[{"name":"os-frr","version":%q,"installed":"1"}]}`, version)
					case "/api/quagga/bgp/get_neighbor":
						template := finalNeighbor()
						if version == "1.46" {
							removeOldAbsentAdvanced(template)
						}
						delete(template, "family")
						template["multiprotocol"] = "0"
						_ = json.NewEncoder(w).Encode(map[string]any{"neighbor": template})
					case "/api/quagga/bgp/search_neighbor":
						_, _ = fmt.Fprint(w, `{"rows":[],"total":0,"current":1,"rowCount":0}`)
					case "/api/quagga/bgp/add_neighbor", "/api/quagga/bgp/set_neighbor/" + id:
						writes++
						var body map[string]map[string]string
						_ = json.NewDecoder(r.Body).Decode(&body)
						if _, ok := body["neighbor"]["family"]; ok {
							t.Error("legacy family wire field")
						}
						if body["neighbor"]["multiprotocol"] != tc.multi {
							t.Error("legacy multiprotocol", body)
						}
						_, _ = fmt.Fprintf(w, `{"result":"saved","uuid":%q}`, id)
					case "/api/quagga/bgp/get_neighbor/" + id:
						response := finalNeighbor()
						if version == "1.46" {
							removeOldAbsentAdvanced(response)
						}
						delete(response, "family")
						response["multiprotocol"] = tc.multi
						response["address"] = tc.address
						_ = json.NewEncoder(w).Encode(map[string]any{"neighbor": response})
					case "/api/quagga/service/reconfigure":
						_, _ = fmt.Fprint(w, `{"status":"ok"}`)
					default:
						if !frrFixture(w, r) {
							t.Error(r.URL.Path)
						}
					}
				}))
				defer api.Close()
				server, schemas := frrProtocolServer(t, api.URL)
				name := "opnsense_quagga_bgp_neighbor"
				schema := schemas.ResourceSchemas[name]
				values := frrNeighborProtocolValues()
				values["address"] = tc.address
				values["family"] = tc.family
				values["multi_protocol"] = tc.multi == "1"
				delete(values, "id")
				var config *tfprotov6.DynamicValue
				prior := &tfprotov6.DynamicValue{JSON: []byte("null")}
				for _, phase := range []string{"create", "update"} {
					values["description"] = phase
					config = frrProtocolValue(t, schema, values)
					plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: prior, ProposedNewState: config, Config: config})
					if err != nil || frrProtocolErrors(plan.Diagnostics) {
						t.Fatal(err, plan.Diagnostics)
					}
					applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: plan.PlannedState, Config: config})
					if err != nil || frrProtocolErrors(applied.Diagnostics) {
						for _, d := range applied.Diagnostics {
							t.Log(d.Summary, d.Detail)
						}
						t.Fatal(err, applied.Diagnostics)
					}
					attrs := frrProtocolAttrs(t, schema, applied.NewState)
					var multi bool
					_ = attrs["multi_protocol"].As(&multi)
					if multi != (tc.multi == "1") {
						t.Fatal("legacy readback alias")
					}
					prior = applied.NewState
				}
				if writes != 2 {
					t.Fatal("legacy mutation count", writes)
				}
			})
		}
	}
}

func TestFinalPeerGroupRelationReadBoundary(t *testing.T) {
	for _, stale := range []bool{false, true} {
		t.Run(fmt.Sprint(stale), func(t *testing.T) {
			reads := 0
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if frrFixture(w, r) {
					return
				}
				if strings.HasSuffix(r.URL.Path, "/get_peergroup/group") {
					reads++
					if stale {
						_, _ = fmt.Fprint(w, `{"peergroup":{}}`)
					} else {
						_, _ = fmt.Fprint(w, `{"peergroup":{"listenranges":[],"enabled":"1","name":"policy","remote_as_mode":"","remoteas":"64512","family":"ipv4","updatesource":"","nexthopself":"0","defaultoriginate":"0","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":[]}}`)
					}
					return
				}
				response := finalNeighbor()
				response["peergroup"] = "group"
				_ = json.NewEncoder(w).Encode(map[string]any{"neighbor": response})
			}))
			defer api.Close()
			server, schemas := frrProtocolServer(t, api.URL)
			name := "opnsense_quagga_bgp_neighbor"
			schema := schemas.ResourceSchemas[name]
			prior := frrProtocolValue(t, schema, frrNeighborProtocolValues())
			read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: prior})
			if err != nil || frrProtocolErrors(read.Diagnostics) != stale {
				t.Fatal("resource relation boundary", err, read.Diagnostics)
			}
			ds := schemas.DataSourceSchemas[name]
			data, err := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: name, Config: frrProtocolValue(t, ds, map[string]any{"id": "neighbor"})})
			if err != nil || frrProtocolErrors(data.Diagnostics) != stale || reads != 2 {
				t.Fatal("data-source relation boundary", err, data.Diagnostics, reads)
			}
		})
	}
}

func removeOldAbsentAdvanced(v map[string]any) {
	for _, wire := range []string{"remote_as_mode", "localas", "weight", "localip", "linklocalinterface", "nexthopselfall", "multihop", "rrclient", "soft_reconfiguration_inbound", "bfd", "bfd_strict", "connecttimer", "defaultoriginate", "asoverride", "removeprivateas", "allowas_in", "disable_connected_check", "attributeunchanged", "capabilities"} {
		delete(v, wire)
	}
}
