// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

func frrProtocolServer(t *testing.T, uri string) (tfprotov6.ProviderServer, *tfprotov6.GetProviderSchemaResponse) {
	t.Helper()
	server := providerserver.NewProtocol6(provider.New("test")())()
	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	config := frrProtocolValue(t, schemas.Provider, map[string]any{"uri": uri, "api_key": "fixture-key", "api_secret": "fixture-secret", "insecure": true})
	configured, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: config, TerraformVersion: "1.14.0"})
	if err != nil || frrProtocolErrors(configured.Diagnostics) {
		t.Fatal("configure", err, configured.Diagnostics)
	}
	return server, schemas
}

func frrProtocolValue(t *testing.T, schema *tfprotov6.Schema, values map[string]any) *tfprotov6.DynamicValue {
	t.Helper()
	complete := map[string]any{}
	for _, attr := range schema.Block.Attributes {
		complete[attr.Name] = nil
	}
	for k, v := range values {
		complete[k] = v
	}
	raw, err := json.Marshal(complete)
	if err != nil {
		t.Fatal(err)
	}
	return &tfprotov6.DynamicValue{JSON: raw}
}

func frrProtocolErrors(ds []*tfprotov6.Diagnostic) bool {
	for _, d := range ds {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func frrProtocolAttrs(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue) map[string]tftypes.Value {
	t.Helper()
	value, err := state.Unmarshal(schema.ValueType())
	if err != nil {
		t.Fatal(err)
	}
	var attrs map[string]tftypes.Value
	if err = value.As(&attrs); err != nil {
		t.Fatal(err)
	}
	return attrs
}

func frrFixture(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path == "/api/core/menu/search" {
		_, _ = fmt.Fprint(w, `{}`)
		return true
	}
	if r.URL.Path == "/api/core/firmware/info" {
		_, _ = fmt.Fprint(w, `{"product_version":"26.7.7","package":[{"name":"os-frr","version":"2.0_1","installed":"1"}]}`)
		return true
	}
	if r.URL.Path == "/api/quagga/bgp/get_neighbor" {
		_, _ = fmt.Fprint(w, `{"neighbor":{"family":{"ipv4":{"value":"IPv4","selected":1},"ipv6":{"value":"IPv6","selected":0}}}}`)
		return true
	}
	return false
}

func TestProtocolFRRVersionZeroUpgradePlanApplyRefresh(t *testing.T) {
	for _, tc := range []struct {
		name, address string
		multi         bool
		family        string
	}{{"ipv4", "192.0.2.1", false, "ipv4"}, {"ipv6_only", "2001:db8::1", false, "ipv6"}, {"dual", "192.0.2.1", true, "ipv4,ipv6"}} {
		t.Run(tc.name, func(t *testing.T) {
			writes := 0
			family := tc.family
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if frrFixture(w, r) {
					return
				}
				switch r.URL.Path {
				case "/api/quagga/bgp/get_neighbor/stable-id":
					_, _ = fmt.Fprintf(w, `{"neighbor":{"remote_as_mode":"","localas":"","weight":"","localip":"","linklocalinterface":"","nexthopselfall":"0","multihop":"0","rrclient":"0","soft_reconfiguration_inbound":"0","bfd":"0","bfd_strict":"0","connecttimer":"","defaultoriginate":"0","asoverride":"0","removeprivateas":"","allowas_in":"","disable_connected_check":"0","attributeunchanged":"","capabilities":[],"enabled":"1","description":"","address":%q,"remoteas":"64512","updatesource":"","nexthopself":"0","family":%q,"peergroup":"","keepalive":"","holddown":"","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":[]}}`, tc.address, family)
				case "/api/quagga/bgp/set_neighbor/stable-id":
					writes++
					var body map[string]map[string]string
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if _, ok := body["neighbor"]["multiprotocol"]; ok {
						t.Error("removed wire alias was sent")
					}
					family = body["neighbor"]["family"]
					_, _ = fmt.Fprint(w, `{"result":"saved"}`)
				case "/api/quagga/service/reconfigure":
					_, _ = fmt.Fprint(w, `{"status":"ok"}`)
				default:
					t.Errorf("unexpected API %s", r.URL.Path)
				}
			}))
			defer api.Close()
			server, schemas := frrProtocolServer(t, api.URL)
			name := "opnsense_quagga_bgp_neighbor"
			schema := schemas.ResourceSchemas[name]
			legacy := map[string]any{"id": "stable-id", "enabled": true, "description": "", "address": tc.address, "remote_as": 64512, "update_source": "", "next_hop_self": false, "multi_protocol": tc.multi, "keepalive": nil, "holddown": nil, "linked_prefixlist_in": []string{}, "linked_prefixlist_out": []string{}, "linked_routemap_in": []string{}, "linked_routemap_out": []string{}}
			raw, _ := json.Marshal(legacy)
			upgraded, err := server.UpgradeResourceState(context.Background(), &tfprotov6.UpgradeResourceStateRequest{TypeName: name, Version: 0, RawState: &tfprotov6.RawState{JSON: raw}})
			if err != nil || frrProtocolErrors(upgraded.Diagnostics) {
				t.Fatal(err, upgraded.Diagnostics)
			}
			attrs := frrProtocolAttrs(t, schema, upgraded.UpgradedState)
			var id string
			if err := attrs["id"].As(&id); err != nil || id != "stable-id" {
				t.Fatal("UUID changed", err)
			}
			config := frrProtocolValue(t, schema, map[string]any{"address": tc.address, "remote_as": 64512, "multi_protocol": tc.multi})
			plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: upgraded.UpgradedState, ProposedNewState: upgraded.UpgradedState, Config: config})
			if err != nil || frrProtocolErrors(plan.Diagnostics) {
				t.Fatal(err, plan.Diagnostics)
			}
			applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: upgraded.UpgradedState, PlannedState: plan.PlannedState, Config: config, PlannedPrivate: plan.PlannedPrivate})
			if err != nil || frrProtocolErrors(applied.Diagnostics) {
				t.Fatal(err, applied.Diagnostics)
			}
			refreshed, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: applied.NewState, Private: applied.Private})
			if err != nil || frrProtocolErrors(refreshed.Diagnostics) {
				t.Fatal(err, refreshed.Diagnostics)
			}
			converged, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: refreshed.NewState, ProposedNewState: refreshed.NewState, Config: config, PriorPrivate: refreshed.Private})
			if err != nil || frrProtocolErrors(converged.Diagnostics) {
				t.Fatal(err, converged.Diagnostics)
			}
			before, _ := refreshed.NewState.Unmarshal(schema.ValueType())
			after, _ := converged.PlannedState.Unmarshal(schema.ValueType())
			if !reflect.DeepEqual(before, after) || writes != 1 {
				t.Fatalf("legacy plan did not converge: writes=%d", writes)
			}
		})
	}
}

