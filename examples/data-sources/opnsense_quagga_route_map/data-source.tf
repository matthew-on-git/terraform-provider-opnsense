# Look up an existing FRR route map by UUID
data "opnsense_quagga_route_map" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "quagga_route_map_id" {
  value = data.opnsense_quagga_route_map.example.id
}
