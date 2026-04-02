// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccWireguardServer_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_wireguard_server", opnsense.ReqOpts{GetEndpoint: "/api/wireguard/server/get_server", Monad: "server"}),
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccWireguardServerConfig("tf_test_wg", "51820"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_wireguard_server.test", "id"),
					resource.TestCheckResourceAttr("opnsense_wireguard_server.test", "name", "tf_test_wg"),
					resource.TestCheckResourceAttr("opnsense_wireguard_server.test", "port", "51820"),
					resource.TestCheckResourceAttr("opnsense_wireguard_server.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:            "opnsense_wireguard_server.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key"},
			},
			// Step 3: Update port and verify.
			{
				Config: testAccWireguardServerConfig("tf_test_wg", "51821"),
				Check:  resource.TestCheckResourceAttr("opnsense_wireguard_server.test", "port", "51821"),
			},
		},
	})
}

func testAccWireguardServerConfig(name, port string) string {
	return fmt.Sprintf(`
resource "opnsense_wireguard_server" "test" {
  name = %[1]q
  port = %[2]q
}
`, name, port)
}
