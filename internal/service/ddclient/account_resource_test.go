// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccDDClientAccount_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_ddclient_account", opnsense.ReqOpts{GetEndpoint: "/api/dyndns/accounts/get_item", Monad: "account"}),
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccDDClientAccountConfig("cloudflare", "test.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_ddclient_account.test", "id"),
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "service", "cloudflare"),
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "hostnames", "test.example.com"),
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "check_ip", "web_icanhazip"),
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:            "opnsense_ddclient_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			// Step 3: Update hostnames and verify.
			{
				Config: testAccDDClientAccountConfig("cloudflare", "updated.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "service", "cloudflare"),
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "hostnames", "updated.example.com"),
				),
			},
		},
	})
}

func TestAccDDClientSettings_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDDClientSettingsConfig("cloudflare", "settings.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ddclient_settings.test", "id", "ddclient-settings"),
					resource.TestCheckResourceAttr("opnsense_ddclient_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttr("opnsense_ddclient_settings.test", "backend", "opnsense"),
					resource.TestCheckResourceAttr("opnsense_ddclient_settings.test", "interval", "300"),
					resource.TestCheckResourceAttr("opnsense_ddclient_account.test", "hostnames", "settings.example.com"),
				),
			},
			{
				ResourceName:      "opnsense_ddclient_settings.test",
				ImportState:       true,
				ImportStateId:     "ddclient-settings",
				ImportStateVerify: true,
			},
			{
				Config:             testAccDDClientSettingsConfig("cloudflare", "settings.example.com"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccDDClientAccountConfig(service, hostnames string) string {
	return fmt.Sprintf(`
resource "opnsense_ddclient_account" "test" {
  service   = %[1]q
  hostnames = %[2]q
  check_ip  = "web_icanhazip"
  username  = "testuser"
  password  = "testpass"
}
`, service, hostnames)
}

func testAccDDClientSettingsConfig(service, hostnames string) string {
	return fmt.Sprintf(`
resource "opnsense_ddclient_settings" "test" {
  enabled    = true
  backend    = "opnsense"
  interval   = 300
  verbose    = false
  allow_ipv6 = false
}

resource "opnsense_ddclient_account" "test" {
  service    = %[1]q
  hostnames  = %[2]q
  check_ip   = "web_icanhazip"
  username   = "testuser"
  password   = "testpass"
  depends_on = [opnsense_ddclient_settings.test]
}
`, service, hostnames)
}
