# Look up an existing ipsec connection by UUID
data "opnsense_ipsec_connection" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "ipsec_connection_id" {
  value = data.opnsense_ipsec_connection.example.id
}
