resource "opnsense_ipsec_psk" "site2site" {
  identity = "203.0.113.1"
  key      = "super-secret-psk"
}
