// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccUnboundDomainOverride_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_unbound_domain_override", opnsense.ReqOpts{GetEndpoint: "/api/unbound/settings/get_forward", Monad: "forward"}),
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccUnboundDomainOverrideConfig("example.com", "10.0.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_unbound_domain_override.test", "id"),
					resource.TestCheckResourceAttr("opnsense_unbound_domain_override.test", "domain", "example.com"),
					resource.TestCheckResourceAttr("opnsense_unbound_domain_override.test", "server", "10.0.0.1"),
					resource.TestCheckResourceAttr("opnsense_unbound_domain_override.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_unbound_domain_override.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update server and verify.
			{
				Config: testAccUnboundDomainOverrideConfig("example.com", "10.0.0.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_domain_override.test", "domain", "example.com"),
					resource.TestCheckResourceAttr("opnsense_unbound_domain_override.test", "server", "10.0.0.2"),
				),
			},
		},
	})
}

func testAccUnboundDomainOverrideConfig(domain, server string) string {
	return fmt.Sprintf(`
resource "opnsense_unbound_domain_override" "test" {
  domain = %[1]q
  server = %[2]q
}
`, domain, server)
}
