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

func TestAccUnboundHostAlias_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_unbound_host_alias", opnsense.ReqOpts{GetEndpoint: "/api/unbound/settings/get_host_alias", Monad: "alias"}),
		Steps: []resource.TestStep{
			// Step 1: Create the parent host override and an alias pointing at it.
			{
				Config: testAccUnboundHostAliasConfig("alias1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_unbound_host_alias.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_unbound_host_alias.test", "host"),
					resource.TestCheckResourceAttr("opnsense_unbound_host_alias.test", "hostname", "alias1"),
					resource.TestCheckResourceAttr("opnsense_unbound_host_alias.test", "domain", "example.com"),
					resource.TestCheckResourceAttr("opnsense_unbound_host_alias.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_unbound_host_alias.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update hostname and verify.
			{
				Config: testAccUnboundHostAliasConfig("alias2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_host_alias.test", "hostname", "alias2"),
				),
			},
		},
	})
}

func testAccUnboundHostAliasConfig(hostname string) string {
	return fmt.Sprintf(`
resource "opnsense_unbound_host_override" "parent" {
  hostname = "tfacc-parent"
  domain   = "example.com"
  server   = "192.0.2.10"
}

resource "opnsense_unbound_host_alias" "test" {
  host     = opnsense_unbound_host_override.parent.id
  hostname = %[1]q
  domain   = "example.com"
}
`, hostname)
}
