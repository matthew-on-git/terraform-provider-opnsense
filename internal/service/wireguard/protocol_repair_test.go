// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

func protocolServer(t *testing.T, uri string) (tfprotov6.ProviderServer, *tfprotov6.GetProviderSchemaResponse) {
	t.Helper()
	server := providerserver.NewProtocol6(provider.New("test")())()
	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	config := protocolValue(t, schemas.Provider, map[string]any{"uri": uri, "api_key": "fixture-key", "api_secret": "fixture-secret", "insecure": true})
	configured, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: config, TerraformVersion: "1.14.0"})
	if err != nil || protocolErrors(configured.Diagnostics) {
		t.Fatal("configure", err, configured.Diagnostics)
	}
	return server, schemas
}

func protocolValue(t *testing.T, schema *tfprotov6.Schema, values map[string]any) *tfprotov6.DynamicValue {
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

func protocolErrors(ds []*tfprotov6.Diagnostic) bool {
	for _, d := range ds {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func protocolAttrs(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue) map[string]tftypes.Value {
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

func TestProtocolVersionZeroWireGuardUpgrade(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "get_client"):
			_, _ = fmt.Fprint(w, `{"client":{"enabled":"0","name":"legacy","pubkey":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","psk":"","tunneladdress":{"10.0.0.1/24":{"selected":1},"fd00::1/64":{"selected":1}},"servers":{"00000000-0000-0000-0000-000000000001":{"selected":1},"00000000-0000-0000-0000-000000000002":{"selected":1}},"serveraddress":"example.test","serverport":"51820","keepalive":"25"}}`)
		case strings.Contains(r.URL.Path, "get_server"):
			_, _ = fmt.Fprint(w, `{"server":{"enabled":"0","name":"legacy","pubkey":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","privkey":"","instance":"4","port":"51821","mtu":"","dns":[],"tunneladdress":{"10.0.0.1/24":{"selected":1},"fd00::1/64":{"selected":1}},"peers":[],"disableroutes":"0","gateway":""}}`)
		case strings.Contains(r.URL.Path, "set_"):
			_, _ = fmt.Fprint(w, `{"result":"saved"}`)
		case strings.Contains(r.URL.Path, "reconfigure"):
			_, _ = fmt.Fprint(w, `{"result":"ok"}`)
		default:
			_, _ = fmt.Fprint(w, `{}`)
		}
	}))
	defer api.Close()
	server, schemas := protocolServer(t, api.URL)
	for _, kind := range []string{"peer", "server"} {
		t.Run(kind, func(t *testing.T) {
			name := "opnsense_wireguard_" + kind
			raw := `{"id":"stable-id","enabled":false,"name":"legacy","public_key":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","tunnel_address":"10.0.0.1/24,fd00::1/64","server_address":"example.test","server_port":"51820","keepalive":25,"servers":"00000000-0000-0000-0000-000000000001,00000000-0000-0000-0000-000000000002"}`
			if kind == "server" {
				raw = `{"id":"stable-id","enabled":false,"name":"legacy","private_key":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=","tunnel_address":"10.0.0.1/24,fd00::1/64","port":"51821","description":"legacy description"}`
			}
			response, err := server.UpgradeResourceState(context.Background(), &tfprotov6.UpgradeResourceStateRequest{TypeName: name, Version: 0, RawState: &tfprotov6.RawState{JSON: []byte(raw)}})
			if err != nil || protocolErrors(response.Diagnostics) {
				t.Fatal(err, response.Diagnostics)
			}
			schema := schemas.ResourceSchemas[name]
			attrs := protocolAttrs(t, schema, response.UpgradedState)
			var id string
			if err := attrs["id"].As(&id); err != nil {
				t.Fatal(err)
			}
			if id != "stable-id" {
				t.Fatal("identity changed")
			}
			var members []tftypes.Value
			if err := attrs["tunnel_address"].As(&members); err != nil {
				t.Fatal(err)
			}
			if len(members) != 2 {
				t.Fatal("lost legacy CIDRs")
			}

			assertString := func(field, want string) {
				t.Helper()
				var got string
				if err := attrs[field].As(&got); err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Fatalf("%s not preserved", field)
				}
			}
			assertNumber := func(field string, want int64) {
				t.Helper()
				var number big.Float
				if err := attrs[field].As(&number); err != nil {
					t.Fatal(err)
				}
				got, accuracy := number.Int64()
				if accuracy != big.Exact || got != want {
					t.Fatalf("%s not preserved", field)
				}
			}
			assertMembers := func(field string, want []string) {
				t.Helper()
				var values []tftypes.Value
				if err := attrs[field].As(&values); err != nil {
					t.Fatal(err)
				}
				got := []string{}
				for _, v := range values {
					var text string
					if err := v.As(&text); err != nil {
						t.Fatal(err)
					}
					got = append(got, text)
				}
				sort.Strings(got)
				sort.Strings(want)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("%s members not preserved", field)
				}
			}
			assertMembers("tunnel_address", []string{"10.0.0.1/24", "fd00::1/64"})
			stable := []string{"id", "enabled", "name", "tunnel_address"}
			configValues := map[string]any{"enabled": false, "name": "legacy", "tunnel_address": []string{"10.0.0.1/24", "fd00::1/64"}}
			if kind == "peer" {
				assertMembers("servers", []string{"00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"})
				assertNumber("server_port", 51820)
				assertNumber("keepalive", 25)
				assertString("public_key", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
				assertString("server_address", "example.test")
				stable = append(stable, "public_key", "server_address", "server_port", "keepalive")
				configValues["public_key"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
				configValues["server_address"] = "example.test"
				configValues["server_port"] = 51820
				configValues["keepalive"] = 25
			} else {
				assertNumber("port", 51821)
				assertString("private_key", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
				assertString("description", "legacy description")
				stable = append(stable, "private_key", "port", "description")
				configValues["private_key"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
				configValues["port"] = 51821
				configValues["description"] = "legacy description"
			}
			config := protocolValue(t, schema, configValues)
			plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: response.UpgradedState, ProposedNewState: response.UpgradedState, Config: config})
			if err != nil || protocolErrors(plan.Diagnostics) || len(plan.RequiresReplace) > 0 {
				t.Fatal("upgrade plan", err, plan.Diagnostics)
			}
			planned := protocolAttrs(t, schema, plan.PlannedState)
			for _, field := range stable {
				if !attrs[field].Equal(planned[field]) {
					t.Fatalf("upgrade changed %s", field)
				}
			}
			// Apply the in-place normalization of newly introduced defaults,
			// persist protocol state, refresh and demand a no-change plan. This
			// proves convergence after the real registered v0 upgrade boundary.
			applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: response.UpgradedState, PlannedState: plan.PlannedState, Config: config, PlannedPrivate: plan.PlannedPrivate})
			if err != nil || protocolErrors(applied.Diagnostics) {
				t.Fatal("upgrade apply", err, applied.Diagnostics)
			}
			refreshed, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: applied.NewState, Private: applied.Private})
			if err != nil || protocolErrors(refreshed.Diagnostics) {
				t.Fatal("upgrade refresh", err, refreshed.Diagnostics)
			}
			finalAttrs := protocolAttrs(t, schema, refreshed.NewState)
			for _, field := range stable {
				if !attrs[field].Equal(finalAttrs[field]) {
					t.Fatalf("convergence lost legacy %s", field)
				}
			}
			if kind == "peer" && !attrs["servers"].Equal(finalAttrs["servers"]) {
				t.Fatal("convergence lost association members")
			}
			converged, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: refreshed.NewState, ProposedNewState: refreshed.NewState, Config: config, PriorPrivate: refreshed.Private})
			if err != nil || protocolErrors(converged.Diagnostics) || len(converged.RequiresReplace) > 0 {
				t.Fatal("convergence plan", err, converged.Diagnostics)
			}
			finalPlan, err := converged.PlannedState.Unmarshal(schema.ValueType())
			if err != nil {
				t.Fatal(err)
			}
			finalState, err := refreshed.NewState.Unmarshal(schema.ValueType())
			if err != nil {
				t.Fatal(err)
			}
			if !finalPlan.Equal(finalState) {
				t.Fatal("legacy upgrade did not converge to no-change plan")
			}
		})
	}
}

