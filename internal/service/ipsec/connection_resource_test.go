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

func TestAccIPsecConnection_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckIPsecConnectionDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccIPsecConnectionConfig("tf_test_conn", "10.0.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_ipsec_connection.test", "id"),
					resource.TestCheckResourceAttr("opnsense_ipsec_connection.test", "description", "tf_test_conn"),
					resource.TestCheckResourceAttr("opnsense_ipsec_connection.test", "remote_addrs", "10.0.0.1"),
					resource.TestCheckResourceAttr("opnsense_ipsec_connection.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_ipsec_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update remote address and verify.
			{
				Config: testAccIPsecConnectionConfig("tf_test_conn", "10.0.0.2"),
				Check:  resource.TestCheckResourceAttr("opnsense_ipsec_connection.test", "remote_addrs", "10.0.0.2"),
			},
		},
	})
}

// testAccCheckIPsecConnectionDestroy verifies all IPsec connection resources
// created during the test have been removed from OPNsense.
func testAccCheckIPsecConnectionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_ipsec_connection" {
			continue
		}
		return fmt.Errorf("IPsec connection %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccIPsecConnectionConfig(description, remoteAddrs string) string {
	return fmt.Sprintf(`
resource "opnsense_ipsec_connection" "test" {
  description  = %[1]q
  remote_addrs = %[2]q
}
`, description, remoteAddrs)
}
