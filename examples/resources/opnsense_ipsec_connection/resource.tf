resource "opnsense_ipsec_connection" "site2site" {
  description  = "Site-to-site VPN"
  remote_addrs = "203.0.113.1"
  version      = "2"
}
