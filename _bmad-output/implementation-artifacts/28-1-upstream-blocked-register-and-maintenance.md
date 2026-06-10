---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 28.1: Upstream-Blocked Register and Maintenance Workflow

Status: done

## Story

As a provider maintainer, I want a public upstream-blocked register and maintenance workflow, so that users understand which gaps require OPNsense API work and how those gaps will be revisited.

## Acceptance Criteria

1. Upstream-blocked domains are documented publicly with reason, upstream item, and provider impact.
2. Register includes interface assignment/IP config/PPPoE, gateway group, and system general settings; tunables/sysctl remains Needs research per user clarification.
3. Maintenance workflow defines how to review blocked items after OPNsense releases.
4. Provider docs or README link to the blocked register or summarize it clearly.
5. `make check` passes.

## Tasks / Subtasks

- [x] Create or expand a public blocked register document.
- [x] Add upstream references, current API status, and provider resource names affected.
- [x] Define review cadence tied to OPNsense major releases.
- [x] Link the register from provider docs, README, or support matrix.
- [x] Run `make check`.

## Dev Notes

This story is not only documentation. It protects product credibility by preventing unsupported domains from being mistaken for provider neglect.

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- Loaded support matrix, core gap analysis, resource-gap verification, and post-release epic plan.
- User clarified that `tunables/sysctl` should remain out of upstream-blocked and stay classified as Needs research.
- Created `docs/upstream-blocked.md` with confirmed upstream blockers only: interface assignment/IP config/PPPoE, gateway group, and system general settings.
- Linked the public register from README, provider index template/docs, support matrix, and core gap analysis.
- Marked Story 28.1 done in sprint status and post-release epic tracking.
- Full validation passed: `make check`.
- Code review found stale references that still treated tunables/sysctl as upstream-blocked; updated story AC wording, migration docs, release checklist, PRD, roadmap, README links, blocked-register links, and post-release sequencing.

### Completion Notes List

- Added a public upstream-blocked register with reason, upstream status, provider impact, and OPNsense-major-release review workflow.
- Excluded system tunables/sysctl from the upstream-blocked register and documented it as Needs research because endpoint lifecycle/wrapper semantics remain unresolved.
- Updated support-facing links so users can find the blocked register from README, Registry/provider docs, and planning artifacts.
- Applied code review patches so public docs consistently keep system tunables/sysctl as Needs research rather than upstream-blocked.

### File List

- `_bmad-output/implementation-artifacts/28-1-upstream-blocked-register-and-maintenance.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `README.md`
- `RELEASE.md`
- `docs/index.md`
- `docs/migration-import.md`
- `docs/upstream-blocked.md`
- `templates/index.md.tmpl`

### Change Log

- 2026-06-09: Added public upstream-blocked register and maintenance workflow, linked it from public docs, and kept tunables/sysctl classified as Needs research per user clarification.
- 2026-06-09: Addressed code review findings for stale tunables/sysctl upstream-blocked references and link consistency.

## Senior Developer Review (AI)

### Review Date

2026-06-09

### Review Outcome

Approve after documentation patches.

### Findings

- [x] High: migration guide still classified system tunables/sysctl as upstream-blocked. Fixed to Needs research.
- [x] Medium: release checklist, PRD, and roadmap still grouped tunables/sysctl with upstream-blocked domains. Fixed.
- [x] Low: story AC2 still included tunables/sysctl despite user clarification. Fixed.
- [x] Low: README and blocked-register maintenance workflow used plain paths where links were clearer. Fixed.
- [x] Low: post-release sequencing still framed completed stories as upcoming and implied HASync was implemented. Fixed.

### Validation

- `make check` passed before review patches and again after review patches.
