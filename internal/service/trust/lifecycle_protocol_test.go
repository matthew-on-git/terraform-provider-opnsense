// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type lifecycleAPI struct {
	issuer, commonName, country string
	mu                          sync.Mutex
	kind                        string
	posts                       []map[string]string
	reads                       int
	badRead                     string
	failWrite                   int
	reply                       string
	redirect                    string
	certificate, key, csr       string
	description                 string
}

func (a *lifecycleAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if strings.HasPrefix(r.URL.Path, "/api/trust/") {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "fixture-key" || pass != "fixture-secret" {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
	}
	if a.description == "" {
		a.description = "fixture"
	}
	if a.commonName == "" {
		a.commonName = "example.test"
	}
	if a.country == "" {
		a.country = "US"
	}
	switch r.URL.Path {
	case "/api/core/menu/search":
		_, _ = fmt.Fprint(w, `{}`)
		return
	case "/api/core/firmware/status":
		_, _ = fmt.Fprint(w, `{"product":{"product_name":"OPNsense","product_version":"26.7.8"}}`)
		return
	}
	if strings.Contains(r.URL.Path, "/get/") {
		a.reads++
		if a.badRead != "" {
			_, _ = fmt.Fprint(w, a.badRead)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{a.kind: map[string]any{"refid": "0123456789abc", "caref": a.issuer, "descr": a.description, "crt_payload": a.certificate, "prv_payload": a.key, "csr_payload": a.csr, "commonname": a.commonName, "country": map[string]any{a.country: map[string]any{"selected": 1}}, "action": map[string]any{"manual": map[string]any{"selected": 1}}, "lifetime": "397", "valid_from": 1793664000, "valid_to": "1825200000", "rfc3280_purpose": "id-kp-serverAuth", "name": "/CN=" + a.commonName}})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "wrong method", 400)
		return
	}
	var body map[string]map[string]string
	if strings.Contains(r.URL.Path, "/del/") {
		_, _ = fmt.Fprint(w, `{"result":"deleted"}`)
		return
	}
	if json.NewDecoder(r.Body).Decode(&body) != nil || body[a.kind] == nil {
		http.Error(w, "bad body", 400)
		return
	}
	a.posts = append(a.posts, body[a.kind])
	a.description = body[a.kind]["descr"]
	if v, ok := body[a.kind]["caref"]; ok {
		a.issuer = v
	}
	if v, ok := body[a.kind]["commonname"]; ok {
		a.commonName = v
	}
	if v, ok := body[a.kind]["country"]; ok {
		a.country = v
	}
	if v, ok := body[a.kind]["crt_payload"]; ok {
		a.certificate = v
	}
	if v, ok := body[a.kind]["prv_payload"]; ok {
		a.key = v
	}
	if v, ok := body[a.kind]["csr_payload"]; ok {
		a.csr = v
	}
	if a.redirect != "" {
		http.Redirect(w, r, a.redirect, http.StatusTemporaryRedirect)
		return
	}
	if a.failWrite != 0 {
		http.Error(w, "private sentinel", a.failWrite)
		return
	}
	if a.reply != "" {
		_, _ = fmt.Fprint(w, a.reply)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/add") && body[a.kind]["private_key_location"] == "local" {
		a.key = ""
		_ = json.NewEncoder(w).Encode(map[string]string{"result": "saved", "uuid": "fixture-id", "private_key": "one-time-sentinel"})
		return
	}
	_, _ = fmt.Fprint(w, `{"result":"saved","uuid":"fixture-id"}`)
}

func lifecycleCSR(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{Subject: pkix.Name{CommonName: "example.test"}}, key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: raw}))
}

func lifecyclePlan(t *testing.T, s tfprotov6.ProviderServer, kind string, sc *tfprotov6.Schema, prior, config *tfprotov6.DynamicValue) *tfprotov6.DynamicValue {
	t.Helper()
	// Emulate Core proposal construction: preserve prior computed observations,
	// while configured attributes (including explicit nulls) come from config.
	proposed := config
	priorValue, err := prior.Unmarshal(sc.ValueType())
	if err != nil {
		t.Fatal(err)
	}
	if !priorValue.IsNull() {
		old := settingsAttrs(t, sc, prior)
		next := settingsAttrs(t, sc, config)
		for _, attribute := range sc.Block.Attributes {
			if attribute.Computed && next[attribute.Name].IsNull() {
				next[attribute.Name] = old[attribute.Name]
			}
		}
		value, buildErr := tfprotov6.NewDynamicValue(sc.ValueType(), tftypes.NewValue(sc.ValueType(), next))
		if buildErr != nil {
			t.Fatal(buildErr)
		}
		proposed = &value
	}
	result, err := s.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: "opnsense_trust_" + kind, PriorState: prior, ProposedNewState: proposed, Config: config})
	if err != nil || settingsErrors(result.Diagnostics) {
		t.Fatalf("plan failed: %v %+v", err, result.Diagnostics)
	}
	return result.PlannedState
}