func TestProtocolFailedSingletonSavePersistsPrivateAcrossServerRestart(t *testing.T) {
	saves := 0
	readFails := true
	activations := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/wireguard/general/set":
			saves++
			_, _ = fmt.Fprint(w, `{}`)
		case "/api/wireguard/general/get":
			if readFails {
				_, _ = fmt.Fprint(w, `{}`)
			} else {
				_, _ = fmt.Fprint(w, `{"general":{"enabled":"0"}}`)
			}
		case "/api/wireguard/service/reconfigure":
			activations++
			_, _ = fmt.Fprint(w, `{"result":"ok"}`)
		default:
			_, _ = fmt.Fprint(w, `{}`)
		}
	}))
	defer api.Close()
	server, schemas := protocolServer(t, api.URL)
	schema := schemas.ResourceSchemas["opnsense_wireguard_settings"]
	prior := protocolValue(t, schema, map[string]any{"id": "wireguard_settings", "enabled": true})
	plan := protocolValue(t, schema, map[string]any{"id": "wireguard_settings", "enabled": false})
	config := protocolValue(t, schema, map[string]any{"enabled": false})
	applied, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "opnsense_wireguard_settings", PriorState: prior, PlannedState: plan, Config: config})
	if err != nil || !protocolErrors(applied.Diagnostics) || len(applied.Private) == 0 {
		t.Fatal("failed apply did not persist private", err, applied.Diagnostics)
	}
	// Persist exactly protocol-returned bytes, instantiate a fresh provider server,
	// then replay the engine refresh boundary. No in-memory Private object passes.
	persisted, err := json.Marshal(struct {
		State   *tfprotov6.DynamicValue
		Private []byte
	}{applied.NewState, applied.Private})
	if err != nil {
		t.Fatal(err)
	}
	var stored struct {
		State   *tfprotov6.DynamicValue
		Private []byte
	}
	if err = json.Unmarshal(persisted, &stored); err != nil {
		t.Fatal(err)
	}
	readFails = false
	fresh, _ := protocolServer(t, api.URL)
	refreshed, err := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_wireguard_settings", CurrentState: stored.State, Private: stored.Private})
	if err != nil || protocolErrors(refreshed.Diagnostics) {
		t.Fatal(err, refreshed.Diagnostics)
	}
	if saves != 1 || activations != 1 {
		t.Fatalf("save=%d activation=%d", saves, activations)
	}
	attrs := protocolAttrs(t, schema, refreshed.NewState)
	var enabled bool
	if err := attrs["enabled"].As(&enabled); err != nil {
		t.Fatal(err)
	}
	if enabled {
		t.Fatal("refresh did not reconcile saved disabled state")
	}
}

