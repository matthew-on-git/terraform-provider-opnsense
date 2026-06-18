data "opnsense_ddclient_settings" "daemon" {
  id = "ddclient-settings"
}

output "ddclient_backend" {
  value = data.opnsense_ddclient_settings.daemon.backend
}
