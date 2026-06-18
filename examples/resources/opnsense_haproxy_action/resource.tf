# Route matching host-header traffic to a backend pool.
resource "opnsense_haproxy_action" "route_api" {
  name        = "route-api"
  type        = "use_backend"
  use_backend = opnsense_haproxy_backend.api.id
  linked_acls = [opnsense_haproxy_acl.api_domain.id]
  test_type   = "if"
}

# Redirect cleartext HTTP requests to HTTPS.
resource "opnsense_haproxy_action" "redirect_https" {
  name     = "redirect-to-https"
  type     = "http-request_redirect"
  redirect = "scheme https code 301"
}

# Set a header before forwarding to the backend.
resource "opnsense_haproxy_action" "forwarded_proto" {
  name               = "set-x-forwarded-proto"
  type               = "http-request_set-header"
  set_header_name    = "X-Forwarded-Proto"
  set_header_content = "https"
}
