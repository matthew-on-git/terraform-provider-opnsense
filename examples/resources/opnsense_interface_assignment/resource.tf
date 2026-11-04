resource "opnsense_interface_assignment" "pxe" {
  # The assignment must already exist. Use only a non-management interface
  # whose addressing can safely be changed.
  name               = "opt3"
  enabled            = true
  ipv4_mode          = "staticv4"
  ipv4_address       = "192.0.2.1"
  ipv4_prefix_length = 24
  ipv4_gateway       = ""
  ipv6_mode          = "none"
}

# Existing downstream interface using a delegated IPv6 prefix. Import first;
# replace these values with the intended configuration only after reviewing state.
resource "opnsense_interface_assignment" "delegated" {
  name                         = "opt4"
  ipv6_mode                    = "idassoc6"
  ipv6_track_interface         = "wan"
  ipv6_prefix_id               = "0x10"
  ipv6_interface_id            = "0x1"
  ipv6_identity_association_id = "0"
  ipv6_prefix_range            = ""
}
