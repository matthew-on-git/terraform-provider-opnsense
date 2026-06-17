---
title: Feature-Complete Roadmap — Expanded Epic Plan
date: 2026-05-29
author: BMad Master
status: updated-2026-06-02
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

- **102 resources + 88 data sources are implemented and documented** in the repository.
- The implementation is ahead of the original PRD and the first roadmap snapshot. Wave A and most Wave B resource work are already built.
- The largest remaining provider-owned gap is **data-source parity**: 15 singleton or sensitive special-case resources do not yet have matching data sources.
- No verified provider-owned resource gap remains after system tunables/sysctl shipped. Interface LAGG, Source NAT, Unbound forward, Dnsmasq item resources, OSPF area, and system tunables are already supported; HASync configuration needs `syncitems` model-shape research; Kea DHCPv4 option/DDNS need live endpoint recheck before implementation.
- **Three domains are upstream-blocked**: interface base assignment/IP config/PPPoE, gateway group, and system general settings. Story 5.1 revalidated interface status on 2026-06-12: OPNsense `master` now has an emerging assignment controller, but it is absent from target `stable/26.1`, absent from published API docs, missing ACL coverage, and does not cover IP config or PPPoE. Story 5.6 revalidated gateway-group status on 2026-06-14: OPNsense `master` has model-only `GatewayGroups` evidence, but no published endpoint/API controller was found and checked `stable/26.1` model paths returned 404. Story 5.7 created the system general settings revalidation gate on 2026-06-14: no durable target-release settings API was found; `core/system` is action/status-only and `core/initial_setup` is wizard-only.

## PRD expansion applied / still required

The original PRD (68 FRs) did **not** equal "all core config." These domains are now treated as the expanded feature-complete scope:

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
- **FR106-110** Dnsmasq (alternative backend): host, domain, range, option, tag, boot
- **FR111-113** System services: cron jobs, monit (service/test/alert), syslog destinations
- **FR114-120** Interface types: bridge, LAGG, GRE, GIF, VXLAN, loopback, neighbor
- **FR121-122** Firewall: source NAT, one-to-one NAT
- (Re-confirm original FR57 DHCP options — now buildable via Kea)

---

## Foundation epic

### Epic 13: API Client Singleton Support
Status: done. Singleton get/set support exists and has enabled singleton resources including FRR/BGP, OSPF, RIP/static, Unbound general, and Kea control agent/settings.

---

## Wave A — high-impact, confirmed APIs (ship first)

### Epic 14: OpenVPN
Status: done. Instance, client overwrite, and static key resources are implemented.

### Epic 15: PKI & Trust
Status: done. CA and certificate resources are implemented.

### Epic 16: Users & Access
Status: done. User and group resources plus data sources are implemented.

### Epic 17: Interface Types
Status: done for durable target-release APIs. Bridge, GRE, GIF, VXLAN, loopback, neighbor, and LAGG are implemented; base assignment/IP config/PPPoE remains upstream-blocked outside this epic.

### Epic 18: Firewall Completion
Status: done. One-to-one NAT is implemented. Source NAT is already shipped as `opnsense_firewall_nat_outbound`.

### Epic 19: Dynamic Routing Completion (FRR/quagga)
Status: substantially done. FRR general, BGP global/sub-resources, OSPF area/sub-resources, OSPFv3, RIP, and FRR static resources are implemented.

---

## Wave B — confirmed, second priority

### Epic 20: VPN Completion (IPsec)
Status: done. Local, remote, pool, VTI, manual SPD, and key pair resources are implemented.

### Epic 21: DHCP Completion (Kea)
Status: partial. DHCPv6 settings/subnet/reservation, HA peer, and control agent are implemented. DHCPv4 option and Kea DDNS need live endpoint recheck before implementation.

### Epic 22: DNS Completion
Status: done. Unbound general, host alias, DNSBL/blocklist, Unbound forward, Dnsmasq settings, and Dnsmasq host/domain/range/option/tag/boot item resources are implemented.

### Epic 23: Traffic Shaping
Status: done. Pipe, queue, and rule resources plus data sources are implemented.

### Epic 24: System Services
Status: done. Cron job, Monit service/test/alert, and syslog destination resources are implemented.

---

## Wave C — release runway (extends Epic 12; runs in parallel)

### Epic 25: Data Sources, Docs & Registry Release
- 88 data sources are implemented, including standalone `system_info`; 15 singleton or sensitive special-case resource-matching data sources remain.
- Provider index, tfplugindocs output, examples, GoReleaser, signed checksum workflow, Registry manifest, and changelog exist.
- Support matrix now lives in `support-matrix.md` and should be mirrored in Registry-facing docs.
- First release can proceed once release workflow dry-run/secrets are confirmed and the support matrix is accepted.

---

## Two-front: upstream contribution track

| Domain | Upstream item | Action | Unblocks |
|---|---|---|---|
| Interface base assignment / IP config / PPPoE | PR #8436 plus emerging `master` assignment controller | track target release, API docs, ACL coverage, and IP/PPPoE scope; test before implementation story | `opnsense_interface` or `opnsense_system_interface` for assignment/IP/PPPoE only after stable target-release API exists |
| Gateway group | `master` model-only `GatewayGroups` evidence; no target-release API/controller | track target release, generated API docs, API controllers, ACL/menu coverage, and tier/member semantics; author fresh MVC API if absent | `opnsense_system_gateway_group` |
| System general settings | roadmap: System Settings -> MVC; current `core/initial_setup` evidence is wizard-only, not a day-2 singleton API | watch release notes (CE 26.x), generated API docs, controllers/models, ACL/menu entries, and stable target-release get/set semantics | `opnsense_system_general` |

Each merged endpoint → a new provider release adding that resource. Document blocked items publicly so users see an active roadmap, not gaps.

---

## Epic 6 (codegen) — still in scope

User decision: **keep & build.** With ~80 more resources following the identical four-file pattern, the codegen pipeline (YAML schema → `text/template`) now has a much stronger ROI. Recommended timing: build it **after Epic 13 + one full Wave-A epic** establishes the stable pattern, then use it to accelerate Waves A/B. Re-sequence Epic 6 to land between Wave A and Wave B.

---

## Suggested execution order

1. v0.1.0 is published. Treat release-readiness work as historical unless a regression is found.
2. Execute the post-release epic plan in `post-release-epics.md`.
3. Start with Story 25B.1 to create the exact data-source parity batch plan.
4. Harden public Registry docs through Epic 26.
5. Maintain data-source parity follow-up for the remaining singleton or sensitive special cases through the parity plan.
6. Verify and implement remaining buildable resources through Epic 27.
7. Maintain upstream-blocked transparency through Epic 28.

## Definition of done: "feature complete"

All core-config domains are either (a) shipped as resources/data sources, (b) explicitly listed as Coming after endpoint and durable-semantics verification, (c) explicitly listed as Needs research with the missing evidence called out, or (d) explicitly blocked upstream with a tracked item and public roadmap note. v1.0 = all buildable resources and data-source parity shipped, plus Needs research and upstream-blocked domains resolved or formally documented as OPNsense limitations.
