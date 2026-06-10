resource "opnsense_quagga_ospf_area" "stub_area" {
  area_id = "0.0.0.1"
  type    = "stub"
}
