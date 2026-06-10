resource "opnsense_dnsmasq_boot" "pxe" {
  filename    = "pxelinux.0"
  server_name = "pxe"
  address     = "192.0.2.20"
  description = "PXE boot entry"
}
