# Read the Kea DHCP control agent singleton.
# Identity is kea-ctrl-agent. The lookup never writes or reconfigures.
data "opnsense_kea_ctrl_agent" "example" {
  id = "kea-ctrl-agent"
}

output "kea_ctrl_agent_http_port" {
  value = data.opnsense_kea_ctrl_agent.example.http_port
}
