resource "opnsense_system_route" "k8s_pods" {
  network     = "10.244.0.0/16"
  gateway     = "Null4"
  description = "Kubernetes pod network"
}
