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

func TestAccHAProxyAction_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkHAProxyActionTestResourcesDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccHAProxyActionConfig("tf_test_action", "if"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_haproxy_action.test", "id"),
					resource.TestCheckResourceAttr("opnsense_haproxy_action.test", "name", "tf_test_action"),
					resource.TestCheckResourceAttr("opnsense_haproxy_action.test", "type", "use_backend"),
					resource.TestCheckResourceAttr("opnsense_haproxy_action.test", "test_type", "if"),
				),
			},
			{
				ResourceName:      "opnsense_haproxy_action.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHAProxyActionConfig("tf_test_action", "unless"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_haproxy_action.test", "test_type", "unless"),
				),
			},
		},
	})
}

func checkHAProxyActionTestResourcesDestroyed(t *testing.T) func(*terraform.State) error {
	return func(s *terraform.State) error {
		checks := []func(*terraform.State) error{
			acctest.CheckResourceDestroyed(t, "opnsense_haproxy_action", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getAction", Monad: "action"}),
			acctest.CheckResourceDestroyed(t, "opnsense_haproxy_frontend", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getFrontend", Monad: "frontend"}),
			acctest.CheckResourceDestroyed(t, "opnsense_haproxy_acl", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getAcl", Monad: "acl"}),
			acctest.CheckResourceDestroyed(t, "opnsense_haproxy_backend", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getBackend", Monad: "backend"}),
			acctest.CheckResourceDestroyed(t, "opnsense_haproxy_server", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getServer", Monad: "server"}),
		}
		for _, check := range checks {
			if err := check(s); err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccHAProxyActionConfig(name, testType string) string {
	return fmt.Sprintf(`
resource "opnsense_haproxy_server" "web1" {
  name    = "tf_test_action_web1"
  address = "10.0.0.10"
  port    = 80
}

resource "opnsense_haproxy_backend" "pool" {
  name           = "tf_test_action_pool"
  mode           = "http"
  linked_servers = [opnsense_haproxy_server.web1.id]
}

resource "opnsense_haproxy_acl" "host" {
  name       = "tf_test_action_host"
  expression = "hdr"
  hdr        = "example.test"
}

resource "opnsense_haproxy_action" "test" {
  name        = %[1]q
  type        = "use_backend"
  use_backend = opnsense_haproxy_backend.pool.id
  linked_acls = [opnsense_haproxy_acl.host.id]
  test_type   = %[2]q
}

resource "opnsense_haproxy_frontend" "test" {
  name           = "tf_test_action_frontend"
  bind           = "127.0.0.1:18080"
  mode           = "http"
  default_backend = opnsense_haproxy_backend.pool.id
  linked_actions = [opnsense_haproxy_action.test.id]
}
`, name, testType)
}
