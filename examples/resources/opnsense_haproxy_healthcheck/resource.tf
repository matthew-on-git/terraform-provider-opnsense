# HTTP health check on a custom path
resource "opnsense_haproxy_healthcheck" "http_check" {
  name        = "http-health"
  type        = "http"
  http_method = "get"
  http_uri    = "/health"
  interval    = "5s"
}

# Simple TCP health check
resource "opnsense_haproxy_healthcheck" "tcp_check" {
  name     = "tcp-health"
  type     = "tcp"
  interval = "3s"
}
