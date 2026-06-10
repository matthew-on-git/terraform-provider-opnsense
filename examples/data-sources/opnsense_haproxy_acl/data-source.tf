# Look up an existing haproxy acl by UUID
data "opnsense_haproxy_acl" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "haproxy_acl_id" {
  value = data.opnsense_haproxy_acl.example.id
}
