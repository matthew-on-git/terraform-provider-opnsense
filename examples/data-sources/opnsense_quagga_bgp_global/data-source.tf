# Read the BGP global settings singleton.
# Identity is bgp. The lookup never writes or reconfigures.
data "opnsense_quagga_bgp_global" "example" {
  id = "bgp"
}

output "bgp_as_number" {
  value = data.opnsense_quagga_bgp_global.example.as_number
}
