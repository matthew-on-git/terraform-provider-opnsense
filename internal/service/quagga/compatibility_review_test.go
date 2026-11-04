// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reviewUnknown(t *testing.T, schema *tfprotov6.Schema, value *tfprotov6.DynamicValue, fields map[string]tftypes.Type) *tfprotov6.DynamicValue {
	t.Helper()
	attrs := frrProtocolAttrs(t, schema, value)
	for key, typ := range fields {
		attrs[key] = tftypes.NewValue(typ, tftypes.UnknownValue)
	}
	result, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
	if err != nil {
		t.Fatal(err)
	}
	return &result
}

func TestReviewTransportFamilyAndUnknownAliasPlanning(t *testing.T) {
	var mutations atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if frrFixture(w, r) {
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/search_neighbor"):
			_, _ = fmt.Fprint(w, `{"rows":[],"current":1,"rowCount":0,"total":0}`)
		case strings.HasSuffix(r.URL.Path, "/add_neighbor"):
			mutations.Add(1)
			_, _ = fmt.Fprint(w, `{"result":"saved","uuid":"22222222-2222-4222-8222-222222222222"}`)
		case r.URL.Path == "/api/quagga/service/reconfigure":
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		case strings.Contains(r.URL.Path, "/get_neighbor/"):
			_, _ = fmt.Fprint(w, `{"neighbor":{"remote_as_mode":"","localas":"","weight":"","localip":"","linklocalinterface":"","nexthopselfall":"0","multihop":"0","rrclient":"0","soft_reconfiguration_inbound":"0","bfd":"0","bfd_strict":"0","connecttimer":"","defaultoriginate":"0","asoverride":"0","removeprivateas":"","allowas_in":"","disable_connected_check":"0","attributeunchanged":"","capabilities":[],"enabled":"1","description":"","address":"2001:db8::1","remoteas":"64512","updatesource":"","nexthopself":"0","family":"ipv6","peergroup":"","keepalive":"","holddown":"","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":[]}}`)
		default:
			t.Errorf("unexpected resolved owner %s", r.URL.Path)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	for _, tc := range []struct {
		name, old, next string
		families        []string
		multi, unknown  bool
		want            string
	}{
		{"forward", "192.0.2.1", "2001:db8::1", []string{"ipv4"}, false, false, "ipv6"},
		{"reverse", "2001:db8::1", "192.0.2.1", []string{"ipv6"}, false, false, "ipv4"},
		{"cross_truth", "192.0.2.1", "192.0.2.2", []string{"ipv6"}, false, false, "ipv6"},
		{"dual", "192.0.2.1", "2001:db8::1", []string{"ipv4", "ipv6"}, true, false, "ipv4,ipv6"},
		{"unknown_transport", "192.0.2.1", "", []string{"ipv4"}, false, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := frrNeighborProtocolValues()
			values["address"] = tc.old
			values["family"] = tc.families
			values["multi_protocol"] = tc.multi
			prior := frrProtocolValue(t, schema, values)
			values["address"] = tc.next
			proposed := frrProtocolValue(t, schema, values)
			config := frrProtocolValue(t, schema, map[string]any{"address": tc.next, "remote_as": 64512})
			if tc.unknown {
				proposed = reviewUnknown(t, schema, proposed, map[string]tftypes.Type{"address": tftypes.String})
				config = reviewUnknown(t, schema, config, map[string]tftypes.Type{"address": tftypes.String})
			}
			plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: prior, ProposedNewState: proposed, Config: config})
			if err != nil || frrProtocolErrors(plan.Diagnostics) {
				t.Fatal(err, plan.Diagnostics)
			}
			af := frrProtocolAttrs(t, schema, plan.PlannedState)["family"]
			if tc.unknown {
				if af.IsKnown() {
					t.Fatal("unknown transport inferred a family")
				}
				return
			}
			var selected []tftypes.Value
			if err := af.As(&selected); err != nil {
				t.Fatal(err)
			}
			got := []string{}
			for _, v := range selected {
				var item string
				_ = v.As(&item)
				got = append(got, item)
			}
			if strings.Join(got, ",") != tc.want {
				t.Fatalf("family %v want %s", got, tc.want)
			}
		})
	}
	for _, resolved := range []string{"2001:db8::1", "192.0.2.1"} {
		t.Run("unknown_alias_"+resolved, func(t *testing.T) {
			values := frrNeighborProtocolValues()
			values["family"] = []string{"ipv6"}
			values["multi_protocol"] = false
			config := frrProtocolValue(t, schema, map[string]any{"address": resolved, "remote_as": 64512, "family": []string{"ipv6"}, "multi_protocol": false})
			unknownConfig := reviewUnknown(t, schema, config, map[string]tftypes.Type{"address": tftypes.String})
			proposed := reviewUnknown(t, schema, frrProtocolValue(t, schema, values), map[string]tftypes.Type{"address": tftypes.String})
			null := &tfprotov6.DynamicValue{JSON: []byte("null")}
			plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: proposed, Config: unknownConfig})
			if err != nil || frrProtocolErrors(plan.Diagnostics) {
				t.Fatal("unknown comparison must defer", err, plan.Diagnostics)
			}
			known := frrProtocolAttrs(t, schema, plan.PlannedState)
			known["address"] = tftypes.NewValue(tftypes.String, resolved)
			appliedPlan, e := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), known))
			if e != nil {
				t.Fatal(e)
			}
			// Replan known resolution: matching IPv6 passes; definite mismatch rejects.
			replanned, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: &appliedPlan, Config: config})
			if err != nil {
				t.Fatal(err)
			}
			if frrProtocolErrors(replanned.Diagnostics) != (resolved == "192.0.2.1") {
				t.Fatal("wrong resolved conflict behavior", replanned.Diagnostics)
			}
			before := mutations.Load()
			applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: &appliedPlan, Config: config})
			if err != nil {
				t.Fatal(err)
			}
			if resolved == "192.0.2.1" {
				if !frrProtocolErrors(applied.Diagnostics) || mutations.Load() != before {
					t.Fatal("resolved conflicting family mutated", applied.Diagnostics)
				}
			} else {
				if frrProtocolErrors(applied.Diagnostics) || mutations.Load() != before+1 {
					t.Fatal("resolved matching family rejected", applied.Diagnostics)
				}
			}
			unresolved, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: plan.PlannedState, Config: unknownConfig})
			if err != nil || !frrProtocolErrors(unresolved.Diagnostics) {
				t.Fatal("unresolved apply did not fail closed", err, unresolved.Diagnostics)
			}
		})
	}
}

