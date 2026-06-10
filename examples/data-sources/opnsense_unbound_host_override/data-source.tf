# Look up an existing Unbound host override by UUID
data "opnsense_unbound_host_override" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "unbound_host_override_id" {
  value = data.opnsense_unbound_host_override.example.id
}
