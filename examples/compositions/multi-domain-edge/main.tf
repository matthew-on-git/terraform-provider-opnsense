# Multi-domain HTTPS edge migration example.
#
# This composition models a single OPNsense HAProxy edge serving several hostnames
# through ACME-issued TLS, host-to-backend map routing, explicit ACL routing, and
# internal-only deny rules. Values use documentation-only domains and IP ranges.

terraform {
  required_providers {
    opnsense = {
      source = "matthew-on-git/opnsense"
    }
  }
}

variable "acme_email" {
  type        = string
  description = "Contact email for the ACME account."
  default     = "ops@example.com"
}

resource "opnsense_acme_account" "edge" {
  name  = "multi-domain-edge"
  email = var.acme_email
  ca    = "letsencrypt_test"
}

resource "opnsense_acme_challenge" "dns" {
  name        = "edge-dns01"
  method      = "dns01"
  dns_service = "dns_cf"
  dns_sleep   = 30
}

resource "opnsense_acme_certificate" "edge" {
  name                   = "grafana.example.com"
  description            = "Example multi-domain edge certificate"
  alt_names              = "argocd.example.com,tipsyhive.example.com,thetipsyhive.example,tipsyhive.example"
  account                = opnsense_acme_account.edge.id
  validation_method      = opnsense_acme_challenge.dns.id
  key_length             = "key_ec256"
  auto_renewal           = true
  issuance_timeout       = "180s"
  issuance_poll_interval = "10s"
}

resource "opnsense_haproxy_server" "grafana" {
  name    = "grafana-01"
  address = "192.0.2.10"
  port    = 3000
  weight  = 100
}

resource "opnsense_haproxy_server" "argocd" {
  name    = "argocd-01"
  address = "192.0.2.20"
  port    = 8080
  weight  = 100
}

resource "opnsense_haproxy_server" "tipsyhive" {
  name    = "tipsyhive-01"
  address = "192.0.2.30"
  port    = 8080
  weight  = 100
}

resource "opnsense_haproxy_backend" "grafana" {
  name                 = "grafana-backend"
  mode                 = "http"
  algorithm            = "roundrobin"
  linked_servers       = [opnsense_haproxy_server.grafana.id]
  health_check_enabled = true
  forward_for          = true
}

resource "opnsense_haproxy_backend" "argocd" {
  name                 = "argocd-backend"
  mode                 = "http"
  algorithm            = "roundrobin"
  linked_servers       = [opnsense_haproxy_server.argocd.id]
  health_check_enabled = true
  forward_for          = true
}

resource "opnsense_haproxy_backend" "tipsyhive" {
  name                 = "tipsyhive-backend"
  mode                 = "http"
  algorithm            = "roundrobin"
  linked_servers       = [opnsense_haproxy_server.tipsyhive.id]
  health_check_enabled = true
  forward_for          = true
}

resource "opnsense_haproxy_mapfile" "domain_map" {
  name        = "domain-map"
  description = "Host-to-backend routing map for the migrated edge"
  type        = "dom"
  content     = <<-EOT
    grafana.example.com grafana-backend
    argocd.example.com argocd-backend
    tipsyhive.example.com tipsyhive-backend
    thetipsyhive.example tipsyhive-backend
    tipsyhive.example tipsyhive-backend
  EOT
}

resource "opnsense_haproxy_acl" "is_tipsyhive_host" {
  name       = "is-tipsyhive-host"
  expression = "hdr"
  hdr        = "tipsyhive.example.com"
}

resource "opnsense_haproxy_acl" "is_protected_host" {
  name       = "is-protected-host"
  expression = "custom_acl"
  custom_acl = "hdr(host) -i grafana.example.com argocd.example.com"
}

resource "opnsense_haproxy_acl" "is_external_source" {
  name       = "is-external-source"
  expression = "src"
  negate     = true
  src        = "10.0.0.0/8 172.16.0.0/12 192.168.0.0/16"
}

resource "opnsense_haproxy_action" "deny_external_protected" {
  name        = "deny-external-protected"
  type        = "http-request_deny"
  test_type   = "if"
  operator    = "and"
  linked_acls = [opnsense_haproxy_acl.is_protected_host.id, opnsense_haproxy_acl.is_external_source.id]
  deny_status = 403
}

resource "opnsense_haproxy_action" "set_forwarded_proto" {
  name               = "set-forwarded-proto"
  type               = "http-request_set-header"
  set_header_name    = "X-Forwarded-Proto"
  set_header_content = "https"
}

resource "opnsense_haproxy_action" "route_tipsyhive_host" {
  name        = "route-tipsyhive-host"
  type        = "use_backend"
  use_backend = opnsense_haproxy_backend.tipsyhive.id
  linked_acls = [opnsense_haproxy_acl.is_tipsyhive_host.id]
}

resource "opnsense_haproxy_action" "route_domain_map" {
  name                    = "route-domain-map"
  type                    = "map_use_backend"
  mapfile                 = opnsense_haproxy_mapfile.domain_map.id
  map_use_backend_default = opnsense_haproxy_backend.tipsyhive.id
}

resource "opnsense_haproxy_action" "redirect_https" {
  name     = "redirect-https"
  type     = "http-request_redirect"
  redirect = "scheme https code 301"
}

resource "opnsense_haproxy_frontend" "http_in" {
  name           = "http-in"
  bind           = "0.0.0.0:80"
  mode           = "http"
  linked_actions = [opnsense_haproxy_action.redirect_https.id]
}

resource "opnsense_haproxy_frontend" "https_in" {
  name                = "https-in"
  bind                = "0.0.0.0:443"
  mode                = "http"
  default_backend     = opnsense_haproxy_backend.tipsyhive.id
  ssl_enabled         = true
  certificates        = [opnsense_acme_certificate.edge.cert_ref_id]
  default_certificate = opnsense_acme_certificate.edge.cert_ref_id
  linked_actions = [
    opnsense_haproxy_action.deny_external_protected.id,
    opnsense_haproxy_action.set_forwarded_proto.id,
    opnsense_haproxy_action.route_tipsyhive_host.id,
    opnsense_haproxy_action.route_domain_map.id,
  ]
  forward_for = true
}
