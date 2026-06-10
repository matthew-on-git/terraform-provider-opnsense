# Look up an existing system vlan by UUID
data "opnsense_system_vlan" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "system_vlan_id" {
  value = data.opnsense_system_vlan.example.id
}
