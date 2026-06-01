# DNS management with Unbound: a local A record, an alias for it, and a
# conditional-forward (domain override) to an internal resolver.

resource "opnsense_unbound_general" "this" {
  enabled = true
  port    = "53"
  dnssec  = true
}

resource "opnsense_unbound_host_override" "app" {
  hostname = "app"
  domain   = "internal.example.com"
  rr       = "A"
  server   = "10.0.5.20"
}

resource "opnsense_unbound_host_alias" "app_alias" {
  host     = opnsense_unbound_host_override.app.id
  hostname = "www"
  domain   = "internal.example.com"
}

resource "opnsense_unbound_domain_override" "corp" {
  domain = "corp.example.com"
  server = "10.0.5.53"
}
