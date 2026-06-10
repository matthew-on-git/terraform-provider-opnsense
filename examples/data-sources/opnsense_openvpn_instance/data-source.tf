# Look up an existing openvpn instance by UUID
data "opnsense_openvpn_instance" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "openvpn_instance_id" {
  value = data.opnsense_openvpn_instance.example.id
}