func lifecycleApply(t *testing.T, s tfprotov6.ProviderServer, kind string, prior, plan, config *tfprotov6.DynamicValue) *tfprotov6.ApplyResourceChangeResponse {
	t.Helper()
	result, err := s.ApplyResourceChange(context.Background(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "opnsense_trust_" + kind, PriorState: prior, PlannedState: plan, Config: config})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func lifecycleString(t *testing.T, sc *tfprotov6.Schema, state *tfprotov6.DynamicValue, key string) string {
	t.Helper()
	attrs := settingsAttrs(t, sc, state)
	if attrs[key].IsNull() {
		return ""
	}
	var value string
	if err := attrs[key].As(&value); err != nil {
		t.Fatal(key, err)
	}
	return value
}

func TestTrustLifecycleProtocolActionMatrix(t *testing.T) {
	cert, key := selfSignedPEM(t, "example.test")
	csr := lifecycleCSR(t)
	matrix := []struct {
		kind, action string
		fields       map[string]any
	}{
		{"ca", "internal", map[string]any{"common_name": "example.test"}},
		{"ca", "internal", map[string]any{"common_name": "example.test", "ca_ref": "issuer"}},
		{"ca", "ocsp", map[string]any{"common_name": "example.test", "ca_ref": "issuer", "ocsp_uri": "https://ocsp.example.test"}},
		{"ca", "existing", map[string]any{"certificate": cert, "private_key": key}},
		{"cert", "internal", map[string]any{"common_name": "example.test", "ca_ref": "issuer", "private_key_location": "firewall"}},
		{"cert", "internal", map[string]any{"common_name": "example.test", "ca_ref": "issuer", "private_key_location": "local"}},
		{"cert", "external", map[string]any{"common_name": "example.test"}},
		{"cert", "import", map[string]any{"certificate": cert, "private_key": key}},
		{"cert", "sign_csr", map[string]any{"csr": csr, "ca_ref": "issuer"}},
		{"cert", "import_csr", map[string]any{"certificate": cert}},
		{"cert", "reissue", map[string]any{"private_key": key, "ca_ref": "issuer", "common_name": "example.test"}},
		{"cert", "manual", map[string]any{"certificate": cert, "private_key": key}},
	}
	for i, row := range matrix {
		t.Run(fmt.Sprintf("%s-%s-%d", row.kind, row.action, i), func(t *testing.T) {
			a := &lifecycleAPI{kind: row.kind, certificate: cert, key: key, csr: csr}
			h := httptest.NewServer(a)
			defer h.Close()
			s, schemas := settingsServer(t, h.URL)
			sc := schemas.ResourceSchemas["opnsense_trust_"+row.kind]
			fields := map[string]any{"description": "fixture", "action": row.action, "lifetime": 90, "key_type": "4096", "digest": "sha384"}
			for k, v := range row.fields {
				fields[k] = v
			}
			if row.kind == "cert" {
				fields["cert_type"] = "combined_server_client"
				fields["altnames_dns"] = "one.example.test\ntwo.example.test"
				fields["altnames_ip"] = "192.0.2.1\n2001:db8::1"
			}
			config := settingsValue(t, sc, fields)
			prior := settingsNull(t, sc)
			plan := lifecyclePlan(t, s, row.kind, sc, prior, config)
			result := lifecycleApply(t, s, row.kind, prior, plan, config)
			if settingsErrors(result.Diagnostics) {
				t.Fatalf("create: %+v", result.Diagnostics)
			}
			if row.kind == "ca" {
				wantIssuer, _ := fields["ca_ref"].(string)
				if a.posts[0]["caref"] != wantIssuer || lifecycleString(t, sc, result.NewState, "returned_ca_ref") != wantIssuer {
					t.Fatal("root/intermediate/OCSP issuer distinction lost")
				}
			}
			if len(a.posts) != 1 || a.posts[0]["action"] != row.action {
				t.Fatalf("wrong action/write count: %+v", a.posts)
			}
			if lifecycleString(t, sc, result.NewState, "action") != row.action || lifecycleString(t, sc, result.NewState, "valid_from") != "1793664000" {
				t.Fatal("read overwrote intent or numeric validity")
			}
			attrs := settingsAttrs(t, sc, result.NewState)
			if !attrs["lifetime"].Equal(settingsAttrs(t, sc, config)["lifetime"]) {
				t.Fatal("configured lifetime overwritten")
			}
			if row.kind == "cert" && fields["private_key_location"] == "local" && lifecycleString(t, sc, result.NewState, "generated_private_key") != "one-time-sentinel" {
				t.Fatal("lost one-time key")
			}
			if row.action == "internal" || row.action == "external" || row.action == "sign_csr" || row.action == "reissue" || row.action == "ocsp" {
				if a.posts[0]["lifetime"] != "90" || a.posts[0]["key_type"] != "4096" || a.posts[0]["digest"] != "sha384" {
					t.Fatal("wrong generation wire", a.posts[0])
				}
			}
			if row.kind == "cert" && row.action == "internal" && a.posts[0]["altnames_dns"] != "one.example.test\ntwo.example.test" {
				t.Fatal("SAN delimiter changed")
			}
			// A real framework plan+apply description edit must never repeat actions.
			fields["description"] = "fixture renamed"
			config = settingsValue(t, sc, fields)
			plan = lifecyclePlan(t, s, row.kind, sc, result.NewState, config)
			updated := lifecycleApply(t, s, row.kind, result.NewState, plan, config)
			if settingsErrors(updated.Diagnostics) {
				t.Fatalf("metadata update: %+v", updated.Diagnostics)
			}
			if lifecycleString(t, sc, updated.NewState, "description") != "fixture renamed" {
				t.Fatal("metadata readback not authoritative")
			}
			unchanged := lifecyclePlan(t, s, row.kind, sc, updated.NewState, config)
			oldAttrs, newAttrs := settingsAttrs(t, sc, updated.NewState), settingsAttrs(t, sc, unchanged)
			for name, want := range oldAttrs {
				if !newAttrs[name].Equal(want) {
					t.Fatalf("unchanged plan differs for %s", name)
				}
			}

			want := "manual"
			if row.kind == "ca" {
				want = "existing"
			}
			if len(a.posts) != 2 || a.posts[1]["action"] != want || a.posts[1]["key_type"] != "" {
				t.Fatal("metadata repeated generation", a.posts)
			}
			if row.kind == "cert" && (a.posts[1]["prv_payload"] != "" || a.posts[1]["crt_payload"] != "") {
				t.Fatal("metadata retransmitted inputs")
			}
		})
	}
}

func TestTrustLifecycleProtocolInvalidPrerequisitesNoWrites(t *testing.T) {
	cert, key := selfSignedPEM(t, "example.test")
	_, other := selfSignedPEM(t, "other.example.test")
	rows := []struct {
		kind   string
		fields map[string]any
	}{
		{"ca", map[string]any{"action": "existing", "certificate": cert}},
		{"ca", map[string]any{"action": "existing", "certificate": cert, "private_key": other}},
		{"ca", map[string]any{"action": "existing", "certificate": cert + cert, "private_key": key}},
		{"ca", map[string]any{"action": "ocsp", "common_name": "example.test"}},
		{"ca", map[string]any{"action": "invalid"}},
		{"cert", map[string]any{"action": "internal", "lifetime": -1}},
		{"cert", map[string]any{"action": "import", "certificate": cert, "private_key": other}},
		{"cert", map[string]any{"action": "sign_csr", "ca_ref": "issuer", "csr": "invalid"}},
		{"cert", map[string]any{"action": "sign_csr", "csr": lifecycleCSR(t)}},
		{"cert", map[string]any{"action": "import_csr"}},
		{"cert", map[string]any{"action": "reissue", "ca_ref": "issuer"}},
		{"cert", map[string]any{"action": "external", "common_name": "example.test", "key_type": "invalid"}},
		{"cert", map[string]any{"action": "external", "common_name": "example.test", "digest": "invalid"}},
		{"cert", map[string]any{"action": "internal", "common_name": "example.test", "cert_type": "invalid"}},
		{"cert", map[string]any{"action": "internal", "common_name": "example.test", "private_key_location": "invalid"}},
		{"cert", map[string]any{"action": "external", "common_name": "example.test", "lifetime": 0}},
	}
	for i, row := range rows {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			a := &lifecycleAPI{kind: row.kind}
			h := httptest.NewServer(a)
			defer h.Close()
			s, schemas := settingsServer(t, h.URL)
			sc := schemas.ResourceSchemas["opnsense_trust_"+row.kind]
			row.fields["description"] = "fixture"
			config := settingsValue(t, sc, row.fields)
			prior := settingsNull(t, sc)
			plan := lifecyclePlan(t, s, row.kind, sc, prior, config)
			result := lifecycleApply(t, s, row.kind, prior, plan, config)
			if !settingsErrors(result.Diagnostics) || len(a.posts) != 0 {
				t.Fatal("invalid action mutated", result.Diagnostics, a.posts)
			}
			for _, d := range result.Diagnostics {
				if strings.Contains(d.Detail, key) || strings.Contains(d.Detail, other) {
					t.Fatal("private material leaked")
				}
			}
		})
	}
}

