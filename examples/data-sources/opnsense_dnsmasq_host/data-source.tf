data "opnsense_dnsmasq_host" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "dnsmasq_host_id" {
  value = data.opnsense_dnsmasq_host.example.id
}
