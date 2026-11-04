variable "certificate_pem" {
  type        = string
  description = "Existing PEM certificate to import."
}

variable "private_key_pem" {
  type        = string
  sensitive   = true
  description = "Matching private key supplied through a secret input."
}

resource "opnsense_trust_cert" "example" {
  description = "Example imported certificate"
  certificate = var.certificate_pem
  private_key = var.private_key_pem
}