func TestProtocolSavedKeyRotationPersistsAcrossRestart(t *testing.T) {
	const oldKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	const newKey = "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE="
	for _, kind := range []string{"peer", "server"} {
		t.Run(kind, func(t *testing.T) {
			saved, activated := 0, false
			api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.Contains(r.URL.Path, "set_"):
					saved++
					_, _ = fmt.Fprint(w, `{"result":"saved"}`)
				case strings.Contains(r.URL.Path, "reconfigure"):
					if activated {
						_, _ = fmt.Fprint(w, `{"result":"ok"}`)
					} else {
						_, _ = fmt.Fprint(w, `{"result":"failed"}`)
					}
				case strings.Contains(r.URL.Path, "get_client"):
					_, _ = fmt.Fprint(w, `{"client":{"enabled":"0","name":"rotation","pubkey":"`+oldKey+`","psk":"","tunneladdress":{"10.0.0.2/32":{"selected":1}},"servers":[],"serveraddress":"","serverport":"","keepalive":""}}`)
				case strings.Contains(r.URL.Path, "get_server"):
					_, _ = fmt.Fprint(w, `{"server":{"enabled":"0","name":"rotation","pubkey":"`+oldKey+`","privkey":"","instance":"3","port":"","mtu":"","dns":[],"tunneladdress":[],"peers":[],"disableroutes":"0","gateway":""}}`)
				default:
					_, _ = fmt.Fprint(w, `{}`)
				}
			}))
			defer api.Close()
			server, schemas := protocolServer(t, api.URL)
			name := "opnsense_wireguard_" + kind
			schema := schemas.ResourceSchemas[name]
			field := "psk"
			values := map[string]any{"id": "owned", "enabled": false, "name": "rotation", "public_key": oldKey, "psk": oldKey, "tunnel_address": []string{"10.0.0.2/32"}, "servers": []string{}, "server_address": ""}
			if kind == "server" {
				field = "private_key"
				values = map[string]any{"id": "owned", "enabled": false, "name": "rotation", "instance": "3", "public_key": oldKey, "private_key": oldKey, "dns": []string{}, "tunnel_address": []string{}, "peers": []string{}, "disable_routes": false, "gateway": "", "description": ""}
			}
			prior := protocolValue(t, schema, values)
			values[field] = newKey
			planned := protocolValue(t, schema, values)
			configValues := map[string]any{}
			for k, v := range values {
				configValues[k] = v
			}
			delete(configValues, "id")
			delete(configValues, "instance")
			delete(configValues, "servers")
			config := protocolValue(t, schema, configValues)
			failed, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: planned, Config: config})
			if err != nil || !protocolErrors(failed.Diagnostics) || len(failed.Private) == 0 {
				t.Fatal("expected persisted failed activation", err, failed.Diagnostics)
			}
			attrs := protocolAttrs(t, schema, failed.NewState)
			var key string
			if err = attrs[field].As(&key); err != nil {
				t.Fatal(err)
			}
			if key != newKey {
				t.Fatal("protocol response reverted confirmed saved rotation")
			}
			persisted, err := json.Marshal(struct {
				State   *tfprotov6.DynamicValue
				Private []byte
			}{failed.NewState, failed.Private})
			if err != nil {
				t.Fatal(err)
			}
			var stored struct {
				State   *tfprotov6.DynamicValue
				Private []byte
			}
			if err = json.Unmarshal(persisted, &stored); err != nil {
				t.Fatal(err)
			}
			activated = true
			fresh, _ := protocolServer(t, api.URL)
			recovered, err := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: name, CurrentState: stored.State, Private: stored.Private})
			if err != nil || protocolErrors(recovered.Diagnostics) || saved != 1 {
				t.Fatal("protocol recovery failed/replayed save", err, recovered.Diagnostics)
			}
			attrs = protocolAttrs(t, schema, recovered.NewState)
			if err = attrs[field].As(&key); err != nil {
				t.Fatal(err)
			}
			if key != newKey {
				t.Fatal("blank API read reverted persisted rotation")
			}
		})
	}
}