func TestTrustLifecycleProtocolPartialFailureKeyRetentionAndStrictReads(t *testing.T) {
	a := &lifecycleAPI{kind: "cert", badRead: `{"cert":{}}`}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas["opnsense_trust_cert"]
	config := settingsValue(t, sc, map[string]any{"description": "fixture", "action": "internal", "common_name": "example.test", "private_key_location": "local"})
	prior := settingsNull(t, sc)
	result := lifecycleApply(t, s, "cert", prior, lifecyclePlan(t, s, "cert", sc, prior, config), config)
	value, decodeErr := result.NewState.Unmarshal(sc.ValueType())
	if decodeErr != nil || !value.IsFullyKnown() {
		t.Fatal("provisional saved state contains unknown outputs", decodeErr)
	}
	if settingsErrors(result.Diagnostics) || len(result.Diagnostics) == 0 || lifecycleString(t, sc, result.NewState, "id") != "fixture-id" || lifecycleString(t, sc, result.NewState, "generated_private_key") != "one-time-sentinel" {
		t.Fatal("saved identity/key lost", result.Diagnostics)
	}
	a.badRead = ""
	read, err := s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_trust_cert", CurrentState: result.NewState})
	if err != nil || settingsErrors(read.Diagnostics) || lifecycleString(t, sc, read.NewState, "generated_private_key") != "one-time-sentinel" {
		t.Fatal("refresh lost retained key", err, read.Diagnostics)
	}
	state := read.NewState
	for _, bad := range []string{`{}`, `{"cert":null}`, `{"cert":{}}`, `{"cert":{"refid":"r","valid_from":{}}}`, `{"cert":{"refid":"r","crt_payload":{}}}`, `{"cert":{"refid":"r","in_use":true}}`, `{"cert":{"refid":"r","country":{"US":{"selected":1},"NL":{"selected":1}}}}`, `{"cert":{"refid":"r","prv_payload":{}}}`} {
		a.badRead = bad
		read, err = s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_trust_cert", CurrentState: state})
		if err != nil || !settingsErrors(read.Diagnostics) {
			t.Fatal("malformed read accepted", bad, err, read.Diagnostics)
		}
		if read.NewState != nil && lifecycleString(t, sc, read.NewState, "id") != "fixture-id" {
			t.Fatal("managed identity removed on malformed response")
		}
	}
	a.badRead = `{"cert":[]}`
	read, err = s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_trust_cert", CurrentState: state})
	if err != nil || settingsErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	value, err = read.NewState.Unmarshal(sc.ValueType())
	if err != nil || !value.IsNull() {
		t.Fatal("affirmative deletion not removed", err)
	}
	if len(a.posts) != 1 {
		t.Fatal("read replayed create")
	}
}

