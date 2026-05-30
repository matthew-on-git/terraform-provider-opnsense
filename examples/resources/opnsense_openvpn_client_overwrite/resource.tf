resource "opnsense_openvpn_client_overwrite" "branch_office" {
  common_name     = "branch-office.example.com"
  description     = "Site-to-site route injection for branch office"
  servers         = [opnsense_openvpn_instance.server.id]
  tunnel_network  = "10.10.8.50/32"
  remote_networks = ["192.168.50.0/24"]
}
