resource "opnsense_ipsec_child" "tunnel" {
  connection_id    = opnsense_ipsec_connection.site2site.id
  enabled          = false
  local_ts         = "192.0.2.0/24"
  remote_ts        = "198.51.100.0/24"
  mode             = "tunnel"
  esp_proposals    = "aes256-sha256,aes128-sha256"
  start_action     = "trap|start"
  close_action     = "none"
  dpd_action       = "clear"
  install_policies = false
  request_id       = "" # Explicit clear; omission preserves imported values.
  sha256_96        = false
  rekey_time       = "0" # Cannot be blank; omission on creation uses server 3600.
}
