# Changelog

All notable changes to the OPNsense Terraform provider are documented here, following the [Terraform provider changelog format](https://developer.hashicorp.com/terraform/plugin/best-practices/versioning).

## 0.10.1 (Unreleased)

UPGRADE NOTES:

* NAT port-forward create/update with categories now requires read access to `/api/firewall/category/getItem/<uuid>`. Existing category UUID configuration and state remain unchanged.

BUG FIXES:

* Resolve NAT port-forward category UUIDs to names when writing the persisted `category` field, preserving UUID state and supporting explicit label clearing. Unknown or omitted categories leave existing labels unchanged; failed lookups abort before rule mutation.

## 0.10.0 (October 26, 2026)

FEATURES:

* Added `opnsense_kea_ddns_settings` resource and read-only four-field lookup for compatible stable OPNsense 26.7.7+, fixed `ddns` import identity, omission-preserving partial writes, explicit false/unknown handling, canonical durable save recovery, and zero-API state-only Delete (including pending activation; normal Core planning refresh may recover/activate first). Owns only `ddns.general` enabled/manual_config/server_ip/server_port; no manual files or DNS-delivery claim. Implementation, disabled-appliance proof, merge and publication remain separate gates.

* Added `opnsense_kea_dhcpv4_option` resource and UUID lookup for the eight stored generic-option fields on stable OPNsense 26.7.7+, with mutable unique description, omission-preserving optional intent, strict model/read validation, single-attempt writes and durable UUID-bound save/delete recovery. Ordered subnet/reservation linkage remains separate; the stored force flag is not a packet guarantee. The disabled-appliance 14-step Terraform lifecycle, 17 independent saved-field observations, in-use HTTP500 identity retention, deletion and exact restoration passed. Publication remains separately gated.

UPGRADE NOTES:

* These are independent, opt-in owners. Existing subnet/reservation option links and manual DDNS files/hooks are not adopted. DDNS create preserves omitted appliance settings; Delete abandons Terraform ownership without resetting the singleton, but normal Core planning refresh can recover/activate a pending save before Delete. Review ownership and the documented recovery boundaries before adoption.

## 0.9.0 (October 26, 2026)

FEATURES:

* Added `opnsense_system_gateway_group` resource and UUID-keyed data source for stable OPNsense 26.7.7+, with typed gateway-name tiers, immutable names, strict firmware/model/read validation and durable save/delete activation recovery. Route reconfigure does not apply PF. The 16-step isolated Terraform lifecycle, independent saved-field observations and exact cleanup/restoration passed. Populated first-tier clearing is rejected before any write because the stable API ignores it; later-tier clears are supported. Older versions, OpenTofu and actual traffic failover remain unverified.

## 0.8.1 (Unreleased)

BUG FIXES:

* Accept installed `os-frr` 2.0 numeric package revisions (including `2.0_2`) for existing modern-model capability checks, fixing BGP refresh rejection. New plugin releases such as 2.1, prerelease/malformed versions, uninstalled/contradictory metadata and incompatible models remain rejected. The change has fixture/provider-protocol regression coverage, not a new live appliance or OpenTofu acceptance claim.

## 0.8.0 (Unreleased)

UPGRADE NOTES:

* `opnsense_dnsmasq_settings.dhcp` is now a typed object instead of a set of strings. Remove legacy `dhcp = []` arguments or replace them with the documented object syntax. Empty/null version-0 DHCP state upgrades in place and refreshes from the appliance. Non-empty legacy sets fail with explicit recovery guidance instead of guessing DHCP configuration; back up state and appliance configuration, remove the legacy argument, remove only this singleton from state, and import it again as `dnsmasq`. State removal/import do not change the appliance.

BUG FIXES:

* Correct the nested Dnsmasq DHCP/RA read and write contract, preserve omitted nested settings, keep `%domain` response metadata out of writes, and validate reply delay, lease limits and mutually exclusive logging flags. Regression verification uses tagged 26.7.7-shaped fixtures and the provider protocol; live appliance lifecycle validation is not claimed.

## 0.7.0 (October 24, 2026)

UPGRADE NOTES:

* WireGuard multi-value fields now use Terraform sets; existing scalar state upgrades without replacement, but configuration must use set syntax. Server `peers` owns associations and peer `servers` is computed. Preserve complete memberships and existing PSKs before applying; omitting a peer PSK disables it on update. See the WireGuard resource migration guides.
* BGP compatibility now detects the installed `os-frr` package and model, requiring read permission for `/api/core/firmware/info`. Unknown or contradictory plugin capabilities fail closed. The FRR 2.0 live proof covers `os-frr` 2.0_1 on OPNsense 26.7.7; it does not establish live compatibility with every legacy plugin or OpenTofu.

FEATURES:

* Added `opnsense_interface_assignment` to adopt existing interface assignments and manage the documented safe IP/DHCP settings through the expanded OPNsense 26.7.6+ API.
* Added Kea DHCPv6 delegated-prefix-pool and HA-peer resources/data sources, and expanded DHCPv4/DHCPv6 subnet option and field coverage.
* Added an OpenVPN static-key data source and `opnsense_openvpn_generate_key` ephemeral resource. Ephemeral key generation requires Terraform 1.10+ and OPNsense 26.7.6+; its output cannot feed ordinary persisted resource attributes.
* Added an Unbound settings data source and expanded general, DNS blocklist, domain override and forwarding configuration coverage.
* Added core WireGuard settings resource/data source and expanded typed peer/server fields and relationships.

IMPROVEMENTS:

* Expanded firewall alias, filter-rule and NAT port-forward field coverage, including matching data-source fields.
* Expanded OpenVPN instance and client-overwrite field coverage and key handling.
* Added explicit IPv4/IPv6 BGP family support and installed-plugin-aware legacy/FRR 2.0 serialization, logical grouped-policy identity and canonical import handling.
* Raised the build/CI/release Go security baseline to 1.26.10 and aligned DevRail tooling to 1.12.24.

BUG FIXES:

* Preserve BGP family intent and UUIDs during version-0 state migration, grouped-policy associations on read/import, and saved-mutation activation recovery after interrupted or ambiguous API responses.
* Preserve WireGuard server membership during peer updates, normalize valid empty selectors, and retain saved write-only key rotations through activation failure/restart recovery. Data sources do not expose secrets.
* Harden lifecycle validation and state recovery for the expanded firewall, Kea, OpenVPN and Unbound surfaces.

## 0.6.0 (October 10, 2026)

FEATURES:

* Added `sequence` and `reflection` attributes to `opnsense_firewall_nat_port_forward` so NAT order and reflection survive create/recreate.

## 0.5.0 (September 29, 2026)

FEATURES:

* **New Resource:** `opnsense_kea_dhcpv4_settings`
* **New Resource:** `opnsense_kea_dhcpv4_subnet`
* **New Resource:** `opnsense_kea_dhcpv4_reservation`

## 0.4.4 (September 24, 2026)

BUG FIXES:

* Read WireGuard peer `tunnel_address` as an ordered multi-select value so peers with multiple Allowed IPs reconcile correctly.

## 0.4.3 (September 24, 2026)

BUG FIXES:

* Apply firewall filter rule changes through `/api/firewall/filter/apply` instead of the unsupported savepoint rollback flow.

## 0.4.0 (September 24, 2026)

FEATURES:

* Added HAProxy backend `health_check` relation support alongside `health_check_enabled`.
* Added HAProxy backend timeout and custom option fields: `timeout_connect`, `timeout_check`, `timeout_server`, and `custom_options`.
* Added HAProxy frontend `timeout_client` support.

IMPROVEMENTS:

* Preserve HAProxy frontend `linked_actions` order by modeling it as an ordered list and decoding OPNsense responses without sorting selected actions.
* Added a state upgrader for `opnsense_haproxy_frontend` to migrate existing `linked_actions` set state into the new list shape.
* Added support for os-haproxy 5 HTTP request action mapping.

BUG FIXES:

* Preserve WireGuard server descriptions when OPNsense returns an empty description on readback.
* Normalize ACME certificate empty `altNames` readback from object-shaped OPNsense responses.
* Defer HAProxy frontend certificate validation when certificate references are unknown during planning.

DEPENDENCIES:

* Updated the Go, `google.golang.org/grpc`, and `golang.org/x/text` security baseline.

## 0.3.1 (July 20, 2026)

BUG FIXES:

* Handle `altNames` returned as a JSON object in OPNsense 26.7 API responses for `opnsense_acme_certificate`.

DEPENDENCIES:

* Bumped toolchain to Go 1.25.12 to resolve govulncheck standard library vulnerabilities.
* Updated `golang.org/x/crypto` from 0.51.0 to 0.52.0.

## 0.3.0 (June 18, 2026)

FEATURES:

* **New Resource:** `opnsense_ddclient_settings`
* **New Resource:** `opnsense_haproxy_action`
* **New Resource:** `opnsense_haproxy_mapfile`
* **New Resource:** `opnsense_interface_lagg`
* **New Resource:** `opnsense_system_tunable`

* **New Data Source:** `opnsense_ddclient_settings`
* **New Data Source:** `opnsense_haproxy_action`
* **New Data Source:** `opnsense_haproxy_mapfile`
* **New Data Source:** `opnsense_interface_lagg`
* **New Data Source:** `opnsense_system_tunable`

IMPROVEMENTS:

* Added ACME certificate issuance/refid acceptance coverage and improved certificate handling.
* Added HAProxy frontend TLS certificate binding support and validation improvements.
* Added multi-domain edge composition examples covering ACME, DDClient, HAProxy, and migration validation.
* Added Vagrant acceptance runner support for serial full-provider validation against a local OPNsense appliance.
* Updated Registry documentation, examples, templates, support matrices, and planning artifacts for the expanded resource coverage.

BUG FIXES:

* Detect OPNsense reconfigure responses that return HTTP 200 with failed JSON status/result bodies.
* Return typed post-mutation reconfigure errors when OPNsense saves a mutation but the follow-up apply/reconfigure step fails.
* Preserve custom generated acceptance prechecks for Quagga OSPF/OSPFv3 resources.

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
