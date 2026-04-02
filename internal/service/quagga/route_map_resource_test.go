// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccQuaggaRouteMap_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_quagga_route_map", opnsense.ReqOpts{GetEndpoint: "/api/quagga/bgp/get_routemap", Monad: "routemap"}),
		Steps: []resource.TestStep{
			{
				Config: testAccQuaggaRouteMapConfig("permit"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_quagga_route_map.test", "id"),
					resource.TestCheckResourceAttr("opnsense_quagga_route_map.test", "action", "permit"),
				),
			},
			{ResourceName: "opnsense_quagga_route_map.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccQuaggaRouteMapConfig("deny"),
				Check:  resource.TestCheckResourceAttr("opnsense_quagga_route_map.test", "action", "deny"),
			},
		},
	})
}

func testAccQuaggaRouteMapConfig(action string) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_route_map" "test" {
  name   = "tf_test_rmap"
  action = %[1]q
  order  = 10
}
`, action)
}