func TestProtocolFRRSavedCreateReloadFailureRestartRecovery(t *testing.T) {
	testFRRSavedCreateRecovery(t, "reload")
}

func TestProtocolFRRSavedCreateReadFailureRestartRecovery(t *testing.T) {
	testFRRSavedCreateRecovery(t, "read")
}

func testFRRSavedCreateRecovery(t *testing.T, stage string) {
	writes, reloads := 0, 0
	fail := true
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if frrFixture(w, r) {
			return
		}
		switch r.URL.Path {
		case "/api/quagga/bgp/search_neighbor":
			_, _ = fmt.Fprint(w, `{"rows":[],"current":1,"rowCount":0,"total":0}`)
		case "/api/quagga/bgp/add_neighbor":
			writes++
			_, _ = fmt.Fprint(w, `{"result":"saved","uuid":"22222222-2222-4222-8222-222222222222"}`)
		case "/api/quagga/service/reconfigure":
			reloads++
			if fail && stage == "reload" {
				_, _ = fmt.Fprint(w, `{"status":"fixture-secret"}`)
			} else {
				_, _ = fmt.Fprint(w, `{"status":"ok"}`)
			}
		case "/api/quagga/bgp/get_neighbor/22222222-2222-4222-8222-222222222222":
			if fail && stage == "read" {
				_, _ = fmt.Fprint(w, `{"neighbor":{}}`)
				return
			}
			_, _ = fmt.Fprint(w, `{"neighbor":{"remote_as_mode":"","localas":"","weight":"","localip":"","linklocalinterface":"","nexthopselfall":"0","multihop":"0","rrclient":"0","soft_reconfiguration_inbound":"0","bfd":"0","bfd_strict":"0","connecttimer":"","defaultoriginate":"0","asoverride":"0","removeprivateas":"","allowas_in":"","disable_connected_check":"0","attributeunchanged":"","capabilities":[],"enabled":"1","description":"","address":"192.0.2.1","remoteas":"64512","updatesource":"","nexthopself":"0","family":"ipv4","peergroup":"","keepalive":"","holddown":"","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":[]}}`)
		default:
			t.Errorf("unexpected API %s", r.URL.Path)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512})
	proposed := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512, "family": []string{"ipv4"}, "enabled": true, "description": "", "next_hop_self": false, "multi_protocol": false, "peer_group": "", "update_source": "", "linked_prefixlist_in": []string{}, "linked_prefixlist_out": []string{}, "linked_routemap_in": []string{}, "linked_routemap_out": []string{}})
	null := &tfprotov6.DynamicValue{JSON: []byte("null")}
	plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: proposed, Config: config})
	if err != nil || frrProtocolErrors(plan.Diagnostics) {
		t.Fatal(err, plan.Diagnostics)
	}
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: plan.PlannedState, Config: config, PlannedPrivate: plan.PlannedPrivate})
	if err != nil || !frrProtocolErrors(applied.Diagnostics) || len(applied.Private) == 0 {
		t.Fatal("saved reload failure lost private state", err, applied.Diagnostics)
	}
	diagJSON, _ := json.Marshal(applied.Diagnostics)
	if strings.Contains(string(diagJSON), "fixture-secret") {
		t.Fatal("reload response leaked configured API secret")
	}
	fail = false
	fresh, _ := frrProtocolServer(t, api.URL)
	refreshed, err := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: applied.NewState, Private: applied.Private})
	if err != nil || frrProtocolErrors(refreshed.Diagnostics) {
		t.Fatal(err, refreshed.Diagnostics)
	}
	attrs := frrProtocolAttrs(t, schema, refreshed.NewState)
	var id string
	_ = attrs["id"].As(&id)
	if id != "22222222-2222-4222-8222-222222222222" || writes != 1 || reloads != 2 {
		t.Fatalf("mutation repeated or identity lost: id=%s writes=%d reloads=%d", id, writes, reloads)
	}
}

