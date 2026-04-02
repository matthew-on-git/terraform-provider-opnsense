// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccWireguardPeer_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_wireguard_peer", opnsense.ReqOpts{GetEndpoint: "/api/wireguard/client/get_client", Monad: "client"}),
		Steps: []resource.TestStep{
			// Step 1: Create and verify.
			{
				Config: testAccWireguardPeerConfig("tf_test_peer", "dGVzdHB1YmxpY2tleQ==", "10.0.0.2/32"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_wireguard_peer.test", "id"),
					resource.TestCheckResourceAttr("opnsense_wireguard_peer.test", "name", "tf_test_peer"),
					resource.TestCheckResourceAttr("opnsense_wireguard_peer.test", "public_key", "dGVzdHB1YmxpY2tleQ=="),
					resource.TestCheckResourceAttr("opnsense_wireguard_peer.test", "tunnel_address", "10.0.0.2/32"),
					resource.TestCheckResourceAttr("opnsense_wireguard_peer.test", "enabled", "true"),
				),
			},
			// Step 2: Import and verify state matches.
			{
				ResourceName:      "opnsense_wireguard_peer.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update tunnel address and verify.
			{
				Config: testAccWireguardPeerConfig("tf_test_peer", "dGVzdHB1YmxpY2tleQ==", "10.0.0.3/32"),
				Check:  resource.TestCheckResourceAttr("opnsense_wireguard_peer.test", "tunnel_address", "10.0.0.3/32"),
			},
		},
	})
}

// testAccCheckWireguardPeerDestroy verifies all WireGuard peer resources
// created during the test have been removed from OPNsense.
func testAccCheckWireguardPeerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_wireguard_peer" {
			continue
		}
		return fmt.Errorf("WireGuard peer %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccWireguardPeerConfig(name, pubkey, tunnelAddr string) string {
	return fmt.Sprintf(`
resource "opnsense_wireguard_peer" "test" {
  name           = %[1]q
  public_key     = %[2]q
  tunnel_address = %[3]q
}
`, name, pubkey, tunnelAddr)
}
