// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccAcmeAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckAcmeAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmeAccountConfig("tf_test_acme", "letsencrypt_test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_acme_account.test", "id"),
					resource.TestCheckResourceAttr("opnsense_acme_account.test", "name", "tf_test_acme"),
					resource.TestCheckResourceAttr("opnsense_acme_account.test", "ca", "letsencrypt_test"),
				),
			},
			{ResourceName: "opnsense_acme_account.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccCheckAcmeAccountDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_acme_account" {
			continue
		}
		return fmt.Errorf("ACME account %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccAcmeAccountConfig(name, ca string) string {
	return fmt.Sprintf(`
resource "opnsense_acme_account" "test" {
  name  = %[1]q
  ca    = %[2]q
  email = "test@example.com"
}
`, name, ca)
}
