resource "opnsense_ipsec_connection" "site2site" {
  description              = "Disabled advanced IPsec example"
  enabled                  = false
  remote_addrs             = "203.0.113.1"
  local_addresses          = "192.0.2.1"
  version                  = "2"
  aggressive               = false
  dpd_delay                = "0"
  dpd_timeout              = "120"
  ike_lifetime             = "30" # Overtime grace, NOT total IKE lifetime.
  keying_tries             = "0"
  reauthentication_time    = ""
  rekey_time               = "3600"
  local_port               = "" # Daemon port 500; literal "500" is invalid.
  remote_port              = "4500"
  mobike                   = false
  send_certificate         = "never"
  send_certificate_request = false
  udp_encapsulation        = true
  # ip_pools may reference existing appliance pool UUIDs (or built-in radius); omitted values survive import.
}
