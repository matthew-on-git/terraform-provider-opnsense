resource "opnsense_openvpn_instance" "server" {
  role         = "server"
  description  = "Road-warrior VPN"
  protocol     = "udp"
  dev_type     = "tun"
  port         = "1194"
  server       = "10.10.8.0/24"
  topology     = "subnet"
  ca           = opnsense_trust_ca.vpn.id
  cert         = opnsense_trust_cert.vpn_server.id
  data_ciphers = ["AES-256-GCM", "CHACHA20-POLY1305"]
  auth         = "SHA256"
  dns_servers  = ["10.10.8.1"]
  push_route   = ["192.168.1.0/24"]
}
