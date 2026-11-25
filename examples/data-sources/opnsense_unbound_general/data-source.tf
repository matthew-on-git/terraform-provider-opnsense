# Read Unbound general settings.
# Identity is unbound-general. The lookup never writes or reconfigures.
data "opnsense_unbound_general" "example" {
  id = "unbound-general"
}

output "unbound_general_enabled" {
  value = data.opnsense_unbound_general.example.enabled
}
