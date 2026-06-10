data "opnsense_dnsmasq_domain" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "dnsmasq_domain_id" {
  value = data.opnsense_dnsmasq_domain.example.id
}
