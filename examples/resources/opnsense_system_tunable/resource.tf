resource "opnsense_system_tunable" "msgbuf_timestamp" {
  tunable     = "kern.msgbuf_show_timestamp"
  value       = "1"
  description = "Show timestamps in the kernel message buffer"
}
