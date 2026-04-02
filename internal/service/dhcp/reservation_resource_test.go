// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dhcp_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/acctest"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestAccDHCPv4Reservation_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		PreCheck:                 func() { acctest.PreCheck(t) },
		CheckDestroy:             acctest.CheckResourceDestroyed(t, "opnsense_dhcpv4_reservation", opnsense.ReqOpts{GetEndpoint: "/api/kea/dhcpv4/get_reservation", Monad: "reservation"}),
		Steps: []resource.TestStep{
			{
				Config: testAccDHCPv4ReservationConfig("10.99.0.50", "00:11:22:33:44:55"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("opnsense_dhcpv4_reservation.test", "id"),
					resource.TestCheckResourceAttr("opnsense_dhcpv4_reservation.test", "ip_address", "10.99.0.50"),
					resource.TestCheckResourceAttr("opnsense_dhcpv4_reservation.test", "mac_address", "00:11:22:33:44:55"),
				),
			},
			{ResourceName: "opnsense_dhcpv4_reservation.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func testAccCheckDHCPv4ReservationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "opnsense_dhcpv4_reservation" {
			continue
		}
		return fmt.Errorf("DHCPv4 reservation %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccDHCPv4ReservationConfig(ip, mac string) string {
	return fmt.Sprintf(`
resource "opnsense_dhcpv4_subnet" "test" {
  subnet = "10.99.0.0/24"
}

resource "opnsense_dhcpv4_reservation" "test" {
  subnet      = opnsense_dhcpv4_subnet.test.id
  ip_address  = %[1]q
  mac_address = %[2]q
  hostname    = "tf-test-host"
}
`, ip, mac)
}
