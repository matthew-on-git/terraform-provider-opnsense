resource "opnsense_dhcpv4_subnet" "lan" {
  subnet      = "10.0.0.0/24"
  pools       = "10.0.0.100-10.0.0.200"
  description = "LAN DHCP pool"
}
