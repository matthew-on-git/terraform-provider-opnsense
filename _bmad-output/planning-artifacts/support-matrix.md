---
title: Provider Support Matrix
date: 2026-06-02
author: BMad PM
status: current
inputs:
  - prd.md
  - core-config-gap-analysis.md
  - feature-complete-roadmap.md
  - repository implementation inventory
---

# Provider Support Matrix

This matrix is the current source of truth for release positioning. It reconciles the expanded PRD/roadmap with the implementation present in the repository.

## Summary

| Area | Count | Status |
|---|---:|---|
| Resources | 97 | Supported |
| Data sources | 83 | Supported |
| Resource docs | 97 | Supported |
| Data source docs | 83 | Supported |
| Remaining resource gaps | 2 verified provider-owned candidates plus research and upstream-blocked domains | Coming, Needs research, or upstream-blocked |
| Remaining data-source gaps | 15 | Coming |

Data-source parity is tracked in [data-source-parity-plan.md](data-source-parity-plan.md). Remaining resource-gap verification is tracked in [resource-gap-verification.md](resource-gap-verification.md).

## Supported Resources

| Domain | Supported resources |
|---|---|
| ACME | account, certificate, challenge |
| Auth | group, user |
| Cron | job |
| DHCPv4 | reservation, subnet |
| Dnsmasq | boot, domain, host, option, range, settings, tag |
| Dynamic DNS | ddclient account |
| Firewall | alias, category, filter rule, NAT one-to-one, NAT outbound, NAT port forward |
| HAProxy | ACL, backend, frontend, health check, server |
| Interfaces | bridge, GIF, GRE, loopback, neighbor, VXLAN |
| IPsec | child, connection, key pair, local, manual SPD, pool, PSK, remote, VTI |
| Kea | control agent, DHCPv6 reservation, DHCPv6 settings, DHCPv6 subnet, HA peer |
| Monit | alert, service, test |
| OpenVPN | client overwrite, instance, static key |
| Quagga / FRR | BGP AS path, BGP community list, BGP global, BGP neighbor, BGP peer group, BGP redistribution, general, OSPF area/general/interface/neighbor/network/prefix list/redistribution/route map, OSPFv3 general/interface/network/prefix list/redistribution/route map, prefix list, RIP, route map, static general, static route |
| Syslog | destination |
| System | gateway, route, VIP, VLAN |
| Traffic shaper | pipe, queue, rule |
| Trust | CA, certificate |
| Unbound | ACL, DNSBL, domain override, general, host alias, host override |
| WireGuard | peer, server |

## Supported Data Sources

| Domain | Supported data sources |
|---|---|
| Auth | group, user |
| Cron | job |
| ACME | account, certificate, challenge |
| DHCPv4 | reservation, subnet |
| Dnsmasq | boot, domain, host, option, range, tag |
| Dynamic DNS | ddclient account |
| Firewall | alias, category, filter rule, NAT one-to-one, NAT outbound, NAT port forward |
| HAProxy | ACL, backend, frontend, health check, server |
| Interfaces | bridge, GIF, GRE, loopback, neighbor, VXLAN |
| IPsec | child, connection, local, manual SPD, pool, remote, VTI |
| Kea | DHCPv6 reservation, DHCPv6 subnet, HA peer |
| Monit | alert, service, test |
| Quagga / FRR | BGP AS path, BGP community list, BGP neighbor, BGP peer group, BGP redistribution, OSPF area/interface/neighbor/network/prefix list/redistribution/route map, OSPFv3 interface/network/prefix list/redistribution/route map, prefix list, route map, static route |
| Syslog | destination |
| OpenVPN | client overwrite, instance |
| System | gateway, route, system info, VIP, VLAN |
| Traffic shaper | pipe, queue, rule |
| Trust | CA |
| Unbound | ACL, domain override, host alias, host override |
| WireGuard | peer, server |

## Coming: Buildable Provider Work

| Domain | Work |
|---|---|
| Data-source parity | Add read-only data sources for the 15 supported singleton/sensitive special-case resources that do not yet have data-source counterparts. |
| Interfaces | Add LAGG; published API docs confirm `interfaces/lagg_settings`, but implementation still needs live validation with assignable member interfaces. |
| System | Add tunables/sysctl item resource after live validation; current `core/tunables` docs/source confirm persistent item CRUD/search plus `reconfigure`, but safety documentation is required before implementation. |
| Documentation | Fill missing resource templates for generated docs, expand composition examples, and keep the provider index support matrix current. |
| Release hardening | Keep release workflow, Registry manifest, changelog, and provider docs verified for subsequent patch/minor releases. |

## Needs Research

| Domain | Research needed |
|---|---|
| Kea | DHCPv4 option and Kea DDNS have published endpoints but conflicting live endpoint-not-found evidence; move to Coming only after live re-probe confirms target availability. |
| System / HA | HASync configuration needs request/response shape research because current `Hasync.xml` uses dynamic `JsonKeyValueStoreField` `syncitems`; HASync status `services`/`version` are data-source candidates after live validation, while service operations are action candidates only after product/framework decision. |

Source NAT is already supported as `opnsense_firewall_nat_outbound`. Unbound forward is already supported as `opnsense_unbound_domain_override`.

## Upstream-Blocked

Confirmed blockers and the maintenance workflow are documented publicly in [`docs/upstream-blocked.md`](../../docs/upstream-blocked.md).

| Resource/domain | Reason | Action |
|---|---|---|
| Interface base assignment / IP config / PPPoE | No stable OPNsense API in current target release. | Track and test OPNsense PR #8436. |
| Gateway group | No usable gateway-group endpoint. | Candidate upstream MVC API contribution. |
| System general settings | Legacy/non-MVC API status. | Watch OPNsense System Settings MVC roadmap. |

## Release Readiness Snapshot

| Item | Status | Evidence / gap |
|---|---|---|
| GoReleaser config | Ready | `.goreleaser.yml` cross-compiles, archives ZIPs, emits checksums, signs checksum artifact. |
| Release workflow | Ready with secrets | `.github/workflows/release.yml` runs on `v*` tags; requires `GPG_PRIVATE_KEY` and `PASSPHRASE` secrets. |
| Terraform Registry manifest | Ready | `terraform-registry-manifest.json` declares protocol `6.0`. |
| Changelog | Published baseline | `CHANGELOG.md` lists 90 resources and 34 data sources for `0.1.0`; current post-release work has added more data sources. |
| Provider index | Mostly ready | Covers auth, minimum version, permissions, and quickstart; support matrix now added to planning artifacts and should stay reflected in Registry docs. |
| Migration/import guidance | Partial | Per-resource import examples exist; broader dependency-order migration guide remains useful follow-up. |
