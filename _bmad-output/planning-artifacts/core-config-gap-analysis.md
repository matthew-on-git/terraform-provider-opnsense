---
title: OPNsense Core Config — Terraform Provider Gap Analysis
date: 2026-06-02
author: BMad PM (reconciliation)
status: current
purpose: Define the feature-complete finish line and classify every core-config domain as Supported, Coming, Needs research, Upstream-blocked, or Not planned.
---

# Core Config Gap Analysis

## Goal

A feature-complete Terraform provider covers all OPNsense core configuration that has a stable usable API, and clearly labels anything that is not yet buildable because OPNsense itself lacks the required endpoint.

## Status Legend

- **Supported** — resource/data source exists in the provider and has generated docs.
- **Coming** — OPNsense API appears buildable; provider work remains.
- **Needs research** — published or suspected endpoint evidence exists, but target-version availability, wrapper shape, durable Terraform semantics, or product priority are not yet clear.
- **Upstream-blocked** — no usable API in the target OPNsense release; requires OPNsense work first.
- **Not planned** — redundant or intentionally out of scope for the current product definition.

## Current Scorecard

| Area | Count | Status |
|---|---:|---|
| Resources | 102 | Supported |
| Data sources | 88 | Supported |
| Resource docs | 102 | Supported |
| Data source docs | 88 | Supported |
| Resource-matching data-source backlog | 15 | Coming |
| Verified provider-owned resource candidates | 0 plus research follow-ups | Needs research |

## Domain-by-Domain Status

### Interfaces

| Item | Status | Notes |
|---|---|---|
| VLAN | Supported | `opnsense_system_vlan` |
| Virtual IP | Supported | `opnsense_system_vip` |
| Bridge | Supported | Resource + data source |
| GRE | Supported | Resource + data source |
| GIF | Supported | Resource + data source |
| VXLAN | Supported | Resource + data source |
| Loopback | Supported | Resource + data source |
| Neighbor / static ARP/NDP | Supported | Resource + data source |
| LAGG | Supported | Resource + data source; live-validated against Vagrant with dedicated `em4`/`em5` member interfaces. |
| Base assignment / IP config / PPPoE | Upstream-blocked | Story 5.1 revalidated on 2026-06-12: OPNsense `master` contains an emerging `interfaces/assignment` API backed by `NetworkInterface`, but it is absent from `stable/26.1`, absent from published interface API docs, missing ACL coverage, and does not cover IP configuration or PPPoE. Track PR #8436 and target-release availability. |

### Firewall & NAT

| Item | Status | Notes |
|---|---|---|
| Alias | Supported | Resource + data source |
| Category | Supported | Resource only |
| Filter rule | Supported | Resource only; savepoint/apply/cancelRollback path implemented. |
| NAT port forward | Supported | Resource only |
| NAT outbound | Supported | Resource only |
| NAT one-to-one | Supported | Resource + data source |
| Source NAT | Supported | Already shipped as `opnsense_firewall_nat_outbound` using `/api/firewall/source_nat/*_rule`. |
| Schedules | Coming | Verify API and product priority before adding. |

### Routing & Gateways

| Item | Status | Notes |
|---|---|---|
| Static route | Supported | Resource only |
| Gateway | Supported | Resource only |
| Gateway group | Upstream-blocked | Story 5.6 revalidated on 2026-06-14: published docs list only individual gateway/static route endpoints; OPNsense `master` has model-only `GatewayGroups` evidence with tier fields, trigger, pool options, and description, but no published endpoint/API controller was found and checked `stable/26.1` model paths returned 404. |

### Dynamic Routing (FRR / Quagga)

| Item | Status | Notes |
|---|---|---|
| FRR general | Supported | Singleton resource |
| BGP global config | Supported | Singleton resource |
| BGP neighbor | Supported | Resource |
| BGP prefix list | Supported | Resource |
| BGP route map | Supported | Resource |
| BGP AS path | Supported | Resource |
| BGP community list | Supported | Resource |
| BGP peer group | Supported | Resource |
| BGP redistribution | Supported | Resource |
| OSPF general/area/interface/neighbor/network/prefix list/redistribution/route map | Supported | Resources; sub-resource data sources exist. |
| OSPFv3 general/interface/network/prefix list/redistribution/route map | Supported | Resources; sub-resource data sources exist. |
| RIP | Supported | Singleton resource |
| FRR static general + route | Supported | Resources |

### VPN

| Item | Status | Notes |
|---|---|---|
| WireGuard server | Supported | Resource only |
| WireGuard peer | Supported | Resource only |
| OpenVPN instance | Supported | Resource only |
| OpenVPN client overwrite | Supported | Resource only |
| OpenVPN static key | Supported | Resource only |
| IPsec connection | Supported | Resource only |
| IPsec child | Supported | Resource only |
| IPsec PSK | Supported | Resource only |
| IPsec local | Supported | Resource only |
| IPsec remote | Supported | Resource only |
| IPsec pool | Supported | Resource + data source |
| IPsec VTI | Supported | Resource + data source |
| IPsec manual SPD | Supported | Resource + data source |
| IPsec key pair | Supported | Resource only |

### DHCP and Dynamic DNS

