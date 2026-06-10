# Look up an existing haproxy healthcheck by UUID
data "opnsense_haproxy_healthcheck" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "haproxy_healthcheck_id" {
  value = data.opnsense_haproxy_healthcheck.example.id
}
