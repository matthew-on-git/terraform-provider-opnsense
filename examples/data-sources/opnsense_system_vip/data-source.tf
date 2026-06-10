# Look up an existing system vip by UUID
data "opnsense_system_vip" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "system_vip_id" {
  value = data.opnsense_system_vip.example.id
}
