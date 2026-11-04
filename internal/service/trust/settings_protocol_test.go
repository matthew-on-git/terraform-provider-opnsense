// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

const settingsType = "opnsense_trust_settings"

func settingsServer(t *testing.T, uri string) (tfprotov6.ProviderServer, *tfprotov6.GetProviderSchemaResponse) {
	t.Helper()
	server := providerserver.NewProtocol6(provider.New("test")())()
	schemas, err := server.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}
	config := settingsValue(t, schemas.Provider, map[string]any{"uri": uri, "api_key": "fixture-key", "api_secret": "fixture-secret", "insecure": true})
	result, err := server.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{Config: config, TerraformVersion: "1.14.0"})
	if err != nil || settingsErrors(result.Diagnostics) {
		t.Fatalf("configure: %v %+v", err, result.Diagnostics[0])
	}
	if schemas.ResourceSchemas[settingsType] == nil || schemas.DataSourceSchemas[settingsType] == nil {
		t.Fatal("missing exported registration")
	}
	return server, schemas
}

func settingsValue(t *testing.T, schema *tfprotov6.Schema, fields map[string]any) *tfprotov6.DynamicValue {
	t.Helper()
	values := map[string]any{}
	for _, attr := range schema.Block.Attributes {
		values[attr.Name] = nil
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

func settingsNull(t *testing.T, schema *tfprotov6.Schema) *tfprotov6.DynamicValue {
	t.Helper()
	value, err := tfprotov6.NewDynamicValue(schema.ValueType(), tftypes.NewValue(schema.ValueType(), nil))
	if err != nil {
		t.Fatal(err)
	}
	return &value
}

func settingsErrors(diagnostics []*tfprotov6.Diagnostic) bool {
	for _, d := range diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func settingsAttrs(t *testing.T, schema *tfprotov6.Schema, state *tfprotov6.DynamicValue) map[string]tftypes.Value {
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

type settingsAPI struct {
	mu                                   sync.Mutex
	fields                               map[string]any
	wire                                 map[string]string
	reads, saves, activations            int
	failure                              string
	badRead                              bool
	saveReply, activationReply, redirect string
	activationStatus                     int
	dropActivation                       bool
	cancel                               context.CancelFunc
}

func (a *settingsAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.fields == nil {
		a.fields = map[string]any{"store_intermediate_certs": "1", "install_crls": "1", "fetch_crls": "0", "enable_legacy_sect": "1", "enable_config_constraints": "0", "CipherString": map[string]any{"AES256-SHA": map[string]any{"value": "AES256-SHA", "selected": 1}, "AES128-SHA": map[string]any{"value": "AES128-SHA", "selected": 0}}, "Ciphersuites": "opaque", "SignatureAlgorithms": "opaque", "groups": "opaque", "MinProtocol": "TLSv1.2", "MinProtocol_DTLS": "DTLSv1.2"}
	}
	switch r.URL.Path {
	case "/api/core/menu/search":
		_, _ = fmt.Fprint(w, `{}`)
	case "/api/core/firmware/status":
		_, _ = fmt.Fprint(w, `{"product":{"product_name":"OPNsense","product_version":"26.7.8"}}`)
	case "/api/trust/settings/get":
		a.reads++
		if a.badRead {
			_, _ = fmt.Fprint(w, `{"trust":{}}`)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"trust": a.fields})
	case "/api/trust/settings/set":
		a.saves++
		var body map[string]map[string]string
		if json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 1 || body["trust"] == nil {
			http.Error(w, "bad shape", 400)
			return
		}
		a.wire = body["trust"]
		if a.failure == "validation" {
			_, _ = fmt.Fprint(w, `{"result":"failed","validations":{"trust.install_crls":"invalid"}}`)
			return
		}
		for k, v := range a.wire {
			a.fields[k] = v
		}
		if a.failure == "mismatch" || a.failure == "confirmed mismatch" {
			a.fields["install_crls"] = "1"
		}
		if a.failure == "read" {
			a.badRead = true
		}
		if a.failure == "connection lost" {
			if hijacker, ok := w.(http.Hijacker); ok {
				conn, _, err := hijacker.Hijack()
				if err == nil {
					_ = conn.Close()
					return
				}
			}
			http.Error(w, "fixture cannot hijack", 500)
			return
		}

		if a.failure == "uncertain" || a.failure == "mismatch" {
			_, _ = fmt.Fprint(w, `{`)
			return
		}
		if a.cancel != nil {
			a.cancel()
		}
		if a.saveReply != "" {
			_, _ = fmt.Fprint(w, a.saveReply)
			return
		}
		_, _ = fmt.Fprint(w, `{"result":"saved"}`)
	case "/api/trust/settings/reconfigure":
		a.activations++
		if a.redirect != "" {
			http.Redirect(w, r, a.redirect, http.StatusTemporaryRedirect)
			return
		}
		if a.activationStatus != 0 {
			http.Error(w, "activation unavailable", a.activationStatus)
			return
		}
		if a.dropActivation {
			if hijacker, ok := w.(http.Hijacker); ok {
				conn, _, err := hijacker.Hijack()
				if err == nil {
					_ = conn.Close()
					return
				}
			}
			http.Error(w, "cannot drop", 500)
			return
		}
		if a.activationReply != "" {
			_, _ = fmt.Fprint(w, a.activationReply)
			return
		}
		if a.failure == "activation" {
			_, _ = fmt.Fprint(w, `{"status":"failed"}`)
			return
		}
		_, _ = fmt.Fprint(w, `{"status":"ok"}`)
	default:
		http.NotFound(w, r)
	}
}

func settingsImport(t *testing.T, s tfprotov6.ProviderServer, _ *tfprotov6.Schema) *tfprotov6.DynamicValue {
	t.Helper()
	v, err := s.ImportResourceState(context.Background(), &tfprotov6.ImportResourceStateRequest{TypeName: settingsType, ID: "trust_settings"})
	if err != nil || settingsErrors(v.Diagnostics) {
		t.Fatal(err, v)
	}
	read, err := s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: settingsType, CurrentState: v.ImportedResources[0].State})
	if err != nil || settingsErrors(read.Diagnostics) {
		t.Fatal(err, read)
	}
	return read.NewState
}

func settingsApply(t *testing.T, s tfprotov6.ProviderServer, prior, plan, config *tfprotov6.DynamicValue, p []byte) *tfprotov6.ApplyResourceChangeResponse {
	t.Helper()
	r, e := s.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: settingsType, PriorState: prior, PlannedState: plan, Config: config, PlannedPrivate: p})
	if e != nil {
		t.Fatal(e)
	}
	return r
}

