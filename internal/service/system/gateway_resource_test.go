// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccSystemGateway_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_system_gateway", opnsense.ReqOpts{GetEndpoint: "/api/routing/settings/get_gateway", Monad: "gateway_item"}),
		Steps: []resource.TestStep{
			{
				Config: testAccSystemGatewayConfig("tf_test_gw", "192.168.56.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_system_gateway.test", "id"),
					resource.TestCheckResourceAttr("opnsense_system_gateway.test", "name", "tf_test_gw"),
					resource.TestCheckResourceAttr("opnsense_system_gateway.test", "gateway", "192.168.56.1"),
					resource.TestCheckResourceAttr("opnsense_system_gateway.test", "enabled", "true"),
				),
			},
			{ResourceName: "opnsense_system_gateway.test", ImportState: true, ImportStateVerify: true},
			{
				Config: testAccSystemGatewayConfig("tf_test_gw", "192.168.56.254"),
				Check:  resource.TestCheckResourceAttr("opnsense_system_gateway.test", "gateway", "192.168.56.254"),
			},
		},
	})
}

func testAccSystemGatewayConfig(name, gw string) string {
	return fmt.Sprintf(`
resource "opnsense_system_gateway" "test" {
  name      = %[1]q
  interface = "lan"
  gateway   = %[2]q
}
`, name, gw)
}
