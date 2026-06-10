# Look up an existing BGP AS path rule by UUID
data "opnsense_quagga_bgp_aspath" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "quagga_bgp_aspath_id" {
  value = data.opnsense_quagga_bgp_aspath.example.id
}