func TestProtocolConfiguredUnknownKeysRejectPriorReuse(t *testing.T) {
	for _, kind := range []string{"peer", "server"} {
		fields := []string{"public_key", "psk"}
		if kind == "server" {
			fields = []string{"public_key", "private_key"}
		}
		for _, field := range fields {
			t.Run(kind+"/"+field, func(t *testing.T) {
				traffic := 0
				api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/api/wireguard/") {
						traffic++
					}
					_, _ = fmt.Fprint(w, `{}`)
				}))
				defer api.Close()
				server, schemas := protocolServer(t, api.URL)
				name := "opnsense_wireguard_" + kind
				schema := schemas.ResourceSchemas[name]
				values := map[string]any{"id": "owned", "enabled": false, "name": "configured", "public_key": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", "psk": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", "tunnel_address": []string{"10.0.0.2/32"}, "servers": []string{}, "server_address": ""}
				if kind == "server" {
					values = map[string]any{"id": "owned", "enabled": false, "name": "configured", "instance": "4", "public_key": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", "private_key": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", "dns": []string{}, "tunnel_address": []string{}, "peers": []string{}, "disable_routes": false, "gateway": "", "description": ""}
				}
				prior := protocolValue(t, schema, values)
				attrs := protocolAttrs(t, schema, prior)
				attrs[field] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
				unknown, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
				if err != nil {
					t.Fatal(err)
				}
				failed, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: name, PriorState: prior, PlannedState: prior, Config: &unknown})
				if err != nil || !protocolErrors(failed.Diagnostics) || traffic != 0 {
					t.Fatal("configured unknown key escaped protocol boundary", err, failed.Diagnostics)
				}
			})
		}
	}
}

// Omitted write-only intent must not manufacture an empty-key update after
// import/refresh. Explicit blank remains an intentional key removal.
func TestProtocolPeerOmittedPSKDoesNotDefaultToClear(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("unexpected mutation: %s", r.URL.Path)
		}
		_, _ = fmt.Fprint(w, `{}`)
	}))
	defer api.Close()
	server, schemas := protocolServer(t, api.URL)
	const name = "opnsense_wireguard_peer"
	schema := schemas.ResourceSchemas[name]
	key := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	for _, tc := range []struct {
		name              string
		prior, configured any
		want              any
	}{
		{"imported-null-omitted", nil, nil, nil},
		{"known-empty-omitted", "", nil, ""},
		{"known-secret-omitted", key, nil, key},
		{"explicit-clear", key, "", ""},
		{"explicit-key", nil, key, key},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := map[string]any{"id": "owned", "enabled": false, "name": "peer", "public_key": key, "psk": tc.prior, "tunnel_address": []string{"192.0.2.2/32"}, "servers": []string{}, "server_address": ""}
			prior := protocolValue(t, schema, values)
			values["psk"] = tc.configured
			proposed := protocolValue(t, schema, values)
			delete(values, "id")
			delete(values, "servers")
			config := protocolValue(t, schema, values)
			plan, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: name, PriorState: prior, ProposedNewState: proposed, Config: config})
			if err != nil || protocolErrors(plan.Diagnostics) {
				t.Fatal(err, plan.Diagnostics)
			}
			got := protocolAttrs(t, schema, plan.PlannedState)["psk"]
			want := tftypes.NewValue(tftypes.String, tc.want)
			if !got.Equal(want) {
				t.Fatalf("PSK plan = %v, want %v", got, want)
			}
		})
	}
}
