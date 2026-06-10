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
| 28.4 Implement System Tunables Resource | backlog | TBD - create story after live validation confirms safe CRUD/reconfigure behavior |

## Recommended Sequence

1. Treat 25B.1, 25B.2, 25B.3, 26.1, 26.2, 26.3, 27.1, 27.2, 27.3, 27.4, 28.1, 28.2, and 28.3 as completed historical work.
2. Create Story 28.4 only after live validation confirms safe tunables CRUD/reconfigure behavior on the target appliance.
3. Keep data-source parity, interface LAGG validation, tunables safety validation, and upstream-blocked review as follow-up maintenance until new evidence changes their classification.
