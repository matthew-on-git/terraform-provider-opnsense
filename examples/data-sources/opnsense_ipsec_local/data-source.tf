# Look up an existing ipsec local by UUID
data "opnsense_ipsec_local" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "ipsec_local_id" {
  value = data.opnsense_ipsec_local.example.id
}
