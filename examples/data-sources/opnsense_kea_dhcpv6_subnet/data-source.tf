# Look up an existing Kea DHCPv6 subnet by UUID
data "opnsense_kea_dhcpv6_subnet" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "kea_dhcpv6_subnet_id" {
  value = data.opnsense_kea_dhcpv6_subnet.example.id
}
