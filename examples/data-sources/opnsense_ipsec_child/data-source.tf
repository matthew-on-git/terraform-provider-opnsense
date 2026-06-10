# Look up an existing ipsec child by UUID
data "opnsense_ipsec_child" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "ipsec_child_id" {
  value = data.opnsense_ipsec_child.example.id
}
