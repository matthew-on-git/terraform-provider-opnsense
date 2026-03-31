// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccIPsecChild_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckIPsecChildDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccIPsecChildConfig("tunnel", "10.0.0.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_ipsec_child.test", "id"),
					resource.TestCheckResourceAttr("opnsense_ipsec_child.test", "mode", "tunnel"),
					resource.TestCheckResourceAttr("opnsense_ipsec_child.test", "local_ts", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("opnsense_ipsec_child.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_ipsec_child.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update local traffic selector and verify.
			{
				Config: testAccIPsecChildConfig("tunnel", "192.168.0.0/24"),
				Check:  resource.TestCheckResourceAttr("opnsense_ipsec_child.test", "local_ts", "192.168.0.0/24"),
			},
		},
	})
}

// testAccCheckIPsecChildDestroy verifies all IPsec child SA resources
// created during the test have been removed from OPNsense.
func testAccCheckIPsecChildDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_ipsec_child" {
			continue
		}
		return fmt.Errorf("IPsec child SA %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccIPsecChildConfig(mode, localTS string) string {
	return fmt.Sprintf(`
resource "opnsense_ipsec_connection" "test" {
  description  = "tf_test_conn_for_child"
  remote_addrs = "10.0.0.1"
}

resource "opnsense_ipsec_child" "test" {
  connection = opnsense_ipsec_connection.test.id
  mode       = %[1]q
  local_ts   = %[2]q
  remote_ts  = "10.0.1.0/24"
}
`, mode, localTS)
}
