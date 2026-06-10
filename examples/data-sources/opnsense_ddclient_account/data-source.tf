# Look up an existing dynamic DNS account by UUID
data "opnsense_ddclient_account" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "ddclient_account_id" {
  value = data.opnsense_ddclient_account.example.id
}
