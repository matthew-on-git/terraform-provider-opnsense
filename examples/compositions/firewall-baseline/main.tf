# A firewall baseline: reusable aliases plus a small allow/deny rule set.

resource "opnsense_firewall_alias" "internal_networks" {
  name        = "internal_networks"
  type        = "network"
  description = "Internal network ranges"
  content     = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
}

resource "opnsense_firewall_alias" "web_ports" {
  name        = "web_ports"
  type        = "port"
  description = "Public web ports"
  content     = ["80", "443"]
}

resource "opnsense_firewall_filter_rule" "allow_web" {
  action           = "pass"
  direction        = "in"
  protocol         = "tcp"
  destination_port = "443"
  description      = "Allow inbound HTTPS"
  log              = true
}

resource "opnsense_firewall_filter_rule" "block_rfc1918_in" {
  action      = "block"
  direction   = "in"
  source_net  = "10.0.0.0/8"
  log         = true
  description = "Drop spoofed internal sources on WAN"
}
