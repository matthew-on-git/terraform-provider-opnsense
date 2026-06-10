# Look up an existing Unbound ACL by UUID
data "opnsense_unbound_acl" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "unbound_acl_id" {
  value = data.opnsense_unbound_acl.example.id
}
