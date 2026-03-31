resource "opnsense_acme_account" "letsencrypt" {
  name  = "letsencrypt-prod"
  ca    = "letsencrypt"
  email = "admin@example.com"
}
