---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 27.2: Implement Dnsmasq Item Resource Gap

Status: done

## Story

As an operator,
I want Terraform resources for Dnsmasq host, domain override, DHCP tag, range, option, and boot entries,
so that Dnsmasq DNS/DHCP configuration can be managed as discrete UUID-backed Terraform objects instead of only through singleton service settings.

## Acceptance Criteria

1. The selected Story 27.1 candidate is the Dnsmasq item-resource gap because it provides the highest confirmed endpoint coverage without hardware-specific live-validation requirements.
2. `opnsense_dnsmasq_host`, `opnsense_dnsmasq_domain`, `opnsense_dnsmasq_tag`, `opnsense_dnsmasq_range`, `opnsense_dnsmasq_option`, and `opnsense_dnsmasq_boot` resources are implemented unless model extraction proves one item family is not safely representable; any skipped family is documented with evidence.
3. Each implemented Dnsmasq item follows the existing generated item-resource pattern: YAML schema, generated model/schema/resource/test/data-source files, service exports registration, CRUD, UUID import, state read-back, drift detection, delete handling, and `POST /api/dnsmasq/service/reconfigure`.
4. Dnsmasq item data sources are registered for every implemented UUID item resource when the generator emits them.
5. Registry docs and examples exist for every implemented resource and generated data source; generated `docs/index.md`, `docs/resources/*`, and `docs/data-sources/*` are refreshed from templates/examples rather than hand-edited divergently.
6. Support matrix, core gap analysis, resource-gap verification handoff, PRD/roadmap counts where current-facing, and README support counts are updated to reflect the exact implemented resource/data-source count.
7. Source NAT and Unbound forward are not duplicated; Kea DHCPv4 option/DDNS are not implemented in this story because they remain `Needs research` pending live endpoint recheck.
8. `make check` passes.

## Tasks / Subtasks

- [x] Selection and preflight verification (AC: 1, 2, 7)
  - [x] Confirm current OPNsense Dnsmasq API page still lists `add/get/set/del/search` for `host`, `domain`, `tag`, `range`, `option`, and `boot` under `/api/dnsmasq/settings`.
  - [x] Confirm `Dnsmasq.xml` model arrays and wrapper names map to endpoint item names: `hosts` -> `host`, `domainoverrides` -> `domain`, `dhcp_tags` -> `tag`, `dhcp_ranges` -> `range`, `dhcp_options` -> `option`, `dhcp_boot` -> `boot`.
  - [x] Decide whether all six item families are safely representable with current generator field types; if not, document the skipped family and reason before implementing the rest.
  - [x] Do not select LAGG, OSPF area, HASync, source NAT, Unbound forward, or Kea items in this story.
- [x] Extend Dnsmasq code generation input (AC: 2, 3, 4)
  - [x] Add item resource definitions to `internal/generate/schemas/dnsmasq.yaml` for `host`, `domain`, `tag`, `range`, `option`, and `boot`.
  - [x] Use `kind: item`, `reconfigure: /api/dnsmasq/service/reconfigure`, and the exact endpoint set for each implemented item.
  - [x] Use wrapper/monad keys matching endpoint item names: `host`, `domain`, `tag`, `range`, `option`, `boot`.
  - [x] Use field types supported by `internal/generate/main.go`: `bool`, `int`, `string`, `selectmap`, `selectmaplist`, and `csvset`; extend the generator only if a required model field cannot be represented correctly and the extension is minimal/reusable.
  - [x] Include `test_value` for required or otherwise hard-to-generate fields so generated acceptance tests compile and are meaningful.
