# Provider Ecosystem Coverage Audit

Pinned baseline: `browningluke/opnsense` v0.26.0 at `33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba` (47 resources, 45 data sources, 1 ephemeral resource, and provider retry configuration).

Current local inventory: **113 resources and 108 data sources**, with 113/108 generated docs. There are **11 resource-only same-name data-source gaps** and the data-only surfaces are `interface`, `interface_all`, `interfaces_overview`, `interfaces_overview_all`, `system_info`, `unbound_settings`.

## Ownership Direction

Use one provider per appliance and object domain in steady state. Never configure two providers to manage the same OPNsense object; close, document, or explicitly block gaps before migration.

## Classification Summary

- `covered`: 58
- `covered-different-name`: 16
- `covered-broader`: 2
- `missing-buildable`: 12
- `missing-upstream-blocked`: 0
- `intentionally-different`: 6
- `needs-field-audit`: 0

## Domain Summary

- **Cron:** 2 surfaces (`covered`=2)
- **Dnsmasq:** 2 surfaces (`covered-broader`=2)
- **Firewall/NAT:** 12 surfaces (`covered`=10, `covered-different-name`=2)
- **IPsec:** 6 surfaces (`covered`=2, `covered-different-name`=2, `missing-buildable`=2)
- **Interfaces:** 8 surfaces (`covered`=4, `covered-different-name`=4)
- **Kea:** 20 surfaces (`covered`=12, `covered-different-name`=2, `intentionally-different`=6)
- **OpenVPN:** 7 surfaces (`covered`=7)
- **Provider:** 1 surfaces (`covered`=1)
- **Quagga/BGP:** 10 surfaces (`covered`=4, `covered-different-name`=2, `missing-buildable`=4)
- **Routing:** 2 surfaces (`covered-different-name`=2)
- **Trust:** 6 surfaces (`missing-buildable`=6)
- **Unbound:** 12 surfaces (`covered`=12)
- **WireGuard:** 6 surfaces (`covered`=4, `covered-different-name`=2)

## Surface Audit

### Provider Configuration

#### `provider-configuration:retry-settings` — covered

- **Domain / local mapping:** Provider / `retries`, `min_backoff`, `max_backoff`
- **Baseline fields:** min_backoff, max_backoff, retries
- **Shared semantics:** released v0.13.0 exposes retries, min_backoff, max_backoff and matching environment fallbacks
- **Baseline-only:** none for these pinned retry controls
- **Local-only:** bounded read concurrency and strict range validation; existing defaults preserved
- **Lifecycle:** provider configuration passed to shared authenticated TLS/retry clients; no mutation ownership
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/internal/provider/provider.go; internal/provider/provider.go; pkg/opnsense/client.go

### Resource

#### `resource:cron_job` — covered

- **Domain / local mapping:** Cron / `cron_job`
- **Baseline fields:** command, days, description, enabled, hours, id, minutes, months, parameters, weekdays, who
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/cron_job.md; docs/resources/cron_job.md

#### `resource:dnsmasq_host` — covered-broader

- **Domain / local mapping:** Dnsmasq / `dnsmasq_host`
- **Baseline fields:** alias_records, client_id, cname_records, comment, description, domain, hardware_addresses, hostname, id, ip_addresses, is_ignored, is_local_domain, tag
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** lease_time plus normalized singular field names
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/dnsmasq_host.md; docs/resources/dnsmasq_host.md

#### `resource:firewall_alias` — covered

- **Domain / local mapping:** Firewall/NAT / `firewall_alias`
- **Baseline fields:** categories, content, description, enabled, id, interface, ip_protocol, name, path_expression, stats, type, update_freq
- **Shared semantics:** same stable OPNsense alias endpoints and fields, including interface, path_expression, counters/stats, fractional update_freq, and lossless multi-protocol state
- **Baseline-only:** none material after Story 31.0a field and compatibility parity
- **Local-only:** deprecated proto scalar retained beside canonical ip_protocol set for backward compatibility
- **Lifecycle:** UUID-backed create/read/update/delete with direct alias reconfigure; live OPNsense 26.7.7 run story-31-0a-live-26.7.7-review1-rerun2 proved advanced create, import, zero-diff plan, out-of-band drift detection/repair, delete, and zero-match cleanup
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/firewall_alias.md; docs/resources/firewall_alias.md

#### `resource:firewall_category` — covered

- **Domain / local mapping:** Firewall/NAT / `firewall_category`
- **Baseline fields:** auto, color, id, name
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/firewall_category.md; docs/resources/firewall_category.md

#### `resource:firewall_filter` — covered

- **Domain / local mapping:** Firewall/NAT / `firewall_filter_rule`
- **Baseline fields:** categories, description, enabled, filter, id, interface, internal_tagging, no_xmlrpc_sync, priority, sequence, source_routing, stateful_firewall, traffic_shaping
- **Shared semantics:** same stable OPNsense filter-rule endpoints and advanced tagging, priority/TOS, reply-to, state limits/timeouts/rates, schedules, TCP/ICMP, shaping, and HA-sync fields
- **Baseline-only:** none material after Story 31.0a advanced filter parity
- **Local-only:** none material
- **Lifecycle:** UUID-backed create/read/update/delete with one direct filter apply; live OPNsense 26.7.7 run story-31-0a-live-26.7.7-review1-rerun2 proved advanced create, import, zero-diff plan, out-of-band drift detection, configured update, delete, and zero-match cleanup
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/firewall_filter.md; docs/resources/firewall_filter_rule.md

#### `resource:firewall_nat` — covered-different-name

- **Domain / local mapping:** Firewall/NAT / `firewall_nat_outbound`
- **Baseline fields:** description, destination, disable_nat, enabled, id, interface, ip_protocol, log, protocol, sequence, source, target
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** categories, static_nat_port, and target_port
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/firewall_nat.md; docs/resources/firewall_nat_outbound.md

#### `resource:firewall_nat_one_to_one` — covered

- **Domain / local mapping:** Firewall/NAT / `firewall_nat_one_to_one`
- **Baseline fields:** categories, description, destination, enabled, external_net, id, interface, log, nat_reflection, sequence, source, type
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/firewall_nat_one_to_one.md; docs/resources/firewall_nat_one_to_one.md

#### `resource:firewall_nat_port_forward` — covered

- **Domain / local mapping:** Firewall/NAT / `firewall_nat_port_forward`
- **Baseline fields:** description, destination, enabled, id, interface, ip_protocol, log, nat_reflection, protocol, sequence, source, target
- **Shared semantics:** same stable destination-NAT endpoints and fields, including lossless multi-interface port-forward membership
- **Baseline-only:** none material after Story 31.0a multi-interface parity
- **Local-only:** deprecated interface scalar retained beside canonical interfaces set for backward compatibility
- **Lifecycle:** UUID-backed create/read/update/delete with one direct d_nat apply; live OPNsense 26.7.7 run story-31-0a-live-26.7.7-review1-rerun2 proved multi-interface create, import, zero-diff plan, membership update, delete, and zero-match cleanup
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/firewall_nat_port_forward.md; docs/resources/firewall_nat_port_forward.md

#### `resource:interfaces_vip` — covered-different-name

- **Domain / local mapping:** Interfaces / `system_vip`
- **Baseline fields:** description, gateway, id, interface, mode, network
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/interfaces_vip.md; docs/resources/system_vip.md

#### `resource:interfaces_vlan` — covered-different-name

- **Domain / local mapping:** Interfaces / `system_vlan`
- **Baseline fields:** description, device, id, parent, priority, tag
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/interfaces_vlan.md; docs/resources/system_vlan.md

