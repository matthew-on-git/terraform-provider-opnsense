# Manage a basic HAProxy backend server
resource "opnsense_haproxy_server" "web1" {
  name    = "web-server-1"
  address = "10.0.0.10"
  port    = 80
  weight  = 100
}

# Manage an SSL-enabled backend server
resource "opnsense_haproxy_server" "api1" {
  name       = "api-server-1"
  address    = "10.0.0.20"
  port       = 443
  ssl        = true
  ssl_verify = true
  weight     = 50
}

# Manage a backup server (failover only)
resource "opnsense_haproxy_server" "backup" {
  name    = "backup-server"
  address = "10.0.0.99"
  port    = 80
  mode    = "backup"
}
