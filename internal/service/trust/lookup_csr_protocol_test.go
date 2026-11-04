// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestTrustCertificateLookupCSRPresentAndAbsent(t *testing.T) {
	csr := lifecycleCSR(t)
	a := &lifecycleAPI{kind: "cert", csr: csr}
	h := httptest.NewServer(a)
	defer h.Close()
	s, schemas := settingsServer(t, h.URL)
	const typ = "opnsense_trust_cert"
	schema := schemas.DataSourceSchemas[typ]
	config := settingsValue(t, schema, map[string]any{"id": "fixture-id"})

	for _, present := range []bool{true, false} {
		name := "present"
		if !present {
			name = "absent"
			a.mu.Lock()
			a.csr = ""
			a.mu.Unlock()
		}
		t.Run(name, func(t *testing.T) {
			result, err := s.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: typ, Config: config})
			if err != nil {
				t.Fatal(err)
			}
			if settingsErrors(result.Diagnostics) {
				t.Fatalf("lookup diagnostics: %+v", result.Diagnostics)
			}
			value := settingsAttrs(t, schema, result.State)["csr"]
			if present {
				if value.IsNull() || !value.IsKnown() || lifecycleString(t, schema, result.State, "csr") != csr {
					t.Fatal("lookup did not return the exact existing CSR")
				}
			} else if !value.IsNull() {
				t.Fatal("lookup fabricated a CSR when absent")
			}
		})
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.reads != 2 || len(a.posts) != 0 {
		t.Fatalf("lookup must only read: reads=%d writes=%d", a.reads, len(a.posts))
	}
}
