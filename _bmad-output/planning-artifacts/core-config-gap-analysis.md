---
title: OPNsense Core Config — Terraform Provider Gap Analysis
date: 2026-05-29
author: BMad Master (audit)
status: draft
purpose: Define "feature complete" for the provider and classify every core-config domain as Built / Buildable-now / Upstream-blocked / Verify.
---

# Core Config Gap Analysis

## Goal

A **feature-complete** Terraform provider covering **all OPNsense core configuration**. Partial coverage is treated as failure (adoption bar). This document defines the finish line and classifies the work.

## Confidence legend

- **[C]** Confirmed against docs.opnsense.org API reference this session.
- **[V]** From OPNsense domain knowledge — **needs live/API-doc verification** before building.

## Status legend

- ✅ **Built** — resource exists in provider today (31 resources).
- 🟢 **Buildable now** — OPNsense MVC API exists; provider work only.
- 🔴 **Upstream-blocked** — no usable API; requires OPNsense core PR (see register).
- 🟡 **Verify** — likely buildable but API not yet confirmed.

---

## Domain-by-domain

### Interfaces
| Item | Status | Notes |
|---|---|---|
| VLAN | ✅ Built | |
| Virtual IP (VIP) | ✅ Built | |
| Bridge | 🟢 Buildable [C] | full CRUD `/api/interfaces/bridge_settings` |
| LAGG | 🟢 Buildable [C] | full CRUD |
| GRE | 🟢 Buildable [C] | full CRUD |
| GIF | 🟢 Buildable [C] | full CRUD |
| VXLAN | 🟢 Buildable [C] | full CRUD |
| Loopback | 🟢 Buildable [C] | full CRUD |
| Neighbor (static ARP/NDP) | 🟢 Buildable [C] | full CRUD |
| **Base assignment / IP config / PPPoE** | 🔴 Blocked [C] | no API; PR #8436 in progress, "needs a lot of time" |

### Firewall & NAT
| Item | Status | Notes |
|---|---|---|
| Alias (+ data source) | ✅ Built | |
| Category | ✅ Built | |
| Filter rule | ✅ Built | |
| NAT port forward | ✅ Built | |
| NAT outbound | ✅ Built | |
| Source NAT | 🟢 Buildable [C] | `/api/firewall/source_nat` full CRUD |
| One-to-One NAT | 🟢 Buildable [C] | `/api/firewall/one_to_one` full CRUD |
| Schedules | 🟡 Verify [V] | likely legacy, confirm |

### Routing & Gateways
| Item | Status | Notes |
|---|---|---|
| Static route | ✅ Built | |
| Gateway | ✅ Built | |
| **Gateway group** | 🔴 Blocked [C] | no gateway-group endpoint on routes API |

### Dynamic routing (FRR / quagga plugin)
| Item | Status | Notes |
|---|---|---|
| BGP neighbor | ✅ Built | |
| BGP prefix list | ✅ Built | |
| BGP route map | ✅ Built | |
| FRR general | 🟢 Buildable [C] | `/api/quagga/general` get/set — **singleton** |
| BGP global config | 🟢 Buildable [C] | `/api/quagga/bgp` get/set — **singleton** |
| BGP aspath / communitylist / peergroup / redistribution | 🟢 Buildable [C] | sub-resources on bgp controller |
| OSPF (general, neighbor, network, prefixlist, routemap, interface) | 🟡 Verify [V] | quagga ospf controller expected |
| OSPFv3 / RIP / FRR static | 🟡 Verify [V] | confirm controllers |

### VPN
| Item | Status | Notes |
|---|---|---|
| WireGuard server | ✅ Built | |
| WireGuard peer | ✅ Built | |
| IPsec connection (phase 1) | ✅ Built | |
| IPsec child (phase 2) | ✅ Built | |
| IPsec PSK | ✅ Built | |
| **OpenVPN instance (server/client)** | 🟢 Buildable [C] | `/api/openvpn/instances` full CRUD — MAJOR gap |
| OpenVPN client overwrites | 🟢 Buildable [C] | full CRUD |
| OpenVPN static keys | 🟢 Buildable [C] | on instances controller |
| IPsec locals / remotes / pools / VTI | 🟡 Verify [V] | swanctl MVC sub-resources |

### DHCP
| Item | Status | Notes |
|---|---|---|
| Kea DHCPv4 subnet | ✅ Built | |
| Kea DHCPv4 reservation | ✅ Built | |
| Kea DHCPv4 general / HA peer | 🟡 Verify [V] | |
| Kea DHCPv6 (subnet/reservation) | 🟡 Verify [V] | |
| DHCP options | 🟢 Buildable [C] | **CORRECTED:** Kea dhcpv4 DOES have `add_option`/`del_option`/`get_option`/`search_option`. PXE 66/67/150 = field-level check |

### DNS
| Item | Status | Notes |
|---|---|---|
| Unbound host override | ✅ Built | |
| Unbound domain override | ✅ Built | |
| Unbound ACL | ✅ Built | |
| Unbound general / forwarding / DoT | 🟡 Verify [V] | |
| Dnsmasq | 🟡 Verify [V] | has API |
| Dynamic DNS account | ✅ Built | |
| DDNS "provider" | 🔴 N/A [C] | redundant — `service` field on account |

