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

func TestAccQuaggaPrefixList_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_quagga_prefix_list", opnsense.ReqOpts{GetEndpoint: "/api/quagga/bgp/get_prefixlist", Monad: "prefixlist"}),
		Steps: []resource.TestStep{
			{
				Config: testAccQuaggaPrefixListConfig("10.0.0.0/8"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_quagga_prefix_list.test", "id"),
					resource.TestCheckResourceAttr("opnsense_quagga_prefix_list.test", "action", "permit"),
					resource.TestCheckResourceAttr("opnsense_quagga_prefix_list.test", "network", "10.0.0.0/8"),
				),
			},
			{ResourceName: "opnsense_quagga_prefix_list.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccQuaggaPrefixListConfig("172.16.0.0/12"),
				Check:  resource.TestCheckResourceAttr("opnsense_quagga_prefix_list.test", "network", "172.16.0.0/12"),
			},
		},
	})
}

func testAccQuaggaPrefixListConfig(network string) string {
	return fmt.Sprintf(`
resource "opnsense_quagga_prefix_list" "test" {
  name     = "tf_test_pfx"
  sequence = 10
  action   = "permit"
  network  = %[1]q
}
`, network)
}
