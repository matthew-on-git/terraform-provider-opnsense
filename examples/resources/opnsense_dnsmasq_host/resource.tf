resource "opnsense_dnsmasq_host" "pxe_server" {
  host        = "pxe"
  domain      = "example.com"
  ip          = "192.0.2.10"
  description = "PXE server host override"
}