- [x] Generate and register implementation artifacts (AC: 3, 4)
  - [x] Run code generation with `go run ./internal/generate` after editing YAML.
  - [x] Register all new `new<GoType>Resource` constructors in `internal/service/dnsmasq/exports.go`.
  - [x] Register all generated `new<GoType>DataSource` constructors in `internal/service/dnsmasq/exports.go`.
  - [x] Verify `internal/provider/provider.go` already aggregates `dnsmasq.Resources()` and `dnsmasq.DataSources()`; do not add duplicate provider-level registration.
  - [x] Inspect generated `*_resource.gen.go`, `*_model.gen.go`, `*_schema.gen.go`, `*_data_source.gen.go`, and `*_resource_gen_test.go` for endpoint, monad, required field, and conversion correctness.
- [x] Add documentation templates and examples (AC: 5)
  - [x] Add `examples/resources/opnsense_dnsmasq_<item>/resource.tf` and `import.sh` for each implemented item.
  - [x] Add `templates/resources/dnsmasq_<item>.md.tmpl` with subcategory `Dnsmasq`, example usage, schema markdown, and UUID import instructions.
  - [x] Add `examples/data-sources/opnsense_dnsmasq_<item>/data-source.tf` for each generated data source.
  - [x] Add `templates/data-sources/dnsmasq_<item>.md.tmpl` with subcategory `Dnsmasq` and UUID lookup example.
  - [x] Regenerate Registry docs with `go generate ./tools` and verify generated `docs/resources/dnsmasq_*.md` and `docs/data-sources/dnsmasq_*.md` include examples and import guidance where applicable.
- [x] Update planning and release-facing docs (AC: 6, 7)
  - [x] Update `_bmad-output/planning-artifacts/support-matrix.md` supported Dnsmasq rows and counts.
  - [x] Update `_bmad-output/planning-artifacts/core-config-gap-analysis.md` Dnsmasq status and release matrix counts.
  - [x] Update `_bmad-output/planning-artifacts/resource-gap-verification.md` to mark implemented Dnsmasq item families as Supported and leave only any skipped family as Coming/Needs research with evidence.
  - [x] Update `_bmad-output/planning-artifacts/prd.md`, `_bmad-output/planning-artifacts/feature-complete-roadmap.md`, `README.md`, `templates/index.md.tmpl`, and `docs/index.md` only where they make current-facing support count or Dnsmasq gap claims.
  - [x] Keep source NAT and Unbound forward classified as Supported under existing resource names; keep Kea DHCPv4 option/DDNS as Needs research.
- [x] Validate implementation (AC: 3, 4, 5, 8)
  - [x] Run targeted searches to confirm all implemented `opnsense_dnsmasq_*` resources and data sources are registered and documented.
  - [x] Run `make check` and fix failures without suppressing checks.
  - [x] Record exact implemented item families, final counts, validation result, and any skipped item evidence in this story's Dev Agent Record.

### Review Findings

- [x] [Review][Patch] Fix Dnsmasq domain override wrapper/monad mismatch [internal/generate/schemas/dnsmasq.yaml:75]
- [x] [Review][Patch] Add safeguards for Dnsmasq option type-specific fields that upstream clears [internal/generate/schemas/dnsmasq.yaml:154]
- [x] [Review][Patch] Correct Dnsmasq support-matrix rows so resources are not duplicated and data sources are listed [support-matrix.md:38]
- [x] [Review][Patch] Align current-facing core gap notes for HAProxy and ACME data-source support [core-config-gap-analysis.md:159]

## Dev Notes

### Selected Target

Implement the Dnsmasq item-resource gap from Story 27.1. This is the preferred Story 27.2 target because it has confirmed published API endpoint coverage and does not require special hardware or ambiguous durable-resource semantics.

The Dnsmasq batch is a resource gap, not a duplicate of `opnsense_dnsmasq_settings`. The existing singleton manages service-wide Dnsmasq settings only. The new resources manage UUID-backed array items under the same Dnsmasq model.

### Endpoint Contract

Current OPNsense API docs show Dnsmasq as a **Core API** page, not a plugin API page: `https://docs.opnsense.org/development/api/core/dnsmasq.html`.

Use these endpoint families:

