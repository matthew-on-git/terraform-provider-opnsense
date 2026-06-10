# Look up an existing FRR static route by UUID
data "opnsense_quagga_static_route" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "quagga_static_route_id" {
  value = data.opnsense_quagga_static_route.example.id
}
