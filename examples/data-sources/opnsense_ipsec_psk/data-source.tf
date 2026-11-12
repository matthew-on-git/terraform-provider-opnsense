# Look up public metadata for an existing IPsec pre-shared key by UUID.
# The pre-shared key value is not lookup state.
data "opnsense_ipsec_psk" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "ipsec_psk_identity" {
  value = data.opnsense_ipsec_psk.example.identity
}
