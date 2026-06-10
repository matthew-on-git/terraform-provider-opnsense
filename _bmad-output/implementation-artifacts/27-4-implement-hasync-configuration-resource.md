---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 27.4: Implement HASync Configuration Resource

Status: done

## Story

As an operator,
I want a Terraform resource for OPNsense HA synchronization configuration,
so that primary/backup synchronization settings can be managed declaratively through Terraform when the core API exposes stable durable configuration.

## Acceptance Criteria

1. HASync configuration is model-reviewed before implementation: endpoint availability, wrapper key, model fields, field classes, defaults, and durable Terraform semantics are verified from current OPNsense docs/source.
2. If model review confirms safe durable configuration semantics, a singleton HASync configuration resource is implemented under the project-consistent service package and naming convention.
3. The resource uses the confirmed `GET /api/core/hasync/get`, `POST /api/core/hasync/set`, and `POST /api/core/hasync/reconfigure` lifecycle or the current equivalent endpoints if source review shows different paths.
4. The implementation follows existing singleton resource patterns: YAML schema if codegen can safely represent the model, generated model/schema/resource/test files, package exports registration, singleton import, state read-back, drift detection, and reconfigure after set.
5. If the model cannot be safely represented or durable semantics are not clear, no resource is shipped; the story updates planning docs with evidence and keeps HASync configuration classified as Coming or Needs research as appropriate.
6. Registry docs/examples and current-facing support counts are updated only if a resource is implemented.
7. HASync status/actions are not implemented in this story; they remain a separate research/data-source/action story.
8. `make check` passes.

## Tasks / Subtasks

- [x] Model-review gate (AC: 1, 3, 5, 7)
  - [x] Locate current OPNsense source for HASync controller and model, expected from prior verification as `core/hasync` and `Hasync.xml`.
  - [x] Confirm the current endpoint paths for get, set, and reconfigure.
  - [x] Confirm request/response wrapper key; do not assume `hasync` without reading controller/model behavior.
  - [x] Extract durable configuration fields, field classes, defaults, required constraints, and sensitive/write-only fields.
  - [x] Verify this story excludes `core/hasync_status` status/action endpoints.
  - [x] Decide whether the model is safe for Terraform CRUD-style singleton management; document evidence before implementation.
- [x] Implement resource if safe (AC: 2, 3, 4)
  - [x] Skipped implementation after model review found unsupported dynamic `JsonKeyValueStoreField` `syncitems` semantics.
  - [x] No package/type naming, generator schema, manual resource, tests, exports, or provider registration were added.
- [x] Add docs/examples if implemented (AC: 6)
  - [x] Skipped resource docs/examples because no HASync resource was implemented.
  - [x] Verified support counts and generated docs remain unchanged.
- [x] Update planning and release-facing docs (AC: 5, 6, 7)
  - [x] If implemented, update support counts and classify HASync configuration as Supported.
  - [x] If not implemented, update `_bmad-output/planning-artifacts/resource-gap-verification.md` and related current-facing docs with the exact blocker.
  - [x] Keep HASync status/actions classified separately as Needs research unless this story only improves wording.
- [x] Validate (AC: 1-8)
  - [x] Run focused tests for the touched package and `./internal/provider`.
  - [x] Run `make check` and fix failures without suppressing checks.
  - [x] Record implementation decision, evidence, final counts if changed, and validation result in this story's Dev Agent Record.

## Dev Notes

### Source Context

Story 27.1 classified HASync configuration as Coming with a model-review gate. The published-core-API evidence recorded in `resource-gap-verification.md` says `core/hasync/get`, `core/hasync/set`, and `core/hasync/reconfigure` exist using `Hasync.xml`, but the model fields and wrapper shape were not reviewed deeply enough to implement during Story 27.2.

Direct guessed docs URLs for `core/hasync.html` and `core/hasync_status.html` returned 404 during story creation, so implementation must verify from the current published API index and/or OPNsense source tree instead of relying on a guessed page path.

### Existing Patterns to Reuse