| Terraform resource | Monad | Add | Get | Set | Delete | Search |
|---|---|---|---|---|---|
| `opnsense_dnsmasq_host` | `host` | `/api/dnsmasq/settings/add_host` | `/api/dnsmasq/settings/get_host` | `/api/dnsmasq/settings/set_host` | `/api/dnsmasq/settings/del_host` | `/api/dnsmasq/settings/search_host` |
| `opnsense_dnsmasq_domain` | `domain` | `/api/dnsmasq/settings/add_domain` | `/api/dnsmasq/settings/get_domain` | `/api/dnsmasq/settings/set_domain` | `/api/dnsmasq/settings/del_domain` | `/api/dnsmasq/settings/search_domain` |
| `opnsense_dnsmasq_tag` | `tag` | `/api/dnsmasq/settings/add_tag` | `/api/dnsmasq/settings/get_tag` | `/api/dnsmasq/settings/set_tag` | `/api/dnsmasq/settings/del_tag` | `/api/dnsmasq/settings/search_tag` |
| `opnsense_dnsmasq_range` | `range` | `/api/dnsmasq/settings/add_range` | `/api/dnsmasq/settings/get_range` | `/api/dnsmasq/settings/set_range` | `/api/dnsmasq/settings/del_range` | `/api/dnsmasq/settings/search_range` |
| `opnsense_dnsmasq_option` | `option` | `/api/dnsmasq/settings/add_option` | `/api/dnsmasq/settings/get_option` | `/api/dnsmasq/settings/set_option` | `/api/dnsmasq/settings/del_option` | `/api/dnsmasq/settings/search_option` |
| `opnsense_dnsmasq_boot` | `boot` | `/api/dnsmasq/settings/add_boot` | `/api/dnsmasq/settings/get_boot` | `/api/dnsmasq/settings/set_boot` | `/api/dnsmasq/settings/del_boot` | `/api/dnsmasq/settings/search_boot` |

All implemented items must use `POST /api/dnsmasq/service/reconfigure` after mutations.

### Model Field Extraction

Use `Dnsmasq.xml` as the model source: `https://github.com/opnsense/core/blob/master/src/opnsense/mvc/app/models/OPNsense/Dnsmasq/Dnsmasq.xml`.

Known array-to-resource mappings from current model XML:

| Endpoint item | Model array | Notable fields |
|---|---|---|
| `host` | `hosts` | `host`, `domain`, `local`, `ip`, `cnames`, `client_id`, `hwaddr`, `lease_time`, `ignore`, `set_tag`, `descr`, `comments`, `aliases` |
| `domain` | `domainoverrides` | `sequence`, `domain`, `ipset`, `srcip`, `port`, `ip`, `descr` |
| `tag` | `dhcp_tags` | `tag` |
| `range` | `dhcp_ranges` | `interface`, `set_tag`, `start_addr`, `end_addr`, `subnet_mask`, `constructor`, `mode`, `prefix_len`, `lease_time`, `domain_type`, `domain`, `nosync`, `ra_mode`, `ra_priority`, `ra_mtu`, `ra_interval`, `ra_router_lifetime`, `description` |
| `option` | `dhcp_options` | `type`, `option`, `option6`, `interface`, `tag`, `set_tag`, `value`, `force`, `description` |
| `boot` | `dhcp_boot` | `interface`, `tag`, `filename`, `servername`, `address`, `description` |

The generator does not understand every XML field class directly. Map fields to existing generator types only when the API response shape supports it:

- `BooleanField` -> `bool`.
- `IntegerField`, `AutoNumberField`, `PortField` -> `int` unless the API echoes non-integer strings; use `string` if uncertain.
- `OptionField` and single `ModelRelationField` -> `selectmap` when the API response is a selected-map object.
- `InterfaceField` or `ModelRelationField` with `Multiple` -> `selectmaplist` when the API response is selected-map-list; otherwise use `csvset` only for plain CSV responses.
- `TextField`, `HostnameField`, `NetworkField`, `MacAddressField`, `JsonKeyValueStoreField`, custom dotted fields, and uncertain fields -> start as `string` unless live/API inspection proves selected-map semantics.

