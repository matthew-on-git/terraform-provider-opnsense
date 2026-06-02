// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// selfSignedPEM generates a throwaway self-signed certificate and its private
// key, returned as PEM strings, for use in the acceptance test.
func selfSignedPEM(t *testing.T, cn string) (certPEM, keyPEM string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %s", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn, Organization: []string{"tfacc"}},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * 365 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %s", err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %s", err)
	}
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}))
	return certPEM, keyPEM
}

func TestAccTrustCert_basic(t *testing.T) {
	certPEM, keyPEM := selfSignedPEM(t, "tf-acc-cert")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_trust_cert", opnsense.ReqOpts{GetEndpoint: "/api/trust/cert/get", Monad: "cert"}),
		Steps: []resource.TestStep{
			// Step 1: Import the certificate and verify.
			{
				Config: testAccTrustCertConfig("tf-acc-cert", certPEM, keyPEM),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_trust_cert.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_trust_cert.test", "refid"),
					resource.TestCheckResourceAttr("opnsense_trust_cert.test", "description", "tf-acc-cert"),
				),
			},
			// Step 2: Import and verify (the PEM payloads are write-only).
			{
				ResourceName:            "opnsense_trust_cert.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate", "private_key"},
			},
			// Step 3: Update the description and verify.
			{
				Config: testAccTrustCertConfig("tf-acc-cert-2", certPEM, keyPEM),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_trust_cert.test", "description", "tf-acc-cert-2"),
				),
			},
		},
	})
}

func testAccTrustCertConfig(descr, certPEM, keyPEM string) string {
	return fmt.Sprintf(`
resource "opnsense_trust_cert" "test" {
  description = %[1]q
  certificate = %[2]q
  private_key = %[3]q
}
`, descr, certPEM, keyPEM)
}
