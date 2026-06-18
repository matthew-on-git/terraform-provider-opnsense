---
title: Post-v0.1.0 Epic Plan
date: 2026-06-02
author: BMad PM
status: current
inputs:
  - prd.md
  - feature-complete-roadmap.md
  - core-config-gap-analysis.md
  - support-matrix.md
---

# Post-v0.1.0 Epic Plan

The first provider version is published. Post-release work should focus on data-source parity, public documentation quality, remaining verified resource gaps, and upstream-blocked transparency.

## Epic 25B: Data Source Parity

Goal: close straightforward data-source parity gaps and leave only singleton or sensitive special-case follow-up work. Epic 25B completed the move from 34 to 76 supported data sources, leaving 15 resource-matching data-source gaps.

| Story | Status | File |
|---|---|---|
| 25B.1 Data Source Parity Inventory and Implementation Batches | done | `_bmad-output/implementation-artifacts/25b-1-data-source-parity-inventory-and-batches.md` |
| 25B.2 High-Reference Data Sources | done | `_bmad-output/implementation-artifacts/25b-2-high-reference-data-sources.md` |
| 25B.3 Routing, DNS, DHCP, ACME, and Trust Data Sources | done | `_bmad-output/implementation-artifacts/25b-3-routing-dns-dhcp-acme-trust-data-sources.md` |

## Epic 26: Registry Docs Hardening

Goal: improve the published Registry experience now that users can install the provider.

| Story | Status | File |
|---|---|---|
| 26.1 Published Registry Documentation Audit | done | `_bmad-output/implementation-artifacts/26-1-registry-docs-audit.md` |
| 26.2 Public Support Matrix in Registry Docs | done | `_bmad-output/implementation-artifacts/26-2-public-support-matrix-registry-docs.md` |
| 26.3 Full-Appliance Migration and Import Guide | done | `_bmad-output/implementation-artifacts/26-3-full-appliance-migration-guide.md` |

## Epic 27: Remaining Buildable Resource Gaps

Goal: verify and implement remaining provider-owned resource gaps without speculating on unconfirmed endpoints.

| Story | Status | File |
|---|---|---|
| 27.1 Verify Remaining Buildable Resource Gaps | done | `_bmad-output/implementation-artifacts/27-1-verify-buildable-resource-gaps.md` |
| 27.2 Implement Highest-Value Verified Resource Gap | done | `_bmad-output/implementation-artifacts/27-2-implement-highest-value-verified-gap.md` |
| 27.3 Implement OSPF Area Resource | done | `_bmad-output/implementation-artifacts/27-3-implement-ospf-area-resource.md` |
| 27.4 HASync Configuration Model Review (no resource shipped) | done | `_bmad-output/implementation-artifacts/27-4-implement-hasync-configuration-resource.md` |

## Epic 28: Upstream-Blocked Tracking

Goal: keep OPNsense-side blockers visible, honest, and periodically reviewed.

| Story | Status | File |
|---|---|---|
| 28.1 Upstream-Blocked Register and Maintenance Workflow | done | `_bmad-output/implementation-artifacts/28-1-upstream-blocked-register-and-maintenance.md` |
| 28.2 Research HASync Status and Actions | done | `_bmad-output/implementation-artifacts/28-2-research-hasync-status-actions.md` |
| 28.3 Research System Tunables and Sysctl API Lifecycle | done | `_bmad-output/implementation-artifacts/28-3-research-system-tunables-sysctl.md` |
| 28.4 Implement System Tunables Resource | done | `_bmad-output/implementation-artifacts/28-4-implement-system-tunables-resource.md` |

## Epic 29: Appliance Migration Blockers — HAProxy Routing, TLS & ACME Issuance

Goal: close the specific gaps that block migrating a real multi-domain OPNsense edge off Ansible onto this provider. Driven by the downstream `opnsense-manager` appliance, whose `https-in` frontend routes many domains (incl. CNAME aliases) via a domain map + actions, applies internal-only deny rules and an HTTP→HTTPS redirect, and binds an ACME-issued certificate. Verified missing/incomplete against **v0.2.0** source.

| Story | Status | File |
|---|---|---|
| 29.1 HAProxy Action Resource | done | `_bmad-output/implementation-artifacts/29-1-haproxy-action-resource.md` |
| 29.2 HAProxy Map File Resource | done | `_bmad-output/implementation-artifacts/29-2-haproxy-mapfile-resource.md` |
| 29.3 HAProxy Frontend TLS Certificate Binding | done | `_bmad-output/implementation-artifacts/29-3-haproxy-frontend-tls-certificate-binding.md` |
| 29.4 ACME Certificate Issuance & Refid Output | done | `_bmad-output/implementation-artifacts/29-4-acme-certificate-issuance-and-refid.md` |
| 29.5 ddclient Daemon Settings Resource | done | `_bmad-output/implementation-artifacts/29-5-ddclient-daemon-settings-resource.md` |
| 29.6 Multi-Domain Edge Composition & Migration Validation | done | `_bmad-output/implementation-artifacts/29-6-multi-domain-edge-composition-and-migration-validation.md` |

**Notes / corrections to prior tracking:**
- 29.1 / 29.2 / 29.3 were deferred in Story 4.2 ("What NOT to Build") to a Story 4.3 that only shipped ACLs — the action, map file, and frontend cert-binding were never created.
- 29.4 corrects Story 8.2: despite being titled "...with Issuance" and marked done, `certificate_resource.go` is plain CRUD (no `/sign`, no status poll, no refid output).
- 29.5 supersedes the wrongly-cancelled Story 9.5 — daemon settings (`enabled`/`daemon_delay`/`backend`/...) are a distinct `/dyndns/settings` object, not the account-level `service` field.
- **Out of scope (still blocked):** DHCP PXE/TFTP options remain in Stories 11.3 / 21.4 pending live Kea `dhcpv4 *_option` endpoint verification; a full appliance cutover keeps DHCP on Ansible until then.

**Recommended build order:** 29.1 → 29.2 → 29.4 → 29.3 → 29.5 (independent) → 29.6 (capstone). 29.3 wants 29.4's `cert_ref_id`; 29.6 depends on 29.1–29.4.

## Recommended Sequence

1. Treat 25B.1, 25B.2, 25B.3, 26.1, 26.2, 26.3, 27.1, 27.2, 27.3, 27.4, 28.1, 28.2, and 28.3 as completed historical work.
2. Implement Story 28.4 only after controlled live validation confirms safe tunables CRUD/reconfigure behavior on the target appliance.
3. Keep data-source parity, tunables safety validation, and upstream-blocked review as follow-up maintenance until new evidence changes their classification.
4. Epic 29 unblocks the HAProxy/ACME edge migration; build in the order above. Re-verify each OPNsense endpoint/field shape against a live 25.x appliance during implementation (the Dev Notes flag the fields to confirm).
