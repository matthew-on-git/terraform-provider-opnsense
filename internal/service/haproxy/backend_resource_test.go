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

func TestAccHAProxyBackend_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_haproxy_backend", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getBackend", Monad: "backend"}),
		Steps: []resource.TestStep{
			// Step 1: Create server + backend with server linking.
			{
				Config: testAccHAProxyBackendConfig("tf_test_backend", "roundrobin"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_haproxy_backend.test", "id"),
					resource.TestCheckResourceAttr("opnsense_haproxy_backend.test", "name", "tf_test_backend"),
					resource.TestCheckResourceAttr("opnsense_haproxy_backend.test", "algorithm", "roundrobin"),
					resource.TestCheckResourceAttr("opnsense_haproxy_backend.test", "mode", "http"),
					resource.TestCheckResourceAttr("opnsense_haproxy_backend.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_haproxy_backend.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update algorithm and verify.
			{
				Config: testAccHAProxyBackendConfig("tf_test_backend", "leastconn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_haproxy_backend.test", "algorithm", "leastconn"),
				),
			},
		},
	})
}

func testAccCheckHAProxyBackendDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_haproxy_backend" {
			continue
		}
		return fmt.Errorf("HAProxy backend %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccHAProxyBackendConfig(name, algorithm string) string {
	return fmt.Sprintf(`
resource "opnsense_haproxy_server" "web1" {
  name    = "tf_test_be_web1"
  address = "10.0.0.10"
  port    = 80
}

resource "opnsense_haproxy_backend" "test" {
  name           = %[1]q
  mode           = "http"
  algorithm      = %[2]q
  linked_servers = [opnsense_haproxy_server.web1.id]
}
`, name, algorithm)
}
