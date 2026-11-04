# Import existing settings before applying. Omitted arguments preserve appliance policy.
resource "opnsense_trust_settings" "policy" {
  install_crls  = false
  cipher_string = [] # Explicitly clear the selected ciphers.
}
