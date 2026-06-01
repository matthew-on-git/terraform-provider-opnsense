# Customer onboarding: terminate TLS for a new customer domain with an
# automatically-issued Let's Encrypt certificate and route it to a backend pool.

resource "opnsense_acme_account" "letsencrypt" {
  name  = "letsencrypt-prod"
  ca    = "letsencrypt"
  email = "ops@example.com"
}

resource "opnsense_acme_challenge" "http" {
  name   = "http-challenge"
  method = "http01"
}

resource "opnsense_acme_certificate" "customer" {
  name              = "shop.customer.com"
  alt_names         = "www.shop.customer.com"
  account           = opnsense_acme_account.letsencrypt.id
  validation_method = opnsense_acme_challenge.http.id
  key_length        = "key_4096"
}

resource "opnsense_haproxy_server" "customer_app" {
  name    = "customer-app-1"
  address = "10.20.0.10"
  port    = 8080
}

resource "opnsense_haproxy_backend" "customer_pool" {
  name           = "customer-pool"
  mode           = "http"
  algorithm      = "roundrobin"
  linked_servers = [opnsense_haproxy_server.customer_app.id]
  forward_for    = true
}

resource "opnsense_haproxy_acl" "customer_sni" {
  name       = "customer-sni"
  expression = "ssl_fc_sni"
  ssl_fc_sni = "shop.customer.com"
}

resource "opnsense_haproxy_frontend" "https" {
  name            = "https-frontend"
  bind            = "0.0.0.0:443"
  mode            = "http"
  default_backend = opnsense_haproxy_backend.customer_pool.id
  forward_for     = true
  description     = "Customer HTTPS entrypoint"
}