func TestProtocolFRROwnersSerializeDiscoveryAndMutation(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	secondDiscovery := make(chan struct{}, 1)
	var discoveries atomic.Int32
	var mutations atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/core/firmware/info" {
			if discoveries.Add(1) > 1 {
				select {
				case secondDiscovery <- struct{}{}:
				default:
				}
			}
		}
		if frrFixture(w, r) {
			return
		}
		switch r.URL.Path {
		case "/api/quagga/bgp/search_neighbor":
			_, _ = fmt.Fprint(w, `{"rows":[],"current":1,"rowCount":0,"total":0}`)
		case "/api/quagga/bgp/add_neighbor":
			mutations.Add(1)
			_, _ = fmt.Fprint(w, `{"result":"saved","uuid":"11111111-1111-4111-8111-111111111111"}`)
		case "/api/quagga/service/reconfigure":
			close(entered)
			<-release
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		case "/api/quagga/bgp/get_neighbor/11111111-1111-4111-8111-111111111111":
			_, _ = fmt.Fprint(w, `{"neighbor":{"remote_as_mode":"","localas":"","weight":"","localip":"","linklocalinterface":"","nexthopselfall":"0","multihop":"0","rrclient":"0","soft_reconfiguration_inbound":"0","bfd":"0","bfd_strict":"0","connecttimer":"","defaultoriginate":"0","asoverride":"0","removeprivateas":"","allowas_in":"","disable_connected_check":"0","attributeunchanged":"","capabilities":[],"enabled":"1","description":"","address":"192.0.2.1","remoteas":"64512","updatesource":"","nexthopself":"0","family":"ipv4","peergroup":"","keepalive":"","holddown":"","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":[]}}`)
		default:
			t.Errorf("unexpected owner handshake request %s", r.URL.Path)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	null := &tfprotov6.DynamicValue{JSON: []byte("null")}
	values := frrNeighborProtocolValues()
	values["id"] = nil
	config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512})
	planned, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: frrProtocolValue(t, schema, values), Config: config})
	if err != nil || frrProtocolErrors(planned.Diagnostics) {
		t.Fatal(err, planned.Diagnostics)
	}
	done := make(chan bool, 2)
	go func() {
		result, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: planned.PlannedState, Config: config, PlannedPrivate: planned.PlannedPrivate})
		done <- err == nil && !frrProtocolErrors(result.Diagnostics)
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("create did not reach saved-mutation reload")
	}
	canceledCtx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	canceledDone := make(chan bool, 1)
	go func() {
		result, err := server.ApplyResourceChange(canceledCtx, &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: planned.PlannedState, Config: config, PlannedPrivate: planned.PlannedPrivate})
		canceledDone <- err != nil || frrProtocolErrors(result.Diagnostics)
	}()
	select {
	case failed := <-canceledDone:
		if !failed {
			t.Error("canceled owner waiter succeeded")
		}
	case <-time.After(time.Second):
		close(release)
		t.Fatal("canceled owner waiter ignored cancellation")
	}
	if discoveries.Load() != 1 || mutations.Load() != 1 {
		t.Fatal("canceled owner waiter reached discovery")
	}
	started := make(chan struct{})
	go func() {
		close(started)
		result, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: frrProtocolValue(t, schema, frrNeighborProtocolValues())})
		done <- err == nil && !frrProtocolErrors(result.Diagnostics)
	}()
	<-started
	select {
	case <-secondDiscovery:
		close(release)
		t.Fatal("another owner discovered capabilities while save/reload remained in progress")
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	for range 2 {
		select {
		case ok := <-done:
			if !ok {
				t.Fatal("serialized owner failed")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("owner handshake deadlocked")
		}
	}
	if discoveries.Load() != 2 {
		t.Fatal("expected independently validated discovery for both owner operations")
	}
}

func frrNeighborProtocolValues() map[string]any {
	return map[string]any{"id": "11111111-1111-4111-8111-111111111111", "address": "192.0.2.1", "remote_as": 64512, "family": []string{"ipv4"}, "enabled": true, "description": "", "next_hop_self": false, "multi_protocol": false, "update_source": "", "peer_group": "", "linked_prefixlist_in": []string{}, "linked_prefixlist_out": []string{}, "linked_routemap_in": []string{}, "linked_routemap_out": []string{}}
}

func TestProtocolFRRRevalidatesPackageBeforeMutation(t *testing.T) {
	version := "2.0_1"
	writes := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/core/firmware/info" {
			_, _ = fmt.Fprintf(w, `{"package":[{"name":"os-frr","installed":"1","version":%q}]}`, version)
			return
		}
		if frrFixture(w, r) {
			return
		}
		writes++
		_, _ = fmt.Fprint(w, `{"result":"saved","uuid":"11111111-1111-4111-8111-111111111111"}`)
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	values := frrNeighborProtocolValues()
	values["id"] = nil
	config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512, "family": []string{"ipv4"}})
	null := &tfprotov6.DynamicValue{JSON: []byte("null")}
	planned, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: frrProtocolValue(t, schema, values), Config: config})
	if err != nil || frrProtocolErrors(planned.Diagnostics) {
		t.Fatal(err, planned.Diagnostics)
	}
	version = "2.1"
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: planned.PlannedState, Config: config, PlannedPrivate: planned.PlannedPrivate})
	if err != nil || !frrProtocolErrors(applied.Diagnostics) || writes != 0 {
		t.Fatal("package changed after planning reached mutation", err, applied.Diagnostics, writes)
	}
}