func TestSettingsProtocolImportPlanPartialClearLookupAndDestroy(t *testing.T) {
	a := &settingsAPI{}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas[settingsType]
	prior := settingsImport(t, s, sc)
	for _, id := range []string{"", "settings", "trust"} {
		r, e := s.ImportResourceState(context.Background(), &tfprotov6.ImportResourceStateRequest{TypeName: settingsType, ID: id})
		if e != nil || !settingsErrors(r.Diagnostics) {
			t.Fatal("bad import accepted", id, r)
		}
	}
	omitted := settingsValue(t, sc, nil)
	plan, e := s.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: settingsType, PriorState: prior, ProposedNewState: omitted, Config: omitted})
	if e != nil || settingsErrors(plan.Diagnostics) {
		t.Fatal(e, plan)
	}
	if !settingsAttrs(t, sc, plan.PlannedState)["install_crls"].Equal(settingsAttrs(t, sc, prior)["install_crls"]) {
		t.Fatal("omitted value changed")
	}

	for key, want := range settingsAttrs(t, sc, prior) {
		if !settingsAttrs(t, sc, plan.PlannedState)[key].Equal(want) {
			t.Fatal("omitted configuration changed prior field", key)
		}
	}
	unchanged := settingsApply(t, s, prior, plan.PlannedState, omitted, nil)
	if settingsErrors(unchanged.Diagnostics) || a.saves != 0 || a.activations != 0 {
		t.Fatal("omission caused mutation", unchanged.Diagnostics)
	}
	values := map[string]any{"id": "trust_settings", "store_intermediate_certs": true, "install_crls": false, "fetch_crls": false, "enable_legacy_sect": true, "enable_config_constraints": false, "cipher_string": []string{}}
	config := settingsValue(t, sc, map[string]any{"install_crls": false, "cipher_string": []string{}})
	result := settingsApply(t, s, prior, settingsValue(t, sc, values), config, nil)
	if settingsErrors(result.Diagnostics) {
		t.Fatal(result.Diagnostics)
	}
	if a.saves != 1 || a.activations != 1 || len(a.wire) != 2 || a.wire["install_crls"] != "0" || a.wire["CipherString"] != "" {
		t.Fatal("partial clear failed", a.wire)
	}
	if a.fields["Ciphersuites"] != "opaque" || a.fields["MinProtocol"] != "TLSv1.2" {
		t.Fatal("unowned policy changed")
	}
	lookup, e := s.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: settingsType, Config: settingsValue(t, schemas.DataSourceSchemas[settingsType], nil)})
	if e != nil || settingsErrors(lookup.Diagnostics) {
		t.Fatal(e, lookup)
	}

	for key, want := range settingsAttrs(t, sc, result.NewState) {
		if !settingsAttrs(t, schemas.DataSourceSchemas[settingsType], lookup.State)[key].Equal(want) {
			t.Fatal("lookup value differs", key)
		}
	}
	a.fields["fetch_crls"] = "1"
	a.fields["CipherString"] = "AES128-SHA"
	changed, e := s.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: settingsType, Config: settingsValue(t, schemas.DataSourceSchemas[settingsType], nil)})
	if e != nil || settingsErrors(changed.Diagnostics) {
		t.Fatal(e, changed)
	}
	expected := map[string]any{"id": "trust_settings", "store_intermediate_certs": true, "install_crls": false, "fetch_crls": true, "enable_legacy_sect": true, "enable_config_constraints": false, "cipher_string": []string{"AES128-SHA"}}
	for key, want := range settingsAttrs(t, schemas.DataSourceSchemas[settingsType], settingsValue(t, schemas.DataSourceSchemas[settingsType], expected)) {
		if !settingsAttrs(t, schemas.DataSourceSchemas[settingsType], changed.State)[key].Equal(want) {
			t.Fatal("changed lookup wrong", key)
		}
	}
	before := a.reads
	destroyed := settingsApply(t, s, result.NewState, settingsNull(t, sc), settingsNull(t, sc), result.Private)
	if settingsErrors(destroyed.Diagnostics) || a.reads != before || a.saves != 1 || a.activations != 1 {
		t.Fatal("destroy touched API", destroyed)
	}
	created := settingsApply(t, s, settingsNull(t, sc), config, config, nil)
	if !settingsErrors(created.Diagnostics) || a.saves != 1 {
		t.Fatal("create must require import")
	}
}

