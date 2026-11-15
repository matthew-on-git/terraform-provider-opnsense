# Read the FRR general settings singleton.
# Identity is general. The lookup never writes or reconfigures.
data "opnsense_quagga_general" "example" {
  id = "general"
}

output "frr_profile" {
  value = data.opnsense_quagga_general.example.profile
}
