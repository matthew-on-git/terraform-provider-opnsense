resource "opnsense_system_vlan" "mgmt" {
  parent_interface = "vtnet0"
  tag              = 100
  device           = "vlan0100"
  description      = "Management VLAN"
}
