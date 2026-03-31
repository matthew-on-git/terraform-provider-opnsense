resource "opnsense_acme_certificate" "web" {
  name              = "www.example.com"
  alt_names         = "example.com,api.example.com"
  account           = opnsense_acme_account.letsencrypt.id
  validation_method = opnsense_acme_challenge.http.id
  key_length        = "key_4096"
}
