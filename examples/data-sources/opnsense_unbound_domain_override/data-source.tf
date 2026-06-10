# Look up an existing Unbound domain override by UUID
data "opnsense_unbound_domain_override" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "unbound_domain_override_id" {
  value = data.opnsense_unbound_domain_override.example.id
}
