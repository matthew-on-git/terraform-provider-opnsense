// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dhcp_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccDHCPv4Subnet_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_dhcpv4_subnet", opnsense.ReqOpts{GetEndpoint: "/api/kea/dhcpv4/get_subnet", Monad: "subnet"}),
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

func testAccDHCPv4SubnetConfig(subnet string) string {
	return fmt.Sprintf(`
resource "opnsense_dhcpv4_subnet" "test" {
  subnet = %[1]q
}
`, subnet)
}
