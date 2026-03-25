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

func TestAccFirewallNatOutbound_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckFirewallNatOutboundDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccFirewallNatOutboundConfig("10.0.0.0/24", "wanip"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_firewall_nat_outbound.test", "id"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat_outbound.test", "source_net", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat_outbound.test", "target", "wanip"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat_outbound.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat_outbound.test", "interface", "wan"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_firewall_nat_outbound.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update source and verify.
			{
				Config: testAccFirewallNatOutboundConfig("10.1.0.0/24", "wanip"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_nat_outbound.test", "source_net", "10.1.0.0/24"),
				),
			},
		},
	})
}

func testAccCheckFirewallNatOutboundDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_firewall_nat_outbound" {
			continue
		}
		return fmt.Errorf("outbound NAT rule %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccFirewallNatOutboundConfig(sourceNet, target string) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_nat_outbound" "test" {
  interface   = "wan"
  source_net  = %[1]q
  target      = %[2]q
  description = "Terraform test SNAT rule"
}
`, sourceNet, target)
}
