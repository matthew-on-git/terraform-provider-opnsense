---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 27.1: Verify Remaining Buildable Resource Gaps

Status: done

## Story

As a provider maintainer, I want to verify the remaining buildable resource gaps against current OPNsense APIs, so that implementation work targets real endpoints and avoids speculative resources.

## Acceptance Criteria

1. Each remaining Coming resource is verified against current OPNsense docs or a live appliance.
2. Verification records endpoint paths, CRUD/search/reconfigure behavior, wrapper key, and singleton vs UUID lifecycle.
3. Items without confirmed API support are reclassified as Upstream-blocked, Not planned, or Needs research.
4. `core-config-gap-analysis.md` and `support-matrix.md` are updated with final classifications.
5. `make check` passes.

## Tasks / Subtasks

- [x] Verify firewall source NAT endpoint and relation to existing NAT resources.
- [x] Verify interface LAGG endpoint.
- [x] Verify Kea DHCPv4 option and Kea DDNS endpoints.
- [x] Verify Dnsmasq host/domain/range/option/tag endpoints.
- [x] Verify Unbound forward endpoint.
- [x] Verify OSPF area and HASync priority/status.
- [x] Update planning artifacts with verified classifications.
- [x] Run `make check`.

## Dev Notes

Use current OPNsense target version assumptions from PRD/provider docs. Do not implement in this story; this is a verification gate for Story 27.2.

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- Loaded current story, sprint tracker, PRD, support matrix, core gap analysis, post-release epic plan, generated resource docs, and local schemas.
- Fetched OPNsense published API documentation on 2026-06-02 for `interfaces`, `firewall`, `kea`, `dnsmasq`, `unbound`, `quagga`, `core`, `routes`, and `routing` API modules.
- Ran `make check`; first attempt timed out under the tool harness without a project result, second attempt passed.
- Code review found provisional endpoint/wrapper evidence presented too strongly, Kea planning conflicts, missing public `Needs research` status, stale Story 27.2 candidate order, and possible Dnsmasq `boot` / tunables follow-ups.

### Completion Notes List

- Added `_bmad-output/planning-artifacts/resource-gap-verification.md` as the Story 27.1 evidence and handoff artifact for Story 27.2.
- Reclassified firewall source NAT as already Supported via `opnsense_firewall_nat_outbound`.
- Reclassified Unbound forward as already Supported via `opnsense_unbound_domain_override`.
- Verified LAGG, Dnsmasq item resources, OSPF area, and HASync configuration as Coming based on published OPNsense API docs, with implementation gates for live member validation or model-wrapper review where needed.
- Reclassified Kea DHCPv4 option and Kea DDNS from Upstream-blocked to Needs research because current published docs expose endpoints but earlier live probing found endpoint-not-found.
- Classified HASync status/actions as Needs research because published endpoints are operational/status-oriented rather than durable configuration CRUD.
- Added review follow-ups for Dnsmasq `boot` item verification and tunables/sysctl endpoint-lifecycle verification.
- Updated `core-config-gap-analysis.md`, `support-matrix.md`, provider index support wording, Story 27.2 handoff, PRD, and roadmap to remove stale source NAT/Unbound-forward Coming language.
- `make check` passed on 2026-06-04 after review patches: lint, format, test, security, scan, and docs all passed.

### File List

- `_bmad-output/implementation-artifacts/27-1-verify-buildable-resource-gaps.md`
- `_bmad-output/implementation-artifacts/27-2-implement-highest-value-verified-gap.md`
- `_bmad-output/implementation-artifacts/26-2-public-support-matrix-registry-docs.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `docs/index.md`
- `templates/index.md.tmpl`

## Change Log

- 2026-06-02: Verified remaining provider-owned resource gaps, updated planning classifications, and moved story to review.
- 2026-06-04: Applied code-review patches for provisional verification wording, Kea Needs research status, public status taxonomy, stale handoffs, Dnsmasq boot follow-up, and tunables/sysctl research tracking.
- 2026-06-04: Ran `make check` successfully and moved story to done.
