# Read the Unbound DNSBL compatibility owner.
# Identity is unbound-dnsbl. The lookup never writes or reconfigures.
data "opnsense_unbound_dnsbl" "example" {
  id = "unbound-dnsbl"
}

output "unbound_dnsbl_enabled" {
  value = data.opnsense_unbound_dnsbl.example.enabled
}
