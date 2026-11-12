data "opnsense_dnsmasq_settings" "example" {
  id = "dnsmasq"
}

output "dnsmasq_enabled" {
  value = data.opnsense_dnsmasq_settings.example.enable
}
