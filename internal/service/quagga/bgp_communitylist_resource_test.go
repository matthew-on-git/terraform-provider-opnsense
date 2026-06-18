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

func TestAccQuaggaBGPCommunityList_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { preCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_quagga_bgp_communitylist", opnsense.ReqOpts{GetEndpoint: "/api/quagga/bgp/get_communitylist", Monad: "communitylist"}),
		Steps: []resource.TestStep{
			{
				Config: testAccQuaggaBGPCommunityListConfig(10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_quagga_bgp_communitylist.test", "id"),
					resource.TestCheckResourceAttr("opnsense_quagga_bgp_communitylist.test", "seq_number", "10"),
				),
			},
			{ResourceName: "opnsense_quagga_bgp_communitylist.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccQuaggaBGPCommunityListConfig(seq int) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_bgp_communitylist" "test" {
  number     = 100
  seq_number = %d
  action     = "permit"
  community  = "65010:100"
}
`, seq)
}
