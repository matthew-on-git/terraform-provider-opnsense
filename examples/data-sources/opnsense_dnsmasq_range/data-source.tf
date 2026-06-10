data "opnsense_dnsmasq_range" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "dnsmasq_range_id" {
  value = data.opnsense_dnsmasq_range.example.id
}
