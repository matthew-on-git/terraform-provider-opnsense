# Look up an existing ACME account by UUID
data "opnsense_acme_account" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "acme_account_id" {
  value = data.opnsense_acme_account.example.id
}
