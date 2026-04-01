// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var aliasReqOpts = opnsense.ReqOpts{
	GetEndpoint: "/api/firewall/alias/getItem",
	Monad:       "alias",
}

func TestAccFirewallAlias_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_firewall_alias", aliasReqOpts),
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccFirewallAliasConfig("tf_test_alias", "host", "10.0.0.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_firewall_alias.test", "id"),
					resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "name", "tf_test_alias"),
					resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "type", "host"),
					resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_firewall_alias.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update content and verify.
			{
				Config: testAccFirewallAliasConfig("tf_test_alias", "host", "10.0.0.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "name", "tf_test_alias"),
				),
			},
		},
	})
}

func testAccFirewallAliasConfig(name, aliasType, content string) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_alias" "test" {
  name    = %[1]q
  type    = %[2]q
  content = [%[3]q]
}
`, name, aliasType, content)
}
