# Match requests where Host header starts with a domain
resource "opnsense_haproxy_acl" "api_domain" {
  name       = "api-domain"
  expression = "hdr_beg"
  hdr_beg    = "api.example.com"
}

# Match requests to a specific path prefix
resource "opnsense_haproxy_acl" "admin_path" {
  name       = "admin-path"
  expression = "path_beg"
  path_beg   = "/admin"
}

# Match by SNI for SSL passthrough
resource "opnsense_haproxy_acl" "ssl_domain" {
  name       = "ssl-domain"
  expression = "ssl_fc_sni"
  ssl_fc_sni = "secure.example.com"
}

# Match by source IP
resource "opnsense_haproxy_acl" "trusted_src" {
  name       = "trusted-source"
  expression = "src"
  src        = "10.0.0.0/8"
}
