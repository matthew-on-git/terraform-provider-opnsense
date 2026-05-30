// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccQuaggaBGPGlobal_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		// Singleton: no CheckDestroy — appliance config persists after destroy.
		Steps: []resource.TestStep{
			{
				Config: testAccQuaggaBGPGlobalConfig(65010),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_quagga_bgp_global.test", "id", "bgp"),
					resource.TestCheckResourceAttr("opnsense_quagga_bgp_global.test", "as_number", "65010"),
				),
			},
			{ResourceName: "opnsense_quagga_bgp_global.test", ImportState: true, ImportStateId: "bgp", ImportStateVerify: true},
			{
				Config: testAccQuaggaBGPGlobalConfig(65020),
				Check:  resource.TestCheckResourceAttr("opnsense_quagga_bgp_global.test", "as_number", "65020"),
			},
		},
	})
}

func testAccQuaggaBGPGlobalConfig(asn int) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_bgp_global" "test" {
  enabled   = true
  as_number = %d
  router_id = "10.0.0.1"
  networks  = ["10.0.0.0/24"]
}
`, asn)
}
