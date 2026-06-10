# Look up an existing Unbound host alias by UUID
data "opnsense_unbound_host_alias" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "unbound_host_alias_id" {
  value = data.opnsense_unbound_host_alias.example.id
}