### System
| Item | Status | Notes |
|---|---|---|
| **General settings** (hostname/domain/DNS/NTP) | 🔴 Blocked [C] | legacy; on roadmap (System Settings → MVC) |
| Tunables / sysctl | 🔴 Blocked [V] | on roadmap, no API yet |
| Users / groups / privileges | 🟡 Verify [V] | auth module partial API |
| Certificates / Trust (CA/cert/CSR) | 🟡 Verify [V] | Trust → MVC on roadmap |
| High Availability (HASync) | 🟡 Verify [V] | |
| Cron | 🟡 Verify [V] | has API |

### Services & shaping
| Item | Status | Notes |
|---|---|---|
| Traffic Shaper (pipes/queues/rules) | 🟡 Verify [V] | MVC API expected |
| Monit | 🟡 Verify [V] | has API |
| Syslog | 🟡 Verify [V] | has API |

### Plugins (beyond core)
HAProxy (5) ✅, ACME (3) ✅, ddclient (1) ✅ already built. Other plugins are out of "core config" scope for v1.

### Cross-cutting
| Item | Status | Notes |
|---|---|---|
| Data sources | 🟢 Buildable | only 1 of 31 resources has one — Epic 12.1 |
| `system_info` data source | 🟢 Buildable [C] | `/api/core/firmware` |
| Registry docs / index | 🟢 Buildable | Epic 12.2/12.3 |
| CI release + v0.x tag | 🟢 Buildable | Epic 12.4/12.5 |

---

## Scorecard (after Wave B verification, 2026-05-29)

- **Built:** 31 resources + 1 data source
- **Buildable now (CONFIRMED via API docs):** ~80+ new resources across: OpenVPN (3), firewall source/one-to-one NAT (2), 7 interface types, certs/trust CA+cert (2), users+groups (2), traffic shaper pipe/queue/rule (3), Kea DHCPv6 + HA peers + DHCPv4 options + ctrl_agent/ddns singletons (~7), IPsec locals/remotes/pools/VTI/manual_spd/key_pairs (6), OSPF (8), OSPFv3 (6), RIP (1), FRR static (2), quagga BGP global/general + aspath/communitylist/peergroup/redistribution (6), dnsmasq host/domain/range/option/tag (5), cron job (1), monit service/test/alert (3), syslog destination (1), unbound general/forward/host_alias/blocklist (4), + the entire data-source layer (~30) and release tooling
- **Upstream-blocked (CONFIRMED — only 3):** interface base assignment (PR #8436), gateway group (no API), system general + tunables (roadmap)
- **Net result:** the provider can reach ~100+ resources covering **nearly all** OPNsense core config. Only 3 domains are genuinely blocked.

### Wave B verification results (all CONFIRMED buildable [C])
| Domain | Resources unlocked |
|---|---|
| Trust | CA, Cert (CRL has no add; CSR not exposed) |
| Auth | User, Group (Priv = get/set singleton, optional) |
| Traffic Shaper | pipe, queue, rule |
| Kea | DHCPv6 subnet/reservation, HA peer (v4+v6), **DHCPv4 option**, ctrl_agent (s), ddns (s) |
| IPsec | local, remote, pool, VTI, manual_spd, key_pair |
| OSPF | general (s) + area, interface, neighbor, network, prefixlist, routemap, redistribution |
| OSPFv3 | general (s) + interface, network, prefixlist, redistribution, routemap |
| RIP | general (s) only |
| FRR static | general (s) + route |
| Dnsmasq | host, domain, range, option, tag (alt DNS/DHCP backend) |
| Cron | job |
| Monit | service, test, alert |
| Syslog | destination |
| Unbound (extend) | general (s), forward, host_alias, blocklist (DoT not exposed) |

**Singletons** (get/set, no UUID — all need the client singleton extension first): FRR general, BGP global, OSPF general, OSPFv3 general, RIP, FRR static, Unbound general, Kea ctrl_agent, Kea ddns, DDNS settings.

## Recommended build waves (the buildable-now backlog)

1. **Wave A — high-impact resources w/ confirmed APIs:** OpenVPN instances + client overwrites + static keys; firewall source NAT + one-to-one NAT; the 7 interface types (bridge/LAGG/GRE/GIF/VXLAN/loopback/neighbor); FRR general + BGP global singletons (needs client singleton support).
2. **Wave B — verify-then-build:** OSPF/RIP, certs/trust, users/groups, traffic shaper, Kea v6, dnsmasq, IPsec sub-resources, cron, monit, syslog.
3. **Wave C — release runway (Epic 12):** all data sources, `system_info`, registry docs, composition examples, CI release, v0.x tag with the support matrix.

## Upstream-blocked register (the two-front workstream)

| Resource | Upstream item | Action |
|---|---|---|
| Interface assignment | PR #8436 (for Terraform) | track / test / contribute |
| System general + tunables | Roadmap: System Settings → MVC | watch release notes |
| Gateway group | none tracked | candidate for fresh MVC PR (model on VIP #6105) |

(DHCP options removed — confirmed buildable via Kea `add_option`.)

## Release matrix (for the registry README — honest labeling)

- **Supported:** everything Built + Wave A/B as shipped.
- **Coming:** verified-but-not-yet-built (Wave B remainder).
- **Upstream-blocked:** interface assignment, system general, gateway group, tunables — with links to the OPNsense issues/PRs.
