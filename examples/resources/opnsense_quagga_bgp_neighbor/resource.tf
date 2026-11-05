# MetalLB BGP peering session
resource "opnsense_quagga_bgp_neighbor" "metallb" {
  address             = "192.0.2.2"
  remote_as           = 65001
  family              = ["ipv4"]
  next_hop_self       = true
  weight              = 0
  connect_timer       = 30
  attribute_unchanged = "as-path"
  capabilities        = ["dynamic"]
  # Authentication is config-owned Sensitive input; omission never adopts it.
  description = "MetalLB speaker"
}
