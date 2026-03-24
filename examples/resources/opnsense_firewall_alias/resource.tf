# Manage a host alias for web servers
resource "opnsense_firewall_alias" "web_servers" {
  name        = "web_servers"
  type        = "host"
  description = "Web server addresses"
  content     = ["10.0.0.10", "10.0.0.11", "10.0.0.12"]
}

# Manage a network alias for internal subnets
resource "opnsense_firewall_alias" "internal_networks" {
  name        = "internal_networks"
  type        = "network"
  description = "Internal network ranges"
  content     = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
}

# Manage a port alias for common services
resource "opnsense_firewall_alias" "common_ports" {
  name        = "common_ports"
  type        = "port"
  description = "Common service ports"
  content     = ["80", "443", "8080"]
}