#### `resource:ipsec_auth_local` — covered-different-name

- **Domain / local mapping:** IPsec / `ipsec_local`
- **Baseline fields:** auth_id, authentication, certificates, description, eap_id, enabled, id, ipsec_connection, public_keys, round
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/ipsec_auth_local.md; docs/resources/ipsec_local.md

#### `resource:ipsec_auth_remote` — covered-different-name

- **Domain / local mapping:** IPsec / `ipsec_remote`
- **Baseline fields:** auth_id, authentication, certificates, description, eap_id, enabled, id, ipsec_connection, public_keys, round
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/ipsec_auth_remote.md; docs/resources/ipsec_remote.md

#### `resource:ipsec_child` — missing-buildable

- **Domain / local mapping:** IPsec / `ipsec_child`
- **Baseline fields:** close_action, description, dpd_action, enabled, id, install_policies, ipsec_connection, local_networks, mode, proposals, rekey_time, remote_networks, request_id, sha256_96, start_action
- **Shared semantics:** Pinned core 26.7.8 Swanctl.xml/ConnectionsController.php; existing connections-controller endpoints and UUID ownership; remote_addresses maps to remote_addrs, child audit names map to existing connection_id/local_ts/remote_ts/esp_proposals without changing their types.
- **Baseline-only:** All 21 audited advanced controls are implemented and guarded-live verified on 26.7.8_2. Classification remains missing-buildable for explicitly deferred inherited legacy omission/read-completeness and unobservable proposal preference-order limitations, not for an unexecuted advanced lifecycle.
- **Local-only:** none material
- **Lifecycle:** Offline protocol/race/Core controls and full guarded 26.7.8_2 disabled-object lifecycle/activation PASS at c03b2e0: persistent imports, all21lookups, omission/no-change, drift/repair, clears, Terraform deletion, zero leftovers, exact raw restoration and retirement. No negotiated IKE/traffic, older-firmware or OpenTofu execution claim. See implementation-artifacts/31-0g-live-proof.json.
- **Sensitive/write-only:** none on these rows; PSKs are owned by the separate sensitive ipsec_psk surface
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/ipsec_child.md; docs/resources/ipsec_child.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0g-ipsec-field-parity.md`

#### `resource:ipsec_connection` — missing-buildable

- **Domain / local mapping:** IPsec / `ipsec_connection`
- **Baseline fields:** aggressive, description, dpd_delay, dpd_timeout, enabled, id, ike_lifetime, ip_pools, keying_tries, local_addresses, local_port, mobike, proposals, reauthentication_time, rekey_time, remote_addresses, remote_port, send_certificate, send_certificate_request, udp_encapsulation, unique, version
- **Shared semantics:** Pinned core 26.7.8 Swanctl.xml/ConnectionsController.php; existing connections-controller endpoints and UUID ownership; remote_addresses maps to remote_addrs, child audit names map to existing connection_id/local_ts/remote_ts/esp_proposals without changing their types.
- **Baseline-only:** All 21 audited advanced controls are implemented and guarded-live verified on 26.7.8_2. Classification remains missing-buildable for explicitly deferred inherited legacy omission/read-completeness and unobservable proposal preference-order limitations, not for an unexecuted advanced lifecycle.
- **Local-only:** none material
- **Lifecycle:** Offline protocol/race/Core controls and full guarded 26.7.8_2 disabled-object lifecycle/activation PASS at c03b2e0: persistent imports, all21lookups, omission/no-change, drift/repair, clears, Terraform deletion, zero leftovers, exact raw restoration and retirement. No negotiated IKE/traffic, older-firmware or OpenTofu execution claim. See implementation-artifacts/31-0g-live-proof.json.
- **Sensitive/write-only:** none on these rows; PSKs are owned by the separate sensitive ipsec_psk surface
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/ipsec_connection.md; docs/resources/ipsec_connection.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0g-ipsec-field-parity.md`

#### `resource:ipsec_psk` — covered

- **Domain / local mapping:** IPsec / `ipsec_psk`
- **Baseline fields:** description, id, identity_local, identity_remote, pre_shared_key, type
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** pre_shared_key is sensitive and write-only; preserve configured state because the API does not safely return the secret
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/ipsec_psk.md; docs/resources/ipsec_psk.md

#### `resource:ipsec_vti` — covered

- **Domain / local mapping:** IPsec / `ipsec_vti`
- **Baseline fields:** description, enabled, id, local_ip, remote_ip, request_id, tunnel_local_ip, tunnel_local_ip2, tunnel_remote_ip, tunnel_remote_ip2
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/ipsec_vti.md; docs/resources/ipsec_vti.md

#### `resource:kea_dhcpv4_peer` — covered-different-name

- **Domain / local mapping:** Kea / `kea_ha_peer`
- **Baseline fields:** id, name, role, url
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/kea_dhcpv4_peer.md; docs/resources/kea_ha_peer.md

#### `resource:kea_dhcpv4_reservation` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv4_reservation`
- **Baseline fields:** description, hostname, id, ip_address, mac_address, subnet_id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/kea_dhcpv4_reservation.md; docs/resources/kea_dhcpv4_reservation.md

#### `resource:kea_dhcpv4_subnet` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv4_subnet`
- **Baseline fields:** auto_collect, description, dns_servers, domain_name, domain_search, id, match_client_id, next_server, ntp_servers, pools, routers, static_routes, subnet, tftp_bootfile, tftp_server, time_servers
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** newline-compatible pools, allocator, ping-check, deterministic comma-pair static routes, and pre-mutation validation
- **Lifecycle:** Disposable OPNsense 26.7.7 run 20261018T233224Z proved advanced DHCPv4 create/read/import, populated-to-explicit-empty update, out-of-band drift detection and repair, stable no-change plans, exact API read-back, affirmative reconfigure, and strict zero-match cleanup
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/kea_dhcpv4_subnet.md; docs/resources/kea_dhcpv4_subnet.md

#### `resource:kea_dhcpv6_pd_pool` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv6_pd_pool`
- **Baseline fields:** delegated_len, description, id, prefix, prefix_len, subnet_id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** aligned-prefix and delegation-length validation plus recoverable create identity
- **Lifecycle:** Disposable OPNsense 26.7.7 run 20261018T233224Z proved generated and parity PD-pool create/read/import/update/delete, CIDR resource stability, parent relationship handling, runtime compatibility gating, affirmative reconfigure, and strict zero-match cleanup
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/kea_dhcpv6_pd_pool.md; docs/resources/kea_dhcpv6_pd_pool.md

#### `resource:kea_dhcpv6_peer` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv6_peer`
- **Baseline fields:** id, name, role, url
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** absolute HTTP/HTTPS URL validation and recoverable create identity
- **Lifecycle:** Disposable OPNsense 26.7.7 run 20261018T233224Z proved generated and parity peer create/read/import/update, role drift detection and repair, stable no-change plans, runtime compatibility gating, affirmative reconfigure, and strict zero-match cleanup
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/kea_dhcpv6_peer.md; docs/resources/kea_dhcpv6_peer.md

#### `resource:kea_dhcpv6_reservation` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv6_reservation`
- **Baseline fields:** description, domain_search, duid, hostname, id, ip_address, subnet_id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/kea_dhcpv6_reservation.md; docs/resources/kea_dhcpv6_reservation.md

