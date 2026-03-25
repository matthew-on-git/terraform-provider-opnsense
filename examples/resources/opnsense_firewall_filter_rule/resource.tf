# Allow inbound HTTPS traffic
resource "opnsense_firewall_filter_rule" "allow_https" {
  action           = "pass"
  direction        = "in"
  protocol         = "tcp"
  destination_port = "443"
  description      = "Allow HTTPS"
  log              = true
}

# Block all traffic from a specific network
resource "opnsense_firewall_filter_rule" "block_untrusted" {
  action     = "block"
  direction  = "in"
  source_net = "10.99.0.0/24"
  source_not = false
  log        = true
  description = "Block untrusted network"
}

# Allow DNS traffic to specific servers
resource "opnsense_firewall_filter_rule" "allow_dns" {
  action           = "pass"
  direction        = "out"
  protocol         = "TCP/UDP"
  destination_net  = "1.1.1.1"
  destination_port = "53"
  description      = "Allow DNS to Cloudflare"
}
