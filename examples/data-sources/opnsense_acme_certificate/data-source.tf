# Look up an existing ACME certificate by UUID
data "opnsense_acme_certificate" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "acme_certificate_id" {
  value = data.opnsense_acme_certificate.example.id
}
