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

func TestAccQuaggaBGPPeerGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_quagga_bgp_peergroup", opnsense.ReqOpts{GetEndpoint: "/api/quagga/bgp/get_peergroup", Monad: "peergroup"}),
		Steps: []resource.TestStep{
			{
				Config: testAccQuaggaBGPPeerGroupConfig("spine"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_quagga_bgp_peergroup.test", "id"),
					resource.TestCheckResourceAttr("opnsense_quagga_bgp_peergroup.test", "name", "spine"),
				),
			},
			{ResourceName: "opnsense_quagga_bgp_peergroup.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccQuaggaBGPPeerGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_bgp_peergroup" "test" {
  name           = %[1]q
  remote_as_mode = "external"
  family         = ["IPv4"]
}
`, name)
}