- Singleton generator pattern: existing singleton resources in `internal/generate/schemas/*` and generated `*_resource.gen.go` files.
- Core/system package conventions: inspect `internal/service/system`, `internal/service/unbound`, and similar singleton resources before choosing final naming.
- Sensitive/write-only field handling: inspect trust/IPsec/auth stories and generator flags if HASync has credentials or passwords.

### Guardrails

- Do not implement HASync status/actions in this story.
- Do not ship a resource if the API only exposes operational actions or status without durable config semantics.
- Do not guess wrapper keys; verify them in controller/source or live API responses.
- Do not update support counts unless a resource is actually implemented and documented.
- Run all tooling through existing project/container workflows; do not install tools on the host.

### References

- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `internal/generate/schemas/*` singleton examples
- `internal/service/system/*` and other singleton service packages

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- Created from post-Story 27.2 follow-up request on 2026-06-05.
- Loaded resource-gap verification, support matrix, core gap analysis, and post-release epic plan.
- Direct guessed HASync documentation pages returned 404, so story requires current API/source re-verification before implementation.
- Fetched current OPNsense core API index and confirmed `core/hasync/get`, `core/hasync/set`, and `core/hasync/reconfigure`; `core/hasync_status` remains separate status/action API surface.
- Fetched current `HasyncController.php`; controller extends `ApiMutableModelControllerBase`, declares internal model name `hasync`, and implements `reconfigureAction()` using `interface pfsync configure` on POST.
- Fetched current `Hasync.xml`; durable fields are `disablepreempt`, `disconnectppps`, `pfsyncinterface`, `pfsyncpeerip`, `pfsyncversion`, `pfsyncdefer`, `synchronizetoip`, `verifypeer`, `username`, `password`, and `syncitems`.
- Model-review blocker: `syncitems` is `JsonKeyValueStoreField` with `Multiple=Y` and `ConfigdPopulateAct=system ha options`; current generator supports `bool`, `int`, `string`, `selectmap`, `selectmaplist`, and `csvset`, but not this dynamic key/value-store shape.
- No HASync resource was implemented because omitting or guessing `syncitems` would risk losing/drifting the core application synchronization selection.
- Full validation passed: `make check`.
- Code review found documentation consistency issues only; applied patches for skipped conditional tasks, stale PRD count wording, post-release status, README taxonomy, HASync wrapper uncertainty, and Coming/Needs research taxonomy.

### Completion Notes List

- Created model-review-gated developer guide for HASync configuration singleton implementation.
- Completed the HASync model-review gate and intentionally did not ship a resource.
- Reclassified HASync configuration from buildable Coming to Needs research until `JsonKeyValueStoreField` request/response shape and a safe Terraform representation are verified or implemented.
- Support counts remain unchanged after Story 27.4: 97 resources, 83 data sources, 97 resource docs, and 83 data-source docs.
- HASync status/actions remain a separate Needs research item and were not implemented.
- Code review completed with documentation consistency patches; no provider code was added.

### File List

- `_bmad-output/implementation-artifacts/27-4-implement-hasync-configuration-resource.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `docs/index.md`
- `templates/index.md.tmpl`

### Change Log

- 2026-06-05: Created story for HASync configuration model review and implementation.
- 2026-06-09: Completed HASync model review; no resource shipped because `syncitems` uses unsupported dynamic `JsonKeyValueStoreField` semantics.
- 2026-06-09: Addressed code review documentation consistency findings and revalidated with `make check`.

## Senior Developer Review (AI)

### Review Date

2026-06-09

### Review Outcome

Approve after documentation patches.

### Findings

- [x] Medium: stale PRD baseline wording still said 96-resource surface after Story 27.3 moved current baseline to 97 resources. Fixed.
- [x] Medium: post-release epic table still listed Story 27.4 as ready-for-dev. Fixed to done.
- [x] Low: conditional implementation/docs tasks were checked too literally despite no resource being shipped. Reworded as skipped due model-review blocker.
- [x] Low: README and planning docs blurred Coming and Needs research taxonomy after moving HASync configuration to Needs research. Fixed.
- [x] Low: HASync wrapper wording overstated certainty without a captured live payload. Reworded as likely `hasync`, to confirm before implementation.

### Validation

- `go test ./internal/provider` passed.
- `make check` passed before review and again after review patches.