func TestSettingsProtocolDurableRecoveryNoRepeatedSet(t *testing.T) {
	for _, failure := range []string{"activation", "read", "uncertain", "connection lost", "mismatch", "confirmed mismatch", "validation"} {
		t.Run(failure, func(t *testing.T) {
			a := &settingsAPI{}
			h := httptest.NewServer(a)
			defer h.Close()
			s, schemas := settingsServer(t, h.URL)
			sc := schemas.ResourceSchemas[settingsType]
			prior := settingsImport(t, s, sc)
			a.failure = failure
			config := settingsValue(t, sc, map[string]any{"install_crls": false})
			failed := settingsApply(t, s, prior, config, config, nil)
			if !settingsErrors(failed.Diagnostics) {
				t.Fatal("failure not reported")
			}
			if failure == "confirmed mismatch" {
				if a.activations != 0 {
					t.Fatal("contradictory confirmed save activated")
				}
				return
			}
			if failure == "validation" {
				if len(failed.Private) != 0 || a.activations != 0 {
					t.Fatal("validation became pending")
				}
				return
			}
			if len(failed.Private) == 0 {
				t.Fatal("missing durable intent")
			}
			var id string
			if e := settingsAttrs(t, sc, failed.NewState)["id"].As(&id); e != nil || id != "trust_settings" {
				t.Fatal("lost identity")
			}
			reads, activations := a.reads, a.activations
			fresh, _ := settingsServer(t, h.URL)
			destroyed := settingsApply(t, fresh, failed.NewState, settingsNull(t, sc), settingsNull(t, sc), failed.Private)
			if settingsErrors(destroyed.Diagnostics) || a.reads != reads || a.activations != activations || a.saves != 1 {
				t.Fatal("pending destroy touched API")
			}
			if failure == "read" {
				blocked, e := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: settingsType, CurrentState: failed.NewState, Private: failed.Private})
				if e != nil || !settingsErrors(blocked.Diagnostics) || len(blocked.Private) == 0 || a.saves != 1 || a.activations != activations {
					t.Fatal("malformed recovery lost intent or mutated", e, blocked)
				}
			}

			a.badRead = false
			a.failure = ""
			recovered, e := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: settingsType, CurrentState: failed.NewState, Private: failed.Private})
			if e != nil || settingsErrors(recovered.Diagnostics) || a.saves != 1 {
				t.Fatal("recovery failed/repeated set", e, recovered)
			}
			want := activations + 1
			if failure == "mismatch" {
				want = activations
			}
			if a.activations != want {
				t.Fatal("wrong recovery activation count", a.activations, want)
			}
			again, e := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: settingsType, CurrentState: recovered.NewState, Private: recovered.Private})
			if e != nil || settingsErrors(again.Diagnostics) || a.activations != want || a.saves != 1 {
				t.Fatal("recovery repeated")
			}
		})
	}
}

