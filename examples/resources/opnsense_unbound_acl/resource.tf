resource "opnsense_unbound_acl" "lan" {
  name     = "lan-access"
  action   = "allow"
  networks = "10.0.0.0/8"
}