func TestTrustLifecycleProtocolExplicitTransitionsAndLocalSetRejection(t *testing.T) {
	cert, key := selfSignedPEM(t, "example.test")
	a := &lifecycleAPI{kind: "cert", certificate: cert, key: key}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas["opnsense_trust_cert"]
	fields := map[string]any{"description": "fixture", "action": "external", "common_name": "example.test", "ca_ref": "issuer"}
	config := settingsValue(t, sc, fields)
	prior := settingsNull(t, sc)
	created := lifecycleApply(t, s, "cert", prior, lifecyclePlan(t, s, "cert", sc, prior, config), config)
	if settingsErrors(created.Diagnostics) {
		t.Fatal(created.Diagnostics)
	}
	prior = created.NewState
	for _, action := range []string{"sign_csr", "import_csr", "reissue", "reissue", "internal"} {
		fields["action"] = action
		fields["operation_token"] = fmt.Sprint(len(a.posts))
		fields["csr"] = lifecycleCSRWithKey(t, key)
		fields["certificate"] = cert
		fields["private_key"] = key
		if action == "internal" {
			fields["private_key_location"] = "local"
		}
		config = settingsValue(t, sc, fields)
		plan := lifecyclePlan(t, s, "cert", sc, prior, config)
		before := len(a.posts)
		result := lifecycleApply(t, s, "cert", prior, plan, config)
		if action == "internal" {
			if !settingsErrors(result.Diagnostics) || len(a.posts) != before {
				t.Fatal("local set accepted")
			}
			continue
		}
		if settingsErrors(result.Diagnostics) || len(a.posts) != before+1 || a.posts[before]["action"] != action {
			t.Fatal("explicit action missing", action, result.Diagnostics, a.posts)
		}
		prior = result.NewState
	}
}

func TestTrustLifecycleProtocolMutationSingleAttemptAndRedirectRejection(t *testing.T) {
	for _, mode := range []string{"500", "redirect", "validation", "malformed", "duplicate", "lost"} {
		t.Run(mode, func(t *testing.T) {
			a := &lifecycleAPI{kind: "cert"}
			targetPosts := 0
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				targetPosts++
				_, _ = fmt.Fprint(w, `{"result":"saved","uuid":"unsafe"}`)
			}))
			defer target.Close()
			switch mode {
			case "500":
				a.failWrite = 500
			case "redirect":
				a.redirect = target.URL
			case "validation":
				a.reply = `{"result":"failed","validations":{"private":"private sentinel"}}`
			case "malformed":
				a.reply = `private sentinel`
			case "duplicate":
				a.reply = `{"result":"saved","result":"saved","uuid":"ambiguous"}`
			case "lost":
				a.reply = `{"result":"saved"}`
			}
			h := httptest.NewServer(a)
			defer h.Close()
			s, schemas := settingsServer(t, h.URL)
			sc := schemas.ResourceSchemas["opnsense_trust_cert"]
			config := settingsValue(t, sc, map[string]any{"description": "fixture", "action": "external", "common_name": "example.test"})
			prior := settingsNull(t, sc)
			result := lifecycleApply(t, s, "cert", prior, lifecyclePlan(t, s, "cert", sc, prior, config), config)
			if !settingsErrors(result.Diagnostics) || len(a.posts) != 1 || targetPosts != 0 {
				t.Fatal("unsafe mutation retry/redirect", result.Diagnostics, len(a.posts), targetPosts)
			}
			for _, d := range result.Diagnostics {
				if strings.Contains(d.Detail, "private sentinel") {
					t.Fatal("raw error leaked")
				}
			}
		})
	}
}

