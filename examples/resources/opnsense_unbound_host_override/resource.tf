resource "opnsense_unbound_host_override" "nas" {
  hostname = "nas"
  domain   = "home.lan"
  rr       = "A"
  server   = "10.0.0.50"
}
