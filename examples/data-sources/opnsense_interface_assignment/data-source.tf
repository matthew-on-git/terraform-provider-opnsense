data "opnsense_interface_assignment" "example" {
  id = "opt3"
}

output "assignment_ipv4_mode" {
  value = data.opnsense_interface_assignment.example.ipv4_mode
}