func TestProtocolFRRConflictingAndUnknownFamilyFailClosed(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/core/menu/search" {
			_, _ = fmt.Fprint(w, `{}`)
			return
		}
		t.Errorf("invalid family reached appliance %s", r.URL.Path)
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	null := &tfprotov6.DynamicValue{JSON: []byte("null")}
	for _, tc := range []struct {
		name   string
		family any
		multi  any
	}{{"conflict", []string{"ipv6"}, false}, {"empty", []string{}, nil}, {"uppercase", []string{"IPv4"}, nil}} {
		t.Run(tc.name, func(t *testing.T) {
			values := frrNeighborProtocolValues()
			values["id"] = nil
			values["family"] = tc.family
			cfg := map[string]any{"address": "192.0.2.1", "remote_as": 64512, "family": tc.family, "multi_protocol": tc.multi}
			planned, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: frrProtocolValue(t, schema, values), Config: frrProtocolValue(t, schema, cfg)})
			if err != nil || !frrProtocolErrors(planned.Diagnostics) {
				t.Fatal("invalid aliases planned successfully", err, planned.Diagnostics)
			}
		})
	}
	values := frrNeighborProtocolValues()
	v, err := frrProtocolValue(t, schema, values).Unmarshal(schema.ValueType())
	if err != nil {
		t.Fatal(err)
	}
	var attributes map[string]tftypes.Value
	if v.As(&attributes) != nil {
		t.Fatal("cannot construct unknown family")
	}
	attributes["id"] = tftypes.NewValue(tftypes.String, nil)
	attributes["family"] = tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, tftypes.UnknownValue)
	unknown, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attributes))
	if err != nil {
		t.Fatal(err)
	}
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: &unknown, Config: &unknown})
	if err != nil || !frrProtocolErrors(applied.Diagnostics) {
		t.Fatal("configured unknown family became default/clear", err, applied.Diagnostics)
	}
}

func TestProtocolFRRDeleteFailureRestartDoesNotReplay(t *testing.T) {
	for _, tc := range []struct{ name, body string }{{"affirmed", `{"result":"deleted"}`}, {"ambiguous_absence_confirmed", `{}`}} {
		t.Run(tc.name, func(t *testing.T) {
			deletes, reloads := 0, 0
			fail := true
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if frrFixture(w, r) {
					return
				}
				switch r.URL.Path {
				case "/api/quagga/bgp/del_neighbor/11111111-1111-4111-8111-111111111111":
					deletes++
					_, _ = fmt.Fprint(w, tc.body)
				case "/api/quagga/bgp/get_neighbor/11111111-1111-4111-8111-111111111111":
					_, _ = fmt.Fprint(w, `{"neighbor":[]}`)
				case "/api/quagga/service/reconfigure":
					reloads++
					if fail {
						_, _ = fmt.Fprint(w, `{"status":"failed"}`)
					} else {
						_, _ = fmt.Fprint(w, `{"status":"ok"}`)
					}
				default:
					t.Errorf("unexpected delete recovery request %s", r.URL.Path)
				}
			}))
			defer api.Close()
			server, schemas := frrProtocolServer(t, api.URL)
			name := "opnsense_quagga_bgp_neighbor"
			schema := schemas.ResourceSchemas[name]
			null := &tfprotov6.DynamicValue{JSON: []byte("null")}
			response, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: frrProtocolValue(t, schema, frrNeighborProtocolValues()), PlannedState: null, Config: null})
			if err != nil || !frrProtocolErrors(response.Diagnostics) || len(response.Private) == 0 {
				t.Fatal("delete activation failure lost durable private state", err, response.Diagnostics)
			}
			fail = false
			fresh, _ := frrProtocolServer(t, api.URL)
			recovered, err := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: response.NewState, Private: response.Private})
			if err != nil || frrProtocolErrors(recovered.Diagnostics) {
				t.Fatal(err, recovered.Diagnostics)
			}
			current, err := recovered.NewState.Unmarshal(schema.ValueType())
			if err != nil || !current.IsNull() || deletes != 1 || reloads != 2 {
				t.Fatal("delete replayed or recovery lost state", err, deletes, reloads)
			}
		})
	}
}

