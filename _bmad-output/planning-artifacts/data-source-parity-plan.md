---
title: Data Source Parity Plan
date: 2026-11-03
author: BMad Dev
status: current
inputs:
  - support-matrix.md
  - core-config-gap-analysis.md
  - post-release-epics.md
  - internal/service/*/exports.go
  - docs/resources/*.md
  - docs/data-sources/*.md
---

# Data Source Parity Plan

## Purpose

This plan defines the remaining data-source work after v0.1.0. It is an implementation handoff for Stories 25B.2 and 25B.3, not an implementation artifact. No provider code was changed to produce this inventory.

## Inventory Summary

| Check | Command | Result |
|---|---|---:|
| Registered resources | Provider registration metadata | 113 |
| Registered data sources | Provider registration metadata | 106 |
| Resource docs | `find docs/resources -name '*.md'` | 113 |
| Data-source docs | `find docs/data-sources -name '*.md'` | 106 |
| Resource-matching data-source gaps | Resource doc basenames minus data-source doc basenames | 13 |

No mismatch was found between registered implementation counts and generated documentation counts. The current gap is 13, not 113 minus 106, because six runtime/standalone lookups have no same-name resource. The OpenVPN static-key lookup is now published (v0.7.0), so it is not a remaining gap. The canonical names and drift checks now live in `provider-ecosystem-coverage-audit.json` and `internal/coverageaudit`.

## Existing Data Sources

The complete deterministic inventory is the set registered by `OpnsenseProvider.DataSources` in `internal/provider/provider.go` and its domain `DataSources()` constructors. Generated files under `docs/data-sources/` are the independently counted public projection of that registration set. This plan intentionally does not duplicate those names in a hand-maintained table, which can silently omit newly shipped surfaces while leaving the totals unchanged. `provider-ecosystem-coverage-audit.json` and `internal/coverageaudit` enforce registration/documentation parity.

## Missing Data-Source Inventory

| Resource | Domain | Current data source? | Pattern | Priority | Batch | Notes |
|---|---|---|---|---|---|---|
| dnsmasq_settings | Dnsmasq | Yes | Singleton | Low | 4 | Implemented by 31.0q as a read-only public allowlist. Live proof remains separate. |
| interface_assignment | Interfaces | Yes | Name-keyed | Medium | 4 | Implemented by 31.0r as a read-only public allowlist. Identity is `wan`, `lan`, or `optN`, not a UUID. Live proof remains separate. |
| ipsec_key_pair | IPsec | No | UUID-backed-special | Medium | 4 | Sensitive/write-only fields require careful state expectations. |
| ipsec_psk | IPsec | Yes | UUID-backed-special | Medium | 4 | Implemented by 31.0t as a read-only public allowlist. The secret is never lookup state. Live proof remains separate. |
| kea_ctrl_agent | Kea | No | Singleton | Low | 4 | Singleton read pattern may differ from UUID resources. |
| kea_dhcpv4_settings | Kea | No | Singleton | Low | 4 | Singleton/settings pattern. |
| kea_dhcpv6_settings | Kea | No | Singleton | Low | 4 | Singleton/settings pattern. |
| quagga_bgp_global | Quagga / FRR | No | Singleton | Low | 4 | Singleton resource. |
| quagga_general | Quagga / FRR | No | Singleton | Low | 4 | Singleton resource. |
| quagga_ospf6_general | Quagga / FRR | No | Singleton | Low | 4 | Singleton resource; sub-resource data sources exist. |
| quagga_ospf_general | Quagga / FRR | No | Singleton | Low | 4 | Singleton resource; sub-resource data sources exist. |
| quagga_rip | Quagga / FRR | No | Singleton | Low | 4 | Singleton resource. |
| quagga_static | Quagga / FRR | No | Singleton | Low | 4 | Singleton resource. |
| trust_cert | Trust | No | UUID-backed-special | Medium | 4 | Certificate/key material may include write-only fields. |
| unbound_dnsbl | Unbound | No | Singleton | Low | 4 | Singleton/blocklist settings. |
| unbound_general | Unbound | No | Singleton | Low | 4 | Singleton settings. |

## Batch 1: High-Reference Data Sources

Status: completed in Story 25B.2. These rows are retained as implementation history; they are no longer part of the missing inventory.

Goal: implement the data sources that unlock the most brownfield migration and composition value.

| Data source | Files to inspect | Likely pattern | Docs/examples needed | Risks |
|---|---|---|---|---|
| haproxy_server | `internal/service/haproxy/server_*`, `internal/service/haproxy/exports.go` | UUID-backed hand-written or generated equivalent | `docs/data-sources/haproxy_server.md`, example data source | Referenced by backend chains. |
| haproxy_backend | `internal/service/haproxy/backend_*` | UUID-backed | Data-source docs/example | Referenced by frontends. |
| haproxy_frontend | `internal/service/haproxy/frontend_*` | UUID-backed | Data-source docs/example | Linked ACL/backend fields. |
| haproxy_acl | `internal/service/haproxy/acl_*` | UUID-backed | Data-source docs/example | Frontend routing references. |
| haproxy_healthcheck | `internal/service/haproxy/healthcheck_*` | UUID-backed | Data-source docs/example | Backend health check references. |
| firewall_category | `internal/service/firewall/category_*` | UUID-backed | Data-source docs/example | Category references in firewall objects. |
| firewall_filter_rule | `internal/service/firewall/filter_rule_*` | UUID-backed | Data-source docs/example | Safety-critical; read-only lookup only. |
| firewall_nat_port_forward | `internal/service/firewall/nat_port_forward_*` | UUID-backed | Data-source docs/example | NAT migration. |
| firewall_nat_outbound | `internal/service/firewall/nat_outbound_*` | UUID-backed | Data-source docs/example | NAT migration. |
| system_vlan | `internal/service/system/vlan_*` | UUID-backed | Data-source docs/example | Interface naming may vary by appliance. |
| system_vip | `internal/service/system/vip_*` | UUID-backed | Data-source docs/example | CARP/write-only fields may require care. |
| system_route | `internal/service/system/route_*` | UUID-backed | Data-source docs/example | Gateway selected map conversion. |
| system_gateway | `internal/service/system/gateway_*` | UUID-backed | Data-source docs/example | Gateway selected map conversion. |
| wireguard_server | `internal/service/wireguard/server_*` | UUID-backed-special | Data-source docs/example | Private key is sensitive/write-only. |
| wireguard_peer | `internal/service/wireguard/peer_*` | UUID-backed | Data-source docs/example | Peer references. |
| openvpn_instance | `internal/service/openvpn/instance_*` | UUID-backed-special | Data-source docs/example | Key/cert fields may be write-only. |
| openvpn_client_overwrite | `internal/service/openvpn/client_overwrite_*` | UUID-backed | Data-source docs/example | Instance linkage. |
| ipsec_connection | `internal/service/ipsec/connection_*` | UUID-backed | Data-source docs/example | Parent of children. |
| ipsec_child | `internal/service/ipsec/child_*` | UUID-backed | Data-source docs/example | Connection linkage. |
| ipsec_local | `internal/service/ipsec/local_*` | UUID-backed | Data-source docs/example | Connection linkage. |
| ipsec_remote | `internal/service/ipsec/remote_*` | UUID-backed | Data-source docs/example | Connection linkage. |

## Batch 2: Routing Data Sources

Status: completed in Story 25B.3. These rows are retained as implementation history; they are no longer part of the missing inventory.

Goal: fill routing policy lookup gaps after OSPF/OSPFv3 item data sources already exist.

| Data source | Files to inspect | Likely pattern | Docs/examples needed | Risks |
|---|---|---|---|---|
| quagga_bgp_neighbor | `internal/service/quagga/bgp_neighbor_*` | UUID-backed | Data-source docs/example | High migration value. |
| quagga_prefix_list | `internal/service/quagga/prefix_list_*` | UUID-backed | Data-source docs/example | Route policy references. |
| quagga_route_map | `internal/service/quagga/route_map_*` | UUID-backed | Data-source docs/example | Route policy references. |
| quagga_bgp_aspath | `internal/service/quagga/bgp_aspath_*` | UUID-backed | Data-source docs/example | Policy lookup. |
| quagga_bgp_communitylist | `internal/service/quagga/bgp_communitylist_*` | UUID-backed | Data-source docs/example | Policy lookup. |
| quagga_bgp_peergroup | `internal/service/quagga/bgp_peergroup_*` | UUID-backed | Data-source docs/example | Neighbor composition. |
| quagga_bgp_redistribution | `internal/service/quagga/bgp_redistribution_*` | UUID-backed | Data-source docs/example | Routing migration. |
| quagga_static_route | `internal/service/quagga/static_route_*` | UUID-backed | Data-source docs/example | Static routing migration. |

## Batch 3: DNS, DHCP, ACME, Dynamic DNS, and Trust Data Sources

Status: completed in Story 25B.3. These rows are retained as implementation history; they are no longer part of the missing inventory.

Goal: support full-appliance import and common service lookups.

| Data source | Files to inspect | Likely pattern | Docs/examples needed | Risks |
|---|---|---|---|---|
| unbound_host_override | `internal/service/unbound/host_override_*` | UUID-backed | Data-source docs/example | DNS migration. |
| unbound_host_alias | `internal/service/unbound/host_alias_*` | UUID-backed | Data-source docs/example | Host override linkage. |
| unbound_domain_override | `internal/service/unbound/domain_override_*` | UUID-backed | Data-source docs/example | Forward/domain lookup. |
| unbound_acl | `internal/service/unbound/acl_*` | UUID-backed | Data-source docs/example | ACL migration. |
| dhcpv4_subnet | `internal/service/dhcp/subnet_*` | UUID-backed | Data-source docs/example | DHCP migration. |
| dhcpv4_reservation | `internal/service/dhcp/reservation_*` | UUID-backed | Data-source docs/example | Static mapping migration. |
| kea_dhcpv6_subnet | `internal/service/kea/dhcpv6_subnet_*` | UUID-backed | Data-source docs/example | DHCPv6 migration. |
| kea_dhcpv6_reservation | `internal/service/kea/dhcpv6_reservation_*` | UUID-backed | Data-source docs/example | DHCPv6 migration. |
| ddclient_account | `internal/service/ddclient/account_*` | UUID-backed-special | Data-source docs/example | Password fields likely sensitive/write-only. |
| acme_account | `internal/service/acme/account_*` | UUID-backed | Data-source docs/example | Account registration fields. |
| acme_certificate | `internal/service/acme/certificate_*` | UUID-backed | Data-source docs/example | Certificate status/issuance fields. |
| acme_challenge | `internal/service/acme/challenge_*` | UUID-backed | Data-source docs/example | Provider-specific fields. |
| trust_ca | `internal/service/trust/ca_*` | UUID-backed-special | Data-source docs/example | Certificate material sensitivity. |

## Batch 4: Singleton, Sensitive, and Special-Case Data Sources

Goal: keep non-standard cases from blocking straightforward UUID-backed data sources.

| Data source | Files to inspect | Likely pattern | Docs/examples needed | Risks |
|---|---|---|---|---|
| dnsmasq_settings | `internal/service/dnsmasq/settings_data_source.go` | Singleton | Data-source docs/example | Implemented by 31.0q. No UUID lifecycle. `expand_hosts` remains unowned. |
| kea_ctrl_agent | `internal/service/kea/ctrl_agent_*` | Singleton | Data-source docs/example | No UUID lifecycle. |
| kea_dhcpv6_settings | `internal/service/kea/dhcpv6_settings_*` | Singleton | Data-source docs/example | No UUID lifecycle. |
| quagga_general | `internal/service/quagga/general_*` | Singleton | Data-source docs/example | No UUID lifecycle. |
| quagga_bgp_global | `internal/service/quagga/bgp_global_*` | Singleton | Data-source docs/example | No UUID lifecycle. |
| quagga_ospf_general | `internal/service/quagga/ospf_general_*` | Singleton | Data-source docs/example | No UUID lifecycle. |
| quagga_ospf6_general | `internal/service/quagga/ospf6_general_*` | Singleton | Data-source docs/example | No UUID lifecycle. |
| quagga_rip | `internal/service/quagga/rip_*` | Singleton | Data-source docs/example | No UUID lifecycle. |
| quagga_static | `internal/service/quagga/static_general_*` | Singleton | Data-source docs/example | No UUID lifecycle. |
| unbound_general | `internal/service/unbound/general_*` | Singleton | Data-source docs/example | No UUID lifecycle. |
| unbound_dnsbl | `internal/service/unbound/dnsbl_*` | Singleton | Data-source docs/example | No UUID lifecycle. |
| ipsec_psk | `internal/service/ipsec/psk_*` | UUID-backed-special | Data-source docs/example | Implemented by 31.0t. Secret material is never lookup state. |
| ipsec_key_pair | `internal/service/ipsec/keypair_*` | UUID-backed-special | Data-source docs/example | Private key is write-only. |
| trust_cert | `internal/service/trust/cert_*` | UUID-backed-special | Data-source docs/example | Private key/certificate material sensitivity. |

## Deferred and Exception Rules

No candidates are deferred by this inventory. Batch 4 items are not rejected; they are isolated because they need explicit implementation decisions around singleton IDs or write-only/sensitive fields.

Future stories may defer a candidate only with a concrete reason, such as no stable read endpoint, no meaningful data-source lookup semantics, or write-only fields that make the data source misleading.

## Verification Requirements for Implementation Stories

Stories 25B.2 and 25B.3 should use this plan as input and verify each implemented batch with:

```bash
make check
```

Recommended local checks during implementation:

```bash
rg "new.*DataSource" internal/service/<domain>
rg "^# opnsense_" docs/data-sources
```

Batch-specific checks should compare the implemented batch rows against constructor registration, generated data-source docs, and examples. For example:

```bash
rg "new(Haproxy|Firewall|System|Wireguard|Openvpn|Ipsec).*DataSource" internal/service
rg "^# opnsense_(haproxy|firewall|system|wireguard|openvpn|ipsec)_" docs/data-sources
rg "data \"opnsense_(haproxy|firewall|system|wireguard|openvpn|ipsec)_" examples/data-sources
```

Use equivalent domain filters for Batch 2 (`quagga`) and Batch 3 (`unbound`, `dhcp`, `kea`, `ddclient`, `acme`, `trust`). Batch 4 requires an explicit team decision on singleton IDs and sensitive/write-only fields before implementation.

Data-source implementation stories should update this plan as each batch is completed, moving implemented items out of the missing inventory or adding a status column if partial completion is easier to track.

## Next-Story Handoff

Story 25B.2 should implement Batch 1 first. It has the highest user value for migration and composition: HAProxy, firewall, system, WireGuard/OpenVPN, and high-reference IPsec resources.

Story 25B.3 should implement Batch 2 and Batch 3 after Batch 1 is complete.

Batch 4 is intentionally outside Stories 25B.2 and 25B.3. Create a follow-up story after the team agrees on singleton ID behavior and sensitive/write-only field expectations.
