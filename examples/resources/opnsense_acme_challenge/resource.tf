resource "opnsense_acme_challenge" "http" {
  name   = "http-challenge"
  method = "http01"
}
