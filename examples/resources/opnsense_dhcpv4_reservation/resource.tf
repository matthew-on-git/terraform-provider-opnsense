resource "opnsense_dhcpv4_reservation" "printer" {
  subnet      = opnsense_dhcpv4_subnet.lan.id
  ip_address  = "10.0.0.50"
  mac_address = "00:11:22:33:44:55"
  hostname    = "printer"
}