#### `resource:kea_dhcpv6_subnet` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv6_subnet`
- **Baseline fields:** allocator, description, dns_servers, domain_search, id, interface, pd_allocator, pools, subnet
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** interface, allocator, PD allocator, newline-compatible pools, and unknown-safe collection requests
- **Lifecycle:** Disposable OPNsense 26.7.7 run 20261018T233224Z proved IPv6 subnet create/read/import, DNS/search populated-to-empty update, stable no-change plans, PD-parent lifecycle ordering, affirmative reconfigure, and strict zero-match cleanup
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/kea_dhcpv6_subnet.md; docs/resources/kea_dhcpv6_subnet.md

#### `resource:kea_peer` — intentionally-different

- **Domain / local mapping:** Kea / `kea_ha_peer`
- **Baseline fields:** id, name, role, url
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/kea_peer.md; docs/resources/kea_ha_peer.md

#### `resource:kea_reservation` — intentionally-different

- **Domain / local mapping:** Kea / `kea_dhcpv4_reservation`
- **Baseline fields:** description, hostname, id, ip_address, mac_address, subnet_id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/kea_reservation.md; docs/resources/kea_dhcpv4_reservation.md

#### `resource:kea_subnet` — intentionally-different

- **Domain / local mapping:** Kea / `kea_dhcpv4_subnet`
- **Baseline fields:** auto_collect, description, dns_servers, domain_name, domain_search, id, match_client_id, next_server, ntp_servers, pools, routers, static_routes, subnet, tftp_bootfile, tftp_server, time_servers
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/kea_subnet.md; docs/resources/kea_dhcpv4_subnet.md

#### `resource:openvpn_client_overwrite` — covered

- **Domain / local mapping:** OpenVPN / `openvpn_client_overwrite`
- **Baseline fields:** block, common_name, description, dns_domain, dns_domain_search, dns_servers, enabled, id, local_networks, ntp_servers, push_reset, redirect_gateway, register_dns, remote_networks, route_gateway, servers, tunnel_network, tunnel_network_v6, wins_servers
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** strict owner-local selected-map decoding, explicit-empty collection clears, import-safe optional values, and recoverable pending activation
- **Lifecycle:** Digest-bound OPNsense 26.7.7 proof imported the persisted advanced object into isolated Terraform import state, verified every imported attribute, and passed no-diff, drift/repair, explicit-clear, and zero-leftover checks
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/openvpn_client_overwrite.md; docs/resources/openvpn_client_overwrite.md

#### `resource:openvpn_instance` — covered

- **Domain / local mapping:** OpenVPN / `openvpn_instance`
- **Baseline fields:** auth_digest, auth_gen_token, auth_gen_token_renewal, auth_gen_token_secret, auth_mode, bridge_gateway, bridge_pool, carp_depend_on, cert_depth, certificate, certificate_authority, compress_migrate, crl, data_ciphers, data_ciphers_fallback, description, dev_type, dns_domain, dns_domain_search, dns_servers, enabled, fragment, http_proxy, id, ifconfig_pool_persist, keepalive_interval, keepalive_timeout, local, local_group, max_clients, mss_fix, no_pool, ntp_servers, password, port, port_share, protocol, provision_exclusive, push_excluded_routes, push_inactive, push_route, redirect_gateway, register_dns, remote, remote_cert_tls, reneg_sec, role, route, route_metric, server, server_ipv6, strict_user_cn, tls_key, topology, tun_mtu, use_ocsp, username, username_as_common_name, various_flags, various_push_flags, verify_client_cert, verify_x509_name, vpn_id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** verb, API-read vpnid, additive ordered-cipher projection, strict owner-local decoding, import-safe optional values, and recoverable pending activation
- **Lifecycle:** Digest-bound OPNsense 26.7.7 proof imported persisted advanced server and client objects into isolated Terraform import state, verified every imported non-write-only attribute, and passed no-diff, drift/repair, explicit-clear, and zero-leftover checks
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/openvpn_instance.md; docs/resources/openvpn_instance.md

#### `resource:openvpn_static_key` — covered

- **Domain / local mapping:** OpenVPN / `openvpn_static_key`
- **Baseline fields:** description, id, key, mode
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** the baseline read-only lookup returns sensitive key material; no local data source exists
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** OpenVPN key is sensitive; the ephemeral form avoids state while the lookup stores a sensitive value
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/openvpn_static_key.md; docs/resources/openvpn_static_key.md

#### `resource:quagga_bgp_aspath` — covered

- **Domain / local mapping:** Quagga/BGP / `quagga_bgp_aspath`
- **Baseline fields:** action, as, description, enabled, id, number
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/quagga_bgp_aspath.md; docs/resources/quagga_bgp_aspath.md

#### `resource:quagga_bgp_communitylist` — covered

- **Domain / local mapping:** Quagga/BGP / `quagga_bgp_communitylist`
- **Baseline fields:** action, community, description, enabled, id, number, seq_number
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/quagga_bgp_communitylist.md; docs/resources/quagga_bgp_communitylist.md

#### `resource:quagga_bgp_neighbor` — missing-buildable

- **Domain / local mapping:** Quagga/BGP / `quagga_bgp_neighbor`
- **Baseline fields:** as_override, attribute_unchanged, bfd, connect_timer, default_route, description, disable_connected_check, enabled, hold_down, id, keep_alive, link_local_interface, local_ip, md5_password, multi_hop, multi_protocol, next_hop_self, next_hop_self_all, peer_ip, prefix_list_in, prefix_list_out, remote_as, route_map_in, route_map_out, rr_client, update_source, weight
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** No remaining buildable neighbor field additions in this advanced candidate; current 26.7.8_2/os-frr 2.0_4 guarded live lifecycle passed. Adopted numeric-clear design and complete older-plugin/appliance/OpenTofu contract proof remain pending; no full-row parity or release is inferred.
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** Optional Sensitive config-owned password input; never remote-adopted or exposed in lookup. Sensitive state is not encryption; remote secret drift cannot be proven.
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/quagga_bgp_neighbor.md; docs/resources/quagga_bgp_neighbor.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0h-quagga-bgp-field-parity.md`

#### `resource:quagga_bgp_prefixlist` — covered-different-name

- **Domain / local mapping:** Quagga/BGP / `quagga_prefix_list`
- **Baseline fields:** action, description, enabled, id, ip_version, name, network, number
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/quagga_bgp_prefixlist.md; docs/resources/quagga_prefix_list.md

#### `resource:quagga_bgp_routemap` — missing-buildable

- **Domain / local mapping:** Quagga/BGP / `quagga_route_map`
- **Baseline fields:** action, aspaths, community_lists, description, enabled, id, name, prefix_lists, route_map_id, set
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** Existing owners represent AS-path/community/prefix relationships and numeric sequence; current 26.7.8_2/os-frr 2.0_4 grouped policy-order, withdrawal and restored-positive controls passed. Complete row-level contract/older-plugin/appliance/OpenTofu proof remains pending; no duplicate policy owner is required.
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/quagga_bgp_routemap.md; docs/resources/quagga_route_map.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0h-quagga-bgp-field-parity.md`

#### `resource:route` — covered-different-name

- **Domain / local mapping:** Routing / `system_route`
- **Baseline fields:** description, enabled, gateway, id, network
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/route.md; docs/resources/system_route.md

#### `resource:trust_ca` — missing-buildable

