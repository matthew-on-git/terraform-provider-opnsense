---
title: Provider Support Matrix
date: 2026-11-03
author: BMad PM
status: current
inputs:
  - prd.md
  - core-config-gap-analysis.md
  - feature-complete-roadmap.md
  - provider-management-roadmap-2026-09-28.md
  - repository implementation inventory
---

# Provider Support Matrix

This matrix is the current source of truth for release positioning. This 2026-11-03 snapshot reconciles registered implementation, generated documentation, merged changes and published releases through v0.12.0; the Trust continuation remains a draft, with a non-skipped 26.7.8_2 live matrix now passed and broader capability/compatibility gates still open. Pending changes and appliance evidence are separate gates; historical digest-bound evidence is not re-attested by this documentation refresh.

## Summary

| Area | Count | Status |
|---|---:|---|
| Resources | 113 | Released inventory through v0.12.0, including Trust settings |
| Data sources | 100 | Source-tree inventory: 99 published in v0.12.0 plus the pending Trust certificate lookup |
| Resource docs | 113 | Matches registered resources |
| Data source docs | 100 | Generated |
| Ephemeral resources / docs | 1 / 1 | Terraform 1.10+ OpenVPN key generator passed final protocol/non-persistence proof; published in v0.7.0; no blanket OpenTofu ephemeral-lifecycle claim |
| Remaining field/surface gaps | 5 audit-derived P0/P1 story groups (Stories 31.0f-31.0j), with separate research/upstream blockers | Coming, needs research, or upstream-blocked |
| Remaining resource-matching data-source gaps | 15 | Coming |

Data-source parity is tracked in [data-source-parity-plan.md](data-source-parity-plan.md). Remaining resource-gap verification is tracked in [resource-gap-verification.md](resource-gap-verification.md).

## Released scope and evidence boundaries

