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

func TestAccFirewallNatPortForward_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckFirewallNatPortForwardDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccFirewallNatPortForwardConfig("10.0.0.10", "443", "443"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_firewall_nat_port_forward.test", "id"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat_port_forward.test", "target", "10.0.0.10"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat_port_forward.test", "destination_port", "443"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat_port_forward.test", "local_port", "443"),
					resource.TestCheckResourceAttr("opnsense_firewall_nat_port_forward.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_firewall_nat_port_forward.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update target port and verify.
			{
				Config: testAccFirewallNatPortForwardConfig("10.0.0.10", "443", "8443"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_nat_port_forward.test", "local_port", "8443"),
				),
			},
		},
	})
}

func testAccCheckFirewallNatPortForwardDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_firewall_nat_port_forward" {
			continue
		}
		return fmt.Errorf("NAT port-forward rule %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccFirewallNatPortForwardConfig(target, destPort, localPort string) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_nat_port_forward" "test" {
  interface        = "wan"
  protocol         = "tcp"
  destination_net  = "wanip"
  destination_port = %[2]q
  target           = %[1]q
  local_port       = %[3]q
  description      = "Terraform test NAT rule"
}
`, target, destPort, localPort)
}
