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

func TestAccSystemRoute_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckSystemRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemRouteConfig("10.99.0.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_system_route.test", "id"),
					resource.TestCheckResourceAttr("opnsense_system_route.test", "network", "10.99.0.0/24"),
					resource.TestCheckResourceAttr("opnsense_system_route.test", "enabled", "true"),
				),
			},
			{ResourceName: "opnsense_system_route.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccSystemRouteConfig("10.99.1.0/24"),
				Check:  resource.TestCheckResourceAttr("opnsense_system_route.test", "network", "10.99.1.0/24"),
			},
		},
	})
}

func testAccCheckSystemRouteDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_system_route" {
			continue
		}
		return fmt.Errorf("static route %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSystemRouteConfig(network string) string {
	return fmt.Sprintf(`
resource "opnsense_system_route" "test" {
  network     = %[1]q
  gateway     = "Null4"
  description = "Terraform test route"
}
`, network)
}
