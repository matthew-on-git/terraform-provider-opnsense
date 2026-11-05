// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

const (
	ipsecID      = "00000000-0000-4000-8000-000000000031"
	otherIPsecID = "00000000-0000-4000-8000-000000000032"
	authIPsecID  = "00000000-0000-4000-8000-000000000033"
)

// Source-shaped oracle: core 26.7.8 Swanctl.xml/ConnectionsController.php.
// Kept independent of provider models and schema metadata on purpose.
func ipsecFixture(owner string) map[string]string {
	if owner == "connection" {
		return map[string]string{
			"enabled": "0", "description": "disposable fixture", "remote_addrs": "203.0.113.2,203.0.113.1", "version": "2", "proposals": "aes256-sha256-modp2048,aes128-sha256-modp2048", "unique": "no",
			"aggressive": "0", "dpd_delay": "0", "dpd_timeout": "120", "over_time": "30", "pools": "00000000-0000-4000-8000-000000000002,00000000-0000-4000-8000-000000000001", "keyingtries": "0", "local_addrs": "192.0.2.2,192.0.2.1", "local_port": "4500", "mobike": "1", "reauth_time": "", "rekey_time": "3600", "remote_port": "", "send_cert": "ifasked", "send_certreq": "1", "encap": "0",
		}
	}
	return map[string]string{"enabled": "0", "connection": ipsecID, "description": "disposable child", "mode": "tunnel", "local_ts": "192.0.2.0/25,192.0.2.128/25", "remote_ts": "198.51.100.0/25,198.51.100.128/25", "esp_proposals": "aes256-sha256,aes128-sha256", "start_action": "trap|start", "close_action": "none", "dpd_action": "clear", "policies": "1", "reqid": "123", "sha256_96": "0", "rekey_time": "3600", "replay_window": "16"}
}

func ipsecWire(owner string, fields map[string]string) map[string]any {
	result := map[string]any{}
	for key, value := range fields {
		result[key] = value
	}
	selections := []string{"remote_addrs", "proposals", "unique", "pools", "local_addrs", "local_port", "remote_port", "send_cert"}
	if owner == "child" {
		selections = []string{"connection", "mode", "local_ts", "remote_ts", "esp_proposals", "start_action", "close_action", "dpd_action"}
	}
	for _, key := range selections {
		options := map[string]any{}
		if fields[key] != "" {
			for _, value := range strings.Split(fields[key], ",") {
				options[strings.TrimSpace(value)] = map[string]any{"value": "display only", "selected": 1}
			}
		}
		result[key] = options
	}
	if owner == "connection" {
		result["version"] = []map[string]int{{"selected": 0}, {"selected": 0}, {"selected": 1}}
	}
	return result
}

type ipsecAPI struct {
	activationReply                          *string
	deleteReply                              *string
	otherFields, authFields                  map[string]string
	override                                 map[string]any
	omit                                     string
	mu                                       sync.Mutex
	owner                                    string
	fields, last                             map[string]string
	adds, saves, deletes, activations, reads int
	failure                                  string
	malformed                                bool
}

func (a *ipsecAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if r.URL.Path == "/api/core/firmware/status" {
		_, _ = fmt.Fprint(w, `{"product":{"product_name":"OPNsense","product_version":"26.7.8"}}`)
		return
	}
	prefix := "/api/ipsec/connections/"
	switch r.URL.Path {
	case prefix + "get_" + a.owner + "/" + otherIPsecID:
		_ = json.NewEncoder(w).Encode(map[string]any{a.owner: ipsecWire(a.owner, a.otherFields)})
	case prefix + "set_" + a.owner + "/" + otherIPsecID, prefix + "set_local/" + authIPsecID, prefix + "set_remote/" + authIPsecID:
		var body map[string]map[string]string
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			http.Error(w, "invalid envelope", 400)
			return
		}
		target := a.otherFields
		if strings.Contains(r.URL.Path, "set_local/") || strings.Contains(r.URL.Path, "set_remote/") {
			target = a.authFields
		}
		for _, fields := range body {
			for key, value := range fields {
				target[key] = value
			}
		}
		a.saves++
		_, _ = fmt.Fprint(w, `{"result":"saved"}`)
	case prefix + "get_" + a.owner + "/" + ipsecID:
		a.reads++
		if a.fields == nil {
			_, _ = fmt.Fprint(w, `[]`)
			return
		}
		fields := ipsecWire(a.owner, a.fields)
		for key, value := range a.override {
			fields[key] = value
		}
		if a.omit != "" {
			delete(fields, a.omit)
		}
		if a.malformed {
			fields["rekey_time"] = map[string]any{"unexpected": "shape"}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{a.owner: fields})
	case prefix + "add_" + a.owner, prefix + "set_" + a.owner + "/" + ipsecID:
		if r.Method != http.MethodPost {
			http.Error(w, "unexpected method", 400)
			return
		}
		var body map[string]map[string]string
		if json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 1 || body[a.owner] == nil {
			http.Error(w, "invalid envelope", 400)
			return
		}
		a.last = body[a.owner]
		if a.failure == "validation" {
			_, _ = fmt.Fprint(w, `{"result":"failed","validations":{"rekey_time":"credential_payload_should_not_appear"}}`)
			return
		}
		if a.fields == nil {
			a.fields = ipsecDefaults(a.owner)
		}
		for key, value := range a.last {
			a.fields[key] = value
		}
		if a.failure == "readback" {
			a.malformed = true
		}
		if strings.Contains(r.URL.Path, "/add_") {
			a.adds++
			_, _ = fmt.Fprint(w, `{"result":"saved","uuid":"`+ipsecID+`"}`)
		} else {
			a.saves++
			_, _ = fmt.Fprint(w, `{"result":"saved"}`)
		}
	case prefix + "del_" + a.owner + "/" + ipsecID:
		a.deletes++
		if a.deleteReply != nil {
			_, _ = fmt.Fprint(w, *a.deleteReply)
			return
		}
		a.fields = nil
		_, _ = fmt.Fprint(w, `{"result":"deleted"}`)
	case "/api/ipsec/service/reconfigure":
		a.activations++
		if a.activationReply != nil {
			_, _ = fmt.Fprint(w, *a.activationReply)
			return
		}
		if a.failure == "activation" {
			_, _ = fmt.Fprint(w, `{"status":"failed"}`)
		} else {
			_, _ = fmt.Fprint(w, `{"status":"ok"}`)
		}
	default:
		if strings.HasPrefix(r.URL.Path, prefix) {
			http.Error(w, "unexpected object/endpoint", 400)
			return
		}
		_, _ = fmt.Fprint(w, `{}`)
	}
}

