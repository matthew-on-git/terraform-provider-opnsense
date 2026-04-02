// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccIPsecPSK_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_ipsec_psk", opnsense.ReqOpts{GetEndpoint: "/api/ipsec/pre_shared_keys/get_item", Monad: "preSharedKey"}),
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccIPsecPSKConfig("local@example.com", "supersecretkey"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_ipsec_psk.test", "id"),
					resource.TestCheckResourceAttr("opnsense_ipsec_psk.test", "identity", "local@example.com"),
					resource.TestCheckResourceAttr("opnsense_ipsec_psk.test", "key_type", "PSK"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:            "opnsense_ipsec_psk.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"key"},
			},
			// Step 3: Update identity and verify.
			{
				Config: testAccIPsecPSKConfig("updated@example.com", "supersecretkey"),
				Check:  resource.TestCheckResourceAttr("opnsense_ipsec_psk.test", "identity", "updated@example.com"),
			},
		},
	})
}

// testAccCheckIPsecPSKDestroy verifies all IPsec pre-shared key resources
// created during the test have been removed from OPNsense.
func testAccCheckIPsecPSKDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_ipsec_psk" {
			continue
		}
		return fmt.Errorf("IPsec pre-shared key %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccIPsecPSKConfig(identity, key string) string {
	return fmt.Sprintf(`
resource "opnsense_ipsec_psk" "test" {
  identity = %[1]q
  key      = %[2]q
}
`, identity, key)
}