- **Domain / local mapping:** Trust / `trust_ca`
- **Baseline fields:** action, caref, city, common_name, country, crt, crt_payload, description, digest, email, id, key_type, lifetime, name, ocsp_uri, organization, organizational_unit, prv, prv_payload, ref_id, serial, state, valid_from, valid_to
- **Shared semantics:** Released26.7.8 Trust source-backed action/read continuation. Configured operation intent is distinct from inferred editor actions and lifetime template defaults.
- **Baseline-only:** Full non-skipped 26.7.8_2 action/settings/adoption/destroy/restoration matrix passed; see 31-0f-live-proof.json. Keyless CA import/update remains a provider implementation gate despite the patched controller correction; no XML or fabricated key workaround.
- **Local-only:** Separate returned public/sensitive observations preserve configured PEM and operation intent; explicit operation token distinguishes repeat operations from description updates.
- **Lifecycle:** Single-attempt UUID-backed create/update/delete; ordinary description changes do not regenerate/sign/reissue. Successful add identity and available one-time key are retained before read-back.
- **Sensitive/write-only:** All supplied and returned private-key material is Sensitive. Configured PEM inputs retain text; response-only generated keys are retained in resource state, never synthesized by lookup or retransmitted.
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/trust_ca.md; docs/resources/trust_ca.md; _bmad-output/implementation-artifacts/31-0f-trust-source-contract.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md`

#### `resource:trust_cert` — missing-buildable

- **Domain / local mapping:** Trust / `trust_cert`
- **Baseline fields:** action, altnames_dns, altnames_email, altnames_ip, altnames_uri, caref, cert_type, city, common_name, country, crt, crt_payload, csr, csr_payload, description, digest, email, id, in_use, is_user, key_type, lifetime, name, ocsp_uri, organization, organizational_unit, private_key_location, prv, prv_payload, ref_id, rfc3280_purpose, state, valid_from, valid_to
- **Shared semantics:** Released26.7.8 Trust source-backed action/read continuation. Configured operation intent is distinct from inferred editor actions and lifetime template defaults.
- **Baseline-only:** Full non-skipped 26.7.8_2 action/settings/adoption/destroy/restoration matrix passed; see 31-0f-live-proof.json. Keyless CA import/update remains a provider implementation gate despite the patched controller correction; no XML or fabricated key workaround.
- **Local-only:** Separate returned public/sensitive observations preserve configured PEM and operation intent; explicit operation token distinguishes repeat operations from description updates.
- **Lifecycle:** Single-attempt UUID-backed create/update/delete; ordinary description changes do not regenerate/sign/reissue. Successful add identity and available one-time key are retained before read-back.
- **Sensitive/write-only:** All supplied and returned private-key material is Sensitive. Configured PEM inputs retain text; response-only generated keys are retained in resource state, never synthesized by lookup or retransmitted.
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/trust_cert.md; docs/resources/trust_cert.md; _bmad-output/implementation-artifacts/31-0f-trust-source-contract.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md`

#### `resource:trust_settings` — missing-buildable

