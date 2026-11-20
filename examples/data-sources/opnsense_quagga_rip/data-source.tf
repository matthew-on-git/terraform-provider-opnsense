# Read the RIP settings singleton.
# Identity is rip. The lookup never writes or reconfigures.
data "opnsense_quagga_rip" "example" {
  id = "rip"
}

output "rip_enabled" {
  value = data.opnsense_quagga_rip.example.enabled
}
