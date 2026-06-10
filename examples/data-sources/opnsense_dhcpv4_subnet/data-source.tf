# Look up an existing DHCPv4 subnet by UUID
data "opnsense_dhcpv4_subnet" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "dhcpv4_subnet_id" {
  value = data.opnsense_dhcpv4_subnet.example.id
}
