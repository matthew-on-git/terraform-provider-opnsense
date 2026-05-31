resource "opnsense_quagga_bgp_global" "this" {
  enabled              = true
  as_number            = 65010
  router_id            = "10.0.0.1"
  networks             = ["10.0.0.0/24", "10.0.1.0/24"]
  log_neighbor_changes = true

  depends_on = [opnsense_quagga_general.this]
}
