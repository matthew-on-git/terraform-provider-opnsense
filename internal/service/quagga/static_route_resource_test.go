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

func TestAccQuaggaStaticRoute_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { preCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_quagga_static_route", opnsense.ReqOpts{GetEndpoint: "/api/quagga/static/get_route", Monad: "route"}),
		Steps: []resource.TestStep{
			{
				Config: testAccQuaggaStaticRouteConfig("192.0.2.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_quagga_static_route.test", "id"),
					resource.TestCheckResourceAttr("opnsense_quagga_static_route.test", "network", "192.0.2.0/24"),
				),
			},
			{ResourceName: "opnsense_quagga_static_route.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccQuaggaStaticRouteConfig("198.51.100.0/24"),
				Check:  resource.TestCheckResourceAttr("opnsense_quagga_static_route.test", "network", "198.51.100.0/24"),
			},
		},
	})
}

func testAccQuaggaStaticRouteConfig(network string) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_static_route" "test" {
  network = %[1]q
  gateway = "10.0.0.254"
}
`, network)
}