func ipsecServer(t *testing.T, api *ipsecAPI) (tfprotov6.ProviderServer, *tfprotov6.Schema, *tfprotov6.Schema) {
	t.Helper()
	api.otherFields = ipsecFixture(api.owner)
	api.authFields = map[string]string{"enabled": "0", "auth": "pubkey", "id": "peer.example.test", "connection": otherIPsecID}
	httpServer := httptest.NewServer(api)
	t.Cleanup(httpServer.Close)
	server := providerserver.NewProtocol6(provider.New("test")())()
	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{TerraformVersion: "1.14.0", Config: ipsecValue(t, schemas.Provider, map[string]any{"uri": httpServer.URL, "api_key": "fixture-key", "api_secret": "fixture-secret", "insecure": true})})
	if err != nil || ipsecErrors(result.Diagnostics) {
		t.Fatalf("configure: %v %v", err, ipsecDiagnosticText(result.Diagnostics))
	}
	name := "opnsense_ipsec_" + api.owner
	return server, schemas.ResourceSchemas[name], schemas.DataSourceSchemas[name]
}

func ipsecErrors(diagnostics []*tfprotov6.Diagnostic) bool {
	for _, d := range diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func ipsecValue(t *testing.T, schema *tfprotov6.Schema, fields map[string]any) *tfprotov6.DynamicValue {
	t.Helper()
	values := map[string]any{}
	for _, attribute := range schema.Block.Attributes {
		values[attribute.Name] = nil
	}
	for key, value := range fields {
		values[key] = value
	}
	raw, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	return &tfprotov6.DynamicValue{JSON: raw}
}

func ipsecNull(t *testing.T, schema *tfprotov6.Schema) *tfprotov6.DynamicValue {
	t.Helper()
	value, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), nil))
	if err != nil {
		t.Fatal(err)
	}
	return &value
}

func ipsecAttrs(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue) map[string]tftypes.Value {
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

func ipsecString(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue, name string) string {
	t.Helper()
	var value string
	if err := ipsecAttrs(t, schema, state)[name].As(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func ipsecPlan(t *testing.T, server tfprotov6.ProviderServer, owner string, prior, proposed, config *tfprotov6.DynamicValue) *tfprotov6.PlanResourceChangeResponse {
	t.Helper()
	result, err := server.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: "opnsense_ipsec_" + owner, PriorState: prior, ProposedNewState: proposed, Config: config})
	if err != nil || ipsecErrors(result.Diagnostics) {
		t.Fatalf("plan: %v %v", err, ipsecDiagnosticText(result.Diagnostics))
	}
	return result
}

func ipsecApply(t *testing.T, server tfprotov6.ProviderServer, owner string, prior, config *tfprotov6.DynamicValue, plan *tfprotov6.PlanResourceChangeResponse) *tfprotov6.ApplyResourceChangeResponse {
	t.Helper()
	result, err := server.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "opnsense_ipsec_" + owner, PriorState: prior, PlannedState: plan.PlannedState, Config: config, PlannedPrivate: plan.PlannedPrivate})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func ipsecConfig(owner string) map[string]any {
	// Deliberately reverse server sort order: string spelling/order is public API.
	if owner == "connection" {
		return map[string]any{"enabled": false, "description": "disposable fixture", "remote_addrs": "203.0.113.2,203.0.113.1", "version": "2", "proposals": "aes256-sha256-modp2048,aes128-sha256-modp2048", "unique": "no", "aggressive": false, "dpd_delay": "0", "dpd_timeout": "120", "ike_lifetime": "30", "ip_pools": "00000000-0000-4000-8000-000000000002, 00000000-0000-4000-8000-000000000001", "keying_tries": "0", "local_addresses": "192.0.2.2, 192.0.2.1", "local_port": "4500", "mobike": true, "reauthentication_time": "", "rekey_time": "3600", "remote_port": "", "send_certificate": "ifasked", "send_certificate_request": true, "udp_encapsulation": false}
	}
	return map[string]any{"enabled": false, "connection_id": ipsecID, "description": "disposable child", "mode": "tunnel", "local_ts": "192.0.2.0/25,192.0.2.128/25", "remote_ts": "198.51.100.0/25,198.51.100.128/25", "esp_proposals": "aes256-sha256,aes128-sha256", "start_action": "trap|start", "close_action": "none", "dpd_action": "clear", "install_policies": true, "request_id": "123", "sha256_96": false, "rekey_time": "3600"}
}

func ipsecMerge(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue, config map[string]any) *tfprotov6.DynamicValue {
	t.Helper()
	attributes := ipsecAttrs(t, schema, state)
	object := schema.ValueType().(tftypes.Object)
	for key, value := range config {
		attributes[key] = tftypes.NewValue(object.AttributeTypes[key], value)
	}
	merged, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attributes))
	if err != nil {
		t.Fatal(err)
	}
	return &merged
}

