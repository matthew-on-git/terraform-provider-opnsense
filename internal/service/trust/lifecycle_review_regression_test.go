// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestTrustLifecycleReviewStrictEditorFields(t *testing.T) {
	for _, kind := range []string{"ca", "cert"} {
		t.Run(kind, func(t *testing.T) {
			a := &lifecycleAPI{kind: kind}
			h := httptest.NewServer(a)
			defer h.Close()
			server, schemas := settingsServer(t, h.URL)
			schema := schemas.DataSourceSchemas["opnsense_trust_"+kind]
			base := map[string]any{"refid": "0123456789abc", "descr": "fixture", "crt_payload": "", "prv_payload": "", "csr_payload": "", "valid_from": "1793664000", "valid_to": 1825200000}
			for _, field := range []string{"descr", "crt_payload", "prv_payload", "csr_payload", "valid_from", "valid_to"} {
				for _, mode := range []string{"null", "missing", "malformed"} {
					if mode == "missing" && field != "descr" && field != "crt_payload" && field != "prv_payload" {
						continue
					}
					t.Run(field+"-"+mode, func(t *testing.T) {
						record := make(map[string]any)
						for k, v := range base {
							record[k] = v
						}
						switch mode {
						case "null":
							record[field] = nil
						case "missing":
							delete(record, field)
						case "malformed":
							record[field] = map[string]any{}
						}
						if mode == "malformed" && (field == "valid_from" || field == "valid_to") {
							record[field] = "not-a-timestamp"
						}
						body, err := json.Marshal(map[string]any{kind: record})
						if err != nil {
							t.Fatal(err)
						}
						a.mu.Lock()
						a.badRead = string(body)
						a.mu.Unlock()
						got, err := server.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "opnsense_trust_" + kind, Config: settingsValue(t, schema, map[string]any{"id": "fixture-id"})})
						if err != nil || !settingsErrors(got.Diagnostics) {
							t.Fatal("invalid editor accepted", field, mode, err, got.Diagnostics)
						}
					})
				}
			}
		})
	}
}

func TestTrustLifecycleReviewRejectsCompoundOptionsAndWrongCSRLabel(t *testing.T) {
	csr := lifecycleCSR(t)
	cases := []map[string]any{
		{"action": "internal|external"},
		{"digest": "sha256|sha384"},
		{"key_type": "2048|3072"},
		{"cert_type": "usr_cert|server_cert"},
		{"private_key_location": "firewall|local"},
		{"action": "sign_csr", "ca_ref": "issuer", "csr": strings.ReplaceAll(csr, "CERTIFICATE REQUEST", "CERTIFICATE")},
	}
	for _, change := range cases {
		a := &lifecycleAPI{kind: "cert"}
		h := httptest.NewServer(a)
		s, schemas := settingsServer(t, h.URL)
		sc := schemas.ResourceSchemas["opnsense_trust_cert"]
		fields := map[string]any{"description": "fixture", "action": "internal"}
		for k, v := range change {
			fields[k] = v
		}
		config := settingsValue(t, sc, fields)
		prior := settingsNull(t, sc)
		got := lifecycleApply(t, s, "cert", prior, lifecyclePlan(t, s, "cert", sc, prior, config), config)
		a.mu.Lock()
		writes := len(a.posts)
		a.mu.Unlock()
		h.Close()
		if !settingsErrors(got.Diagnostics) || writes != 0 {
			t.Fatal("invalid option reached mutation", change, got.Diagnostics, writes)
		}
	}
}

func TestTrustLifecycleReviewRejectsPathIdentitiesBeforeAPI(t *testing.T) {
	for _, kind := range []string{"ca", "cert"} {
		a := &lifecycleAPI{kind: kind}
		h := httptest.NewServer(a)
		s, schemas := settingsServer(t, h.URL)
		sc := schemas.DataSourceSchemas["opnsense_trust_"+kind]
		for _, id := range []string{"../other", "object?query=other", "object#fragment", "object%2Fother", "object\\path", ""} {
			imported, err := s.ImportResourceState(context.Background(), &tfprotov6.ImportResourceStateRequest{TypeName: "opnsense_trust_" + kind, ID: id})
			if err != nil || !settingsErrors(imported.Diagnostics) {
				t.Fatal("unsafe import accepted", id, err)
			}
			got, err := s.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: "opnsense_trust_" + kind, Config: settingsValue(t, sc, map[string]any{"id": id})})
			if err != nil || !settingsErrors(got.Diagnostics) {
				t.Fatal("unsafe lookup accepted", id, err)
			}
		}
		a.mu.Lock()
		reads, writes := a.reads, len(a.posts)
		a.mu.Unlock()
		h.Close()
		if reads != 0 || writes != 0 {
			t.Fatal("unsafe identity reached appliance", reads, writes)
		}
	}
}
