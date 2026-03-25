# NAT internal network to WAN IP
resource "opnsense_firewall_nat_outbound" "lan_to_wan" {
  interface   = "wan"
  source_net  = "10.0.0.0/24"
  target      = "wanip"
  description = "LAN outbound NAT"
}

# NAT with static port preservation (for SIP/VoIP)
resource "opnsense_firewall_nat_outbound" "voip" {
  interface       = "wan"
  protocol        = "udp"
  source_net      = "10.0.10.0/24"
  target          = "wanip"
  static_nat_port = true
  description     = "VoIP outbound with static port"
}

# No-NAT rule (exclude specific traffic from NAT)
resource "opnsense_firewall_nat_outbound" "vpn_no_nat" {
  interface       = "wan"
  source_net      = "10.0.0.0/24"
  destination_net = "10.1.0.0/24"
  target          = "wanip"
  no_nat          = true
  description     = "Skip NAT for VPN traffic"
}
