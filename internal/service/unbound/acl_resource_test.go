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

func TestAccUnboundACL_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_unbound_acl", opnsense.ReqOpts{GetEndpoint: "/api/unbound/settings/get_acl", Monad: "acl"}),
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccUnboundACLConfig("tf_test_acl", "allow", "10.0.0.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_unbound_acl.test", "id"),
					resource.TestCheckResourceAttr("opnsense_unbound_acl.test", "name", "tf_test_acl"),
					resource.TestCheckResourceAttr("opnsense_unbound_acl.test", "action", "allow"),
					resource.TestCheckResourceAttr("opnsense_unbound_acl.test", "networks", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("opnsense_unbound_acl.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_unbound_acl.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update action and verify.
			{
				Config: testAccUnboundACLConfig("tf_test_acl", "deny", "10.0.0.0/24"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_unbound_acl.test", "name", "tf_test_acl"),
					resource.TestCheckResourceAttr("opnsense_unbound_acl.test", "action", "deny"),
				),
			},
		},
	})
}

func testAccUnboundACLConfig(name, action, networks string) string {
	return fmt.Sprintf(`
resource "opnsense_unbound_acl" "test" {
  name     = %[1]q
  action   = %[2]q
  networks = %[3]q
}
`, name, action, networks)
}