func TestIPsecAdvancedProtocolLifecycle(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		t.Run(owner, func(t *testing.T) {
			api := &ipsecAPI{owner: owner}
			server, schema, lookupSchema := ipsecServer(t, api)
			empty := ipsecNull(t, schema)
			otherBefore, authBefore := maps.Clone(api.otherFields), maps.Clone(api.authFields)
			configured := ipsecConfig(owner)
			config := ipsecValue(t, schema, configured)
			planned := ipsecPlan(t, server, owner, empty, config, config)
			created := ipsecApply(t, server, owner, empty, config, planned)
			if ipsecErrors(created.Diagnostics) {
				t.Fatalf("create: %v", created.Diagnostics)
			}
			if api.adds != 1 || api.activations != 1 || ipsecString(t, schema, created.NewState, "id") != ipsecID {
				t.Fatal("create identity/count mismatch")
			}
			assertIPsecValues(t, schema, created.NewState, configured, false)
			assertIPsecAdvancedWire(t, owner, api.last, expectedIPsecAdvancedWire(owner))
			if api.last["rekey_time"] != "3600" {
				t.Fatalf("advanced create omitted: %v", api.last)
			}
			if _, found := api.last["replay_window"]; found {
				t.Fatal("unowned replay window written")
			}
			read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: created.NewState})
			if err != nil || ipsecErrors(read.Diagnostics) {
				t.Fatalf("read: %v %v", err, read.Diagnostics)
			}
			noChange := ipsecPlan(t, server, owner, read.NewState, read.NewState, config)
			previous, _ := read.NewState.Unmarshal(schema.ValueType())
			next, _ := noChange.PlannedState.Unmarshal(schema.ValueType())
			if !previous.Equal(next) {
				t.Fatalf("perpetual no-op diff: %v\n%v", previous, next)
			}
			lookup, err := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "opnsense_ipsec_" + owner, Config: ipsecValue(t, lookupSchema, map[string]any{"id": ipsecID})})
			if err != nil || ipsecErrors(lookup.Diagnostics) {
				t.Fatalf("lookup: %v %v", err, lookup.Diagnostics)
			}
			assertIPsecValues(t, lookupSchema, lookup.State, configured, true)
			if ipsecString(t, lookupSchema, lookup.State, "rekey_time") != "3600" || api.adds != 1 || api.activations != 1 {
				t.Fatal("lookup lost advanced policy or mutated")
			}
			imported, err := server.ImportResourceState(context.Background(), &tfprotov6.ImportResourceStateRequest{TypeName: "opnsense_ipsec_" + owner, ID: ipsecID})
			if err != nil || ipsecErrors(imported.Diagnostics) || len(imported.ImportedResources) != 1 {
				t.Fatalf("import: %v %v", err, imported.Diagnostics)
			}
			refreshed, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: imported.ImportedResources[0].State})
			if err != nil || ipsecErrors(refreshed.Diagnostics) {
				t.Fatalf("refresh import: %v %v", err, refreshed.Diagnostics)
			}
			// Omitting all advanced attributes on imported state must send none of them.
			omitted := map[string]any{}
			for key, value := range configured {
				if isLegacyIPsec(owner, key) {
					omitted[key] = value
				}
			}
			omitted["description"] = "updated fixture"
			omittedConfig := ipsecValue(t, schema, omitted)
			proposed := ipsecMerge(t, schema, refreshed.NewState, omitted)
			update := ipsecApply(t, server, owner, refreshed.NewState, omittedConfig, ipsecPlan(t, server, owner, refreshed.NewState, proposed, omittedConfig))
			if ipsecErrors(update.Diagnostics) {
				t.Fatalf("omitted update: %v", update.Diagnostics)
			}
			for key := range api.last {
				if !isLegacyWire(owner, key) {
					t.Fatalf("omitted advanced key written: %s", key)
				}
			}
			if api.fields["rekey_time"] != "3600" || (owner == "child" && api.fields["replay_window"] != "16") {
				t.Fatal("omission reset imported policy")
			}
			retained := ipsecAttrs(t, schema, update.NewState)
			importedAttrs := ipsecAttrs(t, schema, refreshed.NewState)
			for key, value := range importedAttrs {
				if !isLegacyIPsec(owner, key) && key != "id" && !value.Equal(retained[key]) {
					t.Fatalf("imported advanced value reset: %s", key)
				}
			}
			if owner == "connection" {
				poolConfig := maps.Clone(omitted)
				poolConfig["ip_pools"] = "radius"
				poolValue := ipsecValue(t, schema, poolConfig)
				update = ipsecApply(t, server, owner, update.NewState, poolValue, ipsecPlan(t, server, owner, update.NewState, ipsecMerge(t, schema, update.NewState, poolConfig), poolValue))
				if ipsecErrors(update.Diagnostics) || api.fields["pools"] != "radius" {
					t.Fatal("populated pool membership update was dropped")
				}
				assertIPsecAdvancedWire(t, owner, api.last, map[string]string{"pools": "radius"})
			}
			// Explicit clears/false/zero must survive; child rekey_time cannot be blank.
			explicit := map[string]any{"description": "updated fixture", "rekey_time": "0"}
			if owner == "connection" {
				explicit["mobike"] = false
				explicit["local_addresses"] = ""
				explicit["dpd_timeout"] = ""
				explicit["send_certificate"] = ""
			} else {
				explicit["install_policies"] = false
				explicit["request_id"] = ""
			}
			for key, value := range omitted {
				if _, exists := explicit[key]; !exists {
					explicit[key] = value
				}
			}
			explicitConfig := ipsecValue(t, schema, explicit)
			proposed = ipsecMerge(t, schema, update.NewState, explicit)
			cleared := ipsecApply(t, server, owner, update.NewState, explicitConfig, ipsecPlan(t, server, owner, update.NewState, proposed, explicitConfig))
			if ipsecErrors(cleared.Diagnostics) {
				t.Fatalf("clear: %v", cleared.Diagnostics)
			}
			expectedClear := map[string]string{"mobike": "0", "local_addrs": "", "dpd_timeout": "", "send_cert": "", "rekey_time": "0"}
			if owner == "child" {
				expectedClear = map[string]string{"policies": "0", "reqid": "", "rekey_time": "0"}
			}
			assertIPsecAdvancedWire(t, owner, api.last, expectedClear)
			flag := "mobike"
			if owner == "child" {
				flag = "policies"
			}
			if api.last[flag] != "0" || api.last["rekey_time"] != "0" {
				t.Fatalf("explicit false/zero not written: %v", api.last)
			}
			api.fields["rekey_time"] = "42"
			drift, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: cleared.NewState})
			if err != nil || ipsecErrors(drift.Diagnostics) {
				t.Fatalf("drift: %v %v", err, drift.Diagnostics)
			}
			if ipsecString(t, schema, drift.NewState, "rekey_time") != "42" {
				t.Fatal("drift not authoritative")
			}
			repaired := ipsecApply(t, server, owner, drift.NewState, explicitConfig, ipsecPlan(t, server, owner, drift.NewState, ipsecMerge(t, schema, drift.NewState, explicit), explicitConfig))
			if ipsecErrors(repaired.Diagnostics) || api.fields["rekey_time"] != "0" {
				t.Fatalf("repair: %v", repaired.Diagnostics)
			}
			destroyed := ipsecApply(t, server, owner, repaired.NewState, empty, &tfprotov6.PlanResourceChangeResponse{PlannedState: empty})
			if ipsecErrors(destroyed.Diagnostics) || api.deletes != 1 {
				t.Fatalf("delete: %v", destroyed.Diagnostics)
			}
			if !maps.Equal(otherBefore, api.otherFields) || !maps.Equal(authBefore, api.authFields) {
				t.Fatal("unrelated second object/auth policy was changed")
			}
			expectedActivations := 5
			if owner == "connection" {
				expectedActivations++
			}
			if api.activations != expectedActivations {
				t.Fatalf("one activation per successful mutation expected, got %d", api.activations)
			}
		})
	}
}

