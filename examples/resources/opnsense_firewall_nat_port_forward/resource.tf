# Forward HTTPS traffic from WAN to an internal web server
resource "opnsense_firewall_nat_port_forward" "web_https" {
  interface        = "wan"
  protocol         = "tcp"
  destination_net  = "wanip"
  destination_port = "443"
  target           = "10.0.0.10"
  local_port       = "443"
  description      = "HTTPS to web server"
}

# Forward SSH on a non-standard port to an internal server
resource "opnsense_firewall_nat_port_forward" "ssh" {
  interface        = "wan"
  protocol         = "tcp"
  destination_net  = "wanip"
  destination_port = "2222"
  target           = "10.0.0.20"
  local_port       = "22"
  description      = "SSH to management server"
  log              = true
}