func TestReviewUpdateSavedFailuresRetainNewStateAndDoNotReplay(t *testing.T) {
	for _, stage := range []string{"reload", "read"} {
		for _, recovery := range []string{"refresh", "retry"} {
			t.Run(stage+"_"+recovery, func(t *testing.T) {
				writes, reloads := 0, 0
				fail := true
				description := "old"
				advanced := map[string]string{}
				id := "22222222-2222-4222-8222-222222222222"
				api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if frrFixture(w, r) {
						return
					}
					switch r.URL.Path {
					case "/api/quagga/bgp/set_neighbor/" + id:
						writes++
						var body map[string]map[string]string
						_ = json.NewDecoder(r.Body).Decode(&body)
						description = body["neighbor"]["description"]
						advanced = body["neighbor"]
						if advanced["password"] != "configured-test-input" || advanced["weight"] != "0" || advanced["attributeunchanged"] != "as-path" {
							t.Error("saved advanced intent differs")
						}
						_, _ = fmt.Fprint(w, `{"result":"saved"}`)
					case "/api/quagga/service/reconfigure":
						reloads++
						if fail && stage == "reload" {
							_, _ = fmt.Fprint(w, `{"status":"failed"}`)
						} else {
							_, _ = fmt.Fprint(w, `{"status":"ok"}`)
						}
					case "/api/quagga/bgp/get_neighbor/" + id:
						if fail && stage == "read" {
							_, _ = fmt.Fprint(w, `{"neighbor":{"description":null}}`)
							return
						}
						response := finalNeighbor()
						response["description"] = description
						for k, v := range advanced {
							if k != "password" {
								response[k] = v
							}
						}
						response["password"] = "remote-not-config-owned"
						_ = json.NewEncoder(w).Encode(map[string]any{"neighbor": response})
					default:
						t.Errorf("unexpected %s", r.URL.Path)
					}
				}))
				defer api.Close()
				server, schemas := frrProtocolServer(t, api.URL)
				name := "opnsense_quagga_bgp_neighbor"
				schema := schemas.ResourceSchemas[name]
				values := frrNeighborProtocolValues()
				values["id"] = id
				values["description"] = "old"
				prior := frrProtocolValue(t, schema, values)
				values["description"] = "new"
				values["password"] = "configured-test-input"
				values["local_ip"] = "192.0.2.2"
				values["weight"] = 0
				values["attribute_unchanged"] = "as-path"
				values["capabilities"] = []string{"dynamic", "fqdn"}
				config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512, "description": "new", "password": "configured-test-input", "local_ip": "192.0.2.2", "weight": 0, "attribute_unchanged": "as-path", "capabilities": []string{"dynamic", "fqdn"}})
				plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: prior, ProposedNewState: frrProtocolValue(t, schema, values), Config: config})
				if err != nil || frrProtocolErrors(plan.Diagnostics) {
					t.Fatal(err, plan.Diagnostics)
				}
				saved, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: plan.PlannedState, Config: config})
				if err != nil || !frrProtocolErrors(saved.Diagnostics) || len(saved.Private) == 0 {
					t.Fatal("missing saved update recovery", err, saved.Diagnostics)
				}
				checkAdvanced := func(state *tfprotov6.DynamicValue) {
					attrs := frrProtocolAttrs(t, schema, state)
					var secret, attribute, identity string
					_ = attrs["id"].As(&identity)
					var weight int64
					_ = attrs["password"].As(&secret)
					_ = attrs["attribute_unchanged"].As(&attribute)
					_ = attrs["weight"].As(&weight)
					var caps []tftypes.Value
					_ = attrs["capabilities"].As(&caps)
					if identity != id || attrs["weight"].IsNull() || !attrs["weight"].IsKnown() || secret != "configured-test-input" || attribute != "as-path" || weight != 0 || len(caps) != 2 {
						t.Fatal("saved advanced/authentication state lost")
					}
				}
				checkAdvanced(saved.NewState)
				attrs := frrProtocolAttrs(t, schema, saved.NewState)
				var got string
				_ = attrs["description"].As(&got)
				if got != "new" {
					t.Fatal("saved new value lost")
				}
				_ = attrs["id"].As(&got)
				if got != id {
					t.Fatal("UUID lost")
				}
				fail = false
				fresh, _ := frrProtocolServer(t, api.URL)
				if recovery == "refresh" {
					read, e := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: saved.NewState, Private: saved.Private})
					checkAdvanced(read.NewState)
					if len(read.Private) != 0 {
						t.Fatal("pending recovery marker retained after successful refresh")
					}
					if e != nil || frrProtocolErrors(read.Diagnostics) {
						t.Fatal(e, read.Diagnostics)
					}
				} else {
					retry, e := fresh.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: saved.NewState, PlannedState: saved.NewState, Config: config, PlannedPrivate: saved.Private})
					checkAdvanced(retry.NewState)
					if len(retry.Private) != 0 {
						t.Fatal("pending recovery marker retained after successful retry")
					}
					if e != nil || frrProtocolErrors(retry.Diagnostics) {
						t.Fatal(e, retry.Diagnostics)
					}
				}
				if writes != 1 || reloads != 2 {
					t.Fatal("saved update replayed", writes, reloads)
				}
			})
		}
	}
}

