resource "opnsense_dnsmasq_option" "pxe_tftp" {
  type        = "set"
  value       = "66,192.0.2.20"
  description = "PXE TFTP server option"
}
