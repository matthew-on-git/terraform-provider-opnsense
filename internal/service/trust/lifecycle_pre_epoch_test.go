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
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Expired imports are parseable X.509 objects. Released Store::parseX509
// exposes their signed Unix validity timestamps without an epoch restriction.
func TestTrustLifecyclePreEpochValidityReads(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "example.test", Country: []string{"US"}},
		NotBefore:             time.Date(1965, time.January, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:              time.Date(1966, time.January, 1, 0, 0, 0, 0, time.UTC),
		BasicConstraintsValid: true,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.NotBefore.Equal(template.NotBefore) || !parsed.NotAfter.Equal(template.NotAfter) || parsed.NotBefore.Unix() >= 0 || parsed.NotAfter.Unix() >= 0 {
		t.Fatal("certificate did not retain genuine pre-epoch validity")
	}
	if err := parsed.CheckSignatureFrom(parsed); err != nil {
		t.Fatal("invalid self-signed certificate", err)
	}
	certificate := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	privateKey := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	from, to := parsed.NotBefore.Unix(), parsed.NotAfter.Unix()
	wantFrom, wantTo := strconv.FormatInt(from, 10), strconv.FormatInt(to, 10)
	for _, kind := range []string{"ca", "cert"} {
		for _, encoding := range []string{"numeric", "string"} {
			t.Run(kind+"/"+encoding, func(t *testing.T) {
				var wireFrom, wireTo any = from, to
				if encoding == "string" {
					wireFrom, wireTo = wantFrom, wantTo
				}
				body, err := json.Marshal(map[string]any{kind: map[string]any{
					"refid": "0123456789abc", "descr": "fixture", "crt_payload": certificate,
					"prv_payload": privateKey, "valid_from": wireFrom, "valid_to": wireTo,
				}})
				if err != nil {
					t.Fatal(err)
				}
				a := &lifecycleAPI{kind: kind, badRead: string(body)}
				var writes atomic.Int64
				h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet {
						writes.Add(1)
					}
					a.ServeHTTP(w, r)
				}))
				defer h.Close()
				s, schemas := settingsServer(t, h.URL)
				typ := "opnsense_trust_" + kind
				resourceSchema, lookupSchema := schemas.ResourceSchemas[typ], schemas.DataSourceSchemas[typ]
				imported, err := s.ImportResourceState(context.Background(), &tfprotov6.ImportResourceStateRequest{TypeName: typ, ID: "fixture-id"})
				if err != nil || settingsErrors(imported.Diagnostics) || len(imported.ImportedResources) != 1 {
					t.Fatal("import failed", err, imported)
				}
				read, err := s.ReadResource(context.Background(), &tfprotov6.ReadResourceRequest{TypeName: typ, CurrentState: imported.ImportedResources[0].State})
				if err != nil || settingsErrors(read.Diagnostics) {
					t.Fatal("pre-epoch resource read failed", err, read)
				}
				lookup, err := s.ReadDataSource(context.Background(), &tfprotov6.ReadDataSourceRequest{TypeName: typ, Config: settingsValue(t, lookupSchema, map[string]any{"id": "fixture-id"})})
				if err != nil || settingsErrors(lookup.Diagnostics) {
					t.Fatal("pre-epoch lookup failed", err, lookup)
				}
				for _, result := range []struct {
					schema *tfprotov6.Schema
					state  *tfprotov6.DynamicValue
					pem    string
				}{
					{resourceSchema, read.NewState, "returned_certificate"},
					{lookupSchema, lookup.State, "certificate"},
				} {
					if lifecycleString(t, result.schema, result.state, "valid_from") != wantFrom || lifecycleString(t, result.schema, result.state, "valid_to") != wantTo || lifecycleString(t, result.schema, result.state, result.pem) != certificate {
						t.Fatal("pre-epoch validity or certificate changed")
					}
				}
				a.mu.Lock()
				reads := a.reads
				a.mu.Unlock()
				if reads != 2 || writes.Load() != 0 {
					t.Fatalf("read-only operations changed: reads=%d writes=%d", reads, writes.Load())
				}
			})
		}
	}
}