func TestReviewCleanupFatalContinuation(t *testing.T) {
	if os.Getenv("FRR_CLEANUP_CHILD") == "1" {
		f := os.NewFile(3, "cleanup-proof")
		t.Cleanup(func() {
			if err := f.Close(); err != nil {
				t.Error(err)
			}
		})
		phases := []func(){func() { t.Fatal("intentional earlier cleanup failure") }}
		for _, phase := range []string{"general", "reload", "disabled-sync", "pf", "final-checks"} {
			phases = append(phases, func() {
				if _, err := fmt.Fprintln(f, phase); err != nil {
					t.Error(err)
				}
			})
		}
		frrRegisterCleanup(t, phases...)
		return
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Error(err)
		}
	}()
	// Tests run in the Linux DevRail container. Execute only this exact test
	// binary and convey the marker through an inherited descriptor, not a path.
	cmd := exec.Command("/proc/self/exe", "-test.run=^TestReviewCleanupFatalContinuation$")
	cmd.ExtraFiles = []*os.File{writer}
	cmd.Env = append(os.Environ(), "FRR_CLEANUP_CHILD=1")
	if cmd.Run() == nil {
		t.Fatal("earlier failure suppressed")
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil || string(data) != "general\nreload\ndisabled-sync\npf\nfinal-checks\n" {
		t.Fatal("later cleanup was skipped or reordered", string(data), err)
	}
}

