# Look up an existing wireguard server by UUID
data "opnsense_wireguard_server" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "wireguard_server_id" {
  value = data.opnsense_wireguard_server.example.id
}
