# Look up an existing DHCPv4 reservation by UUID
data "opnsense_dhcpv4_reservation" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "dhcpv4_reservation_id" {
  value = data.opnsense_dhcpv4_reservation.example.id
}
