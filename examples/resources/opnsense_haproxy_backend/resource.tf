# Create servers for the backend pool
resource "opnsense_haproxy_server" "web1" {
  name    = "web-server-1"
  address = "10.0.0.10"
  port    = 80
}

resource "opnsense_haproxy_server" "web2" {
  name    = "web-server-2"
  address = "10.0.0.11"
  port    = 80
}

# Create a backend pool linking to the servers
resource "opnsense_haproxy_backend" "web_pool" {
  name           = "web-pool"
  mode           = "http"
  algorithm      = "roundrobin"
  linked_servers = [
    opnsense_haproxy_server.web1.id,
    opnsense_haproxy_server.web2.id,
  ]
  forward_for = true
  description = "Web server pool"
}
