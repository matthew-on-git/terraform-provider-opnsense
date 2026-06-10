# Look up an existing BGP redistribution rule by UUID
data "opnsense_quagga_bgp_redistribution" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "quagga_bgp_redistribution_id" {
  value = data.opnsense_quagga_bgp_redistribution.example.id
}
