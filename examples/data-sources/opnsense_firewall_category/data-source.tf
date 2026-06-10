# Look up an existing firewall category by UUID
data "opnsense_firewall_category" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "firewall_category_id" {
  value = data.opnsense_firewall_category.example.id
}