If a required field's API shape is too uncertain for the generator, prefer implementing a smaller safe subset for that item or skipping that item with evidence instead of guessing and shipping broken drift detection.

### Generator and Registration Requirements

Edit `internal/generate/schemas/dnsmasq.yaml`; do not hand-edit generated `.gen.go` files.

Use the two generation steps deliberately:

1. `go run ./internal/generate` refreshes generated Go model/schema/resource/test/data-source files from YAML.
2. `go generate ./tools` refreshes Terraform Registry docs from provider schemas, templates, and examples.

Run these through the existing containerized toolchain when doing implementation work; do not install tools on the host.

The generator currently emits these files for each `kind: item` resource:

- `internal/service/dnsmasq/<name>_model.gen.go`
- `internal/service/dnsmasq/<name>_schema.gen.go`
- `internal/service/dnsmasq/<name>_resource.gen.go`
- `internal/service/dnsmasq/<name>_resource_gen_test.go`
- `internal/service/dnsmasq/<name>_data_source.gen.go`

Register generated constructors in `internal/service/dnsmasq/exports.go`. Current file only registers `newSettingsResource` and no data sources.

`internal/provider/provider.go` already appends `dnsmasq.Resources()` and `dnsmasq.DataSources()`. Do not modify provider-level aggregation unless the package-level API changes.

### Existing Patterns to Reuse

- Existing Dnsmasq singleton: `internal/service/dnsmasq/settings_*` and `internal/generate/schemas/dnsmasq.yaml`.
- Generated item resource pattern: `internal/service/quagga/ospf_network_*` and `internal/generate/schemas/quagga_ospf.yaml`.
- Generated item data source pattern: `internal/service/quagga/ospf_network_data_source.gen.go`.
- Generated acceptance-test pattern: `internal/service/quagga/ospf_network_resource_gen_test.go`.
- Resource doc template with import: `templates/resources/firewall_filter_rule.md.tmpl`.
- Data-source doc template and UUID example: `templates/data-sources/quagga_static_route.md.tmpl` and `examples/data-sources/opnsense_quagga_static_route/data-source.tf`.

### Documentation and Count Updates

If all six Dnsmasq item families are implemented, expected support counts become:

| Capability | Before | After |
|---|---:|---:|
| Resources | 90 | 96 |
| Data sources | 76 | 82 |
| Resource docs | 90 | 96 |
| Data-source docs | 76 | 82 |

If any family is skipped, update counts by the exact number implemented. Because generated item resources also emit data sources, each implemented Dnsmasq item should add one resource, one data source, one resource doc, and one data-source doc after registration/docs generation.

Update these current-facing artifacts when counts or Dnsmasq classifications change:

- `_bmad-output/planning-artifacts/support-matrix.md`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `README.md`
- `templates/index.md.tmpl`
- `docs/index.md`

Historical release artifacts may retain v0.1.0 counts if clearly scoped as historical.

### Scope Boundaries and Anti-Duplication Guardrails

Do not implement these in Story 27.2:

- `opnsense_firewall_source_nat` or similar duplicate source NAT resource; source NAT is already supported as `opnsense_firewall_nat_outbound`.
- A separate Unbound forward resource; Unbound forward is already supported as `opnsense_unbound_domain_override`.
- Kea DHCPv4 option or Kea DDNS; both remain `Needs research` because published docs conflict with live endpoint-not-found evidence.
- LAGG; it still needs live appliance validation with assignable member interfaces.
- HASync status/actions; those are operational/data-source/action candidates, not durable resource CRUD.
- Tunables/sysctl; still needs endpoint lifecycle research.

OSPF area and HASync configuration remain valid follow-up candidates after this story.

### Testing Requirements

