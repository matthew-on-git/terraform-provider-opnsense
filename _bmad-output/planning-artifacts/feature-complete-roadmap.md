---
title: Feature-Complete Roadmap — Expanded Epic Plan
date: 2026-05-29
author: BMad Master
status: draft
supersedes: extends epics.md (Epics 1-12) with Epics 13-25
inputs:
  - core-config-gap-analysis.md
  - epics.md
  - prd.md
goal: A feature-complete OPNsense Terraform provider covering all core config (~100+ resources).
strategy: Ship incrementally to the registry; drive the 3 blocked domains upstream in parallel.
---

# Feature-Complete Roadmap

## Where we are

- **31 resources + 1 data source built and tested** (Epics 1-5, 7-11 substantially done).
- **Gap analysis (2026-05-29)** confirmed ~80+ additional resources are buildable against existing OPNsense APIs.
- **Only 3 domains are genuinely blocked** (need upstream OPNsense work): interface base assignment, gateway group, system general/tunables.

## PRD expansion required

The original PRD (68 FRs) does **not** equal "all core config." These domains must be added as new functional requirements before/with implementation:

- **FR69-71** OpenVPN: instances (server/client), client overwrites, static keys
- **FR72-73** PKI/Trust: certificate authorities, certificates
- **FR74-75** Auth: users, groups
- **FR76-78** Traffic shaping: pipes, queues, rules
- **FR79-84** IPsec extensions: locals, remotes, pools, VTI, manual SPD, key pairs
- **FR85-92** OSPF + OSPFv3 (general, areas, interfaces, neighbors, networks, prefix-lists, route-maps, redistribution)
- **FR93-94** RIP, FRR static routes
- **FR95-96** BGP global config + FRR general settings (closes original FR30/FR31 via singletons)
- **FR97-101** Kea extensions: DHCPv6, HA peers, DHCPv4 options, ctrl_agent, ddns
- **FR102-105** Unbound extensions: general, forward, host alias, blocklist
- **FR106-110** Dnsmasq (alternative backend): host, domain, range, option, tag
- **FR111-113** System services: cron jobs, monit (service/test/alert), syslog destinations
- **FR114-120** Interface types: bridge, LAGG, GRE, GIF, VXLAN, loopback, neighbor
- **FR121-122** Firewall: source NAT, one-to-one NAT
- (Re-confirm original FR57 DHCP options — now buildable via Kea)

---

## Foundation epic (do first — unblocks ~10 resources)

### Epic 13: API Client Singleton Support
Add singleton get/set to `pkg/opnsense` (get/set with no UUID; tolerate missing `/{id}`), with unit tests. Required by every singleton resource (FRR/OSPF/RIP/static general, BGP global, Unbound general, Kea ctrl_agent/ddns, DDNS settings). Small, high-leverage.

---

## Wave A — high-impact, confirmed APIs (ship first)

### Epic 14: OpenVPN
instance (server/client), client overwrite, static key. (3) — biggest missing core domain.

### Epic 15: PKI & Trust
certificate authority, certificate. (2) — unblocks ACME/HAProxy SSL composition stories.

### Epic 16: Users & Access
user, group. (2) — (privileges = optional get/set singleton).

### Epic 17: Interface Types
bridge, LAGG, GRE, GIF, VXLAN, loopback, neighbor. (7) — closes most of the interfaces gap that doesn't need upstream.

### Epic 18: Firewall Completion
source NAT, one-to-one NAT. (2)

### Epic 19: Dynamic Routing Completion (FRR/quagga)
FRR general (s), BGP global (s), BGP aspath/communitylist/peergroup/redistribution, OSPF (general + 7 sub), OSPFv3 (general + 5 sub), RIP (s), FRR static (general + route). (~25) — depends on Epic 13.

---

## Wave B — confirmed, second priority

### Epic 20: VPN Completion (IPsec)
local, remote, pool, VTI, manual SPD, key pair. (6)

### Epic 21: DHCP Completion (Kea)
DHCPv6 subnet/reservation, HA peer (v4+v6), DHCPv4 option, ctrl_agent (s), ddns (s). (~7) — singletons depend on Epic 13.

### Epic 22: DNS Completion
Unbound general (s), forward, host alias, blocklist. Optionally dnsmasq (host/domain/range/option/tag) as alternative backend. (4-9)

### Epic 23: Traffic Shaping
pipe, queue, rule. (3)

### Epic 24: System Services
cron job, monit (service/test/alert), syslog destination. (5)

---

## Wave C — release runway (extends Epic 12; runs in parallel)

### Epic 25: Data Sources, Docs & Registry Release
- Data source for every resource (~30+) + `system_info` data source.
- `templates/index.md.tmpl` + `tfplugindocs` registry docs + composition examples.
- CI release pipeline (GoReleaser + GPG), `make check` gate, structure validation.
- **v0.x tag with an honest support matrix** (Supported / Coming / Upstream-blocked).
- Ship as soon as Wave A is stable — do NOT wait for Wave B.

---

## Two-front: upstream contribution track (the 3 blocked domains)

| Domain | Upstream item | Action | Unblocks |
|---|---|---|---|
| Interface base assignment | PR #8436 (Terraform-motivated) | track, test, contribute; review companion TF provider for overlap | `opnsense_interface` (assignment/IP/PPPoE) |
| Gateway group | none tracked | author fresh MVC PR (model on VIP #6105 / IPsec #6187) | `opnsense_system_gateway_group` |
| System general + tunables | roadmap: System Settings → MVC | watch release notes (CE 26.x), adopt when shipped | `opnsense_system_general`, tunables |

Each merged endpoint → a new provider release adding that resource. Document blocked items publicly so users see an active roadmap, not gaps.

---

## Epic 6 (codegen) — still in scope

User decision: **keep & build.** With ~80 more resources following the identical four-file pattern, the codegen pipeline (YAML schema → `text/template`) now has a much stronger ROI. Recommended timing: build it **after Epic 13 + one full Wave-A epic** establishes the stable pattern, then use it to accelerate Waves A/B. Re-sequence Epic 6 to land between Wave A and Wave B.

---

## Suggested execution order

1. Epic 13 (singleton client) — foundation
2. Epic 14 OpenVPN + Epic 15 Trust + Epic 16 Users — highest adoption value
3. Epic 25 partial: data sources + docs + **first registry release (v0.1.0)** with what's built
4. Epic 6 codegen (pattern now stable) → accelerate the rest
5. Epics 17-24 via codegen, releasing incrementally (v0.2, v0.3, …)
6. Upstream track runs continuously; fold in blocked resources as APIs land → v1.0 when core-complete

## Definition of done: "feature complete"

All core-config domains either (a) shipped as resources, or (b) explicitly blocked upstream with a tracked item and public roadmap note. v1.0 = all buildable resources shipped + the 3 upstream domains resolved or formally documented as OPNsense limitations.
