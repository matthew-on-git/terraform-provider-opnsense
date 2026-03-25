resource "opnsense_quagga_route_map" "set_localpref" {
  name   = "set-localpref"
  action = "permit"
  order  = 10
  set    = "local-preference 200"
}
