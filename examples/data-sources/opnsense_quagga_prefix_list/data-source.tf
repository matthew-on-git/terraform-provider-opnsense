# Look up an existing FRR prefix list by UUID
data "opnsense_quagga_prefix_list" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

output "quagga_prefix_list_id" {
  value = data.opnsense_quagga_prefix_list.example.id
}
