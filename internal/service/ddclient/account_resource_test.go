// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccDDClientAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_ddclient_account", opnsense.ReqOpts{GetEndpoint: "/api/dyndns/accounts/get_item", Monad: "account"}),
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccDDClientAccountConfig("cloudflare", "test.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_ddclient_account.test", "id"),
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "service", "cloudflare"),
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "hostnames", "test.example.com"),
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:            "opnsense_ddclient_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Step 3: Update hostnames and verify.
			{
				Config: testAccDDClientAccountConfig("cloudflare", "updated.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "service", "cloudflare"),
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "hostnames", "updated.example.com"),
				),
			},
		},
	})
}

// testAccCheckDDClientAccountDestroy verifies all ddclient account resources
// created during the test have been removed from OPNsense.
func testAccCheckDDClientAccountDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_ddclient_account" {
			continue
		}
		return fmt.Errorf("ddclient account %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDDClientAccountConfig(service, hostnames string) string {
	return fmt.Sprintf(`
resource "opnsense_ddclient_account" "test" {
  service   = %[1]q
  hostnames = %[2]q
  username  = "testuser"
  password  = "testpass"
}
`, service, hostnames)
}
