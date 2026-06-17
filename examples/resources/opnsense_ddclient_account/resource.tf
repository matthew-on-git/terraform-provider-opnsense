resource "opnsense_ddclient_account" "cloudflare" {
  service   = "cloudflare"
  hostnames = "home.example.com"
  check_ip  = "web_icanhazip"
  username  = "user@example.com"
  password  = "api-token-here"
}
