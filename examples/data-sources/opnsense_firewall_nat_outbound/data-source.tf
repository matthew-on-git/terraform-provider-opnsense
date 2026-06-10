# Look up an existing firewall nat outbound by UUID
data "opnsense_firewall_nat_outbound" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "firewall_nat_outbound_id" {
  value = data.opnsense_firewall_nat_outbound.example.id
}
