---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 27.3: Implement OSPF Area Resource

Status: done

## Story

As an operator,
I want a Terraform resource for Quagga/FRR OSPF areas,
so that OSPF area configuration can be managed as a discrete UUID-backed Terraform object alongside the existing OSPF sub-resources.

## Acceptance Criteria

1. `opnsense_quagga_ospf_area` is implemented as a UUID item resource unless model extraction proves the area item is not safely representable; any skipped implementation is documented with evidence.
2. The resource follows the existing generated Quagga OSPF item-resource pattern: YAML schema, generated model/schema/resource/test/data-source files, service exports registration, CRUD, UUID import, state read-back, drift detection, delete handling, and `POST /api/quagga/service/reconfigure`.
3. A generated `opnsense_quagga_ospf_area` data source is registered if the generator emits it.
4. Registry docs and examples exist for the resource and generated data source; generated `docs/index.md`, `docs/resources/*`, and `docs/data-sources/*` are refreshed from templates/examples rather than hand-edited divergently.
5. Support matrix, core gap analysis, resource-gap verification handoff, PRD/roadmap counts where current-facing, and README support counts are updated by the exact implemented resource/data-source count.
6. Existing OSPF resources and data sources remain registered and unchanged except for required aggregate count/doc updates.
7. `make check` passes.

## Tasks / Subtasks

- [x] Endpoint and model verification (AC: 1, 2)
  - [x] Confirm current OPNsense Quagga API docs still list `add_area`, `get_area`, `set_area`, `del_area`, `search_area`, and `toggle_area` under `quagga/ospfsettings`.
  - [x] Fetch and review `OSPF.xml` from the FRR plugin source to extract the `areas` model fields, field classes, defaults, and required constraints.
  - [x] Confirm wrapper/monad key for area by reviewing `OspfsettingsController.php`; do not assume the wrapper from endpoint name alone.
  - [x] Decide whether all area fields are safely representable with current generator field types; if not, document the skipped fields or stop before shipping broken drift detection.
- [x] Extend Quagga OSPF code generation input (AC: 1, 2, 3)
  - [x] Add an `ospf_area` item definition to `internal/generate/schemas/quagga_ospf.yaml`.
  - [x] Use `kind: item`, `reconfigure: /api/quagga/service/reconfigure`, and the exact area endpoint set.
  - [x] Use supported generator field types only: `bool`, `int`, `string`, `selectmap`, `selectmaplist`, and `csvset`.
  - [x] Include `test_value` for required or otherwise hard-to-generate fields so generated acceptance tests compile and are meaningful.
- [x] Generate and register implementation artifacts (AC: 2, 3, 6)
  - [x] Run `go run ./internal/generate` after editing YAML.
  - [x] Register `newOSPFAreaResource` in `internal/service/quagga/exports.go`.
  - [x] Register generated `newOSPFAreaDataSource` in `internal/service/quagga/exports.go`.
  - [x] Verify `internal/provider/provider.go` already aggregates `quagga.Resources()` and `quagga.DataSources()`; do not add duplicate provider-level registration.
  - [x] Inspect generated `ospf_area_*` files for endpoint, monad, required field, conversion, import, and reconfigure correctness.
- [x] Add documentation templates and examples (AC: 4)
  - [x] Add `examples/resources/opnsense_quagga_ospf_area/resource.tf` and `import.sh`.
  - [x] Add `templates/resources/quagga_ospf_area.md.tmpl` with subcategory `Quagga / FRR`, example usage, schema markdown, and UUID import instructions.
  - [x] Add `examples/data-sources/opnsense_quagga_ospf_area/data-source.tf` if the data source is generated.
  - [x] Add `templates/data-sources/quagga_ospf_area.md.tmpl` if the data source is generated.
  - [x] Regenerate Registry docs with containerized `go generate ./tools`.
