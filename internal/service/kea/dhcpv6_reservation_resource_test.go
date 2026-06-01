// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccKeaDHCPv6Reservation_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_kea_dhcpv6_reservation", opnsense.ReqOpts{GetEndpoint: "/api/kea/dhcpv6/get_reservation", Monad: "reservation"}),
		Steps: []resource.TestStep{
			// Step 1: Create settings + subnet + reservation, and verify.
			{
				Config: testAccKeaDHCPv6ReservationConfig("2001:db8::100", "tfacc"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_kea_dhcpv6_reservation.test", "id"),
					resource.TestCheckResourceAttrSet("opnsense_kea_dhcpv6_reservation.test", "subnet_id"),
					resource.TestCheckResourceAttr("opnsense_kea_dhcpv6_reservation.test", "ip_address", "2001:db8::100"),
					resource.TestCheckResourceAttr("opnsense_kea_dhcpv6_reservation.test", "hostname", "tfacc"),
				),
			},
			// Step 2: Import and verify.
			{
				ResourceName:      "opnsense_kea_dhcpv6_reservation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 3: Update hostname and verify.
			{
				Config: testAccKeaDHCPv6ReservationConfig("2001:db8::100", "tfacc2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("opnsense_kea_dhcpv6_reservation.test", "hostname", "tfacc2"),
				),
			},
		},
	})
}

func testAccKeaDHCPv6ReservationConfig(ip, hostname string) string {
	return fmt.Sprintf(`
resource "opnsense_kea_dhcpv6_settings" "s" {
  enabled    = true
  interfaces = ["lan"]
}

resource "opnsense_kea_dhcpv6_subnet" "net" {
  subnet     = "2001:db8::/64"
  interface  = "lan"
  depends_on = [opnsense_kea_dhcpv6_settings.s]
}

resource "opnsense_kea_dhcpv6_reservation" "test" {
  subnet_id  = opnsense_kea_dhcpv6_subnet.net.id
  ip_address = %[1]q
  duid       = "00:03:00:01:aa:bb:cc:dd:ee:ff"
  hostname   = %[2]q
}
`, ip, hostname)
}