func settingsPlan(t *testing.T, s tfprotov6.ProviderServer, prior, config *tfprotov6.DynamicValue) *tfprotov6.PlanResourceChangeResponse {
	t.Helper()
	p, e := s.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: settingsType, PriorState: prior, ProposedNewState: config, Config: config})
	if e != nil || settingsErrors(p.Diagnostics) {
		t.Fatal(e, p)
	}
	return p
}

func TestSettingsProtocolStaleOmittedPlanBeforePartialSave(t *testing.T) {
	for _, field := range []string{"fetch_crls", "CipherString"} {
		t.Run(field, func(t *testing.T) {
			a := &settingsAPI{}
			h := httptest.NewServer(a)
			defer h.Close()
			s, schemas := settingsServer(t, h.URL)
			sc := schemas.ResourceSchemas[settingsType]
			prior := settingsImport(t, s, sc)
			config := settingsValue(t, sc, map[string]any{"install_crls": false})
			plan := settingsPlan(t, s, prior, config)
			if field == "fetch_crls" {
				a.fields[field] = "1"
			} else {
				a.fields[field] = "AES128-SHA"
			}
			r := settingsApply(t, s, prior, plan.PlannedState, config, plan.PlannedPrivate)
			if !settingsErrors(r.Diagnostics) || a.saves != 0 || a.activations != 0 {
				t.Fatal("stale plan wrote", r)
			}
			found := false
			for _, d := range r.Diagnostics {
				if d.Summary == "Stale Trust settings plan" {
					found = true
				}
			}
			if !found {
				t.Fatal("missing stale-plan diagnostic", r.Diagnostics)
			}
		})
	}
}

func TestSettingsProtocolStrictSaveOutcomeAndPendingUpdate(t *testing.T) {
	for _, reply := range []string{`{"result":"saved","result":"failed"}`, `{"result":"failed","result":"saved"}`, `{"result":"saved","Result":"failed"}`, `{"Result":"failed","result":"saved"}`, `{"result":"saved","status":"failed"}`, `{"result":"saved","Status":"ok"}`} {
		t.Run(reply, func(t *testing.T) {
			a := &settingsAPI{saveReply: reply}
			h := httptest.NewServer(a)
			defer h.Close()
			s, schemas := settingsServer(t, h.URL)
			sc := schemas.ResourceSchemas[settingsType]
			prior := settingsImport(t, s, sc)
			config := settingsValue(t, sc, map[string]any{"install_crls": false})
			p := settingsPlan(t, s, prior, config)
			failed := settingsApply(t, s, prior, p.PlannedState, config, p.PlannedPrivate)
			if !settingsErrors(failed.Diagnostics) || len(failed.Private) == 0 || a.saves != 1 || a.activations != 0 {
				t.Fatal("ambiguous response lost intent", failed)
			}
			a.saveReply = ""
			fresh, _ := settingsServer(t, h.URL)
			nextConfig := settingsValue(t, sc, map[string]any{"install_crls": true})
			next := settingsApply(t, fresh, failed.NewState, nextConfig, nextConfig, failed.Private)
			freshPlan := false
			for _, d := range next.Diagnostics {
				if d.Summary == "Trust settings recovered; fresh plan required" {
					freshPlan = true
				}
			}
			if !settingsErrors(next.Diagnostics) || !freshPlan || a.saves != 1 || a.activations != 1 || a.fields["install_crls"] != "0" {
				t.Fatal("pending Update repeated or changed save", next)
			}
		})
	}
}

