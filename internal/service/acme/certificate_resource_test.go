// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccAcmeCertificate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_acme_certificate", opnsense.ReqOpts{GetEndpoint: "/api/acmeclient/certificates/get", Monad: "certificate"}),
		Steps: []resource.TestStep{
			{
				Config: testAccAcmeCertificateConfig("test.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_acme_certificate.test", "id"),
					resource.TestCheckResourceAttr("opnsense_acme_certificate.test", "name", "test.example.com"),
				),
			},
			{ResourceName: "opnsense_acme_certificate.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccCheckAcmeCertificateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_acme_certificate" {
			continue
		}
		return fmt.Errorf("ACME certificate %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccAcmeCertificateConfig(domain string) string {
	return fmt.Sprintf(`
resource "opnsense_acme_account" "test" {
  name  = "tf_test_cert_acct"
  ca    = "letsencrypt_test"
  email = "test@example.com"
}

resource "opnsense_acme_challenge" "test" {
  name   = "tf_test_cert_challenge"
  method = "http01"
}

resource "opnsense_acme_certificate" "test" {
  name              = %[1]q
  account           = opnsense_acme_account.test.id
  validation_method = opnsense_acme_challenge.test.id
}
`, domain)
}