func TestTrustLifecycleProtocolLookupSensitivityImportAndLegacy(t *testing.T) {
	cert, key := selfSignedPEM(t, "example.test")
	for _, kind := range []string{"ca", "cert"} {
		t.Run(kind, func(t *testing.T) {
			a := &lifecycleAPI{kind: kind, certificate: cert, key: key}
			h := httptest.NewServer(a)
			defer h.Close()
			s, schemas := settingsServer(t, h.URL)
			typ := "opnsense_trust_" + kind
			sc := schemas.ResourceSchemas[typ]
			lookup := schemas.DataSourceSchemas[typ]
			for _, schema := range []*tfprotov6.Schema{sc, lookup} {
				for _, attr := range schema.Block.Attributes {
					if attr.Name == "private_key" || attr.Name == "returned_private_key" || attr.Name == "generated_private_key" {
						if !attr.Sensitive {
							t.Fatal("private field is public", attr.Name)
						}
					}
				}
			}
			imported, err := s.ImportResourceState(context.Background(), &tfprotov6.ImportResourceStateRequest{TypeName: typ, ID: "fixture-id"})
			if err != nil || settingsErrors(imported.Diagnostics) {
				t.Fatal(err, imported.Diagnostics)
			}
			read, err := s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: typ, CurrentState: imported.ImportedResources[0].State})
			if err != nil || settingsErrors(read.Diagnostics) {
				t.Fatal(err, read.Diagnostics)
			}
			for _, name := range []string{"action", "operation_token", "operation_status"} {
				if !settingsAttrs(t, sc, read.NewState)[name].IsNull() {
					t.Fatal("import fabricated operation provenance", name)
				}
			}
			data, err := s.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: typ, Config: settingsValue(t, lookup, map[string]any{"id": "fixture-id"})})
			if err != nil || settingsErrors(data.Diagnostics) {
				t.Fatal(err, data.Diagnostics)
			}
			if lifecycleString(t, lookup, data.State, "certificate") != cert || lifecycleString(t, lookup, data.State, "private_key") != key {
				t.Fatal("lookup lost authoritative payload")
			}
			for _, attr := range lookup.Block.Attributes {
				if attr.Name == "action" || attr.Name == "operation_token" || attr.Name == "generated_private_key" {
					t.Fatal("lookup exposes synthetic operation provenance", attr.Name)
				}
			}
			a.key = ""
			data, err = s.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: typ, Config: settingsValue(t, lookup, map[string]any{"id": "fixture-id"})})
			if err != nil || settingsErrors(data.Diagnostics) || !settingsAttrs(t, lookup, data.State)["private_key"].IsNull() {
				t.Fatal("lookup fabricated absent private key", err)
			}
			if len(a.posts) != 0 {
				t.Fatal("import/lookup mutated appliance")
			}
			// Legacy omission resolves only at creation, never to GET editor action.
			a.key = key
			fields := map[string]any{"description": "fixture"}
			if kind == "ca" {
				fields["common_name"] = "example.test"
				fields["country"] = "US"
			} else {
				fields["certificate"] = "\n" + cert
				fields["private_key"] = "\n" + key
			}
			config := settingsValue(t, sc, fields)
			prior := settingsNull(t, sc)
			result := lifecycleApply(t, s, kind, prior, lifecyclePlan(t, s, kind, sc, prior, config), config)
			if settingsErrors(result.Diagnostics) {
				t.Fatal(result.Diagnostics)
			}
			want := "import"
			if kind == "ca" {
				want = "internal"
			}
			if a.posts[0]["action"] != want || !settingsAttrs(t, sc, result.NewState)["action"].IsNull() {
				t.Fatal("legacy action not preserved")
			}
			if kind == "cert" && (lifecycleString(t, sc, result.NewState, "certificate") != "\n"+cert || lifecycleString(t, sc, result.NewState, "private_key") != "\n"+key) {
				t.Fatal("configured PEM text normalized")
			}
		})
	}
}

func TestTrustLifecycleProtocolUncertainUpdateConsumesIntent(t *testing.T) {
	cert, key := selfSignedPEM(t, "example.test")
	a := &lifecycleAPI{kind: "cert", certificate: cert, key: key}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas["opnsense_trust_cert"]
	fields := map[string]any{"description": "fixture", "action": "internal", "common_name": "example.test", "ca_ref": "issuer"}
	config := settingsValue(t, sc, fields)
	prior := settingsNull(t, sc)
	result := lifecycleApply(t, s, "cert", prior, lifecyclePlan(t, s, "cert", sc, prior, config), config)
	if settingsErrors(result.Diagnostics) {
		t.Fatal(result.Diagnostics)
	}
	fields["action"] = "reissue"
	fields["operation_token"] = "attempt-1"
	config = settingsValue(t, sc, fields)
	a.failWrite = 500
	failed := lifecycleApply(t, s, "cert", result.NewState, lifecyclePlan(t, s, "cert", sc, result.NewState, config), config)
	if !settingsErrors(failed.Diagnostics) || len(a.posts) != 2 || lifecycleString(t, sc, failed.NewState, "operation_status") != "unknown" {
		t.Fatal("uncertain operation not journaled", failed.Diagnostics)
	}
	if lifecycleString(t, sc, failed.NewState, "returned_certificate") != cert {
		t.Fatal("uncertainty discarded prior observations")
	}
	// Refresh is read-only and cannot prove whether a reissue executed.
	a.failWrite = 0
	read, err := s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_trust_cert", CurrentState: failed.NewState})
	if err != nil || settingsErrors(read.Diagnostics) || lifecycleString(t, sc, read.NewState, "operation_status") != "unknown" {
		t.Fatal("refresh fabricated resolved operation", err)
	}
	before := len(a.posts)
	plan := lifecyclePlan(t, s, "cert", sc, read.NewState, config)
	if !settingsAttrs(t, sc, plan)["operation_token"].Equal(settingsAttrs(t, sc, read.NewState)["operation_token"]) {
		t.Fatal("intent lost")
	}
	// Even if Core asks apply again, unchanged intent performs no mutation.
	repeated := lifecycleApply(t, s, "cert", read.NewState, plan, config)
	if settingsErrors(repeated.Diagnostics) || len(a.posts) != before || lifecycleString(t, sc, repeated.NewState, "operation_status") != "unknown" {
		t.Fatal("uncertain reissue repeated or falsely resolved", repeated.Diagnostics)
	}
	// Ordinary description updates retain unknown; a deliberate new token permits retry.
	fields["description"] = "updated"
	config = settingsValue(t, sc, fields)
	metadata := lifecycleApply(t, s, "cert", repeated.NewState, lifecyclePlan(t, s, "cert", sc, repeated.NewState, config), config)
	if settingsErrors(metadata.Diagnostics) || lifecycleString(t, sc, metadata.NewState, "operation_status") != "unknown" {
		t.Fatal("metadata cleared uncertainty")
	}
	fields["operation_token"] = "attempt-2"
	config = settingsValue(t, sc, fields)
	retried := lifecycleApply(t, s, "cert", metadata.NewState, lifecyclePlan(t, s, "cert", sc, metadata.NewState, config), config)
	if settingsErrors(retried.Diagnostics) || a.posts[len(a.posts)-1]["action"] != "reissue" || lifecycleString(t, sc, retried.NewState, "operation_status") != "saved" {
		t.Fatal("deliberate retry missing", retried.Diagnostics)
	}
}

