# Read the FRR static routing settings singleton.
# Identity is static. The lookup never writes or reconfigures.
data "opnsense_quagga_static" "example" {
  id = "static"
}

output "quagga_static_enabled" {
  value = data.opnsense_quagga_static.example.enabled
}