- **Domain / local mapping:** Trust / `trust_settings`
- **Baseline fields:** cipher_string, enable_config_constraints, enable_legacy_sect, fetch_crls, id, install_crls, store_intermediate_certs
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** No missing audited six-field owner/lookup; five additional upstream TLS fields remain deliberately unaudited and unmanaged.
- **Local-only:** Import-first six-field singleton implementation; HTTP-backed framework/protocol tests, not appliance or release evidence.
- **Lifecycle:** Fixed trust_settings import; targeted update/read/reconfigure, durable saved/uncertain recovery without repeated set, state-only Delete. Non-skipped 26.7.8_2 settings lifecycle, all six settings and state-only deletion passed with exact restoration; see 31-0f-live-proof.json. Broader firmware/OpenTofu and keyless provider gates remain open.
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/trust_settings.md; internal/provider/provider.go; internal/service/trust/settings_resource.go; internal/service/trust/settings_protocol_test.go; _bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md`

#### `resource:unbound_acl` — covered

- **Domain / local mapping:** Unbound / `unbound_acl`
- **Baseline fields:** action, description, enabled, id, name, networks
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/unbound_acl.md; docs/resources/unbound_acl.md

#### `resource:unbound_domain_override` — covered

- **Domain / local mapping:** Unbound / `unbound_domain_override`
- **Baseline fields:** description, domain, enabled, id, server
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/unbound_domain_override.md; docs/resources/unbound_domain_override.md

#### `resource:unbound_forward` — covered

- **Domain / local mapping:** Unbound / `unbound_domain_override`
- **Baseline fields:** domain, enabled, id, server_ip, server_port, type, verify_cn
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** description
- **Lifecycle:** digest-bound OPNsense 26.7.7 acceptance proves distinct query and DNS-over-TLS create/read/import/update/delete lifecycles, bidirectional type transitions, authoritative readback, drift repair, and non-mutating zero-object proof after Terraform destroy
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/unbound_forward.md; docs/resources/unbound_domain_override.md; _bmad-output/implementation-artifacts/31-0d-live-evidence.json

#### `resource:unbound_host_alias` — covered

- **Domain / local mapping:** Unbound / `unbound_host_alias`
- **Baseline fields:** description, domain, enabled, hostname, id, override
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/unbound_host_alias.md; docs/resources/unbound_host_alias.md

#### `resource:unbound_host_override` — covered

- **Domain / local mapping:** Unbound / `unbound_host_override`
- **Baseline fields:** description, domain, enabled, hostname, id, mx_host, mx_priority, server, type
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** UUID-backed API create/read/update/delete with the domain reconfigure/apply action where required
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/unbound_host_override.md; docs/resources/unbound_host_override.md

#### `resource:unbound_settings` — covered

- **Domain / local mapping:** Unbound / `unbound_general`, `unbound_dnsbl`
- **Baseline fields:** acls, advanced, dnsbl, forwarding, general, id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** digest-bound OPNsense 26.7.7 acceptance proves import-first general and DNSBL ownership, independent general/advanced/ACL-default/forwarding updates, Terraform DNSBL update and drift repair, state-only singleton destroy, complete aggregate equality, and exact settings restoration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/unbound_settings.md; docs/resources/unbound_general.md; docs/resources/unbound_dnsbl.md; _bmad-output/implementation-artifacts/31-0d-live-evidence.json

#### `resource:wireguard_client` — covered-different-name

- **Domain / local mapping:** WireGuard / `wireguard_peer`
- **Baseline fields:** enabled, id, keep_alive, name, psk, public_key, server_address, server_port, tunnel_address
- **Shared semantics:** Released core WireGuard API parity proved on disposable OPNsense 26.7.7; typed sets preserve multiple addresses and selected-map API truth.
- **Baseline-only:** none within the audited field set
- **Local-only:** API-derived servers backreference; keepalive is the local spelling of keep_alive
- **Lifecycle:** UUID create/read/import/no-change/update/drift repair/populated-to-empty/delete; immediate Terraform-delete zero-object proof and owner-locked recovery boundaries
- **Sensitive/write-only:** PSK is Sensitive, resource-only and prior-state-preserved on blank/omitted API reads
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/wireguard_client.md; docs/resources/wireguard_peer.md; _bmad-output/implementation-artifacts/31-0e-wireguard-live-evidence.md

#### `resource:wireguard_server` — covered

- **Domain / local mapping:** WireGuard / `wireguard_server`
- **Baseline fields:** disable_routes, dns, enabled, gateway, id, instance, mtu, name, peers, port, private_key, public_key, tunnel_address
- **Shared semantics:** Released core WireGuard API parity proved on disposable OPNsense 26.7.7; typed sets preserve multiple addresses and selected-map API truth.
- **Baseline-only:** none within the audited field set
- **Local-only:** compatibility description state; instance is API-assigned and Computed
- **Lifecycle:** UUID create/read/import/no-change/update/drift repair/populated-to-empty/delete; immediate Terraform-delete zero-object proof and owner-locked recovery boundaries
- **Sensitive/write-only:** private_key is Sensitive, resource-only and prior-state-preserved on blank/omitted API reads; public_key is readable
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/wireguard_server.md; docs/resources/wireguard_server.md; _bmad-output/implementation-artifacts/31-0e-wireguard-live-evidence.md

#### `resource:wireguard_settings` — covered

- **Domain / local mapping:** WireGuard / `wireguard_settings`
- **Baseline fields:** enabled, id
- **Shared semantics:** Released core WireGuard API parity proved on disposable OPNsense 26.7.7; typed sets preserve multiple addresses and selected-map API truth.
- **Baseline-only:** none within the audited field set
- **Local-only:** fixed import identity wireguard_settings
- **Lifecycle:** Import-first singleton; real Terraform Core Update while daemon stays disabled; state-only destroy and exact settings restoration proved
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/resources/wireguard_settings.md; internal/provider/provider.go; docs/resources/wireguard_settings.md; _bmad-output/implementation-artifacts/31-0e-wireguard-live-evidence.md

### Data Source

#### `data-source:cron_job` — covered

- **Domain / local mapping:** Cron / `cron_job`
- **Baseline fields:** command, days, description, enabled, hours, id, minutes, months, parameters, weekdays, who
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/cron_job.md; docs/data-sources/cron_job.md

#### `data-source:dnsmasq_host` — covered-broader

- **Domain / local mapping:** Dnsmasq / `dnsmasq_host`
- **Baseline fields:** alias_records, client_id, cname_records, comment, description, domain, hardware_addresses, hostname, id, ip_addresses, is_ignored, is_local_domain, tag
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** lease_time plus normalized singular field names
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/dnsmasq_host.md; docs/data-sources/dnsmasq_host.md

#### `data-source:firewall_alias` — covered

- **Domain / local mapping:** Firewall/NAT / `firewall_alias`
- **Baseline fields:** categories, content, description, enabled, id, interface, ip_protocol, name, path_expression, stats, type, update_freq
- **Shared semantics:** same UUID-backed alias lookup and complete advanced read-back, including interface, path_expression, counters/stats, fractional update_freq, and all selected protocols
- **Baseline-only:** none material after Story 31.0a field parity
- **Local-only:** deprecated proto scalar view retained beside canonical ip_protocol set
- **Lifecycle:** read-only API lookup; live OPNsense 26.7.7 run story-31-0a-live-26.7.7-review1-rerun2 asserted advanced alias data-source parity during create/import/drift/repair lifecycle, followed by zero-match cleanup
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/firewall_alias.md; docs/data-sources/firewall_alias.md

#### `data-source:firewall_category` — covered

- **Domain / local mapping:** Firewall/NAT / `firewall_category`
- **Baseline fields:** auto, color, id, name
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/firewall_category.md; docs/data-sources/firewall_category.md

#### `data-source:firewall_filter` — covered

- **Domain / local mapping:** Firewall/NAT / `firewall_filter_rule`
- **Baseline fields:** categories, description, enabled, filter, id, interface, internal_tagging, no_xmlrpc_sync, priority, sequence, source_routing, stateful_firewall, traffic_shaping
- **Shared semantics:** same UUID-backed filter-rule lookup and complete advanced read-back for tagging, priority/TOS, reply-to, state controls, schedules, TCP/ICMP, shaping, and HA-sync fields
- **Baseline-only:** none material after Story 31.0a advanced filter parity
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; live OPNsense 26.7.7 run story-31-0a-live-26.7.7-review1-rerun2 asserted advanced filter data-source parity during create/import/drift/update lifecycle, followed by zero-match cleanup
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/firewall_filter.md; docs/data-sources/firewall_filter_rule.md

#### `data-source:firewall_nat` — covered-different-name

- **Domain / local mapping:** Firewall/NAT / `firewall_nat_outbound`
- **Baseline fields:** description, destination, disable_nat, enabled, id, interface, ip_protocol, log, protocol, sequence, source, target
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** categories, static_nat_port, and target_port
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/firewall_nat.md; docs/data-sources/firewall_nat_outbound.md

#### `data-source:firewall_nat_one_to_one` — covered

- **Domain / local mapping:** Firewall/NAT / `firewall_nat_one_to_one`
- **Baseline fields:** categories, description, destination, enabled, external_net, id, interface, log, nat_reflection, sequence, source, type
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/firewall_nat_one_to_one.md; docs/data-sources/firewall_nat_one_to_one.md

#### `data-source:firewall_nat_port_forward` — covered

- **Domain / local mapping:** Firewall/NAT / `firewall_nat_port_forward`
- **Baseline fields:** description, destination, enabled, id, interface, ip_protocol, log, nat_reflection, protocol, sequence, source, target
- **Shared semantics:** same UUID-backed destination-NAT lookup and lossless multi-interface port-forward membership
- **Baseline-only:** none material after Story 31.0a multi-interface parity
- **Local-only:** deprecated interface scalar view retained beside canonical interfaces set
- **Lifecycle:** read-only API lookup; live OPNsense 26.7.7 run story-31-0a-live-26.7.7-review1-rerun2 asserted data-source membership parity across multi-interface create and membership update, followed by zero-match cleanup
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/firewall_nat_port_forward.md; docs/data-sources/firewall_nat_port_forward.md

#### `data-source:interface` — covered

- **Domain / local mapping:** Interfaces / `interface`
- **Baseline fields:** capabilities, device, flags, groups, ipv4, ipv6, is_physical, macaddr, media, media_raw, mtu, options, status, supported_media
- **Shared semantics:** complete pinned public typed allowlist, exact device selection and aggregate observations; diagnostic MTU Int64 and overview MTU String
- **Baseline-only:** none in the pinned public field contract; unobserved hardware and older firmware remain unverified
- **Local-only:** strict optional null preservation, complete pagination validation, deterministic tuple ordering and bounded read concurrency
- **Lifecycle:** GET-only observational reads; assignment/IP owners unchanged; fixture/Core coverage is not real appliance proof
- **Sensitive/write-only:** configuration, wireless and unknown backend fields excluded; public observations contain inventory and require protected Terraform state
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/interface.md; internal/provider/provider.go; internal/service/iface/live_interface_model.go; internal/service/iface/live_interface_api.go; docs/data-sources/interface.md; _bmad-output/implementation-artifacts/31-0i-interface-api-contract.md

#### `data-source:interface_all` — covered

- **Domain / local mapping:** Interfaces / `interface_all`
- **Baseline fields:** interfaces
- **Shared semantics:** complete pinned public typed allowlist, exact device selection and aggregate observations; diagnostic MTU Int64 and overview MTU String
- **Baseline-only:** none in the pinned public field contract; unobserved hardware and older firmware remain unverified
- **Local-only:** strict optional null preservation, complete pagination validation, deterministic tuple ordering and bounded read concurrency
- **Lifecycle:** GET-only observational reads; assignment/IP owners unchanged; fixture/Core coverage is not real appliance proof
- **Sensitive/write-only:** configuration, wireless and unknown backend fields excluded; public observations contain inventory and require protected Terraform state
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/interface_all.md; internal/provider/provider.go; internal/service/iface/live_interface_model.go; internal/service/iface/live_interface_api.go; docs/data-sources/interface_all.md; _bmad-output/implementation-artifacts/31-0i-interface-api-contract.md

#### `data-source:interfaces_overview` — covered

- **Domain / local mapping:** Interfaces / `interfaces_overview`
- **Baseline fields:** addr4, addr6, capabilities, description, device, enabled, flags, gateways, groups, identifier, ifctl_nameserver, ifctl_prefix, ifctl_router, ifctl_searchdomain, ipv4, ipv6, is_physical, lagg_hash, lagg_options, lagg_proto, lagg_statistics, link_type, macaddr, macaddr_hw, media, media_raw, mtu, nd6_flags, options, routes, status, supported_media, vlan, vlan_tag
- **Shared semantics:** complete pinned public typed allowlist, exact device selection and aggregate observations; diagnostic MTU Int64 and overview MTU String
- **Baseline-only:** none in the pinned public field contract; unobserved hardware and older firmware remain unverified
- **Local-only:** strict optional null preservation, complete pagination validation, deterministic tuple ordering and bounded read concurrency
- **Lifecycle:** GET-only observational reads; assignment/IP owners unchanged; fixture/Core coverage is not real appliance proof
- **Sensitive/write-only:** configuration, wireless and unknown backend fields excluded; public observations contain inventory and require protected Terraform state
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/interfaces_overview.md; internal/provider/provider.go; internal/service/iface/live_interface_model.go; internal/service/iface/live_interface_api.go; docs/data-sources/interfaces_overview.md; _bmad-output/implementation-artifacts/31-0i-interface-api-contract.md

#### `data-source:interfaces_overview_all` — covered

- **Domain / local mapping:** Interfaces / `interfaces_overview_all`
- **Baseline fields:** interfaces
- **Shared semantics:** complete pinned public typed allowlist, exact device selection and aggregate observations; diagnostic MTU Int64 and overview MTU String
- **Baseline-only:** none in the pinned public field contract; unobserved hardware and older firmware remain unverified
- **Local-only:** strict optional null preservation, complete pagination validation, deterministic tuple ordering and bounded read concurrency
- **Lifecycle:** GET-only observational reads; assignment/IP owners unchanged; fixture/Core coverage is not real appliance proof
- **Sensitive/write-only:** configuration, wireless and unknown backend fields excluded; public observations contain inventory and require protected Terraform state
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/interfaces_overview_all.md; internal/provider/provider.go; internal/service/iface/live_interface_model.go; internal/service/iface/live_interface_api.go; docs/data-sources/interfaces_overview_all.md; _bmad-output/implementation-artifacts/31-0i-interface-api-contract.md

#### `data-source:interfaces_vip` — covered-different-name

- **Domain / local mapping:** Interfaces / `system_vip`
- **Baseline fields:** description, gateway, id, interface, mode, network
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/interfaces_vip.md; docs/data-sources/system_vip.md

#### `data-source:interfaces_vlan` — covered-different-name

- **Domain / local mapping:** Interfaces / `system_vlan`
- **Baseline fields:** description, device, id, parent, priority, tag
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/interfaces_vlan.md; docs/data-sources/system_vlan.md

#### `data-source:kea_dhcpv4_peer` — covered-different-name

- **Domain / local mapping:** Kea / `kea_ha_peer`
- **Baseline fields:** id, name, role, url
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/kea_dhcpv4_peer.md; docs/data-sources/kea_ha_peer.md

#### `data-source:kea_dhcpv4_reservation` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv4_reservation`
- **Baseline fields:** description, hostname, id, ip_address, mac_address, subnet_id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/kea_dhcpv4_reservation.md; docs/data-sources/kea_dhcpv4_reservation.md

