resource "opnsense_interface_lagg" "uplink" {
  members     = ["em2", "em3"]
  protocol    = "failover"
  description = "Terraform managed LAGG"
}