func isLegacyIPsec(owner, key string) bool {
	if key == "connection_id" {
		return owner == "child"
	}
	return isLegacyWire(owner, key)
}

func isLegacyWire(owner, key string) bool {
	keys := []string{"enabled", "description", "remote_addrs", "version", "proposals", "unique"}
	if owner == "child" {
		keys = []string{"enabled", "description", "connection", "mode", "local_ts", "remote_ts", "esp_proposals", "start_action"}
	}
	for _, candidate := range keys {
		if key == candidate {
			return true
		}
	}
	return false
}

func TestIPsecAdvancedProtocolFailureRecovery(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		for _, failure := range []string{"validation", "activation", "readback"} {
			t.Run(owner+"/"+failure, func(t *testing.T) {
				api := &ipsecAPI{owner: owner, failure: failure}
				server, schema, _ := ipsecServer(t, api)
				empty := ipsecNull(t, schema)
				config := ipsecValue(t, schema, ipsecConfig(owner))
				result := ipsecApply(t, server, owner, empty, config, ipsecPlan(t, server, owner, empty, config, config))
				if len(result.Diagnostics) == 0 || ipsecErrors(result.Diagnostics) != (failure == "validation") {
					t.Fatal("rejected add requires error; confirmed UUID recovery requires warning")
				}
				if strings.Contains(ipsecDiagnosticText(result.Diagnostics), "credential_payload_should_not_appear") {
					t.Fatal("server payload leaked into diagnostics")
				}
				if failure == "validation" {
					if api.adds != 0 || api.activations != 0 {
						t.Fatal("rejected mutation activated")
					}
					return
				}
				if ipsecString(t, schema, result.NewState, "id") != ipsecID || api.adds != 1 {
					t.Fatal("saved UUID lost")
				}
				api.failure = ""
				api.malformed = false
				fresh, _, _ := ipsecServer(t, api)
				read, err := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: result.NewState, Private: result.Private})
				if err != nil || ipsecErrors(read.Diagnostics) {
					t.Fatalf("recovery refresh: %v %v", err, read.Diagnostics)
				}
				ipsecPlan(t, server, owner, read.NewState, read.NewState, config)
				if api.adds != 1 || api.saves != 0 {
					t.Fatal("recovery duplicated add/save")
				}
				if failure == "activation" && api.activations != 2 {
					t.Fatal("failed create did not retry activation")
				}
				activations := api.activations
				again, againErr := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: read.NewState, Private: read.Private})
				if againErr != nil || ipsecErrors(again.Diagnostics) || api.activations != activations {
					t.Fatal("consumed create marker reactivated")
				}
			})
		}
	}
}

func TestIPsecAdvancedProtocolStrictRead(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		t.Run(owner, func(t *testing.T) {
			api := &ipsecAPI{owner: owner, fields: ipsecFixture(owner)}
			server, schema, _ := ipsecServer(t, api)
			prior := ipsecValue(t, schema, map[string]any{"id": ipsecID})
			initial, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: prior})
			if err != nil || ipsecErrors(initial.Diagnostics) {
				t.Fatalf("initial: %v %v", err, initial.Diagnostics)
			}
			api.malformed = true
			malformed, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: initial.NewState})
			if err != nil || !ipsecErrors(malformed.Diagnostics) {
				t.Fatalf("malformed response not diagnosed: %v", err)
			}
			before, _ := initial.NewState.Unmarshal(schema.ValueType())
			after, _ := malformed.NewState.Unmarshal(schema.ValueType())
			if !before.Equal(after) {
				t.Fatal("malformed GET replaced prior policy")
			}
			api.malformed = false
			api.fields = nil
			missing, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: initial.NewState})
			if err != nil || ipsecErrors(missing.Diagnostics) {
				t.Fatalf("missing: %v %v", err, missing.Diagnostics)
			}
			value, _ := missing.NewState.Unmarshal(schema.ValueType())
			if !value.IsNull() {
				t.Fatal("missing object retained")
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			canceled, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: initial.NewState})
			if err == nil && !ipsecErrors(canceled.Diagnostics) {
				t.Fatal("cancellation not diagnosed")
			}
			if api.adds+api.saves+api.activations != 0 {
				t.Fatal("read path mutated")
			}
		})
	}
}

