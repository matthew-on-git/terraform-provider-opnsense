---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 28.2: Research HASync Status and Actions

Status: done

## Story

As an operator,
I want the provider roadmap to classify HASync status and service actions correctly,
so that operational endpoints are exposed as data sources or explicit actions only when they fit Terraform semantics.

## Acceptance Criteria

1. Current OPNsense HASync status/action endpoints are verified from published API docs and/or source, including service list, version, and remote service action endpoints.
2. Each endpoint is classified as one of: durable resource candidate, read-only data-source candidate, action/operation candidate, not planned, or needs more live validation.
3. The story produces a planning artifact that documents endpoint paths, response shapes when known, required privileges if discoverable, Terraform semantic fit, and recommendation.
4. No durable Terraform resource is implemented in this story.
5. Support matrix, core gap analysis, and resource-gap verification are updated to reflect the classification decision.
6. If data-source or action implementation is recommended, follow-up story rows are added to the post-release epic plan and sprint status.
7. `make check` passes.

## Tasks / Subtasks

- [x] Verify endpoint/source evidence (AC: 1, 3)
  - [x] Locate current published docs or source for `core/hasync_status` endpoints.
  - [x] Confirm endpoints listed in Story 27.1 handoff: `services`, `version`, and `remote_service/{action}/{service}/{service_id}`.
  - [x] Capture method, path, parameters, response shape, and whether endpoint mutates appliance state.
  - [x] Check whether endpoint behavior depends on HA configuration or remote peer availability.
- [x] Classify Terraform semantics (AC: 2, 4)
  - [x] Decide whether service/version endpoints are data-source candidates.
  - [x] Decide whether remote service actions are Terraform action candidates, explicitly not planned, or require a product decision.
  - [x] Confirm none of these endpoints should be modeled as durable resources.
- [x] Create planning artifact (AC: 3)
  - [x] Add `_bmad-output/planning-artifacts/hasync-status-actions-research.md` or equivalent.
  - [x] Include endpoint table, classification table, risks, required privileges if available, live-validation needs, and recommended next story/stories.
- [x] Update planning docs (AC: 5, 6)
  - [x] Update `_bmad-output/planning-artifacts/resource-gap-verification.md`.
  - [x] Update `_bmad-output/planning-artifacts/core-config-gap-analysis.md`.
  - [x] Update `_bmad-output/planning-artifacts/support-matrix.md`.
  - [x] Add follow-up sprint-status/post-release rows only if the classification yields concrete implementation work.
- [x] Validate (AC: 7)
  - [x] Run `make check` and fix documentation/check failures without suppressing checks.
  - [x] Record final classification, created artifact path, follow-up story keys if any, and validation result in this story's Dev Agent Record.

## Dev Notes

### Source Context

Story 27.1 classified HASync status/actions as Needs research because the published endpoints appear operational/status-oriented rather than durable configuration. The handoff listed these candidate endpoints:

- `GET /api/core/hasync_status/services`
- `GET /api/core/hasync_status/version`
- `GET /api/core/hasync_status/remote_service/{action}/{service}/{service_id}`
- service actions such as `start`, `stop`, `restart`, and `restart_all`

Direct guessed docs URL `https://docs.opnsense.org/development/api/core/hasync_status.html` returned 404 during story creation. Research must use the current API index and/or source tree instead of relying on that guessed path.

### Guardrails

- This is a research/classification story only; do not implement resources, data sources, or actions here.
- Do not classify remote service actions as resources unless durable desired state semantics are demonstrated, which is unlikely.
- If recommending Terraform Plugin Framework actions, verify the provider framework version and Registry documentation implications first.
- Keep HASync configuration singleton separate from this story.

### References

- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `_bmad-output/implementation-artifacts/27-4-implement-hasync-configuration-resource.md`

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- Created from post-Story 27.2 follow-up request on 2026-06-05.
- Loaded resource-gap verification, support matrix, core gap analysis, and post-release epic plan.
- Direct guessed HASync status documentation page returned 404, so story requires current API/source re-verification.
- Fetched current OPNsense core API docs and confirmed `core/hasync_status` endpoints for `services`, `version`, documented `remote_service`, and POST service operations.
- Fetched current `HasyncStatusController.php`; `servicesAction()` returns a cached service recordset, `versionAction()` returns decoded remote version JSON, and `start`/`stop`/`restart`/`restartAll` are POST operations that execute commands on the HA peer.
- Created `_bmad-output/planning-artifacts/hasync-status-actions-research.md` with endpoint table, response-shape evidence, risks, live-validation needs, and recommendations.
- No resources, data sources, or actions were implemented.
- Full validation passed: `make check`.
- Code review found two documentation refinements: post-release sequencing still treated 28.2 as pending, and the documented `remote_service` route needed one explicit AC category. Patched both.

### Completion Notes List

- Created research story for HASync status/action classification.
- Classified `services` and `version` as future read-only data-source candidates only after live response/error-shape validation.
- Classified `start`, `stop`, `restart`, and `restart_all` as future action candidates only after explicit product/framework decision.
- Confirmed none of the `core/hasync_status` endpoints should be modeled as durable Terraform resources.
- Kept HASync status/actions as Needs research; no immediate follow-up implementation story was added because live HA-peer validation and action semantics decisions remain prerequisites.
- Code review patch applied: documented `remote_service` as Needs more live validation due published docs/current source mismatch.

### File List

- `_bmad-output/implementation-artifacts/28-2-research-hasync-status-actions.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/hasync-status-actions-research.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/support-matrix.md`

### Change Log

- 2026-06-05: Created story for HASync status/action research and classification.
- 2026-06-09: Completed HASync status/action research classification; no durable resource, data source, or action implemented.
- 2026-06-09: Addressed code review documentation findings for `remote_service` classification and post-release sequencing.

## Senior Developer Review (AI)

### Review Date

2026-06-09

### Review Outcome

Approve after documentation patches.

### Findings

- [x] Low: post-release recommended sequence still described 28.2 as pending. Fixed.
- [x] Low: `remote_service` endpoint classification was ambiguous against AC2 categories. Fixed to Needs more live validation.

### Validation

- `make check` passed before review patches and again after review patches.
