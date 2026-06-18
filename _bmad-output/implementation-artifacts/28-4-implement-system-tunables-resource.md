---
baseline_commit: fbaf085e8287ae4f00f786484cd1ab622d77716a
---

# Story 28.4: Implement System Tunables Resource

Status: done

## Story

As an operator,
I want an `opnsense_system_tunable` resource,
so that I can manage persistent OPNsense system tunables through Terraform without modeling unsafe runtime-only sysctl state.

## Context

Story 28.3 moved system tunables/sysctl out of upstream-blocked after source and API research confirmed `core/tunables` item CRUD with persistent fields under wrapper/model key `sysctl`. Live non-mutating validation against the current 25.x Vagrant appliance confirmed `get_item`, `search_item`, and `reconfigure` endpoint availability and wrapper shape. Story creation deliberately avoided create/update/delete mutation because even low-risk sysctls can affect appliance runtime behavior; this story must perform that controlled live validation before implementation is marked complete.

## Acceptance Criteria

1. **Given** a supported OPNsense appliance exposes `core/tunables`
   **When** an operator defines `opnsense_system_tunable`
   **Then** the provider creates persistent tunable config through `POST /api/core/tunables/add_item` using wrapper `sysctl` and stores the returned UUID as `id`.

2. **And** the resource supports read, update, delete, and import through `/api/core/tunables/get_item/{uuid}`, `/api/core/tunables/set_item/{uuid}`, `/api/core/tunables/del_item/{uuid}`, and UUID import.

3. **And** every create, update, and delete applies changes through `POST /api/core/tunables/reconfigure`; reconfigure failure surfaces as a Terraform diagnostic instead of being ignored.

4. **And** the Terraform schema manages only persistent configuration fields: `tunable` (String, required), `value` (String, required), and `description` (String, optional/computed default empty, mapped to API field `descr`).

5. **And** volatile API/model fields `default_value` and `type` are not accepted as desired configuration. They may be omitted entirely, or exposed only as computed read-only attributes if live readback proves stable; they must not cause diff churn.

6. **And** the implementation adds data-source parity as `data "opnsense_system_tunable"` by UUID unless a concrete API/readback issue makes parity unsafe; any omission must be documented in the story completion notes.

7. **And** acceptance tests cover create/read, import verification, update, delete cleanup, and an idempotent no-change plan after apply against a safe fixture tunable selected during implementation.

8. **And** acceptance tests skip clearly when `core/tunables` endpoints are missing on an appliance version, using existing acceptance precheck helpers rather than failing obscurely.

9. **And** docs/examples are generated for the resource and data source, including safety notes that tunables can affect networking/kernel behavior and some values may require reboot or subsystem reload beyond `reconfigure`.

10. **And** public support/counting docs are updated if resource or data-source counts change, and `make check` passes before the story is marked done.

## Tasks / Subtasks

