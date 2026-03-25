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

func TestAccHAProxyACL_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             testAccCheckHAProxyACLDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create host header ACL.
			{
				Config: testAccHAProxyACLConfig("tf_test_acl", "hdr_beg", "example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_haproxy_acl.test", "id"),
					resource.TestCheckResourceAttr("opnsense_haproxy_acl.test", "name", "tf_test_acl"),
					resource.TestCheckResourceAttr("opnsense_haproxy_acl.test", "expression", "hdr_beg"),
					resource.TestCheckResourceAttr("opnsense_haproxy_acl.test", "hdr_beg", "example.com"),
					resource.TestCheckResourceAttr("opnsense_haproxy_acl.test", "negate", "false"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_haproxy_acl.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update match value and verify.
			{
				Config: testAccHAProxyACLConfig("tf_test_acl", "hdr_beg", "other.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_haproxy_acl.test", "hdr_beg", "other.com"),
				),
			},
		},
	})
}

func testAccCheckHAProxyACLDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_haproxy_acl" {
			continue
		}
		return fmt.Errorf("HAProxy ACL %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccHAProxyACLConfig(name, expression, hdrBeg string) string {
	return fmt.Sprintf(`
resource "opnsense_haproxy_acl" "test" {
  name       = %[1]q
  expression = %[2]q
  hdr_beg    = %[3]q
}
`, name, expression, hdrBeg)
}
