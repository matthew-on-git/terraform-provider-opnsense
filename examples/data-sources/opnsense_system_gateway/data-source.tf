# Look up an existing system gateway by UUID
data "opnsense_system_gateway" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "system_gateway_id" {
  value = data.opnsense_system_gateway.example.id
}