func TestReviewOraclesRejectWrongSiblingNarrowingAndMisplacedAF(t *testing.T) {
	if frrCanonicalReferences(map[string]string{"linked_routemap_out.#": "1", "linked_routemap_out.0": "sibling-b"}, map[string]json.RawMessage{"linkedRoutemapOut": json.RawMessage(`"sibling-a"`)}) == nil {
		t.Fatal("same-group wrong canonical sibling accepted")
	}
	valid := "address-family ipv4 unicast\nneighbor m_pg next-hop-self\nneighbor m_pg route-map m_rm out\nexit-address-family\naddress-family ipv6 unicast\nneighbor m_pg next-hop-self\nneighbor m_pg route-map m_rm out\nexit-address-family"
	if err := frrAFPolicyProof(valid, "m", false); err != nil {
		t.Fatal(err)
	}
	misplaced := strings.Replace(valid, "address-family ipv6 unicast", "", 1)
	if frrAFPolicyProof(misplaced, "m", false) == nil {
		t.Fatal("global counts hid misplaced AF flags")
	}
	rendered := "pass in proto tcp from {any} to {(self)} port {179} # Pass BGP\npass out proto tcp from {(self)} to {any} port {179} # Pass BGP"
	for _, rules := range []string{"pass in proto tcp from 192.0.2.1 to (self) port = bgp\npass out proto tcp from (self) to any port = bgp", "pass in on vtnet0 proto tcp from any to (self) port = bgp\npass out on vtnet0 proto tcp from (self) to any port = bgp"} {
		if frrBGPFirewallProof(rules, rendered, true) == nil {
			t.Fatal("narrow effective coverage accepted")
		}
	}
}

