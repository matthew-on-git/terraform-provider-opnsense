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

func TestAccSystemGateway_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckSystemGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSystemGatewayConfig("tf_test_gw", "10.0.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_system_gateway.test", "id"),
					resource.TestCheckResourceAttr("opnsense_system_gateway.test", "name", "tf_test_gw"),
					resource.TestCheckResourceAttr("opnsense_system_gateway.test", "gateway", "10.0.0.1"),
					resource.TestCheckResourceAttr("opnsense_system_gateway.test", "enabled", "true"),
				),
			},
			{ResourceName: "opnsense_system_gateway.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccSystemGatewayConfig("tf_test_gw", "10.0.0.254"),
				Check:  resource.TestCheckResourceAttr("opnsense_system_gateway.test", "gateway", "10.0.0.254"),
			},
		},
	})
}

func testAccCheckSystemGatewayDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_system_gateway" {
			continue
		}
		return fmt.Errorf("gateway %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccSystemGatewayConfig(name, gw string) string {
	return fmt.Sprintf(`
resource "opnsense_system_gateway" "test" {
  name      = %[1]q
  interface = "wan"
  gateway   = %[2]q
}
`, name, gw)
}