| Item | Status | Notes |
|---|---|---|
| DHCPv4 subnet | Supported | Resource only |
| DHCPv4 reservation | Supported | Resource only |
| Kea control agent | Supported | Singleton resource |
| Kea DHCPv6 settings | Supported | Resource |
| Kea DHCPv6 subnet | Supported | Resource |
| Kea DHCPv6 reservation | Supported | Resource |
| Kea HA peer | Supported | Resource + data source |
| DHCPv4 option | Needs research | Published OPNsense API docs expose `kea/dhcpv4` option CRUD/search endpoints, but earlier live build returned endpoint-not-found; re-probe live target before implementation. |
| Kea DDNS | Needs research | Published OPNsense API docs expose `kea/ddns` singleton get/set endpoints, but earlier live build returned endpoint-not-found; re-probe live target before implementation. |
| Dynamic DNS account | Supported | Resource only |
| DDNS provider | Not planned | Redundant with provider/service field on account unless a separate endpoint emerges. |

### DNS

| Item | Status | Notes |
|---|---|---|
| Unbound host override | Supported | Resource only |
| Unbound host alias | Supported | Resource only |
| Unbound domain override | Supported | Resource only |
| Unbound ACL | Supported | Resource only |
| Unbound general | Supported | Singleton resource |
| Unbound DNSBL/blocklist | Supported | Resource |
| Unbound forward | Supported | Already shipped as `opnsense_unbound_domain_override` using `/api/unbound/settings/*_forward`. |
| Dnsmasq settings | Supported | Resource only |
| Dnsmasq host/domain/range/option/tag/boot | Supported | UUID item resources and data sources using `dnsmasq/settings` item endpoints plus `dnsmasq/service/reconfigure`. |

### System, Access, and Trust

| Item | Status | Notes |
|---|---|---|
| System info | Supported | Data source |
| Users | Supported | Resource + data source |
| Groups | Supported | Resource + data source |
| Certificate authority | Supported | Resource only |
| Certificate | Supported | Resource only |
| Cron job | Supported | Resource + data source |
| System general settings | Upstream-blocked | Story 5.7 created on 2026-06-14: published core API docs list no durable system general settings endpoint; `core/system` is action/status-only; no `Core/Api/SettingsController.php`, `Core/System.xml`, or `Core/Settings.xml` was found; `core/initial_setup` is wizard-only and unsafe for day-2 Terraform management. |
| Tunables / sysctl | Coming with safety/live-validation gate | Story 28.3 confirmed persistent `core/tunables` item CRUD/search and `reconfigure`; implement only after live validation and safety documentation for kernel/network tunables. |
| High availability / HASync config | Needs research | Published core API docs confirm singleton `core/hasync` get/set/reconfigure endpoints, but `Hasync.xml` uses dynamic `JsonKeyValueStoreField` `syncitems`; verify API shape and safe Terraform representation before implementation. |
| High availability / HASync status/actions | Needs research | Story 28.2 classified `services`/`version` as data-source candidates after live response validation and service operations as action candidates only after product/framework decision; no durable resource semantics. |

### Services and Shaping

| Item | Status | Notes |
|---|---|---|
| HAProxy server/backend/frontend/ACL/health check | Supported | Resources + data sources |
| ACME account/certificate/challenge | Supported | Resources + data sources |
| Traffic shaper pipe/queue/rule | Supported | Resources + data sources |
| Monit service/test/alert | Supported | Resources + data sources |
| Syslog destination | Supported | Resource + data source |

## Cross-Cutting Gaps

| Item | Status | Notes |
|---|---|---|
| Data-source parity | Coming | 15 singleton or sensitive special-case resources still lack matching data sources. |
| Provider index | Supported with follow-up | Auth, version, permissions, and quickstart exist; support matrix should remain visible. |
| Migration/import guidance | Supported | Full-appliance dependency-order migration guide exists in `docs/migration-import.md`. |
| Registry release workflow | Supported with preflight | GoReleaser, signed checksums, Registry manifest, changelog, and tag workflow exist; confirm secrets and dry-run before first tag. |

## Upstream-Blocked Register

The public register and maintenance workflow live in [`docs/upstream-blocked.md`](../../docs/upstream-blocked.md).

| Resource/domain | Upstream item | Action |
|---|---|---|
| Interface assignment / IP config / PPPoE | OPNsense PR #8436 plus emerging `master` assignment controller | Track, test target-release availability, verify docs/ACL coverage, and contribute if needed. |
| Gateway group | `master` model-only `GatewayGroups` evidence; no target-release API/controller | Track generated API docs, `stable/*` branch availability, API controllers, ACL/menu entries, and model semantics; candidate fresh MVC API contribution if absent. |
| System general settings | OPNsense System Settings MVC roadmap; current `core/initial_setup` evidence is wizard-only, not a durable singleton API | Watch release notes, API docs, controllers/models, ACL/menu entries, and adopt only after stable target-release get/set semantics are available. |

## Release Matrix

- **Supported:** 102 resources and 88 data sources listed in `support-matrix.md` and generated docs.
- **Coming:** data-source parity for singleton and sensitive special cases.
- **Needs research:** Kea DHCPv4 option, Kea DDNS, HASync configuration, HASync status/actions, and emerging OPNsense `master` interface assignment API semantics before any future unblocked story.
- **Upstream-blocked:** interface assignment/IP config/PPPoE, gateway group, and system general settings.
