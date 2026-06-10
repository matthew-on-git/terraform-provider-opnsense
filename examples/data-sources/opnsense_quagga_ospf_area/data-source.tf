data "opnsense_quagga_ospf_area" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "quagga_ospf_area_id" {
  value = data.opnsense_quagga_ospf_area.example.id
}