- Required final command: `make check`.
- The generated acceptance tests are guarded by `acctest.PreCheck(t)` and only run when acceptance-test environment variables are present; keep them compiling and structurally meaningful.
- For generated item tests, include enough `test_value` fields in YAML for valid minimum HCL. Avoid values that are likely to mutate production-like services destructively.
- Run targeted checks before `make check`, for example:

```bash
rg "new(Dnsmasq|DNSMasq|DNSMASQ)|new.*Dnsmasq|new.*DNS" internal/service/dnsmasq/exports.go
rg "opnsense_dnsmasq_(host|domain|tag|range|option|boot)" internal/service/dnsmasq docs examples templates
rg "Resources \| 9[0-9]|Data sources \| 8[0-9]|Dnsmasq" _bmad-output/planning-artifacts README.md templates/index.md.tmpl docs/index.md
```

Interpret generated docs/templates carefully: `docs/` should be regenerated, not manually diverged from `templates/`.

### Previous Story Intelligence

Story 27.1 completed endpoint verification and review patches:

- Dnsmasq host/domain/range/option/tag were classified as Coming with published API support.
- Code review found Dnsmasq `boot` endpoints were also exposed; current OPNsense docs confirm `add/get/set/del/search_boot`.
- Source NAT and Unbound forward were reclassified as already Supported; do not duplicate them.
- Kea DHCPv4 option and Kea DDNS were reclassified to Needs research because live probing returned endpoint-not-found.
- Story 27.1 added Dnsmasq boot and tunables/sysctl as follow-up verification items.
- `make check` passed after Story 27.1 review patches.

Story 26.1-26.3 docs hardening completed:

- Provider index now includes support matrix and import/migration guidance.
- Full migration/import guide exists at `docs/migration-import.md`.
- Registry docs are generated from templates/examples; do not hand-edit generated docs without matching template/example changes.

### Architecture Compliance

- Terraform Plugin Framework v6 patterns apply throughout.
- CRUD operations must use `opnsense.Add`, `opnsense.Get`, `opnsense.Update`, and `opnsense.Delete` through generated `ReqOpts`.
- OPNsense mutation requests require the correct request body wrapper key (`Monad`).
- Every mutation must trigger the standard Dnsmasq reconfigure endpoint.
- Reads must populate state from API read-back, never by echoing the Terraform plan.
- UUID import must use `resource.ImportStatePassthroughID` for item resources.
- Missing UUID reads should remove resource state via existing generated `NotFoundError` handling.
- Generated data sources should read by UUID and expose computed attributes only.
- All tools must run through the existing containerized Makefile/toolchain; do not install host tools.

### References

- `_bmad-output/planning-artifacts/resource-gap-verification.md` — Story 27.1 endpoint and candidate handoff.
- `_bmad-output/planning-artifacts/support-matrix.md` — current release-positioning source of truth.
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md` — current gap classifications.
- `_bmad-output/planning-artifacts/prd.md` — current implementation baseline and post-MVP follow-up.
- `_bmad-output/planning-artifacts/post-release-epics.md` — Epic 27 sequence.
- `_bmad-output/implementation-artifacts/27-1-verify-buildable-resource-gaps.md` — previous-story intelligence and review patches.
- `internal/generate/schemas/dnsmasq.yaml` — Dnsmasq generator input to edit.
- `internal/generate/main.go` and `internal/generate/templates.go` — generator capabilities and emitted file patterns.
- `internal/service/dnsmasq/exports.go` — Dnsmasq resource/data-source registration.
- `internal/service/quagga/ospf_network_*` — generated item resource/data-source/test reference pattern.
- `templates/resources/firewall_filter_rule.md.tmpl` — resource doc template with import example.
- `templates/data-sources/quagga_static_route.md.tmpl` — data-source doc template with UUID lookup example.
- OPNsense Dnsmasq API docs: `https://docs.opnsense.org/development/api/core/dnsmasq.html`.
- OPNsense Dnsmasq model: `https://github.com/opnsense/core/blob/master/src/opnsense/mvc/app/models/OPNsense/Dnsmasq/Dnsmasq.xml`.

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- Ultimate context engine analysis completed for Story 27.2 on 2026-06-04.
- Loaded sprint tracker, Story 27.1, resource-gap verification, support matrix, core gap analysis, PRD, architecture, generator code, Dnsmasq singleton implementation, and generated Quagga item-resource patterns.
- Fetched OPNsense API index and Dnsmasq API page; confirmed Dnsmasq is a core API and that `boot` endpoints are listed with host/domain/range/option/tag endpoints.
- Fetched current `Dnsmasq.xml` model and extracted model arrays/fields for host/domain/tag/range/option/boot.
- Added Dnsmasq item schema definitions in `internal/generate/schemas/dnsmasq.yaml` and ran `go run ./internal/generate`; generator reported `generated 43 resources`.
- Ran containerized `go generate ./tools` to refresh Registry docs from templates/examples after Dnsmasq template/example additions.
- Targeted searches confirmed all six `opnsense_dnsmasq_*` resource/data-source constructors are registered, provider aggregation already includes `dnsmasq.Resources()` and `dnsmasq.DataSources()`, generated docs include examples, and resource docs include UUID import guidance.
- Ran `make check`; result passed for lint, format, test, security, scan, and docs.
- Code review found four patch findings; applied all four and reran `make check` successfully.

