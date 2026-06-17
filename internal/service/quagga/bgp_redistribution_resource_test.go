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

func TestAccQuaggaBGPRedistribution_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { preCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_quagga_bgp_redistribution", opnsense.ReqOpts{GetEndpoint: "/api/quagga/bgp/get_redistribution", Monad: "redistribution"}),
		Steps: []resource.TestStep{
			{
				Config: testAccQuaggaBGPRedistributionConfig("connected"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_quagga_bgp_redistribution.test", "id"),
					resource.TestCheckResourceAttr("opnsense_quagga_bgp_redistribution.test", "redistribute", "connected"),
				),
			},
			{ResourceName: "opnsense_quagga_bgp_redistribution.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccQuaggaBGPRedistributionConfig(src string) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_bgp_redistribution" "test" {
  redistribute = %[1]q
}
`, src)
}
