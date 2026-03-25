// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccHAProxyHealthcheck_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckHAProxyHealthcheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHAProxyHealthcheckConfig("tf_test_hc", "http", "/health"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_haproxy_healthcheck.test", "id"),
					resource.TestCheckResourceAttr("opnsense_haproxy_healthcheck.test", "name", "tf_test_hc"),
					resource.TestCheckResourceAttr("opnsense_haproxy_healthcheck.test", "type", "http"),
					resource.TestCheckResourceAttr("opnsense_haproxy_healthcheck.test", "http_uri", "/health"),
				),
			},
			{
				ResourceName:      "opnsense_haproxy_healthcheck.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHAProxyHealthcheckConfig("tf_test_hc", "http", "/status"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_haproxy_healthcheck.test", "http_uri", "/status"),
				),
			},
		},
	})
}

func testAccCheckHAProxyHealthcheckDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_haproxy_healthcheck" {
			continue
		}
		return fmt.Errorf("HAProxy health check %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccHAProxyHealthcheckConfig(name, checkType, httpURI string) string {
	return fmt.Sprintf(`
resource "opnsense_haproxy_healthcheck" "test" {
  name     = %[1]q
  type     = %[2]q
  http_uri = %[3]q
}
`, name, checkType, httpURI)
}
