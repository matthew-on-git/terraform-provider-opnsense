// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// rsaKeyPairPEM generates a throwaway RSA key pair as PEM strings (public in
// PKIX/SPKI form, private in PKCS#8 form) for use in the acceptance test.
func rsaKeyPairPEM(t *testing.T) (pubPEM, privPEM string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %s", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %s", err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal private key: %s", err)
	}
	pubPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))
	privPEM = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}))
	return pubPEM, privPEM
}

func TestAccIPsecKeyPair_basic(t *testing.T) {
	pubPEM, privPEM := rsaKeyPairPEM(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_ipsec_key_pair", opnsense.ReqOpts{GetEndpoint: "/api/ipsec/key_pairs/getItem", Monad: "keyPair"}),
		Steps: []resource.TestStep{
			// Step 1: Create the key pair and verify.
			{
				Config: testAccIPsecKeyPairConfig("tfacckp", pubPEM, privPEM),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_ipsec_key_pair.test", "id"),
					resource.TestCheckResourceAttr("opnsense_ipsec_key_pair.test", "name", "tfacckp"),
					resource.TestCheckResourceAttr("opnsense_ipsec_key_pair.test", "key_type", "rsa"),
				),
			},
			// Step 2: Import and verify (the PEM payloads are write-only).
			{
				ResourceName:            "opnsense_ipsec_key_pair.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"public_key", "private_key"},
			},
		},
	})
}

func testAccIPsecKeyPairConfig(name, pubPEM, privPEM string) string {
	return fmt.Sprintf(`
resource "opnsense_ipsec_key_pair" "test" {
  name        = %[1]q
  key_type    = "rsa"
  public_key  = %[2]q
  private_key = %[3]q
}
`, name, pubPEM, privPEM)
}
