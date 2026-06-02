# A WireGuard server with a single road-warrior peer.

resource "opnsense_wireguard_server" "vpn" {
  name = "wg-vpn"
  # Generate with `wg genkey`; supply via a variable or secret store, not in HCL.
  private_key    = var.wireguard_server_private_key
  tunnel_address = "10.10.0.1/24"
  port           = "51820"
  description    = "Road-warrior WireGuard server"
}

resource "opnsense_wireguard_peer" "laptop" {
  name           = "laptop"
  public_key     = var.laptop_public_key # peer's `wg pubkey` output
  tunnel_address = "10.10.0.2/32"
  keepalive      = 25

  depends_on = [opnsense_wireguard_server.vpn]
}

variable "wireguard_server_private_key" {
  type        = string
  sensitive   = true
  description = "WireGuard server private key (output of `wg genkey`)."
}

variable "laptop_public_key" {
  type        = string
  description = "Public key of the laptop peer (output of `wg pubkey`)."
}
