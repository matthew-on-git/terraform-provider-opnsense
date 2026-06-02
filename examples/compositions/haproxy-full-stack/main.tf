# A complete HAProxy stack: a server pool with a health check behind an
# HTTP frontend that routes by host header.

resource "opnsense_haproxy_healthcheck" "http_check" {
  name        = "http-health"
  type        = "http"
  http_method = "get"
  http_uri    = "/health"
  interval    = "5s"
}

resource "opnsense_haproxy_server" "web1" {
  name    = "web-1"
  address = "10.0.0.10"
  port    = 80
  weight  = 100
}

resource "opnsense_haproxy_server" "web2" {
  name    = "web-2"
  address = "10.0.0.11"
  port    = 80
  weight  = 100
}

resource "opnsense_haproxy_backend" "web_pool" {
  name                 = "web-pool"
  mode                 = "http"
  algorithm            = "roundrobin"
  linked_servers       = [opnsense_haproxy_server.web1.id, opnsense_haproxy_server.web2.id]
  health_check_enabled = true
  forward_for          = true

  # Health check is defined above; enable backend health checking here.
  depends_on = [opnsense_haproxy_healthcheck.http_check]
}

resource "opnsense_haproxy_acl" "site" {
  name       = "site-host"
  expression = "hdr_beg"
  hdr_beg    = "www.example.com"
}

resource "opnsense_haproxy_frontend" "http" {
  name            = "http-frontend"
  bind            = "0.0.0.0:80"
  mode            = "http"
  default_backend = opnsense_haproxy_backend.web_pool.id
  forward_for     = true
}