func TestTrustLifecycleProtocolInputEditsRequireExplicitIntentAndUnknownRejects(t *testing.T) {
	cert, key := selfSignedPEM(t, "example.test")
	for _, kind := range []string{"ca", "cert"} {
		t.Run(kind, func(t *testing.T) {
			a := &lifecycleAPI{kind: kind, certificate: cert, key: key}
			h := httptest.NewServer(a)
			defer h.Close()
			s, schemas := settingsServer(t, h.URL)
			sc := schemas.ResourceSchemas["opnsense_trust_"+kind]
			fields := map[string]any{"description": "fixture", "common_name": "example.test", "action": "internal", "key_type": "2048"}
			config := settingsValue(t, sc, fields)
			prior := settingsNull(t, sc)
			created := lifecycleApply(t, s, kind, prior, lifecyclePlan(t, s, kind, sc, prior, config), config)
			if settingsErrors(created.Diagnostics) {
				t.Fatal(created.Diagnostics)
			}
			fields["key_type"] = "4096"
			config = settingsValue(t, sc, fields)
			before := len(a.posts)
			blocked := lifecycleApply(t, s, kind, created.NewState, lifecyclePlan(t, s, kind, sc, created.NewState, config), config)
			if !settingsErrors(blocked.Diagnostics) || len(a.posts) != before {
				t.Fatal("generation input silently accepted")
			}
			// An unknown configured action must never become empty/default on apply.
			attrs := settingsAttrs(t, sc, config)
			attrs["action"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
			unknown, err := tfprotov6.NewDynamicValue(sc.ValueType(), tftypes.NewValue(sc.ValueType(), attrs))
			if err != nil {
				t.Fatal(err)
			}
			blocked = lifecycleApply(t, s, kind, created.NewState, &unknown, &unknown)
			if !settingsErrors(blocked.Diagnostics) || len(a.posts) != before {
				t.Fatal("unknown action defaulted into write")
			}
			if kind == "ca" {
				a.key = ""
				fields["key_type"] = "2048"
				fields["description"] = "renamed"
				config = settingsValue(t, sc, fields)
				blocked = lifecycleApply(t, s, kind, created.NewState, lifecyclePlan(t, s, kind, sc, created.NewState, config), config)
				if !settingsErrors(blocked.Diagnostics) || len(a.posts) != before {
					t.Fatal("keyless CA metadata regenerated or wrote")
				}
				if !strings.Contains(blocked.Diagnostics[0].Detail, "26.7.8") {
					t.Fatal("missing focused upstream diagnosis")
				}
			}
		})
	}
}

func TestTrustLifecycleProtocolHydrationAndMismatchedSignedCSR(t *testing.T) {
	cert, key := selfSignedPEM(t, "example.test")
	a := &lifecycleAPI{kind: "cert", certificate: cert, key: key, csr: lifecycleCSR(t)}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas["opnsense_trust_cert"]
	fields := map[string]any{"description": "fixture", "common_name": "example.test", "action": "internal", "ca_ref": "issuer"}
	config := settingsValue(t, sc, fields)
	prior := settingsNull(t, sc)
	created := lifecycleApply(t, s, "cert", prior, lifecyclePlan(t, s, "cert", sc, prior, config), config)
	if settingsErrors(created.Diagnostics) {
		t.Fatal(created.Diagnostics)
	}
	fields["action"] = "reissue"
	config = settingsValue(t, sc, fields)
	reissued := lifecycleApply(t, s, "cert", created.NewState, lifecyclePlan(t, s, "cert", sc, created.NewState, config), config)
	if settingsErrors(reissued.Diagnostics) || a.posts[1]["prv_payload"] != key {
		t.Fatal("stored firewall key not hydrated", reissued.Diagnostics)
	}
	fields["action"] = "sign_csr"
	config = settingsValue(t, sc, fields)
	signed := lifecycleApply(t, s, "cert", reissued.NewState, lifecyclePlan(t, s, "cert", sc, reissued.NewState, config), config)
	if settingsErrors(signed.Diagnostics) || a.posts[2]["csr_payload"] != a.csr {
		t.Fatal("existing CSR not hydrated", signed.Diagnostics)
	}
	fields["action"] = "import_csr"
	fields["certificate"] = cert
	config = settingsValue(t, sc, fields)
	before := len(a.posts)
	blocked := lifecycleApply(t, s, "cert", signed.NewState, lifecyclePlan(t, s, "cert", sc, signed.NewState, config), config)
	if !settingsErrors(blocked.Diagnostics) || len(a.posts) != before {
		t.Fatal("mismatched signed CSR accepted")
	}
	// A local-only key is not on the appliance and must never be substituted.
	a.key = ""
	a.csr = ""
	fields["action"] = "reissue"
	fields["operation_token"] = "local-key-retry"
	delete(fields, "certificate")
	config = settingsValue(t, sc, fields)
	blocked = lifecycleApply(t, s, "cert", signed.NewState, lifecyclePlan(t, s, "cert", sc, signed.NewState, config), config)
	if !settingsErrors(blocked.Diagnostics) || len(a.posts) != before {
		t.Fatal("missing firewall key fabricated")
	}
}

func TestTrustLifecycleProtocolCAPartialCreateAndDelete(t *testing.T) {
	a := &lifecycleAPI{kind: "ca", badRead: `{"ca":{}}`}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas["opnsense_trust_ca"]
	fields := map[string]any{"description": "fixture", "common_name": "example.test"}
	config := settingsValue(t, sc, fields)
	prior := settingsNull(t, sc)
	failed := lifecycleApply(t, s, "ca", prior, lifecyclePlan(t, s, "ca", sc, prior, config), config)
	if settingsErrors(failed.Diagnostics) || len(failed.Diagnostics) == 0 || lifecycleString(t, sc, failed.NewState, "id") != "fixture-id" {
		t.Fatal("CA confirmed UUID lost")
	}
	a.badRead = ""
	read, err := s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_trust_ca", CurrentState: failed.NewState})
	if err != nil || settingsErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	deleted := lifecycleApply(t, s, "ca", read.NewState, settingsNull(t, sc), settingsNull(t, sc))
	if settingsErrors(deleted.Diagnostics) {
		t.Fatal(deleted.Diagnostics)
	}
}

func TestTrustLifecycleProtocolLegacyReplacementCompatibility(t *testing.T) {
	cert, key := selfSignedPEM(t, "example.test")
	otherCert, otherKey := selfSignedPEM(t, "other.example.test")
	for _, kind := range []string{"ca", "cert"} {
		t.Run(kind, func(t *testing.T) {
			a := &lifecycleAPI{kind: kind, certificate: cert, key: key}
			h := httptest.NewServer(a)
			defer h.Close()
			s, schemas := settingsServer(t, h.URL)
			sc := schemas.ResourceSchemas["opnsense_trust_"+kind]
			fields := map[string]any{"description": "fixture"}
			if kind == "ca" {
				fields["common_name"] = "example.test"
				fields["country"] = "US"
			} else {
				fields["certificate"] = cert
				fields["private_key"] = key
			}
			config := settingsValue(t, sc, fields)
			prior := settingsNull(t, sc)
			created := lifecycleApply(t, s, kind, prior, lifecyclePlan(t, s, kind, sc, prior, config), config)
			if settingsErrors(created.Diagnostics) {
				t.Fatal(created.Diagnostics)
			}
			if kind == "ca" {
				fields["common_name"] = "other.example.test"
			} else {
				fields["certificate"] = otherCert
				fields["private_key"] = otherKey
			}
			config = settingsValue(t, sc, fields)
			planned, err := s.PlanResourceChange(context.Background(), &tfprotov6.PlanResourceChangeRequest{TypeName: "opnsense_trust_" + kind, PriorState: created.NewState, ProposedNewState: config, Config: config})
			if err != nil || settingsErrors(planned.Diagnostics) || len(planned.RequiresReplace) == 0 {
				t.Fatal("legacy replacement behavior lost", err, planned.Diagnostics)
			}
		})
	}
}

func TestTrustLifecycleProtocolLegacyDescriptionOnlyCA(t *testing.T) {
	a := &lifecycleAPI{kind: "ca"}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas["opnsense_trust_ca"]
	config := settingsValue(t, sc, map[string]any{"description": "fixture"})
	prior := settingsNull(t, sc)
	result := lifecycleApply(t, s, "ca", prior, lifecyclePlan(t, s, "ca", sc, prior, config), config)
	if settingsErrors(result.Diagnostics) || len(a.posts) != 1 || a.posts[0]["action"] != "internal" {
		t.Fatal("legacy optional subject rejected", result.Diagnostics)
	}
	if _, sent := a.posts[0]["country"]; sent {
		t.Fatal("omitted country replaced model default")
	}
	if _, sent := a.posts[0]["commonname"]; sent {
		t.Fatal("omitted subject became empty configured write")
	}
}

func TestTrustLifecycleProtocolIntentRemovalNeverGenerates(t *testing.T) {
	cert, key := selfSignedPEM(t, "example.test")
	for _, kind := range []string{"ca", "cert"} {
		for _, change := range []string{"remove-action", "empty-action", "remove-token", "empty-token", "token-without-action"} {
			t.Run(kind+"/"+change, func(t *testing.T) {
				a := &lifecycleAPI{kind: kind, certificate: cert, key: key}
				h := httptest.NewServer(a)
				defer h.Close()
				s, schemas := settingsServer(t, h.URL)
				sc := schemas.ResourceSchemas["opnsense_trust_"+kind]
				fields := map[string]any{"description": "fixture", "action": "internal", "operation_token": "initial", "common_name": "example.test"}
				config := settingsValue(t, sc, fields)
				prior := settingsNull(t, sc)
				created := lifecycleApply(t, s, kind, prior, lifecyclePlan(t, s, kind, sc, prior, config), config)
				if settingsErrors(created.Diagnostics) {
					t.Fatal(created.Diagnostics)
				}
				switch change {
				case "remove-action":
					delete(fields, "action")
				case "empty-action":
					fields["action"] = ""
				case "remove-token":
					delete(fields, "operation_token")
				case "empty-token":
					fields["operation_token"] = ""
				case "token-without-action":
					delete(fields, "action")
					fields["operation_token"] = "new"
				}
				config = settingsValue(t, sc, fields)
				before := len(a.posts)
				result := lifecycleApply(t, s, kind, created.NewState, lifecyclePlan(t, s, kind, sc, created.NewState, config), config)
				if len(a.posts) != before {
					t.Fatal("intent removal/default regenerated or wrote")
				}
				if change == "token-without-action" && !settingsErrors(result.Diagnostics) {
					t.Fatal("token selected implicit default action")
				}
				// Explicit empty action is rejected on Create, not inferred as a default.
				fields["action"] = ""
				config = settingsValue(t, sc, fields)
				result = lifecycleApply(t, s, kind, prior, lifecyclePlan(t, s, kind, sc, prior, config), config)
				if !settingsErrors(result.Diagnostics) || len(a.posts) != before {
					t.Fatal("empty create action selected default")
				}
			})
		}
	}
}

func TestTrustLifecycleProtocolCAAuthoritativeLegacyDrift(t *testing.T) {
	cert, key := selfSignedPEM(t, "example.test")
	other, _ := selfSignedPEM(t, "other.example.test")
	a := &lifecycleAPI{kind: "ca", certificate: cert, key: key}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	sc := schemas.ResourceSchemas["opnsense_trust_ca"]
	fields := map[string]any{"description": "fixture", "action": "internal"}
	config := settingsValue(t, sc, fields)
	prior := settingsNull(t, sc)
	created := lifecycleApply(t, s, "ca", prior, lifecyclePlan(t, s, "ca", sc, prior, config), config)
	if settingsErrors(created.Diagnostics) {
		t.Fatal(created.Diagnostics)
	}
	a.certificate = other
	a.commonName = "other.example.test"
	a.country = "GB"
	read, err := s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: "opnsense_trust_ca", CurrentState: created.NewState})
	if err != nil || settingsErrors(read.Diagnostics) {
		t.Fatal(err, read.Diagnostics)
	}
	if lifecycleString(t, sc, read.NewState, "certificate") != other || lifecycleString(t, sc, read.NewState, "common_name") != "other.example.test" || lifecycleString(t, sc, read.NewState, "country") != "GB" {
		t.Fatal("legacy CA observations froze across external replacement")
	}
	a.certificate = cert // Simulate the newly generated authoritative PEM, not the prior observation.
	fields["operation_token"] = "regenerate"
	fields["common_name"] = "regenerated.example.test"
	fields["country"] = "DE"
	config = settingsValue(t, sc, fields)
	updated := lifecycleApply(t, s, "ca", read.NewState, lifecyclePlan(t, s, "ca", sc, read.NewState, config), config)
	if settingsErrors(updated.Diagnostics) || lifecycleString(t, sc, updated.NewState, "common_name") != "regenerated.example.test" || lifecycleString(t, sc, updated.NewState, "country") != "DE" || lifecycleString(t, sc, updated.NewState, "certificate") != cert {
		t.Fatal("regeneration retained stale legacy subject", updated.Diagnostics)
	}
}

