# Look up public metadata for an existing IPsec key pair by UUID.
# Public and private key PEM payloads are not lookup state.
data "opnsense_ipsec_key_pair" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "ipsec_key_pair_name" {
  value = data.opnsense_ipsec_key_pair.example.name
}