func TestReviewOmittedUpdateFieldsRetainImportedValues(t *testing.T) {
	for _, kind := range []string{"peergroup", "routemap", "bgp"} {
		t.Run(kind, func(t *testing.T) {
			id := "22222222-2222-4222-8222-222222222222"
			name := "opnsense_quagga_bgp_peergroup"
			field, wire := "linked_routemap_out", "linkedRoutemapOut"
			want := "b"
			values := map[string]any{"id": id, "enabled": true, "name": "old", "remote_as_mode": "", "remote_as": 64512, "family": []string{"ipv6"}, "update_source": "", "next_hop_self": false, "default_originate": false, "linked_prefixlist_in": []string{}, "linked_prefixlist_out": []string{}, "linked_routemap_in": []string{}, "linked_routemap_out": []string{"b"}}
			get := `{"peergroup":{"listenranges":[],"enabled":"1","name":"new","remote_as_mode":"","remoteas":"64512","family":"ipv6","updatesource":"","nexthopself":"0","defaultoriginate":"0","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":"b"}}`
			configValues := map[string]any{"name": "new", "remote_as": 64512}
			if kind == "routemap" {
				name = "opnsense_quagga_route_map"
				field, wire = "match_prefix", "match2"
				values = map[string]any{"id": id, "enabled": true, "description": "old", "name": "policy", "action": "permit", "order": 10, "match_prefix": []string{"b"}, "match_aspath": []string{}, "match_community": []string{}, "set": ""}
				configValues = map[string]any{"name": "policy", "action": "permit", "order": 10, "description": "new"}
				get = `{"routemap":{"enabled":"1","description":"new","name":"policy","action":"permit","id":"10","match":"","match2":"b","match3":"","set":""}}`
			}
			if kind == "bgp" {
				name = "opnsense_quagga_bgp_global"
				id = "bgp"
				field, wire = "networks", "networks"
				want = "198.18.31.0/24"
				values = map[string]any{"id": id, "enabled": true, "as_number": 65111, "router_id": "198.18.31.1", "distance": 0, "graceful_restart": false, "network_import_check": true, "enforce_first_as": true, "log_neighbor_changes": false, "networks": []string{want}, "maximum_paths": 0, "maximum_paths_ibgp": 0}
				configValues = map[string]any{"enabled": true, "as_number": 65111, "router_id": "198.18.31.2"}
				get = `{"bgp":{"enabled":"1","asnumber":"65111","routerid":"198.18.31.2","distance":"","graceful":"0","networkimportcheck":"1","enforce_first_as":"1","logneighborchanges":"0","networks":"198.18.31.0/24","maximumpaths":"","maximumpathsibgp":""}}`
			}
			writes := 0
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if frrFixture(w, r) {
					return
				}
				switch {
				case strings.Contains(r.URL.Path, "/search_"):
					_, _ = fmt.Fprint(w, `{"rows":[{"uuid":"b"}],"current":1,"rowCount":1,"total":1}`)
				case strings.HasSuffix(r.URL.Path, "/get_prefixlist/b"):
					_, _ = fmt.Fprint(w, `{"prefixlist":{"name":"prefix","version":"IPv4"}}`)
				case strings.HasSuffix(r.URL.Path, "/get_routemap/b"):
					_, _ = fmt.Fprint(w, `{"routemap":{"name":"policy","id":"10"}}`)
				case strings.Contains(r.URL.Path, "/set_") || strings.HasSuffix(r.URL.Path, "/bgp/set"):
					writes++
					var body map[string]map[string]string
					_ = json.NewDecoder(r.Body).Decode(&body)
					if body[kind][wire] != want {
						t.Errorf("omitted %s was cleared: %v", field, body)
					}
					if kind == "peergroup" && body[kind]["family"] != "ipv6" {
						t.Error("omitted imported family reset")
					}
					_, _ = fmt.Fprint(w, `{"result":"saved"}`)
				case r.URL.Path == "/api/quagga/service/reconfigure":
					_, _ = fmt.Fprint(w, `{"status":"ok"}`)
				case strings.Contains(r.URL.Path, "/get_") || strings.HasSuffix(r.URL.Path, "/bgp/get"):
					_, _ = fmt.Fprint(w, get)
				default:
					t.Errorf("unexpected %s", r.URL.Path)
				}
			}))
			defer api.Close()
			server, schemas := frrProtocolServer(t, api.URL)
			schema := schemas.ResourceSchemas[name]
			prior := frrProtocolValue(t, schema, values)
			for k, v := range configValues {
				values[k] = v
			}
			planned := reviewUnknown(t, schema, frrProtocolValue(t, schema, values), map[string]tftypes.Type{field: tftypes.Set{ElementType: tftypes.String}})
			if kind == "peergroup" {
				planned = reviewUnknown(t, schema, planned, map[string]tftypes.Type{"family": tftypes.Set{ElementType: tftypes.String}})
			}
			config := frrProtocolValue(t, schema, configValues)
			applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: planned, Config: config})
			if err != nil || frrProtocolErrors(applied.Diagnostics) {
				t.Fatal(err, applied.Diagnostics)
			}
			if writes != 1 {
				t.Fatal("unrelated update did not write exactly once")
			}
			unknownConfig := reviewUnknown(t, schema, config, map[string]tftypes.Type{field: tftypes.Set{ElementType: tftypes.String}})
			rejected, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: planned, Config: unknownConfig})
			if err != nil {
				t.Fatal(err)
			}
			if !frrProtocolErrors(rejected.Diagnostics) || writes != 1 {
				t.Fatal("configured unknown replaced by prior state")
			}
		})
	}
}

