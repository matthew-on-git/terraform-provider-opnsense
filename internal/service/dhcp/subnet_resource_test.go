// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dhcp_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccDHCPv4Subnet_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckDHCPv4SubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPv4SubnetConfig("10.99.0.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_dhcpv4_subnet.test", "id"),
					resource.TestCheckResourceAttr("opnsense_dhcpv4_subnet.test", "subnet", "10.99.0.0/24"),
				),
			},
			{ResourceName: "opnsense_dhcpv4_subnet.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccCheckDHCPv4SubnetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_dhcpv4_subnet" {
			continue
		}
		return fmt.Errorf("DHCPv4 subnet %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDHCPv4SubnetConfig(subnet string) string {
	return fmt.Sprintf(`
resource "opnsense_dhcpv4_subnet" "test" {
  subnet = %[1]q
}
`, subnet)
}
