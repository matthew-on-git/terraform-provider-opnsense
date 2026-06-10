# Look up an existing wireguard peer by UUID
data "opnsense_wireguard_peer" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "wireguard_peer_id" {
  value = data.opnsense_wireguard_peer.example.id
}
