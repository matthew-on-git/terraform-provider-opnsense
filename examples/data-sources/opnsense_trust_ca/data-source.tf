# Look up an existing trust CA by UUID
data "opnsense_trust_ca" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "trust_ca_id" {
  value = data.opnsense_trust_ca.example.id
}
