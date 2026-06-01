# A WireGuard server with a single road-warrior peer.

resource "opnsense_wireguard_server" "vpn" {
  name           = "wg-vpn"
  private_key    = "QENU8wQqd0g0Hbn6N1Y8b3pZ5g5w2yE4xj0X3o2Hm0=" # example only
  tunnel_address = "10.10.0.1/24"
  port           = "51820"
  description    = "Road-warrior WireGuard server"
}

resource "opnsense_wireguard_peer" "laptop" {
  name           = "laptop"
  public_key     = "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=" # example only
  tunnel_address = "10.10.0.2/32"
  keepalive      = 25

  depends_on = [opnsense_wireguard_server.vpn]
}
