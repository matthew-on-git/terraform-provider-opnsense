# Look up an existing haproxy server by UUID
data "opnsense_haproxy_server" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "haproxy_server_id" {
  value = data.opnsense_haproxy_server.example.id
}
