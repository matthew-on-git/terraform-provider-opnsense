# Look up an existing BGP peer group by UUID
data "opnsense_quagga_bgp_peergroup" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "quagga_bgp_peergroup_id" {
  value = data.opnsense_quagga_bgp_peergroup.example.id
}
