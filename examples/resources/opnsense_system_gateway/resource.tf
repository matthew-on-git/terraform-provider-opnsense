resource "opnsense_system_gateway" "wan_gw" {
  name      = "WAN_GW"
  interface = "wan"
  gateway   = "192.168.1.1"
}
