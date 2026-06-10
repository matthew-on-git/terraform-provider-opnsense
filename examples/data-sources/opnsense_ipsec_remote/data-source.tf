# Look up an existing ipsec remote by UUID
data "opnsense_ipsec_remote" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "ipsec_remote_id" {
  value = data.opnsense_ipsec_remote.example.id
}