#### `data-source:kea_dhcpv4_subnet` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv4_subnet`
- **Baseline fields:** auto_collect, description, dns_servers, domain_name, domain_search, id, match_client_id, next_server, ntp_servers, pools, routers, static_routes, subnet, tftp_bootfile, tftp_server, time_servers
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** allocator, ping-check, and deterministic static-route pair decoding
- **Lifecycle:** Disposable OPNsense 26.7.7 run 20261018T233224Z proved UUID lookup and exact advanced API truth for DNS/domain/search, routers, static-route pairs, NTP/time, PXE/TFTP fields, and explicit clears
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/kea_dhcpv4_subnet.md; docs/data-sources/kea_dhcpv4_subnet.md

#### `data-source:kea_dhcpv6_pd_pool` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv6_pd_pool`
- **Baseline fields:** delegated_len, description, id, prefix, prefix_len, subnet_id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** Disposable OPNsense 26.7.7 run 20261018T233224Z proved UUID lookup and address-only API truth for prefix, prefix/delegated lengths, parent subnet, and description under the runtime compatibility gate
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/kea_dhcpv6_pd_pool.md; docs/data-sources/kea_dhcpv6_pd_pool.md

#### `data-source:kea_dhcpv6_peer` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv6_peer`
- **Baseline fields:** id, name, role, url
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** Disposable OPNsense 26.7.7 run 20261018T233224Z proved UUID lookup and exact API-truth name, URL, and selected role under the runtime 26.7.7 compatibility gate
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/kea_dhcpv6_peer.md; docs/data-sources/kea_dhcpv6_peer.md

#### `data-source:kea_dhcpv6_reservation` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv6_reservation`
- **Baseline fields:** description, domain_search, duid, hostname, id, ip_address, subnet_id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/kea_dhcpv6_reservation.md; docs/data-sources/kea_dhcpv6_reservation.md

#### `data-source:kea_dhcpv6_subnet` — covered

- **Domain / local mapping:** Kea / `kea_dhcpv6_subnet`
- **Baseline fields:** allocator, description, dns_servers, domain_search, id, interface, pd_allocator, pools, subnet
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** interface, allocator, PD allocator, and newline-compatible pools
- **Lifecycle:** Disposable OPNsense 26.7.7 run 20261018T233224Z proved UUID lookup and exact API-truth DNS/search members, interface/allocator relationships, and explicit clears
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/kea_dhcpv6_subnet.md; docs/data-sources/kea_dhcpv6_subnet.md

#### `data-source:kea_peer` — intentionally-different

- **Domain / local mapping:** Kea / `kea_ha_peer`
- **Baseline fields:** id, name, role, url
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/kea_peer.md; docs/data-sources/kea_ha_peer.md

#### `data-source:kea_reservation` — intentionally-different

- **Domain / local mapping:** Kea / `kea_dhcpv4_reservation`
- **Baseline fields:** description, hostname, id, ip_address, mac_address, subnet_id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/kea_reservation.md; docs/data-sources/kea_dhcpv4_reservation.md

#### `data-source:kea_subnet` — intentionally-different

- **Domain / local mapping:** Kea / `kea_dhcpv4_subnet`
- **Baseline fields:** auto_collect, description, dns_servers, domain_name, domain_search, id, match_client_id, next_server, ntp_servers, pools, routers, static_routes, subnet, tftp_bootfile, tftp_server, time_servers
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/kea_subnet.md; docs/data-sources/kea_dhcpv4_subnet.md

#### `data-source:openvpn_client_overwrite` — covered

- **Domain / local mapping:** OpenVPN / `openvpn_client_overwrite`
- **Baseline fields:** block, common_name, description, dns_domain, dns_domain_search, dns_servers, enabled, id, local_networks, ntp_servers, push_reset, redirect_gateway, register_dns, remote_networks, route_gateway, servers, tunnel_network, tunnel_network_v6, wins_servers
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** strict owner-local selected-map decoding and complete shared resource/data-source read model
- **Lifecycle:** Digest-bound OPNsense 26.7.7 proof compared the complete client-overwrite resource and data-source states and passed import/no-diff, drift/repair, explicit-clear, and zero-leftover checks
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/openvpn_client_overwrite.md; docs/data-sources/openvpn_client_overwrite.md

#### `data-source:openvpn_instance` — covered