func TestReviewNullScalarReadRetainsPriorState(t *testing.T) {
	for _, field := range []string{"description", "keepalive", "holddown"} {
		t.Run(field, func(t *testing.T) {
			var malformed atomic.Bool
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if frrFixture(w, r) {
					return
				}
				response := map[string]any{"remote_as_mode": "", "localas": "", "weight": "", "localip": "", "linklocalinterface": "", "nexthopselfall": "0", "multihop": "0", "rrclient": "0", "soft_reconfiguration_inbound": "0", "bfd": "0", "bfd_strict": "0", "connecttimer": "", "defaultoriginate": "0", "asoverride": "0", "removeprivateas": "", "allowas_in": "", "disable_connected_check": "0", "attributeunchanged": "", "capabilities": []any{}, "enabled": "1", "description": "retain-me", "address": "192.0.2.1", "remoteas": "64512", "updatesource": "", "nexthopself": "0", "family": "ipv4", "peergroup": "", "keepalive": "30", "holddown": "90", "linkedPrefixlistIn": []any{}, "linkedPrefixlistOut": []any{}, "linkedRoutemapIn": []any{}, "linkedRoutemapOut": map[string]any{"": map[string]any{"selected": 1}}}
				if malformed.Load() {
					response[field] = nil
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"neighbor": response}); err != nil {
					t.Error(err)
				}
			}))
			defer api.Close()
			server, schemas := frrProtocolServer(t, api.URL)
			name := "opnsense_quagga_bgp_neighbor"
			schema := schemas.ResourceSchemas[name]
			values := frrNeighborProtocolValues()
			values["description"] = "retain-me"
			values["keepalive"] = 30
			values["holddown"] = 90
			prior := frrProtocolValue(t, schema, values)
			valid, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: prior})
			if err != nil || frrProtocolErrors(valid.Diagnostics) {
				t.Fatal("complete control response invalid", err, valid.Diagnostics)
			}
			malformed.Store(true)
			read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: prior})
			if err != nil || !frrProtocolErrors(read.Diagnostics) {
				t.Fatal("null read accepted", err, read.Diagnostics)
			}
			old, _ := prior.Unmarshal(schema.ValueType())
			current, e := read.NewState.Unmarshal(schema.ValueType())
			if e != nil || !old.Equal(current) {
				t.Fatal("failed malformed read changed prior state", e)
			}
		})
	}
}

func TestReviewGlobalZeroRejectsUntaggedObject(t *testing.T) {
	if os.Getenv("FRR_ZERO_CHILD") == "1" {
		api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["searchPhrase"] != "" {
				t.Error("zero oracle filtered away untagged objects")
			}
			_, _ = fmt.Fprint(w, `{"rows":[{"uuid":"untagged"}],"total":1}`)
		}))
		defer api.Close()
		t.Setenv("OPNSENSE_URI", api.URL)
		t.Setenv("OPNSENSE_API_KEY", "fixture")
		t.Setenv("OPNSENSE_API_SECRET", "fixture")
		frrAssertNamespaceZero(t, "neighbor", "FINAL_CLEANUP")
		return
	}
	cmd := exec.Command("/proc/self/exe", "-test.run=^TestReviewGlobalZeroRejectsUntaggedObject$")
	cmd.Env = append(os.Environ(), "FRR_ZERO_CHILD=1")
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(output), "1 global leftovers") {
		t.Fatalf("untagged leftover hidden: %s", output)
	}
}

func TestReviewPeerGroupCreateDefaultsRemainIPv4(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if frrFixture(w, r) {
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/search_peergroup"):
			_, _ = fmt.Fprint(w, `{"rows":[],"current":1,"rowCount":0,"total":0}`)
		case strings.HasSuffix(r.URL.Path, "/add_peergroup"):
			var body map[string]map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["peergroup"]["family"] != "ipv4" || body["peergroup"]["linkedRoutemapOut"] != "" {
				t.Errorf("wrong create defaults: %v", body)
			}
			_, _ = fmt.Fprint(w, `{"result":"saved","uuid":"22222222-2222-4222-8222-222222222222"}`)
		case r.URL.Path == "/api/quagga/service/reconfigure":
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		default:
			_, _ = fmt.Fprint(w, `{"peergroup":{"listenranges":[],"enabled":"1","name":"group","remote_as_mode":"","remoteas":"64512","family":"ipv4","updatesource":"","nexthopself":"0","defaultoriginate":"0","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":[]}}`)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_peergroup"
	schema := schemas.ResourceSchemas[name]
	config := frrProtocolValue(t, schema, map[string]any{"name": "group", "remote_as": 64512})
	null := &tfprotov6.DynamicValue{JSON: []byte("null")}
	plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: config, Config: config})
	if err != nil || frrProtocolErrors(plan.Diagnostics) {
		t.Fatal(err, plan.Diagnostics)
	}
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: plan.PlannedState, Config: config})
	if err != nil || frrProtocolErrors(applied.Diagnostics) {
		t.Fatal(err, applied.Diagnostics)
	}
}

