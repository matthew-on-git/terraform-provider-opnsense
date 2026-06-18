resource "opnsense_acme_certificate" "web" {
  name              = "www.example.com"
  alt_names         = "example.com,api.example.com"
  account           = opnsense_acme_account.letsencrypt.id
  validation_method = opnsense_acme_challenge.http.id
  key_length        = "key_4096"

  # ACME issuance is asynchronous. The provider signs and polls until OPNsense
  # reports status_code = "200" with a non-empty cert_ref_id.
  issuance_timeout       = "180s"
  issuance_poll_interval = "10s"
}

resource "opnsense_haproxy_frontend" "https" {
  name                = "public_https"
  bind                = "0.0.0.0:443"
  mode                = "http"
  default_backend     = opnsense_haproxy_backend.web.id
  ssl_enabled         = true
  certificates        = [opnsense_acme_certificate.web.cert_ref_id]
  default_certificate = opnsense_acme_certificate.web.cert_ref_id
}
