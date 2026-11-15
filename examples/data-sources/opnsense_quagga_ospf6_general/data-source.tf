# Read the OSPFv3 general settings singleton.
# Identity is ospf6. The lookup never writes or reconfigures.
data "opnsense_quagga_ospf6_general" "example" {
  id = "ospf6"
}

output "ospfv3_enabled" {
  value = data.opnsense_quagga_ospf6_general.example.enabled
}