func TestReviewUnknownTransportKeepsDefiniteAliasConflicts(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !frrFixture(w, r) {
			t.Errorf("conflict reached owner %s", r.URL.Path)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	null := &tfprotov6.DynamicValue{JSON: []byte("null")}
	for _, multi := range []bool{true, false} {
		values := frrNeighborProtocolValues()
		family := []string{"ipv6"}
		if !multi {
			family = []string{"ipv4", "ipv6"}
		}
		values["family"] = family
		values["multi_protocol"] = multi
		config := frrProtocolValue(t, schema, map[string]any{"address": "2001:db8::1", "remote_as": 64512, "family": family, "multi_protocol": multi})
		config = reviewUnknown(t, schema, config, map[string]tftypes.Type{"address": tftypes.String})
		proposed := reviewUnknown(t, schema, frrProtocolValue(t, schema, values), map[string]tftypes.Type{"address": tftypes.String})
		plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: proposed, Config: config})
		if err != nil || !frrProtocolErrors(plan.Diagnostics) {
			t.Fatal("definite conflict wrongly deferred", err, plan.Diagnostics)
		}
	}
}

func TestReviewPartialUnknownPolicySetCannotMutate(t *testing.T) {
	var mutations atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !frrFixture(w, r) {
			mutations.Add(1)
			t.Errorf("partial unknown reached owner %s", r.URL.Path)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	prior := frrProtocolValue(t, schema, frrNeighborProtocolValues())
	attrs := frrProtocolAttrs(t, schema, prior)
	attrs["linked_routemap_out"] = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, tftypes.UnknownValue)})
	partial, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
	if err != nil {
		t.Fatal(err)
	}
	config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512})
	configured := frrProtocolAttrs(t, schema, config)
	configured["linked_routemap_out"] = attrs["linked_routemap_out"]
	unknownConfig, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), configured))
	if err != nil {
		t.Fatal(err)
	}
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: &partial, Config: &unknownConfig})
	if err != nil || !frrProtocolErrors(applied.Diagnostics) || mutations.Load() != 0 {
		t.Fatal("partial unknown was defaulted/cleared", err, applied.Diagnostics)
	}
}

func TestReviewDelayedTransportResolutionDoesNotReuseStaleFamily(t *testing.T) {
	var writes atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if frrFixture(w, r) {
			return
		}
		switch {
		case strings.Contains(r.URL.Path, "/set_neighbor/"):
			writes.Add(1)
			var body map[string]map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["neighbor"]["family"] != "ipv6" {
				t.Errorf("stale family sent: %v", body)
			}
			_, _ = fmt.Fprint(w, `{"result":"saved"}`)
		case r.URL.Path == "/api/quagga/service/reconfigure":
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		default:
			_, _ = fmt.Fprint(w, `{"neighbor":{"remote_as_mode":"","localas":"","weight":"","localip":"","linklocalinterface":"","nexthopselfall":"0","multihop":"0","rrclient":"0","soft_reconfiguration_inbound":"0","bfd":"0","bfd_strict":"0","connecttimer":"","defaultoriginate":"0","asoverride":"0","removeprivateas":"","allowas_in":"","disable_connected_check":"0","attributeunchanged":"","capabilities":[],"enabled":"1","description":"","address":"2001:db8::1","remoteas":"64512","updatesource":"","nexthopself":"0","family":"ipv6","peergroup":"","keepalive":"","holddown":"","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":[]}}`)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	values := frrNeighborProtocolValues()
	prior := frrProtocolValue(t, schema, values)
	proposed := reviewUnknown(t, schema, prior, map[string]tftypes.Type{"address": tftypes.String})
	config := reviewUnknown(t, schema, frrProtocolValue(t, schema, map[string]any{"address": "2001:db8::1", "remote_as": 64512}), map[string]tftypes.Type{"address": tftypes.String})
	plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: prior, ProposedNewState: proposed, Config: config})
	if err != nil || frrProtocolErrors(plan.Diagnostics) {
		t.Fatal(err, plan.Diagnostics)
	}
	attrs := frrProtocolAttrs(t, schema, plan.PlannedState)
	attrs["address"] = tftypes.NewValue(tftypes.String, "2001:db8::1")
	resolved, e := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
	if e != nil {
		t.Fatal(e)
	}
	knownConfig := frrProtocolValue(t, schema, map[string]any{"address": "2001:db8::1", "remote_as": 64512})
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: &resolved, Config: knownConfig})
	if err != nil || frrProtocolErrors(applied.Diagnostics) || writes.Load() != 1 {
		t.Fatal(err, applied.Diagnostics)
	}
}