func TestIPsecAdvancedProtocolValidation(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		t.Run(owner, func(t *testing.T) {
			api := &ipsecAPI{owner: owner}
			server, schema, _ := ipsecServer(t, api)
			invalid := map[string][]any{"rekey_time": {"-1", "500001", "1.5", " 1"}}
			if owner == "connection" {
				invalid["local_port"] = []any{"500", "0"}
				invalid["remote_port"] = []any{"500"}
				invalid["keying_tries"] = []any{"1001"}
				invalid["send_certificate"] = []any{"yes"}
				invalid["dpd_delay"] = []any{"500001"}
				invalid["dpd_timeout"] = []any{"-1"}
				invalid["ike_lifetime"] = []any{"500001"}
				invalid["reauthentication_time"] = []any{"-1"}
			} else {
				invalid["request_id"] = []any{"0", "65536"}
				invalid["rekey_time"] = append(invalid["rekey_time"], "")
				invalid["close_action"] = []any{"clear"}
				invalid["dpd_action"] = []any{"none"}
			}
			for key, values := range invalid {
				for _, value := range values {
					config := ipsecConfig(owner)
					config[key] = value
					result, err := server.ValidateResourceConfig(context.Background(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "opnsense_ipsec_" + owner, Config: ipsecValue(t, schema, config)})
					if err != nil || !ipsecErrors(result.Diagnostics) {
						t.Errorf("%s=%v accepted: %v", key, value, err)
					}
				}
			}
			if api.adds+api.saves+api.activations != 0 {
				t.Fatal("validation mutated")
			}
		})
	}
}

func TestIPsecAdvancedSchemaCompatibility(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		t.Run(owner, func(t *testing.T) {
			api := &ipsecAPI{owner: owner}
			_, schema, lookup := ipsecServer(t, api)
			known := ipsecConfig(owner)
			count := 0
			lookupTypes := map[string]tftypes.Type{}
			for _, attribute := range lookup.Block.Attributes {
				lookupTypes[attribute.Name] = attribute.Type
			}
			for _, attribute := range schema.Block.Attributes {
				if attribute.Name == "id" {
					continue
				}
				if !attribute.Type.Equal(lookupTypes[attribute.Name]) {
					t.Errorf("lookup mismatch: %s", attribute.Name)
				}
				if _, exists := known[attribute.Name]; !exists {
					t.Errorf("uncovered attribute: %s", attribute.Name)
				}
				if !isLegacyIPsec(owner, attribute.Name) {
					count++
					if !attribute.Optional || !attribute.Computed {
						t.Errorf("not Optional+Computed: %s", attribute.Name)
					}
				}
				if strings.Contains(attribute.Name, "psk") || strings.Contains(attribute.Name, "secret") || strings.Contains(attribute.Name, "password") {
					t.Errorf("secret leaked: %s", attribute.Name)
				}
			}
			expected := 15
			if owner == "child" {
				expected = 6
			}
			if count != expected {
				t.Fatalf("additions %d != %d", count, expected)
			}
		})
	}
}

func ipsecDiagnosticText(diagnostics []*tfprotov6.Diagnostic) string {
	var text []string
	for _, diagnostic := range diagnostics {
		text = append(text, diagnostic.Summary+": "+diagnostic.Detail)
	}
	return strings.Join(text, "; ")
}

func TestIPsecAdvancedProtocolUpdateFailures(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		for _, failure := range []string{"validation", "activation", "readback"} {
			t.Run(owner+"/"+failure, func(t *testing.T) {
				api := &ipsecAPI{owner: owner, fields: ipsecFixture(owner)}
				server, schema, _ := ipsecServer(t, api)
				read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: ipsecValue(t, schema, map[string]any{"id": ipsecID})})
				if err != nil || ipsecErrors(read.Diagnostics) {
					t.Fatalf("read: %v", err)
				}
				configured := ipsecConfig(owner)
				configured["rekey_time"] = "0"
				config := ipsecValue(t, schema, configured)
				api.failure = failure
				result := ipsecApply(t, server, owner, read.NewState, config, ipsecPlan(t, server, owner, read.NewState, ipsecMerge(t, schema, read.NewState, configured), config))
				if !ipsecErrors(result.Diagnostics) || ipsecString(t, schema, result.NewState, "id") != ipsecID {
					t.Fatal("update failure lost identity or diagnostic")
				}
				if failure == "validation" && api.activations != 0 {
					t.Fatal("rejected save activated")
				}
				if ipsecString(t, schema, result.NewState, "rekey_time") != "3600" {
					t.Fatal("failed readback replaced confirmed prior state")
				}
				api.failure = ""
				api.malformed = false
				recovery, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: result.NewState, Private: result.Private})
				if err != nil || ipsecErrors(recovery.Diagnostics) {
					t.Fatalf("recovery: %v %s", err, ipsecDiagnosticText(recovery.Diagnostics))
				}
				if failure == "activation" && api.activations != 2 {
					t.Fatal("pending activation not retried")
				}
				if api.adds != 0 {
					t.Fatal("update recovery recreated object")
				}
			})
		}
	}
}

func TestIPsecAdvancedProtocolUnknownBeforeMutation(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		t.Run(owner, func(t *testing.T) {
			api := &ipsecAPI{owner: owner}
			server, schema, _ := ipsecServer(t, api)
			empty := ipsecNull(t, schema)
			known := ipsecValue(t, schema, ipsecConfig(owner))
			attrs := ipsecAttrs(t, schema, known)
			attrs["rekey_time"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
			unknown, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attrs))
			if err != nil {
				t.Fatal(err)
			}
			plan := ipsecPlan(t, server, owner, empty, &unknown, &unknown)
			result := ipsecApply(t, server, owner, empty, &unknown, plan)
			if !ipsecErrors(result.Diagnostics) || api.adds+api.saves+api.activations != 0 {
				t.Fatal("unknown configured value reached mutation")
			}
		})
	}
}

