# Look up an existing firewall nat port forward by UUID
data "opnsense_firewall_nat_port_forward" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "firewall_nat_port_forward_id" {
  value = data.opnsense_firewall_nat_port_forward.example.id
}
