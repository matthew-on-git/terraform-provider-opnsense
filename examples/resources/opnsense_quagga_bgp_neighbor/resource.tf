# MetalLB BGP peering session
resource "opnsense_quagga_bgp_neighbor" "metallb" {
  address       = "10.0.0.2"
  remote_as     = 65001
  next_hop_self = true
  description   = "MetalLB speaker"
}