func TestSettingsProtocolActivationSingleAttemptStrictAndPending(t *testing.T) {
	var targets atomic.Int64
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { targets.Add(1); _, _ = fmt.Fprint(w, `{"status":"ok"}`) }))
	defer target.Close()
	for _, mode := range []string{"redirect", "500", "drop", `{"status":"ok","result":"failed"}`, `{"status":"failed","status":"ok"}`, `{"status":"ok","Status":"failed"}`, `{"status":"ok","Result":"ok"}`, `{"status":"ok"}`} {
		t.Run(mode, func(t *testing.T) {
			a := &settingsAPI{}
			switch mode {
			case "redirect":
				a.redirect = target.URL
			case "500":
				a.activationStatus = 500
			case "drop":
				a.dropActivation = true
			default:
				a.activationReply = mode
			}
			h := httptest.NewServer(a)
			defer h.Close()
			s, schemas := settingsServer(t, h.URL)
			sc := schemas.ResourceSchemas[settingsType]
			prior := settingsImport(t, s, sc)
			config := settingsValue(t, sc, map[string]any{"install_crls": false})
			p := settingsPlan(t, s, prior, config)
			r := settingsApply(t, s, prior, p.PlannedState, config, p.PlannedPrivate)
			a.mu.Lock()
			saves, activations := a.saves, a.activations
			a.mu.Unlock()
			if saves != 1 || activations != 1 || targets.Load() != 0 {
				t.Fatal("activation replayed or redirected")
			}
			if mode == `{"status":"ok"}` {
				if settingsErrors(r.Diagnostics) || len(r.Private) != 0 {
					t.Fatal("normal activation failed", r)
				}
				return
			}
			if !settingsErrors(r.Diagnostics) || len(r.Private) == 0 {
				t.Fatal("failed activation lost pending", r)
			}
			a.redirect = ""
			a.activationStatus = 0
			a.dropActivation = false
			a.activationReply = ""
			fresh, _ := settingsServer(t, h.URL)
			nextConfig := settingsValue(t, sc, map[string]any{"install_crls": true})
			next := settingsApply(t, fresh, r.NewState, nextConfig, nextConfig, r.Private)
			found := false
			for _, d := range next.Diagnostics {
				if strings.Contains(d.Summary, "fresh plan required") {
					found = true
				}
			}
			if !found || a.saves != 1 || a.activations != 2 {
				t.Fatal("confirmed pending update replayed save", next)
			}
		})
	}
}

func TestSettingsProtocolCanceledAfterSaveRetainsPrivate(t *testing.T) {
	a := &settingsAPI{}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas[settingsType]
	prior := settingsImport(t, s, sc)
	config := settingsValue(t, sc, map[string]any{"install_crls": false})
	p := settingsPlan(t, s, prior, config)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a.cancel = cancel
	r, e := s.ApplyResourceChange(ctx, &tfprotov6.ApplyResourceChangeRequest{TypeName: settingsType, PriorState: prior, PlannedState: p.PlannedState, Config: config, PlannedPrivate: p.PlannedPrivate})
	a.mu.Lock()
	saves := a.saves
	a.mu.Unlock()
	if e != nil || !settingsErrors(r.Diagnostics) || len(r.Private) == 0 || saves != 1 {
		t.Fatal("canceled save lost intent", e, r)
	}
	a.cancel = nil
	fresh, _ := settingsServer(t, h.URL)
	recovered, e := fresh.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: settingsType, CurrentState: r.NewState, Private: r.Private})
	if e != nil || settingsErrors(recovered.Diagnostics) || a.saves != 1 || a.activations != 1 {
		t.Fatal("canceled save recovery failed", e, recovered)
	}
}

