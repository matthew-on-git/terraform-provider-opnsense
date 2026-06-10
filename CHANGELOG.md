# Changelog

All notable changes to the OPNsense Terraform provider are documented here, following the [Terraform provider changelog format](https://developer.hashicorp.com/terraform/plugin/best-practices/versioning).

## 0.2.0 (June 10, 2026)

FEATURES:

* **New Resource:** `opnsense_dnsmasq_boot`
* **New Resource:** `opnsense_dnsmasq_domain`
* **New Resource:** `opnsense_dnsmasq_host`
* **New Resource:** `opnsense_dnsmasq_option`
* **New Resource:** `opnsense_dnsmasq_range`
* **New Resource:** `opnsense_dnsmasq_tag`
* **New Resource:** `opnsense_quagga_ospf_area`

* **New Data Source:** `opnsense_acme_account`
* **New Data Source:** `opnsense_acme_certificate`
* **New Data Source:** `opnsense_acme_challenge`
* **New Data Source:** `opnsense_ddclient_account`
* **New Data Source:** `opnsense_dhcpv4_reservation`
* **New Data Source:** `opnsense_dhcpv4_subnet`
* **New Data Source:** `opnsense_dnsmasq_boot`
* **New Data Source:** `opnsense_dnsmasq_domain`
* **New Data Source:** `opnsense_dnsmasq_host`
* **New Data Source:** `opnsense_dnsmasq_option`
* **New Data Source:** `opnsense_dnsmasq_range`
* **New Data Source:** `opnsense_dnsmasq_tag`
* **New Data Source:** `opnsense_firewall_category`
* **New Data Source:** `opnsense_firewall_filter_rule`
* **New Data Source:** `opnsense_firewall_nat_outbound`
* **New Data Source:** `opnsense_firewall_nat_port_forward`
* **New Data Source:** `opnsense_haproxy_acl`
* **New Data Source:** `opnsense_haproxy_backend`
* **New Data Source:** `opnsense_haproxy_frontend`
* **New Data Source:** `opnsense_haproxy_healthcheck`
* **New Data Source:** `opnsense_haproxy_server`
* **New Data Source:** `opnsense_ipsec_child`
* **New Data Source:** `opnsense_ipsec_connection`
* **New Data Source:** `opnsense_ipsec_local`
* **New Data Source:** `opnsense_ipsec_remote`
* **New Data Source:** `opnsense_kea_dhcpv6_reservation`
* **New Data Source:** `opnsense_kea_dhcpv6_subnet`
* **New Data Source:** `opnsense_openvpn_client_overwrite`
* **New Data Source:** `opnsense_openvpn_instance`
* **New Data Source:** `opnsense_quagga_bgp_aspath`
* **New Data Source:** `opnsense_quagga_bgp_communitylist`
* **New Data Source:** `opnsense_quagga_bgp_neighbor`
* **New Data Source:** `opnsense_quagga_bgp_peergroup`
* **New Data Source:** `opnsense_quagga_bgp_redistribution`
* **New Data Source:** `opnsense_quagga_ospf_area`
* **New Data Source:** `opnsense_quagga_prefix_list`
* **New Data Source:** `opnsense_quagga_route_map`
* **New Data Source:** `opnsense_quagga_static_route`
* **New Data Source:** `opnsense_system_gateway`
* **New Data Source:** `opnsense_system_route`
* **New Data Source:** `opnsense_system_vip`
* **New Data Source:** `opnsense_system_vlan`
* **New Data Source:** `opnsense_trust_ca`
* **New Data Source:** `opnsense_unbound_acl`
* **New Data Source:** `opnsense_unbound_domain_override`
* **New Data Source:** `opnsense_unbound_host_alias`
* **New Data Source:** `opnsense_unbound_host_override`
* **New Data Source:** `opnsense_wireguard_peer`
* **New Data Source:** `opnsense_wireguard_server`

IMPROVEMENTS:

* Expanded generated Registry documentation and examples for 97 resources and 83 data sources.
* Added migration/import guidance for brownfield appliance adoption.
* Added a public upstream-blocked register and support-matrix planning artifacts.
* Added post-release research classifications for HASync configuration, HASync status/actions, and system tunables/sysctl.
* Added OpenCode BMad command shims under `.opencode/commands`.
* Added `toolchain go1.25.11` so validation uses the expected Go toolchain.

BUG FIXES:

* Added Dnsmasq option validation for fields OPNsense clears for incompatible option types.
* Corrected OSPF area no-summary option values to match OPNsense API values.

## 0.1.0 (June 1, 2026)

FEATURES:

* Initial release of the OPNsense Terraform provider — manage OPNsense appliance configuration (firewall, NAT, HAProxy, ACME, DNS, DHCP, VPN, dynamic routing, interfaces, certificates, and users) through its MVC API.
* Provider configuration via HCL or `OPNSENSE_*` environment variables, with credential validation against `/api/core/firmware/status`.

* **New Resource:** `opnsense_acme_account`
* **New Resource:** `opnsense_acme_certificate`
* **New Resource:** `opnsense_acme_challenge`
* **New Resource:** `opnsense_auth_group`
* **New Resource:** `opnsense_auth_user`
* **New Resource:** `opnsense_cron_job`
* **New Resource:** `opnsense_ddclient_account`
* **New Resource:** `opnsense_dhcpv4_reservation`
* **New Resource:** `opnsense_dhcpv4_subnet`
* **New Resource:** `opnsense_dnsmasq_settings`
* **New Resource:** `opnsense_firewall_alias`
* **New Resource:** `opnsense_firewall_category`
* **New Resource:** `opnsense_firewall_filter_rule`
* **New Resource:** `opnsense_firewall_nat_one_to_one`
* **New Resource:** `opnsense_firewall_nat_outbound`
* **New Resource:** `opnsense_firewall_nat_port_forward`
* **New Resource:** `opnsense_haproxy_acl`
* **New Resource:** `opnsense_haproxy_backend`
* **New Resource:** `opnsense_haproxy_frontend`
* **New Resource:** `opnsense_haproxy_healthcheck`
* **New Resource:** `opnsense_haproxy_server`
* **New Resource:** `opnsense_interface_bridge`
* **New Resource:** `opnsense_interface_gif`
* **New Resource:** `opnsense_interface_gre`
* **New Resource:** `opnsense_interface_loopback`
* **New Resource:** `opnsense_interface_neighbor`
* **New Resource:** `opnsense_interface_vxlan`
* **New Resource:** `opnsense_ipsec_child`
* **New Resource:** `opnsense_ipsec_connection`
* **New Resource:** `opnsense_ipsec_key_pair`
* **New Resource:** `opnsense_ipsec_local`
* **New Resource:** `opnsense_ipsec_manual_spd`
* **New Resource:** `opnsense_ipsec_pool`
* **New Resource:** `opnsense_ipsec_psk`
* **New Resource:** `opnsense_ipsec_remote`
* **New Resource:** `opnsense_ipsec_vti`
* **New Resource:** `opnsense_kea_ctrl_agent`
* **New Resource:** `opnsense_kea_dhcpv6_reservation`
* **New Resource:** `opnsense_kea_dhcpv6_settings`
* **New Resource:** `opnsense_kea_dhcpv6_subnet`
* **New Resource:** `opnsense_kea_ha_peer`
* **New Resource:** `opnsense_monit_alert`
* **New Resource:** `opnsense_monit_service`
* **New Resource:** `opnsense_monit_test`
* **New Resource:** `opnsense_openvpn_client_overwrite`
* **New Resource:** `opnsense_openvpn_instance`
* **New Resource:** `opnsense_openvpn_static_key`
* **New Resource:** `opnsense_quagga_bgp_aspath`
* **New Resource:** `opnsense_quagga_bgp_communitylist`
* **New Resource:** `opnsense_quagga_bgp_global`
* **New Resource:** `opnsense_quagga_bgp_neighbor`
* **New Resource:** `opnsense_quagga_bgp_peergroup`
* **New Resource:** `opnsense_quagga_bgp_redistribution`
* **New Resource:** `opnsense_quagga_general`
* **New Resource:** `opnsense_quagga_ospf6_general`
* **New Resource:** `opnsense_quagga_ospf6_interface`
* **New Resource:** `opnsense_quagga_ospf6_network`
* **New Resource:** `opnsense_quagga_ospf6_prefixlist`
* **New Resource:** `opnsense_quagga_ospf6_redistribution`
* **New Resource:** `opnsense_quagga_ospf6_routemap`
* **New Resource:** `opnsense_quagga_ospf_general`
* **New Resource:** `opnsense_quagga_ospf_interface`
* **New Resource:** `opnsense_quagga_ospf_neighbor`
* **New Resource:** `opnsense_quagga_ospf_network`
* **New Resource:** `opnsense_quagga_ospf_prefixlist`
* **New Resource:** `opnsense_quagga_ospf_redistribution`
* **New Resource:** `opnsense_quagga_ospf_routemap`
* **New Resource:** `opnsense_quagga_prefix_list`
* **New Resource:** `opnsense_quagga_rip`
* **New Resource:** `opnsense_quagga_route_map`
* **New Resource:** `opnsense_quagga_static`
* **New Resource:** `opnsense_quagga_static_route`
* **New Resource:** `opnsense_syslog_destination`
* **New Resource:** `opnsense_system_gateway`
* **New Resource:** `opnsense_system_route`
* **New Resource:** `opnsense_system_vip`
* **New Resource:** `opnsense_system_vlan`
* **New Resource:** `opnsense_trafficshaper_pipe`
* **New Resource:** `opnsense_trafficshaper_queue`
* **New Resource:** `opnsense_trafficshaper_rule`
* **New Resource:** `opnsense_trust_ca`
* **New Resource:** `opnsense_trust_cert`
* **New Resource:** `opnsense_unbound_acl`
* **New Resource:** `opnsense_unbound_dnsbl`
* **New Resource:** `opnsense_unbound_domain_override`
* **New Resource:** `opnsense_unbound_general`
* **New Resource:** `opnsense_unbound_host_alias`
* **New Resource:** `opnsense_unbound_host_override`
* **New Resource:** `opnsense_wireguard_peer`
* **New Resource:** `opnsense_wireguard_server`

* **New Data Source:** `opnsense_auth_group`
* **New Data Source:** `opnsense_auth_user`
* **New Data Source:** `opnsense_cron_job`
* **New Data Source:** `opnsense_firewall_alias`
* **New Data Source:** `opnsense_firewall_nat_one_to_one`
* **New Data Source:** `opnsense_interface_bridge`
* **New Data Source:** `opnsense_interface_gif`
* **New Data Source:** `opnsense_interface_gre`
* **New Data Source:** `opnsense_interface_loopback`
* **New Data Source:** `opnsense_interface_neighbor`
* **New Data Source:** `opnsense_interface_vxlan`
* **New Data Source:** `opnsense_ipsec_manual_spd`
* **New Data Source:** `opnsense_ipsec_pool`
* **New Data Source:** `opnsense_ipsec_vti`
* **New Data Source:** `opnsense_kea_ha_peer`
* **New Data Source:** `opnsense_monit_alert`
* **New Data Source:** `opnsense_monit_service`
* **New Data Source:** `opnsense_monit_test`
* **New Data Source:** `opnsense_quagga_ospf6_interface`
* **New Data Source:** `opnsense_quagga_ospf6_network`
* **New Data Source:** `opnsense_quagga_ospf6_prefixlist`
* **New Data Source:** `opnsense_quagga_ospf6_redistribution`
* **New Data Source:** `opnsense_quagga_ospf6_routemap`
* **New Data Source:** `opnsense_quagga_ospf_interface`
* **New Data Source:** `opnsense_quagga_ospf_neighbor`
* **New Data Source:** `opnsense_quagga_ospf_network`
* **New Data Source:** `opnsense_quagga_ospf_prefixlist`
* **New Data Source:** `opnsense_quagga_ospf_redistribution`
* **New Data Source:** `opnsense_quagga_ospf_routemap`
* **New Data Source:** `opnsense_syslog_destination`
* **New Data Source:** `opnsense_system_info`
* **New Data Source:** `opnsense_trafficshaper_pipe`
* **New Data Source:** `opnsense_trafficshaper_queue`
* **New Data Source:** `opnsense_trafficshaper_rule`

IMPROVEMENTS:

* n/a (initial release).

BUG FIXES:

* n/a (initial release).

---

## Repository tooling history

Earlier entries below track the DevRail repository scaffold and tooling, not the provider itself.

## [Unreleased]

### Changed

- Updated beta banner to v1 stable

## [1.0.0] - 2026-03-01

### Added

- Makefile with all 7 language ecosystems (Python, Bash, Terraform, Ansible, Ruby, Go, JavaScript/TypeScript)
- `make init` / `make _init` config scaffolding target
- CI workflows: lint, format, test, security, scan, docs
- Pre-commit hooks for all supported languages (commented out by default)
- Agent instruction files (CLAUDE.md, AGENTS.md, .cursorrules, .opencode/agents.yaml)
- DevRail compliance badge in README
- Retrofit guide for adding DevRail to existing repositories
- `.devrail.yml` with all 7 languages listed (commented out)
- `.editorconfig`, `.gitignore`, `DEVELOPMENT.md`, `CHANGELOG.md`, `LICENSE`
