resource "opnsense_quagga_general" "this" {
  enabled       = true
  profile       = "datacenter"
  enable_syslog = true
  syslog_level  = "notifications"
}
