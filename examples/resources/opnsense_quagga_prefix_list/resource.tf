resource "opnsense_quagga_prefix_list" "allow_k8s" {
  name     = "allow-k8s-pods"
  sequence = 10
  action   = "permit"
  network  = "10.244.0.0/16 le 32"
}
