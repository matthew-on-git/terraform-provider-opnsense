resource "opnsense_ddclient_settings" "daemon" {
  enabled    = true
  backend    = "opnsense"
  interval   = 300
  verbose    = false
  allow_ipv6 = true
}

resource "opnsense_ddclient_account" "cloudflare" {
  enabled   = true
  service   = "cloudflare"
  hostnames = "www.example.com"
  username  = "cloudflare-account@example.com"
  password  = var.cloudflare_api_token

  depends_on = [opnsense_ddclient_settings.daemon]
}
