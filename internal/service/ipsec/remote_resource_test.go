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

func TestAccIPsecRemote_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_ipsec_remote", opnsense.ReqOpts{GetEndpoint: "/api/ipsec/connections/get_remote", Monad: "remote"}),
		Steps: []resource.TestStep{
			// Step 1: Create a connection and a PSK remote auth entry referencing it.
			{
				Config: testAccIPsecRemoteConfig("tfacc@remote"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_ipsec_remote.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_ipsec_remote.test", "connection_id"),
					resource.TestCheckResourceAttr("opnsense_ipsec_remote.test", "auth", "psk"),
					resource.TestCheckResourceAttr("opnsense_ipsec_remote.test", "identity", "tfacc@remote"),
					resource.TestCheckResourceAttr("opnsense_ipsec_remote.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_ipsec_remote.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update the identity and verify.
			{
				Config: testAccIPsecRemoteConfig("tfacc2@remote"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_ipsec_remote.test", "identity", "tfacc2@remote"),
				),
			},
		},
	})
}

func testAccIPsecRemoteConfig(identity string) string {
	return fmt.Sprintf(`
resource "opnsense_ipsec_connection" "conn" {
  description  = "tfacc-remote-conn"
  remote_addrs = "192.0.2.1"
}

resource "opnsense_ipsec_remote" "test" {
  connection_id = opnsense_ipsec_connection.conn.id
  auth          = "psk"
  identity      = %[1]q
}
`, identity)
}