| Capability | First published version | Verified scope and remaining limits |
|---|---|---|
| WireGuard typed parity/settings and OpenVPN ephemeral key generation | v0.7.0 | Merged in [PR #63](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/63) / [PR #60](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/60). WireGuard's recorded disposable 26.7.7 proof covers six audit rows, associations, drift repair, zero leftovers and exact restoration; daemon stayed disabled. Ephemeral protocol/non-persistence proof is recorded separately. Neither establishes arbitrary tunnel traffic or OpenTofu ephemeral behavior. |
| FRR 2.0 compatibility | v0.7.0; numeric package revisions in v0.8.1 | [PR #64](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/64) added installed-plugin/model checks, family/state migration and grouped-policy preservation; exact live proof covers os-frr 2.0_1 on OPNsense 26.7.7. [PR #69](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/69) accepts numeric 2.0 package revisions including 2.0_2, with fixture/protocol tests rather than a new appliance run. Future plugin releases and advanced parity remain separate. |
| Dnsmasq typed DHCP settings | v0.8.0 | [PR #68](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/68) corrected nested read/write serialization and added legacy-state migration. Fixture/provider-protocol lifecycle coverage is not live-appliance write proof. Configuration migration is documented. |
| Gateway groups | v0.9.0 | [PR #72](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/72) published the typed owner and UUID lookup. The recorded 16-step Core lifecycle, five recovery/private-replay cases, independent saved-field observations and exact restoration passed on 26.7.7; see `31-0m-gateway-group-live-evidence.md`. Populated first-tier clearing remains rejected before save. Older cores, OpenTofu lifecycle and actual multi-WAN failover are not proven. |
| Kea generic DHCPv4 options | v0.10.0 | [PR #73](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/73) published the eight-field owner and UUID lookup. The disabled-appliance 14-step Core lifecycle and 17 independent observations cover preservation, tracked in-use HTTP500 rejection, deletion and exact restoration; see `31-0n-kea-option-live-evidence.md`. Ordered parent linkage is upstream-blocked on released core 26.7.8; stored force is not packet-proven. |
| Kea DDNS general settings | v0.10.0 | [PR #73](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/73) published the four-field singleton and lookup with fixed `ddns` import, omission-preserving partial writes and durable recovery. Final 14-step disabled-appliance Core proof, state-only Delete and exact restoration are recorded in `31-0o-kea-ddns-live-evidence.md`. Normal planning refresh can recover/activate before Delete. Manual files, active DNS delivery, older cores and OpenTofu lifecycle are not proven. |
| IPv6 identity association / legacy track-interface assignments | v0.11.0 | [PR #77](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/77) adds adoption/updates of existing `idassoc6`/`track6` assignments and six delegation fields. Independent source/race/fixture review passed; live import/no-change-plan evidence remains contributor-reported, not independent activation/packet proof. See the detailed scope below. |
| NAT port-forward category persistence | v0.10.1 | [PR #75](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/75) resolves UUIDs to names for persisted writes while retaining UUID state; adds category-read permission. Regression, race and protocol/fixture coverage pass; no independent live-appliance validation was performed for this patch. |

Published tags v0.9.0, v0.10.0, v0.10.1 and v0.11.0 resolve to their respective merge commits; GitHub release assets and Terraform Registry versions were checked. Published v0.12.0 adds the six-field Trust settings owner and lookup (113 resources/99 data sources). Its tag matches PR79 merge `4bd107fc769586d5904919c8e69562897a87e84d`, all16expected release assets are present, and Terraform Registry lists0.12.0; OpenTofu Registry did not list it at the checked checkpoint. The current continuation source tree adds Trust certificate lookup (113resources/100data sources,15resource-matching gaps); it is not yet published, and the complete Trust action/settings/adoption/destroy/restoration matrix passed on authentic 26.7.8_2; see the [digest-bound sanitized proof](../implementation-artifacts/31-0f-live-proof.json). Keyless provider support and broader firmware/OpenTofu proof remain open.

## IPv6 delegation: released scope and evidence

[PR #77](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/77), reviewed at `4f1ace6a1d788004140a4802f14a7c47b62a4b82`, merged as `2c6ea1eaa2079a81d6794eef02635e8caee761a6` and implements existing-assignment adoption/updates for IPv6 `idassoc6` and legacy `track6`, with six optional/computed delegation fields. All eight hosted checks pass. Source comparison, race-enabled interface tests and a counterfactual adoption test passed; the old validator rejects the same delegation fixture.

The contributor reports OPNsense 26.7.8_2 / OpenTofu 1.13.2 import and a no-change refresh plan. That report was not independently reproduced and does not establish activation, packet behavior or full write/restore acceptance. Assignment creation, automatic mode migration, upstream DHCPv6-client ownership and PPPoE remain outside this PR. v0.11.0 is published: main CI, automatic tagging and release workflow passed; the tag resolves to the PR merge commit, all 16 expected release assets are present, and Terraform Registry lists 0.11.0. This verifies publication, not new appliance activation or packet behavior.

## Supported Resources

| Domain | Supported resources |
|---|---|
| ACME | account, certificate, challenge |
| Auth | group, user |
| Cron | job |
| DHCPv4 | reservation, subnet |
| Dnsmasq | boot, domain, host, option, range, settings, tag |
| Dynamic DNS | ddclient account, ddclient settings |
| Firewall | alias, category, filter rule, NAT one-to-one, NAT outbound, NAT port forward |
| HAProxy | ACL, action, backend, frontend, health check, map file, server |
| Interfaces | assignment/IP settings, bridge, GIF, GRE, LAGG, loopback, neighbor, VXLAN |
| IPsec | child, connection, key pair, local, manual SPD, pool, PSK, remote, VTI |
| Kea | control agent, DDNS settings, DHCPv4 option/reservation/settings/subnet, DHCPv6 peer/PD pool/reservation/settings/subnet, HA peer |
| Monit | alert, service, test |
| OpenVPN | client overwrite, instance, static key |
| Quagga / FRR | BGP AS path, BGP community list, BGP global, BGP neighbor, BGP peer group, BGP redistribution, general, OSPF area/general/interface/neighbor/network/prefix list/redistribution/route map, OSPFv3 general/interface/network/prefix list/redistribution/route map, prefix list, RIP, route map, static general, static route |
| Syslog | destination |
| System | gateway, gateway group, route, tunable, VIP, VLAN |
| Traffic shaper | pipe, queue, rule |
| Trust | CA, certificate |
| Unbound | ACL, DNSBL, domain override, general, host alias, host override |
| WireGuard | peer, server, settings |

## Supported Data Sources

| Domain | Supported data sources |
|---|---|
| Auth | group, user |
| Cron | job |
| ACME | account, certificate, challenge |
| DHCPv4 | reservation, subnet |
| Dnsmasq | boot, domain, host, option, range, tag |
| Dynamic DNS | ddclient account, ddclient settings |
| Firewall | alias, category, filter rule, NAT one-to-one, NAT outbound, NAT port forward |
| HAProxy | ACL, action, backend, frontend, health check, map file, server |
| Interfaces | bridge, GIF, GRE, LAGG, loopback, neighbor, VXLAN |
| IPsec | child, connection, local, manual SPD, pool, remote, VTI |
| Kea | DDNS settings, DHCPv4 option/reservation/subnet, DHCPv6 peer/PD pool/reservation/subnet, HA peer |
| Monit | alert, service, test |
| Quagga / FRR | BGP AS path, BGP community list, BGP neighbor, BGP peer group, BGP redistribution, OSPF area/interface/neighbor/network/prefix list/redistribution/route map, OSPFv3 interface/network/prefix list/redistribution/route map, prefix list, route map, static route |
| Syslog | destination |
| OpenVPN | client overwrite, instance, static key |
| System | gateway, gateway group, route, system info, tunable, VIP, VLAN |
| Traffic shaper | pipe, queue, rule |
| Trust | CA |
| Unbound | ACL, domain override, host alias, host override, settings |
| WireGuard | peer, server, settings |

## Coming: Buildable Provider Work

| Domain | Work |
|---|---|
| Data-source parity | Add read-only data sources for the 15 remaining singleton/sensitive special-case resources that do not yet have same-name data-source counterparts. |
| Documentation | Fill missing resource templates for generated docs, expand composition examples, and keep the provider index support matrix current. |
| Release hardening | Keep release workflow, Registry manifest, changelog, and provider docs verified for subsequent patch/minor releases. |
| Audited field/surface parity | Complete Stories 31.0f–31.0j for trust, IPsec, Quagga/BGP, live interface state, and provider retry controls before positioning/migration/adoption work. |
| Competitive positioning | Refresh README/Registry positioning and adoption examples only after the P0/P1 audit follow-ups complete. |

## Needs Research

| Domain | Research needed |
|---|---|
| Kea | Stable 26.7.7 source and disabled-service direct-API option/DDNS probes are recorded in `kea-option-ddns-api-validation.md`: targeted updates/preservation/false clear, reconfigure, required-field negatives and exact settings restoration. Stories 31.0n/31.0o are published in v0.10.0 with separate final disabled-appliance Core lifecycle evidence. Ordered parent linkage is upstream-blocked: released core 26.7.8 has no authoritative ordered parent read (`kea-ordered-option-linkage-api-research.md`). Older targets and operator integrations remain separate; do not duplicate completed subnet/PXE/DHCPv6 parity. |
| System / HA | HASync configuration needs request/response shape research because current `Hasync.xml` uses dynamic `JsonKeyValueStoreField` `syncitems`; HASync status `services`/`version` are data-source candidates after live validation, while service operations are action candidates only after product/framework decision. |
| Interfaces | Base assignment/IP settings are supported by `opnsense_interface_assignment` on OPNsense 26.7.6+ after disposable 26.7.7 lifecycle/drift/restoration proof. PPPoE semantics remain separate research. |

Source NAT is supported as `opnsense_firewall_nat_outbound`. Story 31.0d Repair 7 implements and digest-bound proves Unbound query and DNS-over-TLS forwarding plus complete read-only settings; exactly four audited parity rows are reconciled and four-layer convergence is clean.

## Upstream-Blocked

Confirmed blockers and the maintenance workflow are documented publicly in [`docs/upstream-blocked.md`](../../docs/upstream-blocked.md).

| Resource/domain | Reason | Action |
|---|---|---|
| System general settings | No stable target-release durable settings API; `core/system` is action/status-only and `core/initial_setup` is wizard-only with broad side effects. | Watch OPNsense System Settings MVC roadmap, generated API docs, controllers/models, ACL/menu entries, and stable get/set semantics. |

## Release Readiness Snapshot

| Item | Status | Evidence / gap |
|---|---|---|
| GoReleaser config | Ready | `.goreleaser.yml` cross-compiles, archives ZIPs, emits checksums, signs checksum artifact. |
| Release workflow | Ready with secrets | `.github/workflows/release.yml` runs on `v*` tags; requires `GPG_PRIVATE_KEY` and `PASSPHRASE` secrets. |
| Terraform Registry manifest | Ready | `terraform-registry-manifest.json` declares protocol `6.0`. |
| Changelog / current release | v0.11.0 published | [PR #77](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/77) merged as `2c6ea1e`; its automatic v0.11.0 tag, release assets and Terraform Registry version are verified. Older entries may retain `(Unreleased)` labels despite existing tags; tags/assets/Registry are checked separately. |
| Provider index | Reconciled | Template and generated index mirror the 112-resource / 98-data-source inventory and released Kea scope. |
| OpenTofu distribution | Registered; installation not independently tested | Provider [PR #6248](https://github.com/opentofu/registry/pull/6248) and signing-key [PR #6258](https://github.com/opentofu/registry/pull/6258) are merged in opentofu/registry. The OpenTofu versions API lists v0.10.1, and its linux/amd64 download metadata advertises signing key E019411104049B47. Registry presence is not resource-specific lifecycle, activation or packet proof. |
| Migration/import guidance | Partial | Per-resource import examples exist; broader dependency-order migration guide remains useful follow-up. |

Story 31.0e proves exactly six WireGuard audit rows on disposable OPNsense 26.7.7: typed peer/server field parity and import-first daemon settings. WireGuard is core on 24.1+; server `peers` owns associations, peer `servers` is its authoritative API backreference. PSK/private keys stay resource-only Sensitive values. The daemon remained disabled, Terraform deletes left zero tagged items, and settings restored exactly. Implementation merged in [PR #63](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/63) and published in v0.7.0; see [digest-bound evidence](../implementation-artifacts/31-0e-wireguard-live-evidence.md). Stories 31.0k/31.0l remain separate stable-API validation work; the FRR2 compatibility gate is unchanged.
