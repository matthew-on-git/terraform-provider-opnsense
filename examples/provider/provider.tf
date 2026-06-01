terraform {
  required_providers {
    opnsense = {
      source = "matthew-on-git/opnsense"
    }
  }
}

# Configure the OPNsense provider. Credentials may also be supplied via the
# OPNSENSE_URI, OPNSENSE_API_KEY, OPNSENSE_API_SECRET, and OPNSENSE_ALLOW_INSECURE
# environment variables (HCL values take priority).
provider "opnsense" {
  uri        = "https://opnsense.example.com"
  api_key    = var.opnsense_api_key
  api_secret = var.opnsense_api_secret
}

# Quickstart: a firewall alias and a rule that references it.
resource "opnsense_firewall_alias" "web_servers" {
  name        = "web_servers"
  type        = "host"
  description = "Public web servers"
  content     = ["10.0.0.10", "10.0.0.11"]
}

resource "opnsense_firewall_filter_rule" "allow_https" {
  action           = "pass"
  direction        = "in"
  protocol         = "tcp"
  destination_port = "443"
  description      = "Allow inbound HTTPS"
}
