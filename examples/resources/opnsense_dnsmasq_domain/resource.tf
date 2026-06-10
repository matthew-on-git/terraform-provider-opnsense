resource "opnsense_dnsmasq_domain" "corp" {
  domain      = "corp.example.com"
  ip          = "192.0.2.53"
  description = "Forward corporate domain queries"
}
