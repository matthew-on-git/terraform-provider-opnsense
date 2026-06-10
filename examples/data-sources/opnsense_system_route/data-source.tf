# Look up an existing system route by UUID
data "opnsense_system_route" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "system_route_id" {
  value = data.opnsense_system_route.example.id
}
