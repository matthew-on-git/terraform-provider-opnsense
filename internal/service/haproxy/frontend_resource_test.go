// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccHAProxyFrontend_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_haproxy_frontend", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getFrontend", Monad: "frontend"}),
		Steps: []resource.TestStep{
			// Step 1: Create server → backend → frontend chain.
			{
				Config: testAccHAProxyFrontendConfig("tf_test_frontend", "0.0.0.0:8080"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_haproxy_frontend.test", "id"),
					resource.TestCheckResourceAttr("opnsense_haproxy_frontend.test", "name", "tf_test_frontend"),
					resource.TestCheckResourceAttr("opnsense_haproxy_frontend.test", "bind", "0.0.0.0:8080"),
					resource.TestCheckResourceAttr("opnsense_haproxy_frontend.test", "mode", "http"),
					resource.TestCheckResourceAttr("opnsense_haproxy_frontend.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_haproxy_frontend.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update bind address and verify.
			{
				Config: testAccHAProxyFrontendConfig("tf_test_frontend", "0.0.0.0:9090"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_haproxy_frontend.test", "bind", "0.0.0.0:9090"),
				),
			},
		},
	})
}

func TestAccHAProxyFrontend_sslCertificate(t *testing.T) {
	certRefID := os.Getenv("OPNSENSE_HAPROXY_CERT_REFID")
	if certRefID == "" {
		t.Skip("OPNSENSE_HAPROXY_CERT_REFID must be set to run SSL certificate binding acceptance test")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_haproxy_frontend", opnsense.ReqOpts{GetEndpoint: "/api/haproxy/settings/getFrontend", Monad: "frontend"}),
		Steps: []resource.TestStep{
			{
				Config: testAccHAProxyFrontendSSLConfig("tf_test_frontend_ssl", "127.0.0.1:18443", certRefID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_haproxy_frontend.test", "id"),
					resource.TestCheckResourceAttr("opnsense_haproxy_frontend.test", "ssl_enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_haproxy_frontend.test", "certificates.#", "1"),
					resource.TestCheckTypeSetElemAttr("opnsense_haproxy_frontend.test", "certificates.*", certRefID),
					resource.TestCheckResourceAttr("opnsense_haproxy_frontend.test", "default_certificate", certRefID),
				),
			},
			{
				Config:   testAccHAProxyFrontendSSLConfig("tf_test_frontend_ssl", "127.0.0.1:18443", certRefID),
				PlanOnly: true,
			},
			{
				ResourceName:      "opnsense_haproxy_frontend.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccHAProxyFrontendConfig(name, bind string) string {
	return fmt.Sprintf(`
resource "opnsense_haproxy_server" "web1" {
  name    = "tf_test_fe_web1"
  address = "10.0.0.10"
  port    = 80
}

resource "opnsense_haproxy_backend" "pool" {
  name           = "tf_test_fe_pool"
  mode           = "http"
  algorithm      = "roundrobin"
  linked_servers = [opnsense_haproxy_server.web1.id]
}

resource "opnsense_haproxy_frontend" "test" {
  name            = %[1]q
  bind            = %[2]q
  mode            = "http"
  default_backend = opnsense_haproxy_backend.pool.id
}
`, name, bind)
}

func testAccHAProxyFrontendSSLConfig(name, bind, certRefID string) string {
	return fmt.Sprintf(`
resource "opnsense_haproxy_server" "web1" {
  name    = "tf_test_fe_ssl_web1"
  address = "10.0.0.10"
  port    = 80
}

resource "opnsense_haproxy_backend" "pool" {
  name           = "tf_test_fe_ssl_pool"
  mode           = "http"
  algorithm      = "roundrobin"
  linked_servers = [opnsense_haproxy_server.web1.id]
}

resource "opnsense_haproxy_frontend" "test" {
  name                = %[1]q
  bind                = %[2]q
  mode                = "http"
  default_backend     = opnsense_haproxy_backend.pool.id
  ssl_enabled         = true
  certificates        = [%[3]q]
  default_certificate = %[3]q
}
`, name, bind, certRefID)
}
