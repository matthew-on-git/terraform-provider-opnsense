resource "opnsense_system_vip" "web_vip" {
  interface   = "lan"
  mode        = "ipalias"
  address     = "10.0.0.100"
  subnet_bits = 24
  description = "Web service VIP"
}
