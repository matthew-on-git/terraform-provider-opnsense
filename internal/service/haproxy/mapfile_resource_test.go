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

func TestAccHAProxyMapfile_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             checkHAProxyMapfileTestResourcesDestroyed(t),
		Steps: []resource.TestStep{
			{
				Config: testAccHAProxyMapfileConfig("tf_test_domain_map", "grafana.example.test tf_test_mapfile_pool"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_haproxy_mapfile.test", "id"),
					resource.TestCheckResourceAttr("opnsense_haproxy_mapfile.test", "name", "tf_test_domain_map"),
					resource.TestCheckResourceAttr("opnsense_haproxy_mapfile.test", "type", "dom"),
					resource.TestCheckResourceAttr("opnsense_haproxy_mapfile.test", "content", "grafana.example.test tf_test_mapfile_pool"),
					resource.TestCheckResourceAttr("opnsense_haproxy_action.test", "type", "map_use_backend"),
					resource.TestCheckResourceAttrPair("opnsense_haproxy_action.test", "mapfile", "opnsense_haproxy_mapfile.test", "id"),
					resource.TestCheckResourceAttrPair("opnsense_haproxy_action.test", "map_use_backend_default", "opnsense_haproxy_backend.pool", "id"),
				),
			},
			{
				ResourceName:      "opnsense_haproxy_mapfile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccHAProxyMapfileConfig("tf_test_domain_map", "grafana.example.test tf_test_mapfile_pool"),
				PlanOnly: true,
			},
			{
				Config: testAccHAProxyMapfileConfig("tf_test_domain_map", "grafana.example.test tf_test_mapfile_pool\nargocd.example.test tf_test_mapfile_pool"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_haproxy_mapfile.test", "content", "grafana.example.test tf_test_mapfile_pool\nargocd.example.test tf_test_mapfile_pool"),
				),
			},
		},
	})
}

func checkHAProxyMapfileTestResourcesDestroyed(t *testing.T) func(*terraform.State) error {
	return func(s *terraform.State) error {
		checks := []func(*terraform.State) error{
			acctest.CheckResourceDestroyed(t, "opnsense_haproxy_action", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getAction", Monad: "action"}),
			acctest.CheckResourceDestroyed(t, "opnsense_haproxy_mapfile", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getMapFile", Monad: "mapfile"}),
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

func testAccHAProxyMapfileConfig(name, content string) string {
	return fmt.Sprintf(`
resource "opnsense_haproxy_server" "web1" {
  name    = "tf_test_mapfile_web1"
  address = "10.0.0.10"
  port    = 80
}

resource "opnsense_haproxy_backend" "pool" {
  name           = "tf_test_mapfile_pool"
  mode           = "http"
  linked_servers = [opnsense_haproxy_server.web1.id]
}

resource "opnsense_haproxy_mapfile" "test" {
  name    = %[1]q
  type    = "dom"
  content = %[2]q
}

resource "opnsense_haproxy_action" "test" {
  name                    = "tf_test_mapfile_action"
  type                    = "map_use_backend"
  mapfile                 = opnsense_haproxy_mapfile.test.id
  map_use_backend_default = opnsense_haproxy_backend.pool.id
}
`, name, content)
}