- **Domain / local mapping:** OpenVPN / `openvpn_instance`
- **Baseline fields:** auth_digest, auth_gen_token, auth_gen_token_renewal, auth_gen_token_secret, auth_mode, bridge_gateway, bridge_pool, carp_depend_on, cert_depth, certificate, certificate_authority, compress_migrate, crl, data_ciphers, data_ciphers_fallback, description, dev_type, dns_domain, dns_domain_search, dns_servers, enabled, fragment, http_proxy, id, ifconfig_pool_persist, keepalive_interval, keepalive_timeout, local, local_group, max_clients, mss_fix, no_pool, ntp_servers, password, port, port_share, protocol, provision_exclusive, push_excluded_routes, push_inactive, push_route, redirect_gateway, register_dns, remote, remote_cert_tls, reneg_sec, role, route, route_metric, server, server_ipv6, strict_user_cn, tls_key, topology, tun_mtu, use_ocsp, username, username_as_common_name, various_flags, various_push_flags, verify_client_cert, verify_x509_name, vpn_id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** verb, API-read vpnid, additive ordered-cipher projection, and strict owner-local selection/numeric decoding
- **Lifecycle:** Digest-bound OPNsense 26.7.7 proof compared complete server, no-pool server, and client resource/data-source states except write-only credentials and passed import/no-diff, drift/repair, explicit-clear, and zero-leftover checks
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/openvpn_instance.md; docs/data-sources/openvpn_instance.md

#### `data-source:openvpn_static_key` — covered

- **Domain / local mapping:** OpenVPN / `openvpn_static_key`
- **Baseline fields:** description, id, key, mode
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** strict pre-request UUID validation and deterministic missing/malformed response diagnostics
- **Lifecycle:** Digest-bound OPNsense 26.7.7 proof performed a strict UUID GET-only lookup and asserted non-empty mode, description, and sensitive key state
- **Sensitive/write-only:** OpenVPN key is sensitive; the ephemeral form avoids state while the lookup stores a sensitive value
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/openvpn_static_key.md; docs/data-sources/openvpn_static_key.md

#### `data-source:quagga_bgp_aspath` — covered

- **Domain / local mapping:** Quagga/BGP / `quagga_bgp_aspath`
- **Baseline fields:** action, as, description, enabled, id, number
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/quagga_bgp_aspath.md; docs/data-sources/quagga_bgp_aspath.md

#### `data-source:quagga_bgp_communitylist` — covered

- **Domain / local mapping:** Quagga/BGP / `quagga_bgp_communitylist`
- **Baseline fields:** action, community, description, enabled, id, number, seq_number
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/quagga_bgp_communitylist.md; docs/data-sources/quagga_bgp_communitylist.md

#### `data-source:quagga_bgp_neighbor` — missing-buildable

- **Domain / local mapping:** Quagga/BGP / `quagga_bgp_neighbor`
- **Baseline fields:** as_override, attribute_unchanged, bfd, connect_timer, default_route, description, disable_connected_check, enabled, hold_down, id, keep_alive, link_local_interface, local_ip, md5_password, multi_hop, multi_protocol, next_hop_self, next_hop_self_all, peer_ip, prefix_list_in, prefix_list_out, remote_as, route_map_in, route_map_out, rr_client, update_source, weight
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** All 19 public advanced neighbor fields are represented locally and checked against current-appliance lookup state; authentication is intentionally config-owned and never exposed by lookup. Complete older-plugin/appliance/OpenTofu contract proof remains pending.
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/quagga_bgp_neighbor.md; docs/data-sources/quagga_bgp_neighbor.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0h-quagga-bgp-field-parity.md`

#### `data-source:quagga_bgp_prefixlist` — covered-different-name

- **Domain / local mapping:** Quagga/BGP / `quagga_prefix_list`
- **Baseline fields:** action, description, enabled, id, ip_version, name, network, number
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/quagga_bgp_prefixlist.md; docs/data-sources/quagga_prefix_list.md

#### `data-source:quagga_bgp_routemap` — missing-buildable

- **Domain / local mapping:** Quagga/BGP / `quagga_route_map`
- **Baseline fields:** action, aspaths, community_lists, description, enabled, id, name, prefix_lists, route_map_id, set
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** Existing owners represent AS-path/community/prefix relationships and numeric sequence; current 26.7.8_2/os-frr 2.0_4 grouped policy-order, withdrawal and restored-positive controls passed. Complete row-level contract/older-plugin/appliance/OpenTofu proof remains pending; no duplicate policy owner is required.
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/quagga_bgp_routemap.md; docs/data-sources/quagga_route_map.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0h-quagga-bgp-field-parity.md`

#### `data-source:route` — covered-different-name

- **Domain / local mapping:** Routing / `system_route`
- **Baseline fields:** description, enabled, gateway, id, network
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/route.md; docs/data-sources/system_route.md

#### `data-source:trust_ca` — missing-buildable

- **Domain / local mapping:** Trust / `trust_ca`
- **Baseline fields:** action, caref, city, common_name, country, crt, crt_payload, description, digest, email, id, key_type, lifetime, name, ocsp_uri, organization, organizational_unit, prv, prv_payload, ref_id, serial, state, valid_from, valid_to
- **Shared semantics:** Released26.7.8 Trust source-backed action/read continuation. Configured operation intent is distinct from inferred editor actions and lifetime template defaults.
- **Baseline-only:** Full non-skipped 26.7.8_2 action/settings/adoption/destroy/restoration matrix passed; see 31-0f-live-proof.json. Keyless CA import/update remains a provider implementation gate despite the patched controller correction; no XML or fabricated key workaround.
- **Local-only:** Separate returned public/sensitive observations preserve configured PEM and operation intent; explicit operation token distinguishes repeat operations from description updates.
- **Lifecycle:** Read-only UUID lookup; no mutation, activation or recovery. One-time local keys cannot be recovered by GET.
- **Sensitive/write-only:** All supplied and returned private-key material is Sensitive. Configured PEM inputs retain text; response-only generated keys are retained in resource state, never synthesized by lookup or retransmitted.
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/trust_ca.md; docs/data-sources/trust_ca.md; _bmad-output/implementation-artifacts/31-0f-trust-source-contract.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md`

#### `data-source:trust_cert` — missing-buildable

- **Domain / local mapping:** Trust / `trust_cert`
- **Baseline fields:** action, altnames_dns, altnames_email, altnames_ip, altnames_uri, caref, cert_type, city, common_name, country, crt, crt_payload, csr, csr_payload, description, digest, email, id, in_use, is_user, key_type, lifetime, name, ocsp_uri, organization, organizational_unit, private_key_location, prv, prv_payload, ref_id, rfc3280_purpose, state, valid_from, valid_to
- **Shared semantics:** Released26.7.8 Trust source-backed action/read continuation. Configured operation intent is distinct from inferred editor actions and lifetime template defaults.
- **Baseline-only:** Full non-skipped 26.7.8_2 action/settings/adoption/destroy/restoration matrix passed; see 31-0f-live-proof.json. Keyless CA import/update remains a provider implementation gate despite the patched controller correction; no XML or fabricated key workaround.
- **Local-only:** Separate returned public/sensitive observations preserve configured PEM and operation intent; explicit operation token distinguishes repeat operations from description updates.
- **Lifecycle:** Read-only UUID lookup; no mutation, activation or recovery. One-time local keys cannot be recovered by GET.
- **Sensitive/write-only:** All supplied and returned private-key material is Sensitive. Configured PEM inputs retain text; response-only generated keys are retained in resource state, never synthesized by lookup or retransmitted.
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/trust_cert.md; docs/data-sources/trust_cert.md; _bmad-output/implementation-artifacts/31-0f-trust-source-contract.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md`

#### `data-source:trust_settings` — missing-buildable

