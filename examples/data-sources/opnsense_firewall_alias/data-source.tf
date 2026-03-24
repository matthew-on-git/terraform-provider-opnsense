# Look up an existing firewall alias by UUID
data "opnsense_firewall_alias" "k8s_services" {
  id = "12345678-1234-1234-1234-123456789012"
}

# Reference the alias attributes in other resources
output "alias_name" {
  value = data.opnsense_firewall_alias.k8s_services.name
}

output "alias_content" {
  value = data.opnsense_firewall_alias.k8s_services.content
}
