// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccKeaDHCPv6Subnet_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_kea_dhcpv6_subnet", opnsense.ReqOpts{GetEndpoint: "/api/kea/dhcpv6/get_subnet", Monad: "subnet6"}),
		Steps: []resource.TestStep{
			// Step 1: Enable the interface in DHCPv6 general settings, then create a subnet on it.
			{
				Config: testAccKeaDHCPv6SubnetConfig("2001:db8::/64"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_kea_dhcpv6_subnet.test", "id"),
					resource.TestCheckResourceAttr("opnsense_kea_dhcpv6_subnet.test", "subnet", "2001:db8::/64"),
					resource.TestCheckResourceAttr("opnsense_kea_dhcpv6_subnet.test", "interface", "lan"),
				),
			},
			// Step 2: Import and verify.
			{
				ResourceName:      "opnsense_kea_dhcpv6_subnet.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update the subnet description and verify.
			{
				Config: testAccKeaDHCPv6SubnetConfigDescr("2001:db8::/64", "tf-acc-subnet"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_kea_dhcpv6_subnet.test", "description", "tf-acc-subnet"),
				),
			},
		},
	})
}

func testAccKeaDHCPv6SubnetConfig(subnet string) string {
	return fmt.Sprintf(`
resource "opnsense_kea_dhcpv6_settings" "s" {
  enabled    = true
  interfaces = ["lan"]
}

resource "opnsense_kea_dhcpv6_subnet" "test" {
  subnet     = %[1]q
  interface  = "lan"
  depends_on = [opnsense_kea_dhcpv6_settings.s]
}
`, subnet)
}

func testAccKeaDHCPv6SubnetConfigDescr(subnet, descr string) string {
	return fmt.Sprintf(`
resource "opnsense_kea_dhcpv6_settings" "s" {
  enabled    = true
  interfaces = ["lan"]
}

resource "opnsense_kea_dhcpv6_subnet" "test" {
  subnet      = %[1]q
  interface   = "lan"
  description = %[2]q
  depends_on  = [opnsense_kea_dhcpv6_settings.s]
}
`, subnet, descr)
}
