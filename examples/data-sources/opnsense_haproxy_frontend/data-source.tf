# Look up an existing haproxy frontend by UUID
data "opnsense_haproxy_frontend" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "haproxy_frontend_id" {
  value = data.opnsense_haproxy_frontend.example.id
}
