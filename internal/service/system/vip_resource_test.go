// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccSystemVip_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckSystemVipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemVipConfig("10.99.99.1", 24),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_system_vip.test", "id"),
					resource.TestCheckResourceAttr("opnsense_system_vip.test", "address", "10.99.99.1"),
				),
			},
			{ResourceName: "opnsense_system_vip.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccSystemVipConfig("10.99.99.2", 24),
				Check:  resource.TestCheckResourceAttr("opnsense_system_vip.test", "address", "10.99.99.2"),
			},
		},
	})
}

func testAccCheckSystemVipDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_system_vip" {
			continue
		}
		return fmt.Errorf("VIP %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSystemVipConfig(addr string, bits int) string {
	return fmt.Sprintf(`
resource "opnsense_system_vip" "test" {
  interface   = "lan"
  mode        = "ipalias"
  address     = %[1]q
  subnet_bits = %[2]d
}
`, addr, bits)
}
