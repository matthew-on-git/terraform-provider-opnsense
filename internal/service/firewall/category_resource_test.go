// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
)

func TestAccFirewallCategory_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_firewall_category", opnsense.ReqOpts{GetEndpoint: "/api/firewall/category/getItem", Monad: "category"}),
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccFirewallCategoryConfig("tf_test_category", "ff0000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_firewall_category.test", "id"),
					resource.TestCheckResourceAttr("opnsense_firewall_category.test", "name", "tf_test_category"),
					resource.TestCheckResourceAttr("opnsense_firewall_category.test", "color", "ff0000"),
					resource.TestCheckResourceAttr("opnsense_firewall_category.test", "auto", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_firewall_category.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update color and verify.
			{
				Config: testAccFirewallCategoryConfig("tf_test_category", "00ff00"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_firewall_category.test", "name", "tf_test_category"),
					resource.TestCheckResourceAttr("opnsense_firewall_category.test", "color", "00ff00"),
				),
			},
		},
	})
}

func testAccFirewallCategoryConfig(name, color string) string {
	return fmt.Sprintf(`
resource "opnsense_firewall_category" "test" {
  name  = %[1]q
  color = %[2]q
}
`, name, color)
}
