# Read the OSPF general settings singleton.
# Identity is ospf. The lookup never writes or reconfigures.
data "opnsense_quagga_ospf_general" "example" {
  id = "ospf"
}

output "ospf_enabled" {
  value = data.opnsense_quagga_ospf_general.example.enabled
}
