resource "opnsense_wireguard_server" "vpn" {
  name           = "wg0"
  port           = "51820"
  tunnel_address = "10.10.0.1/24"
  private_key    = "PRIVATE_KEY_HERE"
}
