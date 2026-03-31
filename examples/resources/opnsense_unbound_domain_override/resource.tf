resource "opnsense_unbound_domain_override" "corp" {
  domain = "corp.example.com"
  server = "10.0.0.53"
}