func assertIPsecValues(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue, want map[string]any, membership bool) {
	t.Helper()
	attrs := ipsecAttrs(t, schema, state)
	for key, expected := range want {
		switch value := expected.(type) {
		case bool:
			var actual bool
			if err := attrs[key].As(&actual); err != nil || actual != value {
				t.Errorf("%s: boolean mismatch: %v", key, err)
			}
		case string:
			var actual string
			if err := attrs[key].As(&actual); err != nil {
				t.Errorf("%s: %v", key, err)
				continue
			}
			if membership && strings.Contains(value, ",") {
				if canonicalIPsecCSV(actual) != canonicalIPsecCSV(value) {
					t.Errorf("%s: selected membership mismatch %q != %q", key, actual, value)
				}
			} else if actual != value {
				t.Errorf("%s: %q != %q", key, actual, value)
			}
		default:
			t.Fatalf("unhandled oracle type %T", expected)
		}
	}
}

func canonicalIPsecCSV(value string) string {
	values := strings.Split(value, ",")
	for index := range values {
		values[index] = strings.TrimSpace(values[index])
	}
	sort.Strings(values)
	return strings.Join(values, ",")
}

func TestIPsecAdvancedProtocolMalformedEveryField(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		t.Run(owner, func(t *testing.T) {
			api := &ipsecAPI{owner: owner, fields: ipsecFixture(owner)}
			server, schema, lookupSchema := ipsecServer(t, api)
			initial, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: ipsecValue(t, schema, map[string]any{"id": ipsecID})})
			if err != nil || ipsecErrors(initial.Diagnostics) {
				t.Fatalf("initial: %v", err)
			}
			fields := []string{"aggressive", "dpd_delay", "dpd_timeout", "over_time", "pools", "keyingtries", "local_addrs", "local_port", "mobike", "reauth_time", "rekey_time", "remote_port", "send_cert", "send_certreq", "encap"}
			if owner == "child" {
				fields = []string{"close_action", "dpd_action", "policies", "reqid", "sha256_96", "rekey_time"}
			}
			for _, field := range fields {
				t.Run(field, func(t *testing.T) {
					api.override = map[string]any{field: 42}
					read, readErr := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: initial.NewState})
					if readErr != nil || !ipsecErrors(read.Diagnostics) {
						t.Fatal("malformed advanced shape not diagnosed")
					}
					prior, _ := initial.NewState.Unmarshal(schema.ValueType())
					after, _ := read.NewState.Unmarshal(schema.ValueType())
					if !prior.Equal(after) {
						t.Fatal("malformed read erased policy")
					}
					lookup, lookupErr := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "opnsense_ipsec_" + owner, Config: ipsecValue(t, lookupSchema, map[string]any{"id": ipsecID})})
					if lookupErr != nil || !ipsecErrors(lookup.Diagnostics) {
						t.Fatal("lookup silently erased malformed policy")
					}
					api.override = nil
					api.omit = field
					missing, missingErr := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: initial.NewState})
					if missingErr != nil || !ipsecErrors(missing.Diagnostics) {
						t.Fatal("missing policy field silently accepted")
					}
					api.omit = ""
				})
			}
			if api.adds+api.saves+api.deletes+api.activations != 0 {
				t.Fatal("malformed reads mutated")
			}
		})
	}
}

func TestIPsecAdvancedProtocolOmittedCreateReadFailure(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		t.Run(owner, func(t *testing.T) {
			api := &ipsecAPI{owner: owner, failure: "readback"}
			server, schema, _ := ipsecServer(t, api)
			empty := ipsecNull(t, schema)
			fields := map[string]any{}
			for key, value := range ipsecConfig(owner) {
				if isLegacyIPsec(owner, key) {
					fields[key] = value
				}
			}
			config := ipsecValue(t, schema, fields)
			result := ipsecApply(t, server, owner, empty, config, ipsecPlan(t, server, owner, empty, config, config))
			if ipsecErrors(result.Diagnostics) || len(result.Diagnostics) == 0 || ipsecString(t, schema, result.NewState, "id") != ipsecID {
				t.Fatal("omitted confirmed create requires untainted retained identity and warning")
			}
			for key, value := range ipsecAttrs(t, schema, result.NewState) {
				if !value.IsKnown() {
					t.Fatalf("failed create retained unknown state for %s", key)
				}
			}
			for key := range api.last {
				if !isLegacyWire(owner, key) {
					t.Fatalf("omitted advanced create key written: %s", key)
				}
			}
			api.failure = ""
			api.malformed = false
			read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: result.NewState, Private: result.Private})
			if err != nil || ipsecErrors(read.Diagnostics) {
				t.Fatalf("recover: %v %s", err, ipsecDiagnosticText(read.Diagnostics))
			}
			ipsecPlan(t, server, owner, read.NewState, read.NewState, config)
			if api.adds != 1 {
				t.Fatal("omitted create recovery duplicated object")
			}
		})
	}
}

// Independent source defaults intentionally differ from desired advanced input.
func ipsecDefaults(owner string) map[string]string {
	if owner == "connection" {
		return map[string]string{"enabled": "1", "description": "", "remote_addrs": "", "version": "2", "proposals": "default", "unique": "no", "aggressive": "0", "dpd_delay": "", "dpd_timeout": "", "over_time": "", "pools": "", "keyingtries": "", "local_addrs": "", "local_port": "", "mobike": "1", "reauth_time": "", "rekey_time": "", "remote_port": "", "send_cert": "", "send_certreq": "1", "encap": "0"}
	}
	return map[string]string{"enabled": "1", "connection": "", "description": "", "mode": "tunnel", "local_ts": "", "remote_ts": "", "esp_proposals": "default", "start_action": "start", "close_action": "none", "dpd_action": "clear", "policies": "1", "reqid": "", "sha256_96": "0", "rekey_time": "3600", "replay_window": "16"}
}

