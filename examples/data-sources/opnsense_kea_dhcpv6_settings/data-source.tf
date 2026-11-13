# Read the Kea DHCPv6 general settings singleton.
# Identity is kea-dhcpv6. The lookup never writes or reconfigures.
data "opnsense_kea_dhcpv6_settings" "example" {
  id = "kea-dhcpv6"
}

output "kea_dhcpv6_valid_lifetime" {
  value = data.opnsense_kea_dhcpv6_settings.example.valid_lifetime
}