func TestProtocolFRRAmbiguousCreateUniqueIdentityRecovery(t *testing.T) {
	created := false
	writes, reloads := 0, 0
	const id = "33333333-3333-4333-8333-333333333333"
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if frrFixture(w, r) {
			return
		}
		switch r.URL.Path {
		case "/api/quagga/bgp/search_neighbor":
			if created {
				_, _ = fmt.Fprintf(w, `{"rows":[{"uuid":%q}],"current":1,"rowCount":1,"total":1}`, id)
			} else {
				_, _ = fmt.Fprint(w, `{"rows":[],"current":1,"rowCount":0,"total":0}`)
			}
		case "/api/quagga/bgp/add_neighbor":
			created = true
			writes++
			_, _ = fmt.Fprint(w, `{"result":"unknown"}`)
		case "/api/quagga/bgp/get_neighbor/" + id:
			_, _ = fmt.Fprint(w, `{"neighbor":{"remote_as_mode":"","localas":"","weight":"","localip":"","linklocalinterface":"","nexthopselfall":"0","multihop":"0","rrclient":"0","soft_reconfiguration_inbound":"0","bfd":"0","bfd_strict":"0","connecttimer":"","defaultoriginate":"0","asoverride":"0","removeprivateas":"","allowas_in":"","disable_connected_check":"0","attributeunchanged":"","capabilities":[],"enabled":"1","description":"","address":"192.0.2.1","remoteas":"64512","updatesource":"","nexthopself":"0","family":"ipv4","keepalive":"","holddown":"","peergroup":"","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":[]}}`)
		case "/api/quagga/service/reconfigure":
			reloads++
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		default:
			t.Errorf("unexpected ambiguous recovery API %s", r.URL.Path)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	values := frrNeighborProtocolValues()
	values["id"] = nil
	config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512})
	null := &tfprotov6.DynamicValue{JSON: []byte("null")}
	plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: frrProtocolValue(t, schema, values), Config: config})
	if err != nil || frrProtocolErrors(plan.Diagnostics) {
		t.Fatal(err, plan.Diagnostics)
	}
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: plan.PlannedState, Config: config, PlannedPrivate: plan.PlannedPrivate})
	if err != nil || !frrProtocolErrors(applied.Diagnostics) || len(applied.Private) == 0 {
		t.Fatal("ambiguous saved create did not retain proven identity", err, applied.Diagnostics)
	}
	fresh, _ := frrProtocolServer(t, api.URL)
	recovered, err := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: applied.NewState, Private: applied.Private})
	if err != nil || frrProtocolErrors(recovered.Diagnostics) {
		t.Fatal(err, recovered.Diagnostics)
	}
	attrs := frrProtocolAttrs(t, schema, recovered.NewState)
	var got string
	_ = attrs["id"].As(&got)
	if got != id || writes != 1 || reloads != 1 {
		t.Fatal("ambiguous save replayed or identity changed", writes, reloads)
	}
}

func TestFRRPFOracleRequiresDirectionsFamiliesAndSourceMeaning(t *testing.T) {
	rendered := "pass in log quick proto tcp from {any} to {(self)} port {179} keep state # Pass BGP\npass out log quick proto tcp from {(self)} to {any} port {179} keep state # Pass BGP"
	expanded := "pass in log quick inet proto tcp from {any} to {(self)} port {179} keep state # Pass BGP\npass in log quick inet6 proto tcp from {any} to {(self)} port {179} keep state # Pass BGP\npass out log quick inet proto tcp from {(self)} to {any} port {179} keep state # Pass BGP\npass out log quick inet6 proto tcp from {(self)} to {any} port {179} keep state # Pass BGP"
	completeRules := "pass in quick proto tcp from any to self port = bgp\npass out quick proto tcp from self to any port = bgp"
	if err := frrBGPFirewallProof(completeRules, expanded, true); err != nil {
		t.Fatal("actual core-expanded inet46 rendering rejected", err)
	}
	for _, incomplete := range []string{
		"pass in log quick inet proto tcp from {any} to {(self)} port {179} keep state # Pass BGP\npass out log quick inet proto tcp from {(self)} to {any} port {179} keep state # Pass BGP",
		"pass in log quick proto tcp from {any} to {(self)} port {179} keep state # Pass BGP",
		strings.ReplaceAll(rendered, "from {any}", "from {192.0.2.1}"),
	} {
		if frrBGPFirewallProof(completeRules, incomplete, true) == nil {
			t.Fatal("incomplete or narrowed generated PF proof accepted")
		}
	}
	for _, rules := range []string{"pass in quick proto tcp from any to self port = bgp\npass out quick proto tcp from self to any port = bgp", "pass in quick inet proto tcp from any to self port = 179\npass out quick inet proto tcp from self to any port = 179\npass in quick inet6 proto tcp from any to self port = 179\npass out quick inet6 proto tcp from self to any port = 179"} {
		if err := frrBGPFirewallProof(rules, rendered, true); err != nil {
			t.Fatal(err)
		}
	}
	for _, rules := range []string{"pass in quick proto tcp from any to self port = 443", "pass in quick inet proto tcp from any to self port = 179\npass out quick inet proto tcp from self to any port = 179", "pass in quick proto tcp from any to self port = 179"} {
		if frrBGPFirewallProof(rules, rendered, true) == nil {
			t.Fatal("incomplete effective PF proof accepted")
		}
	}
	if frrBGPFirewallProof("pass out quick proto tcp from self to any port = bgp", "", false) == nil {
		t.Fatal("effective PF leftovers accepted during opt-out")
	}
}

