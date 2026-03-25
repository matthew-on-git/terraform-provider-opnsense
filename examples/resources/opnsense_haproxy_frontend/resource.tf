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

# HTTP frontend routing to the backend pool
resource "opnsense_haproxy_frontend" "http" {
  name            = "http-frontend"
  bind            = "0.0.0.0:80"
  mode            = "http"
  default_backend = opnsense_haproxy_backend.web_pool.id
  forward_for     = true
  description     = "HTTP traffic"
}

# HTTPS frontend with SSL offloading
resource "opnsense_haproxy_frontend" "https" {
  name            = "https-frontend"
  bind            = "0.0.0.0:443"
  mode            = "http"
  default_backend = opnsense_haproxy_backend.web_pool.id
  ssl_enabled     = true
  forward_for     = true
  description     = "HTTPS traffic with SSL offloading"
}