func expectedIPsecAdvancedWire(owner string) map[string]string {
	if owner == "connection" {
		return map[string]string{"aggressive": "0", "dpd_delay": "0", "dpd_timeout": "120", "over_time": "30", "pools": "00000000-0000-4000-8000-000000000002,00000000-0000-4000-8000-000000000001", "keyingtries": "0", "local_addrs": "192.0.2.2,192.0.2.1", "local_port": "4500", "mobike": "1", "reauth_time": "", "rekey_time": "3600", "remote_port": "", "send_cert": "ifasked", "send_certreq": "1", "encap": "0"}
	}
	return map[string]string{"close_action": "none", "dpd_action": "clear", "policies": "1", "reqid": "123", "sha256_96": "0", "rekey_time": "3600"}
}

func assertIPsecAdvancedWire(t *testing.T, owner string, wire, want map[string]string) {
	t.Helper()
	actual := map[string]string{}
	for key, value := range wire {
		if !isLegacyWire(owner, key) {
			actual[key] = value
		}
	}
	if !maps.Equal(actual, want) {
		t.Fatalf("advanced wire keys/values differ: got %v want %v", actual, want)
	}
}

func ipsecPending(private []byte) bool {
	var values map[string][]byte
	return json.Unmarshal(private, &values) == nil && len(values["ipsec_pending_activation_v1"]) > 0
}

func TestIPsecPendingInvalidUpdateNeverActivates(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		for _, unknown := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/unknown=%t", owner, unknown), func(t *testing.T) {
				api := &ipsecAPI{owner: owner, failure: "activation"}
				server, schema, _ := ipsecServer(t, api)
				config := ipsecValue(t, schema, ipsecConfig(owner))
				empty := ipsecNull(t, schema)
				created := ipsecApply(t, server, owner, empty, config, ipsecPlan(t, server, owner, empty, config, config))
				if ipsecErrors(created.Diagnostics) || len(created.Diagnostics) == 0 || !ipsecPending(created.Private) {
					t.Fatal("expected saved pending create")
				}
				fields := ipsecConfig(owner)
				fields["rekey_time"] = "500001"
				bad := ipsecValue(t, schema, fields)
				if unknown {
					attributes := ipsecAttrs(t, schema, bad)
					attributes["rekey_time"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
					value, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), attributes))
					if err != nil {
						t.Fatal(err)
					}
					bad = &value
				}
				counts := []int{api.adds, api.saves, api.deletes, api.activations, api.reads}
				result := ipsecApply(t, server, owner, created.NewState, bad, &tfprotov6.PlanResourceChangeResponse{PlannedState: bad, PlannedPrivate: created.Private})
				if !ipsecErrors(result.Diagnostics) || !ipsecPending(result.Private) {
					t.Fatal("invalid pending update lost marker/diagnostic")
				}
				if !slices.Equal(counts, []int{api.adds, api.saves, api.deletes, api.activations, api.reads}) {
					t.Fatal("invalid/unknown configured update reached API before rejection")
				}
			})
		}
	}
}

func TestIPsecStrictActivationRecoveryAndDelete(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		for _, deleting := range []bool{false, true} {
			for _, reply := range []string{"", `{}`, `{"status":"unexpected"}`, `{"result":"ok"}`} {
				t.Run(fmt.Sprintf("%s/delete=%t/reply=%q", owner, deleting, reply), func(t *testing.T) {
					api := &ipsecAPI{owner: owner, activationReply: &reply}
					server, schema, _ := ipsecServer(t, api)
					empty := ipsecNull(t, schema)
					config := ipsecValue(t, schema, ipsecConfig(owner))
					var failed *tfprotov6.ApplyResourceChangeResponse
					if deleting {
						api.fields = ipsecFixture(owner)
						read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: ipsecValue(t, schema, map[string]any{"id": ipsecID})})
						if err != nil || ipsecErrors(read.Diagnostics) {
							t.Fatal("initial read failed")
						}
						failed = ipsecApply(t, server, owner, read.NewState, empty, &tfprotov6.PlanResourceChangeResponse{PlannedState: empty})
					} else {
						failed = ipsecApply(t, server, owner, empty, config, ipsecPlan(t, server, owner, empty, config, config))
					}
					if ipsecErrors(failed.Diagnostics) != deleting || len(failed.Diagnostics) == 0 || !ipsecPending(failed.Private) || api.activations != 1 {
						t.Fatal("non-affirmative activation accepted or marker lost")
					}
					fresh, _, _ := ipsecServer(t, api)
					state, private := failed.NewState, failed.Private
					for attempt := 0; attempt < 2; attempt++ {
						retried, err := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: state, Private: private})
						if err != nil || !ipsecErrors(retried.Diagnostics) || !ipsecPending(retried.Private) {
							t.Fatal("failed retry must retain pending marker")
						}
						state, private = retried.NewState, retried.Private
					}
					affirmative := `{"status":"ok"}`
					api.activationReply = &affirmative
					recovered, err := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: state, Private: private})
					if err != nil || ipsecErrors(recovered.Diagnostics) || ipsecPending(recovered.Private) || api.activations != 4 {
						t.Fatal("affirmative retry did not consume marker")
					}
					value, _ := recovered.NewState.Unmarshal(schema.ValueType())
					if deleting {
						if !value.IsNull() || api.deletes != 1 || api.adds+api.saves != 0 {
							t.Fatal("confirmed delete was repeated or missing-object refresh failed")
						}
					} else {
						if value.IsNull() || api.adds != 1 || api.saves+api.deletes != 0 {
							t.Fatal("saved create was repeated")
						}
					}
					nextState := recovered.NewState
					if deleting {
						nextState = failed.NewState
					}
					again, againErr := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: nextState, Private: recovered.Private})
					if againErr != nil || ipsecErrors(again.Diagnostics) || api.activations != 4 {
						t.Fatal("consumed marker reactivated")
					}
				})
			}
		}
	}
}

