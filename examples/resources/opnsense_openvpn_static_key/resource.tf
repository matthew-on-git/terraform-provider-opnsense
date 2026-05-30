resource "opnsense_openvpn_static_key" "tls_crypt" {
  mode        = "crypt"
  description = "tls-crypt key for road-warrior VPN"
  key         = file("${path.module}/tls-crypt.key")
}
