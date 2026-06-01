// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccIPsecLocal_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_ipsec_local", opnsense.ReqOpts{GetEndpoint: "/api/ipsec/connections/get_local", Monad: "local"}),
		Steps: []resource.TestStep{
			// Step 1: Create a connection and a PSK local auth entry referencing it.
			{
				Config: testAccIPsecLocalConfig("tfacc@local"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_ipsec_local.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_ipsec_local.test", "connection_id"),
					resource.TestCheckResourceAttr("opnsense_ipsec_local.test", "auth", "psk"),
					resource.TestCheckResourceAttr("opnsense_ipsec_local.test", "identity", "tfacc@local"),
					resource.TestCheckResourceAttr("opnsense_ipsec_local.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_ipsec_local.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update the identity and verify.
			{
				Config: testAccIPsecLocalConfig("tfacc2@local"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ipsec_local.test", "identity", "tfacc2@local"),
				),
			},
		},
	})
}

func testAccIPsecLocalConfig(identity string) string {
	return fmt.Sprintf(`
resource "opnsense_ipsec_connection" "conn" {
  description  = "tfacc-local-conn"
  remote_addrs = "192.0.2.1"
}

resource "opnsense_ipsec_local" "test" {
  connection_id = opnsense_ipsec_connection.conn.id
  auth       = "psk"
  identity   = %[1]q
}
`, identity)
}
