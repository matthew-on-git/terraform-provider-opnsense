// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccQuaggaPrefixList_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckQuaggaPrefixListDestroy,
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

func testAccCheckQuaggaPrefixListDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_quagga_prefix_list" {
			continue
		}
		return fmt.Errorf("prefix list %s still exists", rs.Primary.ID)
	}
	return nil
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
