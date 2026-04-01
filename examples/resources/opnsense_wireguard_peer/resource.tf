resource "opnsense_wireguard_peer" "client1" {
  name           = "client1"
  public_key     = "PUBLIC_KEY_HERE"
  tunnel_address = "10.10.0.2/32"
  keepalive      = 25
}