func TestIPsecSelectionValidationBeforeMutation(t *testing.T) {
	for field, values := range map[string][]string{"local_addresses": {"192.0.2.1, 192.0.2.1", ",192.0.2.1", "192.0.2.1,", "192.0.2.1, ,192.0.2.2"}, "ip_pools": {"radius, radius", "not-a-pool", "radius,", "radius,,00000000-0000-4000-8000-000000000001"}} {
		for _, value := range values {
			t.Run(field+"/"+value, func(t *testing.T) {
				api := &ipsecAPI{owner: "connection"}
				server, schema, _ := ipsecServer(t, api)
				fields := ipsecConfig("connection")
				fields[field] = value
				config := ipsecValue(t, schema, fields)
				result := ipsecApply(t, server, "connection", ipsecNull(t, schema), config, &tfprotov6.PlanResourceChangeResponse{PlannedState: config})
				if !ipsecErrors(result.Diagnostics) || api.adds+api.saves+api.activations != 0 {
					t.Fatal("invalid selection reached mutation")
				}
			})
		}
	}
}

func TestIPsecSelectionWhitespaceArraysReadAndLookup(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		for _, raw := range []string{"[ ]", "[\n\t ]", `["bad"]`} {
			t.Run(owner+raw, func(t *testing.T) {
				field := "local_addrs"
				attribute := "local_addresses"
				if owner == "child" {
					field = "local_ts"
					attribute = "local_ts"
				}
				api := &ipsecAPI{owner: owner, fields: ipsecFixture(owner), override: map[string]any{field: json.RawMessage(raw)}}
				server, schema, lookupSchema := ipsecServer(t, api)
				read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_" + owner, CurrentState: ipsecValue(t, schema, map[string]any{"id": ipsecID})})
				lookup, lookupErr := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "opnsense_ipsec_" + owner, Config: ipsecValue(t, lookupSchema, map[string]any{"id": ipsecID})})
				invalid := strings.Contains(raw, "bad")
				if err != nil || lookupErr != nil || ipsecErrors(read.Diagnostics) != invalid || ipsecErrors(lookup.Diagnostics) != invalid {
					t.Fatal("array shape handling mismatch")
				}
				if !invalid && (ipsecString(t, schema, read.NewState, attribute) != "" || ipsecString(t, lookupSchema, lookup.State, attribute) != "") {
					t.Fatal("empty array not decoded as empty selection")
				}
			})
		}
	}
}

func TestIPsecPopulatedPoolUpdateWire(t *testing.T) {
	api := &ipsecAPI{owner: "connection", fields: ipsecFixture("connection")}
	api.fields["pools"] = "00000000-0000-4000-8000-000000000001"
	server, schema, _ := ipsecServer(t, api)
	read, err := server.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_ipsec_connection", CurrentState: ipsecValue(t, schema, map[string]any{"id": ipsecID})})
	if err != nil || ipsecErrors(read.Diagnostics) {
		t.Fatal("import-like refresh failed")
	}
	fields := ipsecConfig("connection")
	fields["ip_pools"] = "radius, 00000000-0000-4000-8000-000000000002"
	config := ipsecValue(t, schema, fields)
	result := ipsecApply(t, server, "connection", read.NewState, config, ipsecPlan(t, server, "connection", read.NewState, ipsecMerge(t, schema, read.NewState, fields), config))
	if ipsecErrors(result.Diagnostics) {
		t.Fatal("pool update failed", ipsecDiagnosticText(result.Diagnostics))
	}
	expected := expectedIPsecAdvancedWire("connection")
	expected["pools"] = "radius,00000000-0000-4000-8000-000000000002"
	assertIPsecAdvancedWire(t, "connection", api.last, expected)
	if api.fields["pools"] != expected["pools"] || api.adds != 0 || api.saves != 1 || api.activations != 1 {
		t.Fatal("populated pool update was omitted or misdirected")
	}
}

func TestIPsecPendingDeleteSupersedesActivation(t *testing.T) {
	for _, owner := range []string{"connection", "child"} {
		for _, updating := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/update=%t", owner, updating), func(t *testing.T) {
				api := &ipsecAPI{owner: owner}
				if !updating {
					api.failure = "activation"
				}
				server, schema, _ := ipsecServer(t, api)
				empty := ipsecNull(t, schema)
				fields := ipsecConfig(owner)
				config := ipsecValue(t, schema, fields)
				pending := ipsecApply(t, server, owner, empty, config, ipsecPlan(t, server, owner, empty, config, config))
				if updating {
					if ipsecErrors(pending.Diagnostics) {
						t.Fatal("initial create failed")
					}
					api.failure = "activation"
					fields["rekey_time"] = "0"
					config = ipsecValue(t, schema, fields)
					pending = ipsecApply(t, server, owner, pending.NewState, config, ipsecPlan(t, server, owner, pending.NewState, ipsecMerge(t, schema, pending.NewState, fields), config))
				}
				if ipsecErrors(pending.Diagnostics) != updating || len(pending.Diagnostics) == 0 || !ipsecPending(pending.Private) {
					t.Fatal("pending saved mutation missing")
				}
				api.failure = ""
				beforeActivations, beforeSaves := api.activations, api.saves
				deleted := ipsecApply(t, server, owner, pending.NewState, empty, &tfprotov6.PlanResourceChangeResponse{PlannedState: empty, PlannedPrivate: pending.Private})
				if ipsecErrors(deleted.Diagnostics) || ipsecPending(deleted.Private) || api.deletes != 1 || api.activations != beforeActivations+1 || api.adds != 1 || api.saves != beforeSaves {
					t.Fatal("Delete replayed superseded activation or failed to consume marker")
				}
			})
		}
	}
}