- **Domain / local mapping:** Trust / `trust_settings`
- **Baseline fields:** cipher_string, enable_config_constraints, enable_legacy_sect, fetch_crls, id, install_crls, store_intermediate_certs
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** No missing audited six-field owner/lookup; five additional upstream TLS fields remain deliberately unaudited and unmanaged.
- **Local-only:** Read-only six-field lookup with authoritative values, strict Boolean/selected-choice decoding, fixed computed identity and drift observations; HTTP-backed protocol evidence only.
- **Lifecycle:** GET-only lookup; no import, update, delete, mutation, activation or recovery ownership. Non-skipped 26.7.8_2 lookup/read comparison passed alongside the complete supported Trust matrix; see 31-0f-live-proof.json. Broader firmware/OpenTofu and keyless provider gates remain open.
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/trust_settings.md; internal/provider/provider.go; internal/service/trust/settings_data_source.go; internal/service/trust/settings_protocol_test.go; _bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md
- **Follow-up:** `_bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md`

#### `data-source:unbound_acl` — covered

- **Domain / local mapping:** Unbound / `unbound_acl`
- **Baseline fields:** action, description, enabled, id, name, networks
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/unbound_acl.md; docs/data-sources/unbound_acl.md

#### `data-source:unbound_domain_override` — covered

- **Domain / local mapping:** Unbound / `unbound_domain_override`
- **Baseline fields:** description, domain, enabled, id, server
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/unbound_domain_override.md; docs/data-sources/unbound_domain_override.md

#### `data-source:unbound_forward` — covered

- **Domain / local mapping:** Unbound / `unbound_domain_override`
- **Baseline fields:** domain, enabled, id, server_ip, server_port, type, verify_cn
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** description
- **Lifecycle:** digest-bound OPNsense 26.7.7 acceptance proves authoritative query and DNS-over-TLS reads with complete server_port, type, verify_cn, and existing-field parity across independent lifecycle and import boundaries
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/unbound_forward.md; docs/data-sources/unbound_domain_override.md; _bmad-output/implementation-artifacts/31-0d-live-evidence.json

#### `data-source:unbound_host_alias` — covered

- **Domain / local mapping:** Unbound / `unbound_host_alias`
- **Baseline fields:** description, domain, enabled, hostname, id, override
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/unbound_host_alias.md; docs/data-sources/unbound_host_alias.md

#### `data-source:unbound_host_override` — covered

- **Domain / local mapping:** Unbound / `unbound_host_override`
- **Baseline fields:** description, domain, enabled, hostname, id, mx_host, mx_priority, server, type
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** read-only API lookup; operational interface rows read live state rather than durable configuration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/unbound_host_override.md; docs/data-sources/unbound_host_override.md

#### `data-source:unbound_settings` — covered

- **Domain / local mapping:** Unbound / `unbound_settings`
- **Baseline fields:** acls, advanced, dnsbl, forwarding, general, id
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after semantic name/type normalization
- **Local-only:** none material
- **Lifecycle:** digest-bound OPNsense 26.7.7 acceptance proves complete authoritative general, advanced, ACL-default, forwarding, and every DNSBL item field plus canonical dnsbl_json equality before exact settings restoration
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/unbound_settings.md; docs/resources/unbound_general.md; docs/resources/unbound_dnsbl.md; _bmad-output/implementation-artifacts/31-0d-live-evidence.json

#### `data-source:wireguard_client` — covered-different-name

- **Domain / local mapping:** WireGuard / `wireguard_peer`
- **Baseline fields:** enabled, id, keep_alive, name, psk, public_key, server_address, server_port, tunnel_address
- **Shared semantics:** Released core WireGuard API parity proved on disposable OPNsense 26.7.7; typed sets preserve multiple addresses and selected-map API truth.
- **Baseline-only:** PSK intentionally remains resource-only; read-only data sources never expose write-only secrets
- **Local-only:** API-derived servers backreference; keepalive is the local spelling of keep_alive
- **Lifecycle:** Read-only authoritative API lookup; import/refresh/no-change state comparison proved
- **Sensitive/write-only:** PSK is Sensitive, resource-only and prior-state-preserved on blank/omitted API reads
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/wireguard_client.md; docs/data-sources/wireguard_peer.md; _bmad-output/implementation-artifacts/31-0e-wireguard-live-evidence.md

#### `data-source:wireguard_server` — covered

- **Domain / local mapping:** WireGuard / `wireguard_server`
- **Baseline fields:** disable_routes, dns, enabled, gateway, id, instance, mtu, name, peers, port, private_key, public_key, tunnel_address
- **Shared semantics:** Released core WireGuard API parity proved on disposable OPNsense 26.7.7; typed sets preserve multiple addresses and selected-map API truth.
- **Baseline-only:** private_key intentionally remains resource-only; read-only data sources never expose write-only secrets
- **Local-only:** compatibility description state; instance is API-assigned and Computed
- **Lifecycle:** Read-only authoritative API lookup; import/refresh/no-change state comparison proved
- **Sensitive/write-only:** private_key is Sensitive, resource-only and prior-state-preserved on blank/omitted API reads; public_key is readable
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/wireguard_server.md; docs/data-sources/wireguard_server.md; _bmad-output/implementation-artifacts/31-0e-wireguard-live-evidence.md

#### `data-source:wireguard_settings` — covered

- **Domain / local mapping:** WireGuard / `wireguard_settings`
- **Baseline fields:** enabled, id
- **Shared semantics:** Released core WireGuard API parity proved on disposable OPNsense 26.7.7; typed sets preserve multiple addresses and selected-map API truth.
- **Baseline-only:** none within the audited field set
- **Local-only:** fixed import identity wireguard_settings
- **Lifecycle:** Read-only authoritative API lookup; import/refresh/no-change state comparison proved
- **Sensitive/write-only:** none
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/data-sources/wireguard_settings.md; internal/provider/provider.go; docs/data-sources/wireguard_settings.md; _bmad-output/implementation-artifacts/31-0e-wireguard-live-evidence.md

### Ephemeral Resource

#### `ephemeral-resource:openvpn_generate_key` — covered

- **Domain / local mapping:** OpenVPN / `openvpn_generate_key`
- **Baseline fields:** key, key_type
- **Shared semantics:** same stable OPNsense object/endpoint family; top-level field names are enumerated above and semantic renames are recorded in the local generated schema
- **Baseline-only:** none material after correcting the pinned baseline's query-string path and unsupported client key-type behavior to the authoritative OPNsense contract
- **Local-only:** source-correct path-segment POST contract limited to secret, auth-token, tls-auth, tls-crypt, and tls-crypt-v2-server
- **Lifecycle:** Digest-bound OPNsense 26.7.7 proof opened all five key types through Terraform Core and write-only consumers; exact generated values were absent from the actual saved plan, state, and Terraform log
- **Sensitive/write-only:** OpenVPN key is sensitive; the ephemeral form avoids state while the lookup stores a sensitive value
- **Sources:** https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/docs/ephemeral-resources/openvpn_generate_key.md; internal/provider/provider.go

## Local Inventory Exceptions

Resource-only same-name data-source gaps (11): `kea_ctrl_agent`, `kea_dhcpv4_settings`, `kea_dhcpv6_settings`, `quagga_bgp_global`, `quagga_general`, `quagga_ospf6_general`, `quagga_ospf_general`, `quagga_rip`, `quagga_static`, `unbound_dnsbl`, `unbound_general`.

Data-only: `interface`, `interface_all`, `interfaces_overview`, `interfaces_overview_all`, `system_info`, `unbound_settings`.