func TestTrustLifecycleProtocolDeleteReconcilesOnlyAffirmativeAbsence(t *testing.T) {
	for _, kind := range []string{"ca", "cert"} {
		for _, reply := range []string{"disappeared", "lost-response", "ambiguous", "present", "auth"} {
			t.Run(kind+"/"+reply, func(t *testing.T) {
				a := &lifecycleAPI{kind: kind}
				deleted := false
				gets, posts := 0, 0
				h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.Contains(r.URL.Path, "/del/") {
						deleted = true
						posts++
						if reply == "lost-response" {
							conn, _, err := w.(http.Hijacker).Hijack()
							if err != nil {
								t.Error(err)
								return
							}
							_ = conn.Close()
							return
						}
						_, _ = fmt.Fprint(w, `{"result":"failed"}`)
						return
					}
					if deleted && strings.Contains(r.URL.Path, "/get/") {
						gets++
						switch reply {
						case "disappeared", "lost-response":
							_, _ = fmt.Fprintf(w, `{%q:[]}`, kind)
							return
						case "ambiguous":
							_, _ = fmt.Fprintf(w, `{%q:{}}`, kind)
							return
						case "auth":
							http.Error(w, "private", http.StatusUnauthorized)
							return
						}
					}
					a.ServeHTTP(w, r)
				}))
				defer h.Close()
				s, schemas := settingsServer(t, h.URL)
				sc := schemas.ResourceSchemas["opnsense_trust_"+kind]
				imported, err := s.ImportResourceState(context.Background(), &tfprotov6.ImportResourceStateRequest{TypeName: "opnsense_trust_" + kind, ID: "fixture-id"})
				if err != nil {
					t.Fatal(err)
				}
				result := lifecycleApply(t, s, kind, imported.ImportedResources[0].State, settingsNull(t, sc), settingsNull(t, sc))
				if ((reply == "disappeared" || reply == "lost-response") == settingsErrors(result.Diagnostics)) || gets != 1 || posts != 1 {
					t.Fatal("unconfirmed delete did not distinguish authoritative absence", reply, result.Diagnostics, gets)
				}
			})
		}
	}
}

func lifecycleCSRWithKey(t *testing.T, keyPEM string) string {
	t.Helper()
	block, _ := pem.Decode([]byte(keyPEM))
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{Subject: pkix.Name{CommonName: "example.test"}}, key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: raw}))
}
