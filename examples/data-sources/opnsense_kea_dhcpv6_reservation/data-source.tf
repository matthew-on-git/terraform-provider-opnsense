# Look up an existing Kea DHCPv6 reservation by UUID
data "opnsense_kea_dhcpv6_reservation" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "kea_dhcpv6_reservation_id" {
  value = data.opnsense_kea_dhcpv6_reservation.example.id
}
