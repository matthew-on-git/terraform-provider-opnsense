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

func TestAccFirewallFilterRule_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckFirewallFilterRuleDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccFirewallFilterRuleConfig("pass", "in", "443"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_firewall_filter_rule.test", "id"),
					resource.TestCheckResourceAttr("opnsense_firewall_filter_rule.test", "action", "pass"),
					resource.TestCheckResourceAttr("opnsense_firewall_filter_rule.test", "direction", "in"),
					resource.TestCheckResourceAttr("opnsense_firewall_filter_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_firewall_filter_rule.test", "protocol", "tcp"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_firewall_filter_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update to block and verify.
			{
				Config: testAccFirewallFilterRuleConfig("block", "in", "443"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_filter_rule.test", "action", "block"),
				),
			},
		},
	})
}

func testAccCheckFirewallFilterRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_firewall_filter_rule" {
			continue
		}
		return fmt.Errorf("firewall filter rule %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccFirewallFilterRuleConfig(action, direction, destPort string) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_filter_rule" "test" {
  action           = %[1]q
  direction        = %[2]q
  protocol         = "tcp"
  destination_port = %[3]q
  description      = "Terraform test rule"
}
`, action, direction, destPort)
}