func TestSettingsProtocolConcurrentUpdateReadRecoverySerialized(t *testing.T) {
	a := &settingsAPI{failure: "activation"}
	var armed atomic.Bool
	events := make(chan string, 8)
	permit := make(chan struct{})
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if armed.Load() && strings.HasPrefix(r.URL.Path, "/api/trust/") {
			events <- r.URL.Path
			<-permit
		}
		a.ServeHTTP(w, r)
	}))
	defer h.Close()
	defer close(permit)
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas[settingsType]
	prior := settingsImport(t, s, sc)
	config := settingsValue(t, sc, map[string]any{"install_crls": false})
	p := settingsPlan(t, s, prior, config)
	failed := settingsApply(t, s, prior, p.PlannedState, config, p.PlannedPrivate)
	if len(failed.Private) == 0 {
		t.Fatal("missing pending fixture")
	}
	a.failure = ""
	armed.Store(true)
	updateDone := make(chan *tfprotov6.ApplyResourceChangeResponse, 1)
	go func() { updateDone <- settingsApply(t, s, prior, p.PlannedState, config, nil) }()
	if path := <-events; path != "/api/trust/settings/get" {
		t.Fatal(path)
	}
	readStarted := make(chan struct{})
	readDone := make(chan *tfprotov6.ReadResourceResponse, 1)
	go func() {
		close(readStarted)
		r, e := s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: settingsType, CurrentState: failed.NewState, Private: failed.Private})
		if e != nil {
			t.Error(e)
		}
		readDone <- r
	}()
	<-readStarted
	// Handshake each stage while the competing Read is actively waiting. It must
	// not reach HTTP until Update completes baseline/save/read-back/activation.
	for _, next := range []string{"/api/trust/settings/set", "/api/trust/settings/get", "/api/trust/settings/reconfigure"} {
		select {
		case path := <-events:
			t.Fatal("Read interleaved while Update stage blocked", path)
		case <-time.After(30 * time.Millisecond):
		}
		permit <- struct{}{}
		if path := <-events; path != next {
			t.Fatal("owner interleaved lifecycle", path, next)
		}
	}
	select {
	case path := <-events:
		t.Fatal("Read interleaved activation", path)
	case <-time.After(30 * time.Millisecond):
	}
	permit <- struct{}{}
	updated := <-updateDone
	for _, expected := range []string{"/api/trust/settings/get", "/api/trust/settings/reconfigure"} {
		if path := <-events; path != expected {
			t.Fatal(path, expected)
		}
		permit <- struct{}{}
	}
	read := <-readDone
	if settingsErrors(updated.Diagnostics) || read == nil || settingsErrors(read.Diagnostics) || a.saves != 2 || a.activations != 3 {
		t.Fatal("serialized lifecycle failed", updated, read, a.saves, a.activations)
	}
}

func TestSettingsProtocolCorruptPrivateAndIdentityNoHTTP(t *testing.T) {
	a := &settingsAPI{}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas[settingsType]
	prior := settingsImport(t, s, sc)
	config := settingsValue(t, sc, map[string]any{"install_crls": false})
	reads := a.reads
	for _, raw := range []string{`{"id":"wrong","kind":"saved","payload":{"install_crls":"0"}}`, ` {"id":"trust_settings","kind":"saved","payload":{"install_crls":"0"}}`, `{"id":"trust_settings","Kind":"saved","payload":{"install_crls":"0"}}`, `{"id":"trust_settings","kind":"saved","payload":{"install_crls":"0","install_crls":"1"}}`} {
		private, e := json.Marshal(map[string][]byte{"trust_settings_recovery_v1": []byte(raw)})
		if e != nil {
			t.Fatal(e)
		}
		read, e := s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: settingsType, CurrentState: prior, Private: private})
		if e != nil || !settingsErrors(read.Diagnostics) || string(read.Private) != string(private) {
			t.Fatal("corrupt recovery marker changed or accepted", e, read)
		}
		updated := settingsApply(t, s, prior, config, config, private)
		if !settingsErrors(updated.Diagnostics) || string(updated.Private) != string(private) {
			t.Fatal("corrupt Update marker changed or accepted", updated)
		}
		destroyed := settingsApply(t, s, prior, settingsNull(t, sc), settingsNull(t, sc), private)
		if settingsErrors(destroyed.Diagnostics) {
			t.Fatal("Delete inspected corrupt marker", destroyed)
		}
	}
	wrong := settingsValue(t, sc, map[string]any{"id": "wrong", "store_intermediate_certs": true, "install_crls": true, "fetch_crls": false, "enable_legacy_sect": true, "enable_config_constraints": false, "cipher_string": []string{"AES256-SHA"}})
	read, e := s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: settingsType, CurrentState: wrong})
	if e != nil || !settingsErrors(read.Diagnostics) {
		t.Fatal("bad identity accepted", e, read)
	}
	if a.reads != reads || a.saves != 0 || a.activations != 0 {
		t.Fatal("corrupt marker or identity caused HTTP")
	}
}
