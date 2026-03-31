resource "opnsense_ddclient_account" "cloudflare" {
  service   = "cloudflare"
  hostnames = "home.example.com"
  username  = "user@example.com"
  password  = "api-token-here"
}
