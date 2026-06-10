# Look up an existing BGP neighbor by UUID
data "opnsense_quagga_bgp_neighbor" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "quagga_bgp_neighbor_id" {
  value = data.opnsense_quagga_bgp_neighbor.example.id
}