func TestProtocolFRRSingletonSavedSettingsRestartRecovery(t *testing.T) {
	for _, kind := range []string{"general", "bgp"} {
		for _, recovery := range []string{"refresh", "unchanged_update"} {
			t.Run(kind+"_"+recovery, func(t *testing.T) {
				writes, reloads := 0, 0
				fail := true
				values := map[string]any{"enabled": true, "profile": "traditional", "enable_carp": false, "enable_syslog": true, "enable_snmp": false, "syslog_level": "notifications", "firewall_rules": false}
				get := `{"general":{"enabled":"1","profile":"traditional","enablecarp":"0","enablesyslog":"1","enablesnmp":"0","sysloglevel":"notifications","fwrules":"0"}}`
				name := "opnsense_quagga_general"
				if kind == "bgp" {
					name = "opnsense_quagga_bgp_global"
					values = map[string]any{"enabled": true, "as_number": 65111, "router_id": "198.18.31.1", "distance": 0, "graceful_restart": false, "network_import_check": true, "enforce_first_as": true, "log_neighbor_changes": false, "networks": []string{}, "maximum_paths": 0, "maximum_paths_ibgp": 0}
					get = `{"bgp":{"enabled":"1","asnumber":"65111","routerid":"198.18.31.1","distance":"","graceful":"0","networkimportcheck":"1","enforce_first_as":"1","logneighborchanges":"0","networks":[],"maximumpaths":"","maximumpathsibgp":""}}`
				}
				api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if frrFixture(w, r) {
						return
					}
					switch r.URL.Path {
					case "/api/quagga/" + kind + "/set":
						writes++
						_, _ = fmt.Fprint(w, `{"result":"saved"}`)
					case "/api/quagga/" + kind + "/get":
						_, _ = fmt.Fprint(w, get)
					case "/api/quagga/service/reconfigure":
						reloads++
						if fail {
							_, _ = fmt.Fprint(w, `{"status":"failed"}`)
						} else {
							_, _ = fmt.Fprint(w, `{"status":"ok"}`)
						}
					default:
						t.Errorf("unexpected singleton request %s", r.URL.Path)
					}
				}))
				defer api.Close()
				server, schemas := frrProtocolServer(t, api.URL)
				schema := schemas.ResourceSchemas[name]
				null := &tfprotov6.DynamicValue{JSON: []byte("null")}
				config := frrProtocolValue(t, schema, values)
				plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: null, ProposedNewState: config, Config: config})
				if err != nil || frrProtocolErrors(plan.Diagnostics) {
					t.Fatal(err, plan.Diagnostics)
				}
				applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: null, PlannedState: plan.PlannedState, Config: config, PlannedPrivate: plan.PlannedPrivate})
				if err != nil || !frrProtocolErrors(applied.Diagnostics) || len(applied.Private) == 0 {
					t.Fatal("saved singleton settings lost recovery state", err, applied.Diagnostics)
				}
				fail = false
				fresh, _ := frrProtocolServer(t, api.URL)
				if recovery == "unchanged_update" {
					updated, err := fresh.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: applied.NewState, PlannedState: applied.NewState, Config: config, PlannedPrivate: applied.Private})
					if err != nil || frrProtocolErrors(updated.Diagnostics) {
						t.Fatal(err, updated.Diagnostics)
					}
				} else {
					refreshed, err := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: applied.NewState, Private: applied.Private})
					if err != nil || frrProtocolErrors(refreshed.Diagnostics) {
						t.Fatal(err, refreshed.Diagnostics)
					}
				}
				if writes != 1 || reloads != 2 {
					t.Fatal("singleton recovery repeated a saved mutation", writes, reloads)
				}
			})
		}
	}
}

func TestProtocolFRRExplicitDualFamilyPreservesOmittedStateAfterAliasDefault(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !frrFixture(w, r) {
			t.Errorf("planning must not mutate or read owners: %s", r.URL.Path)
		}
	}))
	defer api.Close()
	server, schemas := frrProtocolServer(t, api.URL)
	name := "opnsense_quagga_bgp_neighbor"
	schema := schemas.ResourceSchemas[name]
	values := map[string]any{"id": "stable-id", "enabled": true, "description": "", "address": "192.0.2.1", "remote_as": 64512, "update_source": "", "next_hop_self": false, "multi_protocol": true, "family": []string{"ipv4", "ipv6"}, "peer_group": "", "linked_prefixlist_in": []string{}, "linked_prefixlist_out": []string{}, "linked_routemap_in": []string{}, "linked_routemap_out": []string{}}
	prior := frrProtocolValue(t, schema, values)
	// Core applies the omitted alias's false schema default before ModifyPlan
	// reconciles explicit dual family. This initially marks other computed
	// fields unknown, even though the final alias value is unchanged.
	values["multi_protocol"] = false
	proposed := frrProtocolValue(t, schema, values)
	config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512, "family": []string{"ipv4", "ipv6"}})
	plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: prior, ProposedNewState: proposed, Config: config})
	if err != nil || frrProtocolErrors(plan.Diagnostics) {
		t.Fatal(err, plan.Diagnostics)
	}
	before, _ := prior.Unmarshal(schema.ValueType())
	after, _ := plan.PlannedState.Unmarshal(schema.ValueType())
	if !before.Equal(after) {
		priorAttrs, plannedAttrs := frrProtocolAttrs(t, schema, prior), frrProtocolAttrs(t, schema, plan.PlannedState)
		for key, value := range priorAttrs {
			if !value.Equal(plannedAttrs[key]) {
				t.Logf("changed attribute %s: %s -> %s", key, value.String(), plannedAttrs[key].String())
			}
		}
		t.Fatal("alias reconciliation left phantom unknown changes in omitted timers/policies/peer group")
	}
	configAttrs := frrProtocolAttrs(t, schema, config)
	configAttrs["peer_group"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	unknownConfig, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), configAttrs))
	if err != nil {
		t.Fatal(err)
	}
	proposedAttrs := frrProtocolAttrs(t, schema, proposed)
	proposedAttrs["peer_group"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	unknownProposed, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), proposedAttrs))
	if err != nil {
		t.Fatal(err)
	}
	unknownPlan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: prior, ProposedNewState: &unknownProposed, Config: &unknownConfig})
	if err != nil || frrProtocolErrors(unknownPlan.Diagnostics) {
		t.Fatal(err, unknownPlan.Diagnostics)
	}
	if frrProtocolAttrs(t, schema, unknownPlan.PlannedState)["peer_group"].IsKnown() {
		t.Fatal("explicitly configured unknown relation was replaced by old state")
	}
}

