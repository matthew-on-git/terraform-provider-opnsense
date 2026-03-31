// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccUnboundHostOverride_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckUnboundHostOverrideDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccUnboundHostOverrideConfig("myhost", "example.com", "10.0.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_unbound_host_override.test", "id"),
					resource.TestCheckResourceAttr("opnsense_unbound_host_override.test", "hostname", "myhost"),
					resource.TestCheckResourceAttr("opnsense_unbound_host_override.test", "domain", "example.com"),
					resource.TestCheckResourceAttr("opnsense_unbound_host_override.test", "server", "10.0.0.1"),
					resource.TestCheckResourceAttr("opnsense_unbound_host_override.test", "rr", "A"),
					resource.TestCheckResourceAttr("opnsense_unbound_host_override.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_unbound_host_override.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update server and verify.
			{
				Config: testAccUnboundHostOverrideConfig("myhost", "example.com", "10.0.0.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_host_override.test", "hostname", "myhost"),
					resource.TestCheckResourceAttr("opnsense_unbound_host_override.test", "server", "10.0.0.2"),
				),
			},
		},
	})
}

// testAccCheckUnboundHostOverrideDestroy verifies all Unbound host override resources
// created during the test have been removed from OPNsense.
func testAccCheckUnboundHostOverrideDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_unbound_host_override" {
			continue
		}
		return fmt.Errorf("Unbound host override %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccUnboundHostOverrideConfig(hostname, domain, server string) string {
	return fmt.Sprintf(`
resource "opnsense_unbound_host_override" "test" {
  hostname = %[1]q
  domain   = %[2]q
  server   = %[3]q
}
`, hostname, domain, server)
}