- [x] Update planning and release-facing docs (AC: 5)
  - [x] Update `_bmad-output/planning-artifacts/support-matrix.md` counts and Dynamic routing rows.
  - [x] Update `_bmad-output/planning-artifacts/core-config-gap-analysis.md` OSPF area classification and release matrix counts.
  - [x] Update `_bmad-output/planning-artifacts/resource-gap-verification.md` to mark OSPF area Supported or document why it was not implemented.
  - [x] Update `_bmad-output/planning-artifacts/prd.md`, `_bmad-output/planning-artifacts/feature-complete-roadmap.md`, `README.md`, `templates/index.md.tmpl`, and `docs/index.md` only where they make current-facing support count or OSPF area gap claims.
- [x] Validate implementation (AC: 1-7)
  - [x] Run targeted searches for `opnsense_quagga_ospf_area` across `internal/service/quagga`, `docs`, `examples`, and `templates`.
  - [x] Run focused tests for `./internal/service/quagga ./internal/provider`.
  - [x] Run `make check` and fix failures without suppressing checks.
  - [x] Record exact counts, validation result, and skipped-field evidence if any in this story's Dev Agent Record.

## Dev Notes

### Source Context

Story 27.1 classified OSPF area as Coming because published Quagga API docs expose `quagga/ospfsettings` area CRUD/search/toggle endpoints. Story 27.2 intentionally did not implement OSPF area because it selected the Dnsmasq item-resource batch.

Current endpoint handoff from `_bmad-output/planning-artifacts/resource-gap-verification.md`:

| Operation | Endpoint |
|---|---|
| Add | `POST /api/quagga/ospfsettings/add_area` |
| Get | `GET /api/quagga/ospfsettings/get_area/{uuid?}` |
| Set | `POST /api/quagga/ospfsettings/set_area/{uuid}` |
| Delete | `POST /api/quagga/ospfsettings/del_area/{uuid}` |
| Search | `GET,POST /api/quagga/ospfsettings/search_area` |
| Toggle | `POST /api/quagga/ospfsettings/toggle_area/{uuid}` |
| Reconfigure | `POST /api/quagga/service/reconfigure` |

Toggle support is listed in upstream docs, but existing generated item resources generally model `enabled` as a normal field and do not call toggle endpoints. Follow existing project patterns unless code inspection proves area behaves differently.

### Existing Patterns to Reuse

- Generator input: `internal/generate/schemas/quagga_ospf.yaml`.
- Existing generated OSPF item resources: `internal/service/quagga/ospf_network_*`, `ospf_interface_*`, `ospf_neighbor_*`, `ospf_prefixlist_*`, `ospf_routemap_*`, `ospf_redistribution_*`.
- Existing Quagga exports: `internal/service/quagga/exports.go`.
- Existing resource docs/examples: `templates/resources/quagga_ospf_network.md.tmpl` if present, otherwise use comparable Quagga templates such as `templates/resources/quagga_static_route.md.tmpl`.
- Generated data-source pattern: `internal/service/quagga/*_data_source.gen.go` and `templates/data-sources/quagga_static_route.md.tmpl`.

### Guardrails

- Do not hand-edit generated `.gen.go` files; edit YAML and regenerate.
- Do not add provider-level registration in `internal/provider/provider.go` unless package aggregation changes.
- Preserve existing OSPF resources and data sources.
- Do not update historical v0.1.0 counts unless the text is explicitly current-facing.
- If the area model has field classes or wrapper behavior that the generator cannot represent safely, stop and document the evidence rather than guessing.

### References

- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `internal/generate/schemas/quagga_ospf.yaml`
- `internal/service/quagga/exports.go`
- OPNsense Quagga API docs: `https://docs.opnsense.org/development/api/plugins/quagga.html`
- OPNsense FRR OSPF model: `https://github.com/opnsense/plugins/blob/master/net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/OSPF.xml`

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- Created from post-Story 27.2 follow-up request on 2026-06-05.
- Loaded resource-gap verification, support matrix, core gap analysis, post-release epic plan, and Quagga OSPF generator schema.
- Fetched current OPNsense Quagga API docs and confirmed area endpoints are listed under `quagga/ospfsettings`.
- Fetched upstream `OSPF.xml`; `areas.area` has representable fields `enabled` (`BooleanField`, default `1`, required), `id` (`NetworkField`, required IPv4 dotted area ID, unique), and `type` (`OptionField`, required, default `stub`, options `stub`, `stub-no-summary`, `nssa`, `nssa-no-summary`).
- Fetched upstream `OspfsettingsController.php`; area CRUD/search/toggle uses model path `areas.area` and request/response wrapper `area`.
- Generated provider code with `go run ./internal/generate`; output: `generated 44 resources`.
- Generated Registry docs with containerized `go generate ./tools` using `ghcr.io/devrail-dev/dev-toolchain:1.12.0`.
- Targeted searches confirmed `opnsense_quagga_ospf_area` across `internal/service/quagga`, `docs`, `examples`, and `templates`.
- Focused tests passed: `go test ./internal/service/quagga ./internal/provider`.
- Full validation passed: `make check`.
- Code review found one valid OSPF area option-value issue; changed no-summary type options to OPNsense XML `value` strings `stub no-summary` and `nssa no-summary`, regenerated code/docs, and reran focused tests plus `make check`.

### Completion Notes List

- Created comprehensive developer guide for implementing `opnsense_quagga_ospf_area`.
- Implemented `opnsense_quagga_ospf_area` as a generated UUID item resource and generated data source with wrapper/monad `area`, Quagga OSPF area endpoints, and `/api/quagga/service/reconfigure`.
- Registered `newOSPFAreaResource` and `newOSPFAreaDataSource` in Quagga exports; provider-level registration already aggregates Quagga package resources/data sources.
- Added resource and data-source examples/templates and generated Registry docs.
- Updated current-facing counts to 97 resources, 83 data sources, 97 resource docs, and 83 data-source docs; OSPF area is now Supported.
- No OSPF area fields were skipped; all upstream area fields are safely represented by supported generator field types.
- Code review patch applied: no-summary OSPF area type values now match OPNsense API-selected values with spaces rather than XML element names with hyphens.

### File List

- `_bmad-output/implementation-artifacts/27-3-implement-ospf-area-resource.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `README.md`
- `docs/data-sources/quagga_ospf_area.md`
- `docs/index.md`
- `docs/resources/quagga_ospf_area.md`
- `examples/data-sources/opnsense_quagga_ospf_area/data-source.tf`
- `examples/resources/opnsense_quagga_ospf_area/import.sh`
- `examples/resources/opnsense_quagga_ospf_area/resource.tf`
- `internal/generate/schemas/quagga_ospf.yaml`
- `internal/service/quagga/data_source_schema_test.go`
- `internal/service/quagga/exports.go`
- `internal/service/quagga/ospf_area_data_source.gen.go`
- `internal/service/quagga/ospf_area_model.gen.go`
- `internal/service/quagga/ospf_area_resource.gen.go`
- `internal/service/quagga/ospf_area_resource_gen_test.go`
- `internal/service/quagga/ospf_area_schema.gen.go`
- `templates/data-sources/quagga_ospf_area.md.tmpl`
- `templates/index.md.tmpl`
- `templates/resources/quagga_ospf_area.md.tmpl`

### Change Log

- 2026-06-05: Created story for OSPF area implementation.
- 2026-06-09: Implemented generated OSPF area resource/data source, docs/examples, current-facing count updates, and validation.
- 2026-06-09: Addressed code review finding for OSPF area no-summary type option values and revalidated with focused tests and `make check`.

## Senior Developer Review (AI)

### Review Date

2026-06-09

### Review Outcome

Approve after patch.

### Findings

- [x] Medium: `internal/generate/schemas/quagga_ospf.yaml` used XML element names `stub-no-summary` and `nssa-no-summary` instead of OPNsense `OptionField` `value` strings `stub no-summary` and `nssa no-summary`. Fixed in the generator schema, regenerated `ospf_area_schema.gen.go`, regenerated Registry docs, and reran validation.

### Validation

- `go run ./internal/generate` passed; output: `generated 44 resources`.
- Containerized `go generate ./tools` passed.
- `go test ./internal/service/quagga ./internal/provider` passed.
- `make check` passed after review patch.