func TestProtocolFRRSingletonAmbiguousUpdateRestartRecovery(t *testing.T) {
	for _, kind := range []string{"general", "bgp"} {
		for _, mode := range []string{"ambiguous", "reload", "read", "ambiguous_read_failure", "validation"} {
			for _, recovery := range []string{"refresh", "unchanged_update"} {
				t.Run(kind+"_"+mode+"_"+recovery, func(t *testing.T) {
					writes, reloads := 0, 0
					fail := true
					values := map[string]any{"enabled": true, "profile": "traditional", "enable_carp": false, "enable_syslog": true, "enable_snmp": false, "syslog_level": "notifications", "firewall_rules": false}
					get := `{"general":{"enabled":"1","profile":{"traditional":{"selected":1},"datacenter":{"selected":0}},"enablecarp":"0","enablesyslog":"1","enablesnmp":"0","sysloglevel":{"notifications":{"selected":1},"errors":{"selected":0}},"fwrules":"0"}}`
					name := "opnsense_quagga_general"
					if kind == "bgp" {
						name = "opnsense_quagga_bgp_global"
						values = map[string]any{"enabled": true, "as_number": 65111, "router_id": "198.18.31.1", "distance": 0, "graceful_restart": false, "network_import_check": true, "enforce_first_as": true, "log_neighbor_changes": false, "networks": []string{}, "maximum_paths": 0, "maximum_paths_ibgp": 0}
						get = `{"bgp":{"enabled":"1","asnumber":"65111","routerid":"198.18.31.1","distance":"","graceful":"0","networkimportcheck":"1","enforce_first_as":"1","logneighborchanges":"0","networks":[],"maximumpaths":"","maximumpathsibgp":""}}`
					}
					api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if frrFixture(w, r) {
							return
						}
						switch r.URL.Path {
						case "/api/quagga/" + kind + "/set":
							writes++
							switch {
							case mode == "validation":
								_, _ = fmt.Fprint(w, `{"result":"failed","validations":{"enabled":"rejected"}}`)
							case strings.HasPrefix(mode, "ambiguous"):
								_, _ = fmt.Fprint(w, `{"result":`)
							default:
								_, _ = fmt.Fprint(w, `{"result":"saved"}`)
							}
						case "/api/quagga/" + kind + "/get":
							if (mode == "read" || mode == "ambiguous_read_failure") && fail {
								_, _ = fmt.Fprint(w, `{}`)
							} else {
								_, _ = fmt.Fprint(w, get)
							}
						case "/api/quagga/service/reconfigure":
							reloads++
							if fail && mode == "reload" {
								_, _ = fmt.Fprint(w, `{"status":"failed"}`)
							} else {
								_, _ = fmt.Fprint(w, `{"status":"ok"}`)
							}
						default:
							t.Errorf("unexpected singleton request %s", r.URL.Path)
						}
					}))
					defer api.Close()
					server, schemas := frrProtocolServer(t, api.URL)
					schema := schemas.ResourceSchemas[name]
					values["id"] = kind
					prior := frrProtocolValue(t, schema, values)
					values["enabled"] = false
					get = strings.Replace(get, `"enabled":"1"`, `"enabled":"0"`, 1)
					config := frrProtocolValue(t, schema, values)
					plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: prior, ProposedNewState: config, Config: config})
					if err != nil || frrProtocolErrors(plan.Diagnostics) {
						t.Fatal(err, plan.Diagnostics)
					}
					applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: plan.PlannedState, Config: config, PlannedPrivate: plan.PlannedPrivate})
					if mode == "ambiguous_read_failure" || mode == "validation" {
						if err != nil || !frrProtocolErrors(applied.Diagnostics) {
							t.Fatal("uncertain/rejected save accepted", err, applied.Diagnostics)
						}
						before, _ := prior.Unmarshal(schema.ValueType())
						after, _ := applied.NewState.Unmarshal(schema.ValueType())
						if !before.Equal(after) {
							t.Fatal("unconfirmed/rejected save fabricated committed desired state")
						}
						fail = false
						fresh, _ := frrProtocolServer(t, api.URL)
						refreshed, refreshErr := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: applied.NewState, Private: applied.Private})
						if refreshErr != nil || frrProtocolErrors(refreshed.Diagnostics) {
							t.Fatal(refreshErr, refreshed.Diagnostics)
						}
						if mode == "ambiguous_read_failure" {
							var enabled bool
							if e := frrProtocolAttrs(t, schema, refreshed.NewState)["enabled"].As(&enabled); e != nil || enabled {
								t.Fatal("refresh did not recover actual saved new settings", e)
							}
						}
						want := 1
						if mode == "validation" {
							want = 0
						}
						if writes != 1 || reloads != want {
							t.Fatal("uncertain/rejected recovery activation", writes, reloads)
						}
						return
					}
					if err != nil || !frrProtocolErrors(applied.Diagnostics) || len(applied.Private) == 0 {
						t.Fatal("saved singleton settings lost recovery state", err, applied.Diagnostics)
					}
					var enabled bool
					if e := frrProtocolAttrs(t, schema, applied.NewState)["enabled"].As(&enabled); e != nil || enabled {
						t.Fatal("saved new settings not retained", e)
					}
					var id string
					if e := frrProtocolAttrs(t, schema, applied.NewState)["id"].As(&id); e != nil || id != kind {
						t.Fatal("singleton identity lost", e, id)
					}
					fail = false
					fresh, _ := frrProtocolServer(t, api.URL)
					if recovery == "unchanged_update" {
						updated, err := fresh.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: applied.NewState, PlannedState: applied.NewState, Config: config, PlannedPrivate: applied.Private})
						if err != nil || frrProtocolErrors(updated.Diagnostics) {
							t.Fatal(err, updated.Diagnostics)
						}
					} else {
						refreshed, err := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: applied.NewState, Private: applied.Private})
						if err != nil || frrProtocolErrors(refreshed.Diagnostics) {
							t.Fatal(err, refreshed.Diagnostics)
						}
					}
					wantReloads := 2
					if mode == "ambiguous" {
						wantReloads = 1
					}
					if writes != 1 || reloads != wantReloads {
						t.Fatal("singleton recovery repeated a saved mutation", writes, reloads)
					}
				})
			}
		}
	}
}