func TestReviewNullElementPolicySetCannotMutate(t *testing.T) {
	var mutations atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !frrFixture(w, r) {
			mutations.Add(1)
			t.Errorf("partial unknown reached owner %s", r.URL.Path)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	prior := frrProtocolValue(t, schema, frrNeighborProtocolValues())
	attrs := frrProtocolAttrs(t, schema, prior)
	attrs["linked_routemap_out"] = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, nil)})
	partial, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
	if err != nil {
		t.Fatal(err)
	}
	config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512})
	configured := frrProtocolAttrs(t, schema, config)
	configured["linked_routemap_out"] = attrs["linked_routemap_out"]
	unknownConfig, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), configured))
	if err != nil {
		t.Fatal(err)
	}
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: &partial, Config: &unknownConfig})
	if err != nil || !frrProtocolErrors(applied.Diagnostics) || mutations.Load() != 0 {
		t.Fatal("partial unknown was defaulted/cleared", err, applied.Diagnostics)
	}
}

func TestFinalFirewallExactPortToken(t *testing.T) {
	rendered := "pass in proto tcp from {any} to {(self)} port {179} # Pass BGP\npass out proto tcp from {(self)} to {any} port {179} # Pass BGP"
	rules := "pass in proto tcp from any to (self) port = 1790\npass out proto tcp from (self) to any port = 1790"
	if frrBGPFirewallProof(rules, rendered, true) == nil {
		t.Fatal("1790 was accepted as BGP")
	}
	for _, port := range []string{"179", "bgp"} {
		if err := frrBGPFirewallProof(strings.ReplaceAll(rules, "1790", port), rendered, true); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReviewEmptyElementPolicySetCannotMutate(t *testing.T) {
	var mutations atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !frrFixture(w, r) {
			mutations.Add(1)
			t.Errorf("partial unknown reached owner %s", r.URL.Path)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	prior := frrProtocolValue(t, schema, frrNeighborProtocolValues())
	attrs := frrProtocolAttrs(t, schema, prior)
	attrs["linked_routemap_out"] = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "")})
	partial, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
	if err != nil {
		t.Fatal(err)
	}
	config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512})
	configured := frrProtocolAttrs(t, schema, config)
	configured["linked_routemap_out"] = attrs["linked_routemap_out"]
	unknownConfig, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), configured))
	if err != nil {
		t.Fatal(err)
	}
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: &partial, Config: &unknownConfig})
	if err != nil || !frrProtocolErrors(applied.Diagnostics) || mutations.Load() != 0 {
		t.Fatal("partial unknown was defaulted/cleared", err, applied.Diagnostics)
	}
}

func TestReviewWhitespaceElementPolicySetCannotMutate(t *testing.T) {
	var mutations atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !frrFixture(w, r) {
			mutations.Add(1)
			t.Errorf("partial unknown reached owner %s", r.URL.Path)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	prior := frrProtocolValue(t, schema, frrNeighborProtocolValues())
	attrs := frrProtocolAttrs(t, schema, prior)
	attrs["linked_routemap_out"] = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, " \t ")})
	partial, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
	if err != nil {
		t.Fatal(err)
	}
	config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512})
	configured := frrProtocolAttrs(t, schema, config)
	configured["linked_routemap_out"] = attrs["linked_routemap_out"]
	unknownConfig, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), configured))
	if err != nil {
		t.Fatal(err)
	}
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: &partial, Config: &unknownConfig})
	if err != nil || !frrProtocolErrors(applied.Diagnostics) || mutations.Load() != 0 {
		t.Fatal("partial unknown was defaulted/cleared", err, applied.Diagnostics)
	}
}
