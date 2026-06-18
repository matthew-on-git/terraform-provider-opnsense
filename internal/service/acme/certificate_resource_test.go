// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccAcmeCertificate_basic(t *testing.T) {
	if os.Getenv("OPNSENSE_ACME_ISSUE") != "1" {
		t.Skip("set OPNSENSE_ACME_ISSUE=1 with OPNSENSE_ACME_CERT_DOMAIN, OPNSENSE_ACME_ACCOUNT_UUID, and OPNSENSE_ACME_VALIDATION_UUID to run ACME issuance acceptance")
	}
	domain := os.Getenv("OPNSENSE_ACME_CERT_DOMAIN")
	account := os.Getenv("OPNSENSE_ACME_ACCOUNT_UUID")
	validationMethod := os.Getenv("OPNSENSE_ACME_VALIDATION_UUID")
	if domain == "" || account == "" || validationMethod == "" {
		t.Skip("OPNSENSE_ACME_CERT_DOMAIN, OPNSENSE_ACME_ACCOUNT_UUID, and OPNSENSE_ACME_VALIDATION_UUID are required for ACME issuance acceptance")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_acme_certificate", opnsense.ReqOpts{GetEndpoint: "/api/acmeclient/certificates/get", Monad: "certificate"}),
		Steps: []resource.TestStep{
			{
				Config: testAccAcmeCertificateConfig(domain, account, validationMethod),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_acme_certificate.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_acme_certificate.test", "cert_ref_id"),
					resource.TestCheckResourceAttr("opnsense_acme_certificate.test", "status_code", "200"),
					resource.TestCheckResourceAttr("opnsense_acme_certificate.test", "name", domain),
				),
			},
			{ResourceName: "opnsense_acme_certificate.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccAcmeCertificateConfig(domain, account, validationMethod string) string {
	return fmt.Sprintf(`
resource "opnsense_acme_certificate" "test" {
  name              = %[1]q
  account           = %[2]q
  validation_method = %[3]q
}
`, domain, account, validationMethod)
}
