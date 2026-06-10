---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 28.3: Research System Tunables and Sysctl API Lifecycle

Status: done

## Story

As an operator,
I want the provider roadmap to classify OPNsense system tunables/sysctl support accurately,
so that the provider only exposes tunable management when endpoint lifecycle and Terraform semantics are stable.

## Acceptance Criteria

1. Current OPNsense tunables/sysctl API evidence is verified from published docs and/or source, including endpoint paths, model files, and target-version availability.
2. Endpoint lifecycle is classified: durable configuration CRUD, operational runtime sysctl, mixed behavior, unavailable, or version-dependent.
3. The story produces a planning artifact documenting endpoint paths, wrappers, model fields, reconfigure/apply lifecycle, live-validation requirements, and recommended provider classification.
4. No Terraform resource is implemented in this story.
5. Support matrix, core gap analysis, and resource-gap verification are updated to classify tunables/sysctl as Coming, Needs research, Upstream-blocked, or Not planned with evidence.
6. If implementation is recommended, follow-up story rows are added to the post-release epic plan and sprint status.
7. `make check` passes.

## Tasks / Subtasks

- [x] Verify endpoint/source evidence (AC: 1, 2, 3)
  - [x] Locate current OPNsense published API docs or source for `core/tunables` or equivalent sysctl/tunable endpoints.
  - [x] Capture method, path, parameters, wrapper key, model file, and response shape.
  - [x] Determine whether endpoints manage persistent configuration, runtime-only sysctl values, or both.
  - [x] Determine whether changes require reconfigure/apply/reboot and whether the provider can detect drift safely.
  - [x] Verify target-version availability against the provider's supported OPNsense version range.
- [x] Classify Terraform semantics (AC: 2, 4, 5)
  - [x] Decide whether tunables are resource candidates, data-source candidates, not planned, or still Needs research.
  - [x] Identify risks: appliance lockout, kernel/network instability, non-idempotent runtime state, validation gaps, and reboot-only effects.
  - [x] Do not implement provider code in this story.
- [x] Create planning artifact (AC: 3)
  - [x] Add `_bmad-output/planning-artifacts/tunables-sysctl-research.md` or equivalent.
  - [x] Include endpoint table, lifecycle classification, field/model notes, safety risks, and recommendation.
- [x] Update planning docs (AC: 5, 6)
  - [x] Update `_bmad-output/planning-artifacts/resource-gap-verification.md`.
  - [x] Update `_bmad-output/planning-artifacts/core-config-gap-analysis.md`.
  - [x] Update `_bmad-output/planning-artifacts/support-matrix.md`.
  - [x] Add follow-up sprint-status/post-release rows only if implementation is recommended.
- [x] Validate (AC: 7)
  - [x] Run `make check` and fix documentation/check failures without suppressing checks.
  - [x] Record final classification, created artifact path, follow-up story keys if any, and validation result in this story's Dev Agent Record.

## Dev Notes

### Source Context

Story 27.1 moved system tunables/sysctl from a simple upstream-blocked assumption to Needs research because current OPNsense docs may expose `core/tunables` endpoints, but endpoint paths, wrapper, lifecycle, and target-version availability were not verified.

Direct guessed docs URL `https://docs.opnsense.org/development/api/core/tunables.html` returned 404 during story creation. Research must use the current API index and/or source tree instead of relying on that guessed path.

### Guardrails

- This is a research/classification story only; do not implement resources or data sources here.
- Tunables can affect kernel/network behavior. Treat safety, validation, and recovery as first-class classification criteria.
- Do not move tunables/sysctl to Coming without durable endpoint evidence and a credible lifecycle story.
- Do not keep tunables/sysctl in Upstream-blocked if current target-version endpoints are confirmed usable.

### References

- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- OPNsense source tree for core tunables/sysctl controllers and models

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- Created from post-Story 27.2 follow-up request on 2026-06-05.
- Loaded resource-gap verification, support matrix, core gap analysis, and post-release epic plan.
- Direct guessed tunables documentation page returned 404, so story requires current API/source re-verification.
- Fetched current OPNsense core API docs and confirmed `core/tunables` item CRUD/search, root get/set, `reconfigure`, and `reset` endpoints.
- Fetched current `TunablesController.php`; controller uses internal model name `sysctl`, wrapper `sysctl`, model path `item`, and `reconfigureAction()` restarts `login` and `sysctl`.
- Fetched current `Tunables.xml`; model mounts at `//sysctl`, uses `TunableField`, and exposes persistent `tunable`, `value`, `descr` plus volatile `default_value` and `type`.
- Created `_bmad-output/planning-artifacts/tunables-sysctl-research.md` with endpoint table, lifecycle classification, safety risks, target-version note, and recommendation.
- No Terraform resource or data source was implemented.
- Full validation passed: `make check`.
- Code review found stale PRD/roadmap count/classification wording and over-precise volatile/import semantics in the research artifact. Patched those and clarified that minimum-version availability remains part of the live-validation gate.

### Completion Notes List

- Created research story for system tunables/sysctl endpoint lifecycle classification.
- Classified system tunables/sysctl as Coming with safety/live-validation gate.
- Confirmed tunables are durable configuration CRUD with operational apply side effects through `reconfigure`; runtime-only sysctl state should not be modeled as desired state.
- Added follow-up backlog row `28-4-implement-system-tunables-resource` to sprint status and post-release epics, gated on live CRUD/reconfigure validation.
- Kept `reset` out of normal resource lifecycle because it restores factory tunables broadly.
- Code review patch applied: target-version availability across the provider minimum range remains unverified and is included in the 28.4 live-validation gate.

### File List

- `_bmad-output/implementation-artifacts/28-3-research-system-tunables-sysctl.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `_bmad-output/planning-artifacts/tunables-sysctl-research.md`
- `RELEASE.md`
- `docs/index.md`
- `docs/migration-import.md`
- `docs/upstream-blocked.md`
- `templates/index.md.tmpl`

### Change Log

- 2026-06-05: Created story for system tunables/sysctl research and classification.
- 2026-06-09: Completed tunables/sysctl endpoint lifecycle research and reclassified as Coming with safety/live-validation gate.
- 2026-06-09: Addressed code review findings for stale docs and target-version/volatile-field wording.

## Senior Developer Review (AI)

### Review Date

2026-06-09

### Review Outcome

Approve after documentation patches.

### Findings

- [x] Medium: roadmap still said tunables/sysctl needed endpoint lifecycle research after Story 28.3 classified it as Coming. Fixed.
- [x] Medium: target-version availability across the provider minimum range was overstated. Clarified as part of the live-validation gate.
- [x] Low: PRD still had stale 96-resource wording. Fixed.
- [x] Low: volatile fields were described as read-only without proving API read-only semantics. Reworded as non-persistent.
- [x] Low: `set_item` non-UUID key behavior was described as tunable-name/import behavior. Reworded as name/key-as-create behavior requiring live validation.

### Validation

- `make check` passed before review patches and again after review patches.
