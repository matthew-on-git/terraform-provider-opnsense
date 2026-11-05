resource "opnsense_trust_ca" "example" {
  description = "Example root CA"
  common_name = "root.example.invalid"
  country     = "US"
  action      = "internal"
  lifetime    = 365
}
