# Look up an existing BGP community list rule by UUID
data "opnsense_quagga_bgp_communitylist" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "quagga_bgp_communitylist_id" {
  value = data.opnsense_quagga_bgp_communitylist.example.id
}
