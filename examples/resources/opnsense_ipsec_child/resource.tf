resource "opnsense_ipsec_child" "tunnel" {
  connection_id = opnsense_ipsec_connection.site2site.id
  local_ts   = "10.0.0.0/24"
  remote_ts  = "10.1.0.0/24"
  mode       = "tunnel"
}
