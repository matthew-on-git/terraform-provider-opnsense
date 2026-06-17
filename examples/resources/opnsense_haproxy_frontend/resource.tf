# Create the full server → backend → frontend chain
resource "opnsense_haproxy_server" "web1" {
  name    = "web-server-1"
  address = "10.0.0.10"
  port    = 80
}

resource "opnsense_haproxy_backend" "web_pool" {
  name           = "web-pool"
  mode           = "http"
  algorithm      = "roundrobin"
  linked_servers = [opnsense_haproxy_server.web1.id]
}

resource "opnsense_haproxy_backend" "api_pool" {
  name           = "api-pool"
  mode           = "http"
  algorithm      = "roundrobin"
  linked_servers = [opnsense_haproxy_server.web1.id]
}

resource "opnsense_haproxy_acl" "api_host" {
  name       = "api-host"
  expression = "hdr"
  hdr        = "api.example.com"
}

resource "opnsense_haproxy_action" "route_api" {
  name        = "route-api"
  type        = "use_backend"
  use_backend = opnsense_haproxy_backend.api_pool.id
  linked_acls = [opnsense_haproxy_acl.api_host.id]
  test_type   = "if"
}

# HTTP frontend routing to the backend pool with an ACL-based action
resource "opnsense_haproxy_frontend" "http" {
  name            = "http-frontend"
  bind            = "0.0.0.0:80"
  mode            = "http"
  default_backend = opnsense_haproxy_backend.web_pool.id
  linked_actions  = [opnsense_haproxy_action.route_api.id]
  forward_for     = true
  description     = "HTTP traffic"
}

# HTTPS frontend with SSL offloading
resource "opnsense_haproxy_frontend" "https" {
  name                = "https-frontend"
  bind                = "0.0.0.0:443"
  mode                = "http"
  default_backend     = opnsense_haproxy_backend.web_pool.id
  ssl_enabled         = true
  certificates        = [var.haproxy_certificate_refid]
  default_certificate = var.haproxy_certificate_refid
  forward_for         = true
  description         = "HTTPS traffic with SSL offloading"
}

variable "haproxy_certificate_refid" {
  description = "HAProxy certificate refid, not the certificate API UUID. Story 29.4 will expose this as opnsense_acme_certificate.cert_ref_id."
  type        = string
}
