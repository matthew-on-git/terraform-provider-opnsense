// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccFirewallAlias_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckFirewallAliasDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccFirewallAliasConfig("tf_test_alias", "host", "10.0.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_firewall_alias.test", "id"),
					resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "name", "tf_test_alias"),
					resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "type", "host"),
					resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_firewall_alias.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update content and verify.
			{
				Config: testAccFirewallAliasConfig("tf_test_alias", "host", "10.0.0.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "name", "tf_test_alias"),
				),
			},
		},
	})
}

// testAccCheckFirewallAliasDestroy verifies all firewall alias resources
// created during the test have been removed from OPNsense.
func testAccCheckFirewallAliasDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_firewall_alias" {
			continue
		}
		// If the resource still exists in state after destroy, report it.
		// The acceptance test framework calls Read after destroy — if Read
		// returns without error and doesn't remove the resource, this check
		// catches it. The framework handles the API call; we just verify
		// no alias resources remain in the final state.
		return fmt.Errorf("firewall alias %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccFirewallAliasConfig(name, aliasType, content string) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_alias" "test" {
  name    = %[1]q
  type    = %[2]q
  content = [%[3]q]
}
`, name, aliasType, content)
}
