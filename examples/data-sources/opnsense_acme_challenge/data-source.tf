# Look up an existing ACME challenge by UUID
data "opnsense_acme_challenge" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "acme_challenge_id" {
  value = data.opnsense_acme_challenge.example.id
}
