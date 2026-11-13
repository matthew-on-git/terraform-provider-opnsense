# Read the Kea DHCPv4 general settings singleton.
# Identity is kea-dhcpv4. The lookup never writes or reconfigures.
data "opnsense_kea_dhcpv4_settings" "example" {
  id = "kea-dhcpv4"
}

output "kea_dhcpv4_valid_lifetime" {
  value = data.opnsense_kea_dhcpv4_settings.example.valid_lifetime
}
