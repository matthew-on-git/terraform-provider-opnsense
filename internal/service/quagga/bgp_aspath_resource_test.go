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

func TestAccQuaggaBGPASPath_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_quagga_bgp_aspath", opnsense.ReqOpts{GetEndpoint: "/api/quagga/bgp/get_aspath", Monad: "aspath"}),
		Steps: []resource.TestStep{
			{
				Config: testAccQuaggaBGPASPathConfig(100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_quagga_bgp_aspath.test", "id"),
					resource.TestCheckResourceAttr("opnsense_quagga_bgp_aspath.test", "number", "100"),
				),
			},
			{ResourceName: "opnsense_quagga_bgp_aspath.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccQuaggaBGPASPathConfig(number int) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_bgp_aspath" "test" {
  number     = %d
  action     = "permit"
  as_pattern = "_65010_"
}
`, number)
}
