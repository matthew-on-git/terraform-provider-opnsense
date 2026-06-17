# Domain map used by a map_use_backend HAProxy action.
resource "opnsense_haproxy_mapfile" "domain_map" {
  name        = "domain-map"
  description = "Host to backend routing map"
  type        = "dom"
  content     = <<-EOT
    grafana.example.com grafana-backend
    argocd.example.com argocd-backend
    tipsyhive.example.com tipsyhive-backend
    www.tipsyhive.example.com tipsyhive-backend
  EOT
}