// Reproduce #67 at the provider refresh boundary, not just the version regexp.
func TestProtocolFRR20PackageRevisionRefresh(t *testing.T) {
	for _, tc := range []struct {
		version               string
		compatibleModel, good bool
	}{
		{"2.0_2", true, true},
		{"2.0_999", true, true},
		{"2.0_2", false, false},
		{"2.1", true, false},
		{"2.0_2-devel", true, false},
	} {
		t.Run(fmt.Sprintf("%s/model=%t", tc.version, tc.compatibleModel), func(t *testing.T) {
			reads := 0
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("refresh mutated appliance: %s %s", r.Method, r.URL.Path)
				}
				if r.URL.Path == "/api/core/firmware/info" {
					_, _ = fmt.Fprintf(w, `{"product_version":"26.7.7","package":[{"name":"os-frr","version":%q,"installed":"1"}]}`, tc.version)
					return
				}
				if r.URL.Path == "/api/quagga/bgp/get_neighbor" && !tc.compatibleModel {
					_, _ = fmt.Fprint(w, `{"neighbor":{"multiprotocol":"0"}}`)
					return
				}
				if frrFixture(w, r) {
					return
				}
				if r.URL.Path == "/api/quagga/bgp/get_neighbor/11111111-1111-4111-8111-111111111111" {
					reads++
					_, _ = fmt.Fprint(w, `{"neighbor":{"remote_as_mode":"","localas":"","weight":"","localip":"","linklocalinterface":"","nexthopselfall":"0","multihop":"0","rrclient":"0","soft_reconfiguration_inbound":"0","bfd":"0","bfd_strict":"0","connecttimer":"","defaultoriginate":"0","asoverride":"0","removeprivateas":"","allowas_in":"","disable_connected_check":"0","attributeunchanged":"","capabilities":[],"enabled":"1","description":"","address":"192.0.2.1","remoteas":"64512","updatesource":"","nexthopself":"0","family":"ipv4","peergroup":"","keepalive":"","holddown":"","linkedPrefixlistIn":[],"linkedPrefixlistOut":[],"linkedRoutemapIn":[],"linkedRoutemapOut":[]}}`)
					return
				}
				t.Errorf("unexpected refresh request %s", r.URL.Path)
				http.NotFound(w, r)
			}))
			defer api.Close()
			server, schemas := frrProtocolServer(t, api.URL)
			name := "opnsense_quagga_bgp_neighbor"
			schema := schemas.ResourceSchemas[name]
			prior := frrProtocolValue(t, schema, frrNeighborProtocolValues())
			response, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: prior})
			if err != nil || frrProtocolErrors(response.Diagnostics) == tc.good {
				t.Fatal(err, response.Diagnostics)
			}
			if !tc.good {
				if reads != 0 {
					t.Fatal("incompatible package/model reached neighbor read")
				}
				return
			}
			if reads != 1 {
				t.Fatal("refresh did not read existing neighbor")
			}
			attrs := frrProtocolAttrs(t, schema, response.NewState)
			if !attrs["id"].Equal(tftypes.NewValue(tftypes.String, "11111111-1111-4111-8111-111111111111")) {
				t.Fatal("refresh changed UUID")
			}
			var family []tftypes.Value
			if err := attrs["family"].As(&family); err != nil || len(family) != 1 || !family[0].Equal(tftypes.NewValue(tftypes.String, "ipv4")) {
				t.Fatal("modern family lost", err)
			}
			config := frrProtocolValue(t, schema, map[string]any{"address": "192.0.2.1", "remote_as": 64512, "family": []string{"ipv4"}})
			plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: response.NewState, ProposedNewState: response.NewState, Config: config})
			if err != nil || frrProtocolErrors(plan.Diagnostics) {
				t.Fatal(err, plan.Diagnostics)
			}
			before, err := response.NewState.Unmarshal(schema.ValueType())
			if err != nil {
				t.Fatal(err)
			}
			after, err := plan.PlannedState.Unmarshal(schema.ValueType())
			if err != nil || !before.Equal(after) {
				t.Fatal("revision refresh produced nonempty plan", err)
			}
		})
	}
}
