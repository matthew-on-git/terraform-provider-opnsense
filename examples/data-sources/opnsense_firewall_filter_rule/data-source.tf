# Look up an existing firewall filter rule by UUID
data "opnsense_firewall_filter_rule" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "firewall_filter_rule_id" {
  value = data.opnsense_firewall_filter_rule.example.id
}