### Completion Notes List

- Rewrote Story 27.2 from a thin ready-for-dev stub into a comprehensive developer guide.
- Selected Dnsmasq item-resource batch as the implementation target and included `boot` based on current API docs.
- Added endpoint table, model field extraction, generator constraints, docs/examples requirements, count-update requirements, and anti-duplication guardrails.
- Implemented all six selected Dnsmasq UUID item families: host, domain, tag, range, option, and boot. No item family was skipped.
- Registered generated resources and data sources for each implemented Dnsmasq item family.
- Added Dnsmasq resource examples, import examples, data-source examples, Registry templates, and regenerated docs.
- Updated current-facing support counts from 90 resources / 76 data sources to 96 resources / 82 data sources, with matching 96 resource docs / 82 data-source docs.
- Kept source NAT and Unbound forward classified as Supported under existing resource names; Kea DHCPv4 option and Kea DDNS remain Needs research.
- Final validation passed with `make check`.
- Resolved code-review findings by changing the Dnsmasq domain wrapper to `domainoverride`, adding Dnsmasq option type-field config validation, fixing Dnsmasq support-matrix rows, and aligning HAProxy/ACME data-source notes.
- Post-review validation passed with `make check`.

### File List

- `README.md`
- `_bmad-output/implementation-artifacts/27-2-implement-highest-value-verified-gap.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `docs/data-sources/dnsmasq_boot.md`
- `docs/data-sources/dnsmasq_domain.md`
- `docs/data-sources/dnsmasq_host.md`
- `docs/data-sources/dnsmasq_option.md`
- `docs/data-sources/dnsmasq_range.md`
- `docs/data-sources/dnsmasq_tag.md`
- `docs/index.md`
- `docs/resources/dnsmasq_boot.md`
- `docs/resources/dnsmasq_domain.md`
- `docs/resources/dnsmasq_host.md`
- `docs/resources/dnsmasq_option.md`
- `docs/resources/dnsmasq_range.md`
- `docs/resources/dnsmasq_tag.md`
- `examples/data-sources/opnsense_dnsmasq_boot/data-source.tf`
- `examples/data-sources/opnsense_dnsmasq_domain/data-source.tf`
- `examples/data-sources/opnsense_dnsmasq_host/data-source.tf`
- `examples/data-sources/opnsense_dnsmasq_option/data-source.tf`
- `examples/data-sources/opnsense_dnsmasq_range/data-source.tf`
- `examples/data-sources/opnsense_dnsmasq_tag/data-source.tf`
- `examples/resources/opnsense_dnsmasq_boot/import.sh`
- `examples/resources/opnsense_dnsmasq_boot/resource.tf`
- `examples/resources/opnsense_dnsmasq_domain/import.sh`
- `examples/resources/opnsense_dnsmasq_domain/resource.tf`
- `examples/resources/opnsense_dnsmasq_host/import.sh`
- `examples/resources/opnsense_dnsmasq_host/resource.tf`
- `examples/resources/opnsense_dnsmasq_option/import.sh`
- `examples/resources/opnsense_dnsmasq_option/resource.tf`
- `examples/resources/opnsense_dnsmasq_range/import.sh`
- `examples/resources/opnsense_dnsmasq_range/resource.tf`
- `examples/resources/opnsense_dnsmasq_tag/import.sh`
- `examples/resources/opnsense_dnsmasq_tag/resource.tf`
- `internal/generate/schemas/dnsmasq.yaml`
- `internal/service/dnsmasq/boot_data_source.gen.go`
- `internal/service/dnsmasq/boot_model.gen.go`
- `internal/service/dnsmasq/boot_resource.gen.go`
- `internal/service/dnsmasq/boot_resource_gen_test.go`
- `internal/service/dnsmasq/boot_schema.gen.go`
- `internal/service/dnsmasq/domain_data_source.gen.go`
- `internal/service/dnsmasq/domain_model.gen.go`
- `internal/service/dnsmasq/domain_resource.gen.go`
- `internal/service/dnsmasq/domain_resource_gen_test.go`
- `internal/service/dnsmasq/domain_schema.gen.go`
- `internal/service/dnsmasq/exports.go`
- `internal/service/dnsmasq/host_data_source.gen.go`
- `internal/service/dnsmasq/host_model.gen.go`
- `internal/service/dnsmasq/host_resource.gen.go`
- `internal/service/dnsmasq/host_resource_gen_test.go`
- `internal/service/dnsmasq/host_schema.gen.go`
- `internal/service/dnsmasq/option_data_source.gen.go`
- `internal/service/dnsmasq/option_model.gen.go`
- `internal/service/dnsmasq/option_resource.gen.go`
- `internal/service/dnsmasq/option_resource_gen_test.go`
- `internal/service/dnsmasq/option_schema.gen.go`
- `internal/service/dnsmasq/option_validators.go`
- `internal/service/dnsmasq/range_data_source.gen.go`
- `internal/service/dnsmasq/range_model.gen.go`
- `internal/service/dnsmasq/range_resource.gen.go`
- `internal/service/dnsmasq/range_resource_gen_test.go`
- `internal/service/dnsmasq/range_schema.gen.go`
- `internal/service/dnsmasq/tag_data_source.gen.go`
- `internal/service/dnsmasq/tag_model.gen.go`
- `internal/service/dnsmasq/tag_resource.gen.go`
- `internal/service/dnsmasq/tag_resource_gen_test.go`
- `internal/service/dnsmasq/tag_schema.gen.go`
- `templates/data-sources/dnsmasq_boot.md.tmpl`
- `templates/data-sources/dnsmasq_domain.md.tmpl`
- `templates/data-sources/dnsmasq_host.md.tmpl`
- `templates/data-sources/dnsmasq_option.md.tmpl`
- `templates/data-sources/dnsmasq_range.md.tmpl`
- `templates/data-sources/dnsmasq_tag.md.tmpl`
- `templates/index.md.tmpl`
- `templates/resources/dnsmasq_boot.md.tmpl`
- `templates/resources/dnsmasq_domain.md.tmpl`
- `templates/resources/dnsmasq_host.md.tmpl`
- `templates/resources/dnsmasq_option.md.tmpl`
- `templates/resources/dnsmasq_range.md.tmpl`
- `templates/resources/dnsmasq_tag.md.tmpl`

### Change Log

- 2026-06-04: Implemented Story 27.2 Dnsmasq item-resource batch; added six resources, six data sources, docs/examples/templates, count updates, and passed `make check`.
- 2026-06-05: Addressed code-review patch findings, reran `make check`, and marked story done.
