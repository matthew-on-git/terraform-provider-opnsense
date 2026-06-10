resource "opnsense_dnsmasq_range" "lan" {
  start_address = "192.0.2.100"
  end_address   = "192.0.2.150"
  domain_type   = "range"
  domain        = "example.com"
  description   = "Example LAN DHCP range"
}
