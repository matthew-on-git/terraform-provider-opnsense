// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccHAProxyServer_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_haproxy_server", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getServer", Monad: "server"}),
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccHAProxyServerConfig("tf_test_server", "10.0.0.10", 80),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_haproxy_server.test", "id"),
					resource.TestCheckResourceAttr("opnsense_haproxy_server.test", "name", "tf_test_server"),
					resource.TestCheckResourceAttr("opnsense_haproxy_server.test", "address", "10.0.0.10"),
					resource.TestCheckResourceAttr("opnsense_haproxy_server.test", "port", "80"),
					resource.TestCheckResourceAttr("opnsense_haproxy_server.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_haproxy_server.test", "ssl", "false"),
					resource.TestCheckResourceAttr("opnsense_haproxy_server.test", "mode", "active"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_haproxy_server.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update port and verify.
			{
				Config: testAccHAProxyServerConfig("tf_test_server", "10.0.0.10", 8080),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_haproxy_server.test", "name", "tf_test_server"),
					resource.TestCheckResourceAttr("opnsense_haproxy_server.test", "port", "8080"),
				),
			},
		},
	})
}

// testAccCheckHAProxyServerDestroy verifies all HAProxy server resources
// created during the test have been removed from OPNsense.
func testAccCheckHAProxyServerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_haproxy_server" {
			continue
		}
		return fmt.Errorf("HAProxy server %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccHAProxyServerConfig(name, address string, port int) string {
	return fmt.Sprintf(`
resource "opnsense_haproxy_server" "test" {
  name    = %[1]q
  address = %[2]q
  port    = %[3]d
}
`, name, address, port)
}
