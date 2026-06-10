# Story 26.1: Published Registry Documentation Audit

Status: done

## Story

As a provider maintainer, I want to audit the published Registry documentation after v0.1.0, so that public docs accurately describe the shipped provider and do not overpromise unsupported domains.

## Acceptance Criteria

1. Published Registry provider page is reviewed against repository docs and planning artifacts.
2. Resource and data-source docs are spot-checked for accuracy, examples, import instructions, and naming consistency.
3. Any stale, misleading, or missing documentation is corrected in templates/examples where possible.
4. A short audit report records findings, fixes, and residual documentation risks.
5. `make check` passes.

## Tasks / Subtasks

- [x] Review published provider index for auth, permissions, version, support matrix, and import guidance.
- [x] Spot-check representative docs across firewall, HAProxy, routing, VPN, DNS, DHCP, system, auth, and trust.
- [x] Fix stale docs in templates/examples/source docs.
- [x] Create a documentation audit report under planning artifacts.
- [x] Run `make check`.

### Review Findings

- [x] [Review][Patch] Update stale audit follow-up now that Stories 26.2, 26.3, and 27.1 are complete [_bmad-output/planning-artifacts/registry-docs-audit.md:66]
- [x] [Review][Patch] Clarify that browser-rendered Registry review remains a residual post-release verification item [_bmad-output/planning-artifacts/registry-docs-audit.md:30]
- [x] [Review][Patch] Refresh post-release recommended sequence to match completed story table [_bmad-output/planning-artifacts/post-release-epics.md:56]
- [x] [Review][Patch] Remove stale source NAT / Unbound forward Coming wording from Story 26.2 context [_bmad-output/implementation-artifacts/26-2-public-support-matrix-registry-docs.md:172]
- [x] [Review][Patch] Update roadmap counts from 34 data sources / 57 gaps to 76 data sources / 15 gaps [_bmad-output/planning-artifacts/feature-complete-roadmap.md:19]

## Dev Notes

The Registry site is JavaScript-rendered; if automated fetch cannot inspect it, compare local generated `docs/` against the Registry manually or via browser-accessible rendered content where available.

## Dev Agent Record

### Agent Model Used

gpt-5.5

### Debug Log References

- Registry web content is JavaScript-rendered; audit used local Registry source files as allowed by Dev Notes.
- No browser-rendered Registry review was recorded; the human Registry spot check remains a residual post-release verification item.
- Structural audit found 90 resource docs, 76 data-source docs, 43 data-source templates, and 43 data-source examples.
- Structural audit found 33 generated data-source docs without custom templates/examples and 59 resource docs without custom import guidance/templates.
- Spot-checked provider index plus representative firewall, HAProxy, WireGuard, routing, traffic shaper, and trust docs.
- Final validation: `make check` passed.
- Code review applied patches for stale follow-up sequencing, stale roadmap counts, outdated source NAT/Unbound-forward Coming wording, and audit scope clarification.
- Review patch validation: `make check` passed on 2026-06-04.

### Completion Notes List

- Created `_bmad-output/planning-artifacts/registry-docs-audit.md` with findings, fixes, residual documentation risks, and recommended follow-up.
- Updated Story 26.2 so support-matrix work reflects 76 data sources after Epic 25B completion instead of the original v0.1.0 count of 34.
- Updated audit follow-up language to reflect that Stories 26.2, 26.3, and 27.1 are now complete.
- Story moved to done after code-review patches and `make check` passed.

### Change Log

- 2026-06-02: Created Registry documentation audit report and moved story to review after `make check` passed.
- 2026-06-04: Applied review patches for stale follow-up/context, reran `make check`, and moved story to done.

### File List

- `_bmad-output/implementation-artifacts/26-1-registry-docs-audit.md`
- `_bmad-output/implementation-artifacts/26-2-public-support-matrix-registry-docs.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `_bmad-output/planning-artifacts/registry-docs-audit.md`
