# Look up an existing openvpn client overwrite by UUID
data "opnsense_openvpn_client_overwrite" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "openvpn_client_overwrite_id" {
  value = data.opnsense_openvpn_client_overwrite.example.id
}