- [x] Task 1: Live mutation probe and safe fixture selection (AC: #1, #2, #3, #7, #8)
  - [x] Confirm `add_item`, `set_item`, `del_item`, and `reconfigure` behavior on the target Vagrant appliance before relying on implementation assumptions.
  - [x] Select a low-risk, reversible fixture tunable/value pair and document why it is safe. Avoid networking, firewall, login, routing, storage, and kernel stability tunables.
  - [x] Verify delete cleanup removes only the Terraform-created item and never calls `/api/core/tunables/reset`.
- [x] Task 2: Implement model and request/response mapping (AC: #1, #2, #4, #5)
  - [x] Add `tunable_model.go` or equivalent in `internal/service/system`.
  - [x] Map API `tunable` → Terraform `tunable`, `value` → `value`, and `descr` → `description`.
  - [x] Accept read responses that include `default_value` and `type` without making them configurable desired state.
- [x] Task 3: Implement resource CRUD (AC: #1, #2, #3)
  - [x] Add `tunable_resource.go` using existing `opnsense.Add`, `Get`, `Update`, and `Delete` helpers.
  - [x] Use `ReqOpts` endpoints `/api/core/tunables/add_item`, `/api/core/tunables/get_item`, `/api/core/tunables/set_item`, `/api/core/tunables/del_item`, `/api/core/tunables/search_item`, `/api/core/tunables/reconfigure`, with `Monad: "sysctl"`.
  - [x] Implement UUID import with `resource.ImportStatePassthroughID`.
- [x] Task 4: Implement resource schema (AC: #4, #5, #9)
  - [x] Add `tunable_schema.go` with Markdown descriptions that warn about tunable safety.
  - [x] Keep `description` optional/computed with default empty string to match existing system resource style.
  - [x] Do not add provider-side allow-lists that can go stale unless live validation proves an OPNsense-defined finite enum.
- [x] Task 5: Implement data-source parity (AC: #6)
  - [x] Add `tunable_data_source.go` reading by required UUID, mirroring existing system data-source patterns.
  - [x] Update `internal/service/system/data_source_schema_test.go` expected constructor count and assertions.
- [x] Task 6: Register resource and data source (AC: #1, #6)
  - [x] Add constructors to `internal/service/system/exports.go`.
- [x] Task 7: Add tests (AC: #7, #8)
  - [x] Add focused unit/schema tests where useful for model conversion or data-source readback.
  - [x] Add `internal/service/system/tunable_resource_test.go` acceptance coverage with endpoint precheck and serial-safe fixture cleanup.
  - [x] Ensure import verification succeeds and an idempotent plan step shows no diff.
- [x] Task 8: Add examples and generated docs (AC: #9, #10)
  - [x] Add examples under `examples/resources/opnsense_system_tunable` and, if built, `examples/data-sources/opnsense_system_tunable`.
  - [x] Add templates under `templates/resources/system_tunable.md.tmpl` and, if built, `templates/data-sources/system_tunable.md.tmpl`.
  - [x] Regenerate docs with the project's existing docs generation path.
- [x] Task 9: Update planning/support artifacts and validate (AC: #10)
  - [x] Update support matrix, roadmap/count docs, and any generated index files affected by the new resource/data source.
  - [x] Run targeted tests for `internal/service/system`.
  - [x] Run `make check` and fix failures without suppressing checks.

### Review Findings

- [x] [Review][Patch] Public support/counting docs are internally inconsistent [core-config-gap-analysis.md:27]
- [x] [Review][Patch] Acceptance precheck only verifies `get_item`, not mutation/apply endpoints [internal/service/system/tunable_resource_test.go:19]
- [x] [Review][Patch] Acceptance destroy check does not prove tunable cleanup [internal/service/system/tunable_resource_test.go:23]
- [x] [Review][Patch] Data-source documentation omits the required tunables safety note [docs/data-sources/system_tunable.md:8]
- [x] [Review][Defer] Mutation succeeds but failed `reconfigure` can orphan or desync Terraform state [pkg/opnsense/crud.go:58] — deferred, pre-existing

## Dev Notes

### API Contract

| Operation | Method | Endpoint | Wrapper / notes |
|---|---|---|---|
| Create | POST | `/api/core/tunables/add_item` | Request wrapper `sysctl`; returns UUID mutation response. |
| Read | GET | `/api/core/tunables/get_item/{uuid}` | Response wrapper `sysctl`. Live non-mutating probe returned fields `tunable`, `value`, `descr`, `default_value`, and `type`. |
| Update | POST | `/api/core/tunables/set_item/{uuid}` | Request wrapper `sysctl`; must call reconfigure through shared helper. |
| Delete | POST | `/api/core/tunables/del_item/{uuid}` | Delete only the managed UUID; do not call broad `reset`. |
| Search | GET/POST | `/api/core/tunables/search_item` | Available for lookup/readback support, not required for basic UUID data source. |
| Apply | POST | `/api/core/tunables/reconfigure` | Live non-mutating probe returned `{"status":"ok"}`; shared CRUD helpers call this after mutations. |
| Reset | POST | `/api/core/tunables/reset` | Out of scope; destructive/broad relative to one Terraform resource. |

Suggested request options:

```go
var tunableReqOpts = opnsense.ReqOpts{
    AddEndpoint:         "/api/core/tunables/add_item",
    GetEndpoint:         "/api/core/tunables/get_item",
    UpdateEndpoint:      "/api/core/tunables/set_item",
    DeleteEndpoint:      "/api/core/tunables/del_item",
    SearchEndpoint:      "/api/core/tunables/search_item",
    ReconfigureEndpoint: "/api/core/tunables/reconfigure",
    Monad:               "sysctl",
}
```

### Schema

| API field | Terraform attr | Type | Required behavior |
|---|---|---|---|
| `tunable` | `tunable` | String | Required. The sysctl/tunable name. |
| `value` | `value` | String | Required. Keep string-typed; OPNsense values may be numeric, boolean-like, or text. |
| `descr` | `description` | String | Optional/computed, default empty string. |
| `default_value` | optional computed-only or omitted | String | Volatile/non-persistent. Do not configure. |
| `type` | optional computed-only or omitted | String | Volatile/non-persistent. Do not configure. |

### Existing Patterns To Reuse

- `internal/service/system/vlan_resource.go`, `vlan_model.go`, `vlan_schema.go`, and `vlan_data_source.go` are the closest hand-written system item-resource pattern.
- `internal/service/system/vip_resource.go` shows the same CRUD/import/delete-not-found handling.
- `pkg/opnsense/crud.go` already wraps request bodies by monad, parses UUID mutation responses, serializes writes through the global mutex, and calls `Reconfigure` after `Add`, `Update`, and `Delete`. Do not create custom HTTP code for standard item CRUD.
- `internal/acctest/acctest.go` includes `SkipIfEndpointMissing`, `PreCheck`, and `CheckResourceDestroyed`; use those for acceptance tests.

### Safety And Fixture Guidance

- Tunables can break networking, packet filtering, routing, login behavior, storage, or kernel stability. The acceptance fixture must be intentionally low-risk and reversible.
- Do not choose a fixture solely because it is familiar; verify the tunable exists and can be safely set on the test appliance.
- Avoid any fixture that could drop the API connection, require reboot to recover, alter interface assignment, affect firewall/NAT, or broadly change kernel networking behavior.
- Record the chosen fixture and live validation result in the Dev Agent Record.

### Project Structure Notes

- Implementation files belong in `internal/service/system/` and should follow existing naming: `tunable_resource.go`, `tunable_model.go`, `tunable_schema.go`, and `tunable_data_source.go` if parity is included.
- Generated Registry docs are produced from `templates/resources/*.md.tmpl` and `templates/data-sources/*.md.tmpl`; do not hand-edit generated docs as the source of truth unless the project pattern requires it.
- Resource registration is through `internal/service/system/exports.go`; provider-level aggregation already consumes service exports.
- Keep all tooling inside the project Makefile/dev-toolchain path. Do not install host tools.

### Testing Requirements

- Unit/package tests: run `go test ./internal/service/system ./pkg/opnsense` or a narrower equivalent while developing.
- Acceptance tests: run system package acceptance serially against the Vagrant appliance, for example `test/scripts/run-vagrant-acceptance.sh --package ./internal/service/system` with the required `OPNSENSE_*` environment variables.
- Whole-project gate: `make check` is mandatory before completion.
- Acceptance tests must not depend on whole-suite ordering and must clean up their created tunable even if update/import steps fail where feasible.

### Previous Story Intelligence

- Story 28.3 established that this is durable configuration CRUD with operational apply side effects, not runtime-only sysctl desired state.
- Story 28.3 explicitly kept `/api/core/tunables/reset` out of normal lifecycle because it restores factory tunables broadly.
- Story 28.3 clarified target-version availability remains a live-validation gate; do not treat upstream source docs alone as enough for implementation completion.
- Recent acceptance remediation found some OPNsense APIs vary by appliance/plugin/version. Use explicit endpoint prechecks for missing endpoints rather than letting acceptance fail with generic API errors.

### References

- `_bmad-output/planning-artifacts/tunables-sysctl-research.md` — source evidence, endpoint table, lifecycle classification, safety risks.
- `_bmad-output/implementation-artifacts/28-3-research-system-tunables-sysctl.md` — predecessor story and completion notes.
- `internal/service/system/vlan_resource.go` — existing hand-written system UUID item resource pattern.
- `internal/service/system/vlan_data_source.go` — existing system UUID data source pattern.
- `pkg/opnsense/crud.go` — shared CRUD/reconfigure helpers.
- `internal/acctest/acctest.go` — acceptance helper patterns.
- `test/README.md` — Vagrant acceptance environment and serial runner guidance.

## Dev Agent Record

### Implementation Plan

- Use the existing hand-written System UUID item-resource pattern from VLAN/VIP instead of adding generator or client behavior.
- Manage only persistent `tunable`, `value`, and `descr` fields; ignore volatile `default_value` and `type` as desired state.
- Add UUID data-source parity because live readback by UUID is stable.
- Use `kern.msgbuf_show_timestamp = 1` as the acceptance fixture because it matches the current/default value and lets update coverage change only description text.

### Debug Log References

- Story created after live non-mutating probe confirmed `core/tunables` `get_item`, `search_item`, and `reconfigure` are reachable on the local 25.x appliance with wrapper `sysctl`.
- Story creation intentionally did not create/update/delete tunables because sysctl mutations can affect appliance runtime behavior.
- 2026-06-17: Development started for explicit user-selected Story 28.4; baseline commit preserved from story creation.
- 2026-06-17: Live mutation probe confirmed `add_item`, `get_item`, `set_item`, `del_item`, and `reconfigure`; cleanup verified no probe items remained.
- 2026-06-17: Red phase `go test ./internal/service/system` failed on missing `TunableResourceModel` and `tunableAPIResponse` before implementation.
- 2026-06-17: Added `opnsense_system_tunable` resource/data source, docs/examples, and support-count updates.
- 2026-06-17: Containerized `go generate ./tools` succeeded; host-side `go generate ./tools` failed due Terraform auto-download OpenPGP key expiry, so generated docs were produced in the dev-toolchain container.
- 2026-06-17: Focused acceptance passed: `go test -run TestAccSystemTunable_basic -count=1 -p 1 ./internal/service/system` in the dev-toolchain container against the local Vagrant API.
- 2026-06-17: Targeted package tests passed: `go test ./internal/service/system ./pkg/opnsense`.
- 2026-06-17: Required full gate passed: `make check`.
- 2026-06-17: Code review recorded 4 patch findings and 1 deferred shared-CRUD lifecycle finding; all patch findings were applied.
- 2026-06-17: Post-review validation passed: `go test ./internal/service/system ./pkg/opnsense`, Vagrant runner `test/scripts/run-vagrant-acceptance.sh --package ./internal/service/system`, and `make check`.

### Completion Notes List

- Added `opnsense_system_tunable` resource with UUID CRUD/import using `/api/core/tunables/*_item` and automatic `/api/core/tunables/reconfigure` through shared CRUD helpers.
- Added `opnsense_system_tunable` data source by UUID.
- Kept `default_value` and `type` out of configurable Terraform desired state to avoid volatile metadata diff churn.
- Added acceptance coverage for create/read, idempotent plan, import, update, and delete cleanup using the low-risk `kern.msgbuf_show_timestamp = 1` fixture.
- Updated Registry docs/examples and public planning/support docs from Coming to Supported with 102 resources and 88 data sources.
- Resolved code-review patches by aligning remaining count docs, expanding acceptance endpoint prechecks, adding precise tunable destroy verification, and adding the data-source safety note.

### File List

- `README.md`
- `RELEASE.md`
- `_bmad-output/implementation-artifacts/28-4-implement-system-tunables-resource.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `docs/data-sources/system_tunable.md`
- `docs/index.md`
- `docs/migration-import.md`
- `docs/resources/system_tunable.md`
- `examples/data-sources/opnsense_system_tunable/data-source.tf`
- `examples/resources/opnsense_system_tunable/import.sh`
- `examples/resources/opnsense_system_tunable/resource.tf`
- `internal/service/system/data_source_schema_test.go`
- `internal/service/system/exports.go`
- `internal/service/system/tunable_data_source.go`
- `internal/service/system/tunable_model.go`
- `internal/service/system/tunable_model_test.go`
- `internal/service/system/tunable_resource.go`
- `internal/service/system/tunable_resource_test.go`
- `internal/service/system/tunable_schema.go`
- `templates/data-sources/system_tunable.md.tmpl`
- `templates/index.md.tmpl`
- `templates/resources/system_tunable.md.tmpl`

## Change Log

- 2026-06-17: Created ready-for-dev story for `opnsense_system_tunable` resource implementation.
- 2026-06-17: Implemented `opnsense_system_tunable` resource/data source, docs/examples, live acceptance coverage, support-count updates, and validation.
- 2026-06-17: Applied code-review patch findings and marked story done after Vagrant acceptance and `make check` passed.
