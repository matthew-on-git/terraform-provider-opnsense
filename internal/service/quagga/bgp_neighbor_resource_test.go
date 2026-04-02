// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccQuaggaBGPNeighbor_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_quagga_bgp_neighbor", opnsense.ReqOpts{GetEndpoint: "/api/quagga/bgp/get_neighbor", Monad: "neighbor"}),
		Steps: []resource.TestStep{
			{
				Config: testAccQuaggaBGPNeighborConfig("10.0.0.2", 65001),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_quagga_bgp_neighbor.test", "id"),
					resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.test", "address", "10.0.0.2"),
					resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.test", "remote_as", "65001"),
				),
			},
			{ResourceName: "opnsense_quagga_bgp_neighbor.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccQuaggaBGPNeighborConfig("10.0.0.3", 65001),
				Check:  resource.TestCheckResourceAttr("opnsense_quagga_bgp_neighbor.test", "address", "10.0.0.3"),
			},
		},
	})
}

func testAccCheckQuaggaBGPNeighborDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_quagga_bgp_neighbor" {
			continue
		}
		return fmt.Errorf("BGP neighbor %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccQuaggaBGPNeighborConfig(addr string, remoteAS int) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_bgp_neighbor" "test" {
  address   = %[1]q
  remote_as = %[2]d
}
`, addr, remoteAS)
}
