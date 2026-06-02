# BGP peering for MetalLB / route advertisement: enable FRR, configure the BGP
# AS and advertised networks, and peer with a downstream speaker.

resource "opnsense_quagga_general" "this" {
  enabled       = true
  profile       = "datacenter"
  enable_syslog = true
  syslog_level  = "notifications"
}

resource "opnsense_quagga_bgp_global" "this" {
  enabled              = true
  as_number            = 65010
  router_id            = "10.0.0.1"
  networks             = ["10.0.0.0/24", "10.0.1.0/24"]
  log_neighbor_changes = true

  depends_on = [opnsense_quagga_general.this]
}

resource "opnsense_quagga_bgp_neighbor" "metallb" {
  address       = "10.0.0.2"
  remote_as     = 65001
  next_hop_self = true
  description   = "MetalLB speaker"

  depends_on = [opnsense_quagga_bgp_global.this]
}
