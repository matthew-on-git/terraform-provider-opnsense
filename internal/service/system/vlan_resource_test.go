// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccSystemVlan_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_system_vlan", opnsense.ReqOpts{GetEndpoint: "/api/interfaces/vlan_settings/get_item", Monad: "vlan"}),
		Steps: []resource.TestStep{
			{
				Config: testAccSystemVlanConfig(100),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_system_vlan.test", "id"),
					resource.TestCheckResourceAttr("opnsense_system_vlan.test", "tag", "100"),
				),
			},
			{ResourceName: "opnsense_system_vlan.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccSystemVlanConfig(200),
				Check:  resource.TestCheckResourceAttr("opnsense_system_vlan.test", "tag", "200"),
			},
		},
	})
}

func testAccCheckSystemVlanDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_system_vlan" {
			continue
		}
		return fmt.Errorf("VLAN %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSystemVlanConfig(tag int) string {
	return fmt.Sprintf(`
resource "opnsense_system_vlan" "test" {
  parent_interface = "vtnet0"
  tag              = %d
  device           = "vlan%04d"
}
`, tag, tag)
}
