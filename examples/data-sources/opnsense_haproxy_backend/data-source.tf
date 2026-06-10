# Look up an existing haproxy backend by UUID
data "opnsense_haproxy_backend" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "haproxy_backend_id" {
  value = data.opnsense_haproxy_backend.example.id
}
