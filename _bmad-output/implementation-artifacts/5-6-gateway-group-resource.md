---
baseline_commit: fbaf085e8287ae4f00f786484cd1ab622d77716a
---

# Story 5.6: Gateway Group Resource API Revalidation Gate

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a provider maintainer,
I want to revalidate whether OPNsense now exposes a stable API for gateway groups,
so that we either unblock a safe future `opnsense_system_gateway_group` implementation or keep the upstream-blocked register honest with current evidence.

## Acceptance Criteria

1. **Given** the current target OPNsense release and upstream source state
   **When** the developer investigates gateway-group support
   **Then** the story records whether a stable upstream API exists, including endpoint paths, controller/model names, wrapper keys, required fields, defaults, validation constraints, tier/member representation, reconfigure/apply behavior, and release/version evidence

2. **And** if no stable target-release gateway-group API exists, the developer must **not** implement `opnsense_system_gateway_group`; instead they update the relevant planning/public docs with the latest blocked evidence

3. **And** if a stable target-release API has shipped, the developer creates a follow-up implementation story or updates this story with exact implementation scope before writing provider code

4. **And** the investigation explicitly compares existing provider coverage so no duplicate resource is created for already-supported routing primitives: `opnsense_system_gateway`, `opnsense_system_route`, dynamic routing resources under `opnsense_quagga_*`, and firewall NAT/routing resources

5. **And** all touched docs/status files remain aligned: `docs/upstream-blocked.md`, `docs/migration-import.md`, `docs/index.md` / `templates/index.md.tmpl`, `_bmad-output/planning-artifacts/support-matrix.md`, `_bmad-output/planning-artifacts/core-config-gap-analysis.md`, `_bmad-output/planning-artifacts/feature-complete-roadmap.md`, `_bmad-output/planning-artifacts/prd.md`, and `_bmad-output/implementation-artifacts/sprint-status.yaml`

6. **And** `make check` passes before completion

## Tasks / Subtasks

- [x] Task 1: Revalidate upstream API availability (AC: #1, #2, #3)
  - [x] 1.1 Check published OPNsense API docs for any gateway-group controller/endpoints under `routes`, `routing`, or related modules
  - [x] 1.2 Check current OPNsense `master` and target `stable/26.1` source for gateway-group controllers, models, forms, ACL/menu entries, and configd actions
  - [x] 1.3 If endpoints appear, capture HTTP methods, paths, wrapper/monad keys, model root, ID semantics, request/response payload shape, tier/member representation, validation constraints, defaults, and reconfigure/apply behavior
  - [x] 1.4 If endpoints are absent or not in the target release, record the blocker and stop before provider implementation
- [x] Task 2: Compare existing provider coverage to prevent duplicate work (AC: #4)
  - [x] 2.1 Confirm `internal/service/system` already owns `opnsense_system_gateway` and `opnsense_system_route`
  - [x] 2.2 Confirm gateway fields such as `weight` and `priority` already live on `opnsense_system_gateway` and are not a substitute for a gateway-group resource
  - [x] 2.3 Confirm this story is about multi-WAN failover/load-balancing gateway groups only, not static routes, individual gateways, FRR/BGP, or firewall NAT
- [x] Task 3: Update public/planning docs based on evidence (AC: #2, #5)
  - [x] 3.1 If still blocked, refresh blocked-register/support docs with current target-release and upstream `master` evidence
  - [x] 3.2 If unblocked, update docs to move gateway groups from Upstream-blocked to Coming and create precise implementation handoff notes
- [x] Task 4: Update sprint/planning status consistently (AC: #5)
  - [x] 4.1 If still blocked, mark this story done as an upstream-blocked revalidation record and leave implementation work uncreated
  - [x] 4.2 If unblocked, create or update a follow-up implementation story before marking this story complete
- [x] Task 5: Run `make check` (AC: #6)

### Review Findings

- [x] [Review][Patch] Missing GatewayGroups field semantics in revalidation record [_bmad-output/implementation-artifacts/5-6-gateway-group-resource.md:70]
- [x] [Review][Patch] Remaining provider-owned resource gap count is inconsistent [_bmad-output/planning-artifacts/core-config-gap-analysis.md:31]
- [x] [Review][Defer] Stale generated header timestamp in sprint status [_bmad-output/implementation-artifacts/sprint-status.yaml:2] — deferred, pre-existing
- [x] [Review][Defer] Generated provider index import guidance is stale versus migration guide [templates/index.md.tmpl:113] — deferred, pre-existing
- [x] [Review][Defer] Blocked interface story marked done while remaining upstream-blocked [_bmad-output/implementation-artifacts/sprint-status.yaml:84] — deferred, pre-existing

## Dev Notes

### Current Classification

This story is a **revalidation gate**, not a build-now resource story. The original Epic 5 requirement asked for `opnsense_system_gateway_group`, but current planning artifacts classify gateway groups as **Upstream-blocked** because no usable stable target-release API is documented or implemented.

Do **not** implement provider code against legacy config XML, UI-only forms, private helpers, local patches, or model-only source evidence. Terraform cannot safely manage durable gateway groups without upstream-supported API semantics.

### Current Upstream Evidence Snapshot

Evidence checked during story creation on 2026-06-14:

| Source | Evidence |
|---|---|
| Published API docs: `development/api/core/routing.html` | Lists `routing/settings` gateway endpoints only: `add_gateway`, `del_gateway`, `get`, `get_gateway`, `reconfigure`, `search_gateway`, `set`, `set_gateway`, `toggle_gateway`; uses `OPNsense/Routing/Gateways.xml`. No gateway-group endpoint is listed. |
| Published API docs: `development/api/core/routes.html` | Lists static route endpoints and gateway status only. No gateway-group endpoint is listed. |
| OPNsense `master`: `OPNsense/Routing/Api/SettingsController.php` | Implements individual gateway CRUD/search/reconfigure only. Delete logic references gateway-group membership via `GatewayGroups`, which means groups can block gateway deletion, but this controller does not expose gateway-group CRUD. |
| OPNsense `stable/26.1`: `OPNsense/Routing/Api/SettingsController.php` | Implements individual gateway CRUD/search/reconfigure only. Gateway delete checks legacy `config.xml` gateway-group references, but no gateway-group CRUD API is exposed. |
| OPNsense `master`: `OPNsense/Routing/GatewayGroups.php` and `GatewayGroups.xml` | A model exists on `master`, with `gateway_group` items, `item`/`item2`/`item3`/`item4`/`item5` tier fields, `trigger`, `poolopts`, and `descr`; this is model evidence only, not an API controller. |
| OPNsense `stable/26.1`: `OPNsense/Routing/GatewayGroups.php` and `GatewayGroups.xml` | Both checked paths returned 404 during story creation; no target-release model evidence was found at those paths. |
| Candidate controller names on `master` | Checked likely paths `GatewayGroupsController.php`, `GatewayGroupController.php`, `GroupsController.php`, and `GroupController.php`; all returned 404 during story creation. |

Model semantics captured from `master` `GatewayGroups.xml` and `GatewayGroups.php` for future handoff only:

| Field / behavior | Evidence |
|---|---|
| Model mount/root | `GatewayGroups.xml` mounts `/gateways/gateway_group+` with item `gateway_group` as an `ArrayField`. No API wrapper/monad key is confirmed because no API controller was found. |
| Required name | `name` is required, must match `/^[a-zA-Z0-9_\-]{1,32}$/`, and has a unique constraint. `GatewayGroups.php` rejects changing the persisted group name. |
| Tier/member representation | Five tier fields exist: `item`, `item2`, `item3`, `item4`, and `item5`. `GatewayGroups.php` reads each tier as comma-separated gateway names and exposes normalized `tiers` only through private model helper behavior, not an API contract. |
| Tier validation | At least one gateway must be selected across the tier fields; validation reports against all tier fields if none are set. |
| Gateway name collision | A gateway group name cannot equal an existing individual gateway name. |
| Trigger setting | `trigger` is required and defaults to `down`; valid values are `down`, `downloss`, `downlatency`, and `downlosslatency`. |
| Pool options | `poolopts` options are default empty, `round-robin`, and `round-robin sticky-address`. |
| Description | `descr` is a `DescriptionField` with no required/default value identified. |
| Populate action | Tier fields use `ConfigdPopulateAct` `interface gateways list -l`, which is UI/model population evidence, not durable API evidence. |
| Reconfigure/apply behavior | No gateway-group-specific API reconfigure endpoint was found; individual gateway routing uses `/api/routing/settings/reconfigure` and configd `interface routes configure`. |

Treat this evidence as a starting point. Re-check live/current source during implementation because upstream status can change.

### If Upstream Has Shipped the API

Before any provider implementation, capture all of the following in this story or a new implementation story:

- Resource name decision: likely `opnsense_system_gateway_group` only if it maps to durable multi-WAN failover/load-balancing groups and does not duplicate `opnsense_system_gateway` fields.
- Endpoint table: add/get/set/delete/search or singleton get/set paths, HTTP methods, whether UUID-backed or name-backed, and reconfigure/apply endpoint.
- Monad/wrapper keys and model root path.
- Field model: `name`, description, trigger setting, pool options, and tier/member fields.
- Tier/member representation: whether API accepts legacy `item`, `item2`, `item3`, `item4`, `item5`, comma-separated gateway names, key-value maps, or a normalized list/map shape.
- Required fields, defaults, and constraints.
- Import semantics: UUID, group name, or synthetic ID.
- Update semantics: whether gateway-group `name` is immutable; `GatewayGroups.php` on `master` rejects name changes when persisted config already contains the group.
- Delete semantics: whether deletion is safe, whether routes/rules/interfaces may reference the group, and whether OPNsense rejects in-use deletion.
- Live validation plan on disposable OPNsense only; gateway-group changes can affect default-route failover/load-balancing behavior.
- Safety notes for WAN/LAN, default gateway switching, dpinger trigger behavior, and multi-WAN production impact.

### Existing Provider Coverage to Preserve

Already-supported routing and gateway-adjacent resources:

| Domain | Provider resource(s) | Location / notes |
|---|---|---|
| Individual gateways | `opnsense_system_gateway` + data source | `internal/service/system/*gateway*`; endpoints under `/api/routing/settings/*_gateway`; schema includes `weight` and `priority` used by gateway groups. |
| Static routes | `opnsense_system_route` + data source | `internal/service/system/*route*`; endpoints under `/api/routes/routes/*route`. |
| Dynamic routing | `opnsense_quagga_*` resources | FRR/BGP/OSPF/RIP/static routing resources are separate from gateway groups. |
| Firewall NAT/routing | `opnsense_firewall_*` resources | NAT rules are separate firewall resources, not gateway-group configuration. |

Do not create alternate resources for these. If docs mention gateways generically, keep the distinction clear: individual gateways and static routes are supported; gateway groups remain blocked unless upstream exposes stable API support.

### Architecture and Implementation Guardrails

If this story becomes unblocked and implementation is explicitly authorized by captured endpoint evidence, follow the established resource pattern:

- Service package: `internal/service/system`.
- Resource name: `opnsense_system_gateway_group` unless the implementation handoff records a better name.
- Files: `gateway_group_resource.go`, `gateway_group_schema.go`, `gateway_group_model.go`, `gateway_group_resource_test.go`, `gateway_group_data_source.go` if lookup semantics are safe, plus `exports.go` registration.
- API client: use existing `pkg/opnsense` generic CRUD/singleton helpers and `ReqOpts`; do not create raw HTTP calls in resource code.
- State: always read back from API after create/update; never set state from plan.
- References: gateway members should reference gateway names or IDs exactly as the upstream model requires; add validators only after evidence confirms the stable field shape.
- Collections: use `types.Set` for unordered gateway member collections unless OPNsense tier ordering is semantically significant; if tier order is significant, model tiers explicitly and document ordering.
- Reconfigure: route to the endpoint confirmed by upstream evidence; do not invent a reconfigure path.
- Tests: include create, import, update, read-back, destroy, and in-use/dependency edge cases when feasible. Acceptance tests must avoid disrupting the active management path.
- Documentation: add templates, examples, generated docs, and migration guidance only when the resource actually ships.

### What NOT to Build

- Do not implement `opnsense_system_gateway_group` while gateway-group API support remains model-only, UI-only, legacy-config-only, or absent from the target release.
- Do not parse or write `config.xml` directly.
- Do not scrape UI forms or rely on private helper methods such as `GatewayGroups::getGroupsConfig()` as a provider API.
- Do not assume the individual gateway `weight`/`priority` fields are sufficient to represent gateway groups.
- Do not model runtime gateway status or dpinger state as durable Terraform resource state.
- Do not add backward-compatibility shims for speculative API shapes.

### Previous Story Intelligence

- Story 5.5 implemented `opnsense_system_gateway` with CRUD, read-back after create/update, import, drift removal on not found, docs, examples, and standard `/api/routing/settings/reconfigure`.
- The gateway resource currently uses `gatewayReqOpts` with monad `gateway`, but upstream docs/source name the model key `gateway_item`. If gateway-group implementation becomes possible, verify wrapper keys from live responses instead of copying the gateway resource blindly.
- `opnsense_system_gateway` already exposes `weight` and `priority`; gateway groups should compose existing gateways rather than duplicate gateway creation.
- Gateway delete behavior upstream already checks gateway-group membership and refuses deleting in-use gateways. A future gateway-group resource must account for Terraform dependency ordering and clear diagnostics when members are still referenced.

### Testing / Validation Requirements

- If still blocked: documentation/status-only changes plus `make check` are sufficient.
- If unblocked and implementation is explicitly authorized by captured endpoint evidence: follow standard provider resource requirements: unit tests for model/API conversion, resource lifecycle tests, import behavior, examples/templates/docs, `go generate ./tools`, and `make check`.
- Acceptance testing for any future implementation must avoid mutating production default routing or active management WAN/LAN paths. Use disposable test appliances or isolated gateway objects.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md` Story 5.6]
- [Source: `_bmad-output/planning-artifacts/prd.md` FR45]
- [Source: `_bmad-output/planning-artifacts/core-config-gap-analysis.md` Routing & Gateways]
- [Source: `_bmad-output/planning-artifacts/support-matrix.md` Upstream-Blocked]
- [Source: `_bmad-output/planning-artifacts/feature-complete-roadmap.md` Two-front upstream contribution track]
- [Source: `docs/upstream-blocked.md` blocked domains register]
- [Source: `docs/migration-import.md` upstream-blocked migration guidance]
- [Source: `_bmad-output/implementation-artifacts/5-5-gateway-resource.md`]
- [Source: `internal/service/system/gateway_resource.go`]
- [Source: `internal/service/system/gateway_schema.go`]
- [Source: `internal/service/system/gateway_model.go`]
- [Source: OPNsense published API docs: `development/api/core/routing.html` and `development/api/core/routes.html`]
- [Source: OPNsense upstream source paths checked under `src/opnsense/mvc/app/controllers/OPNsense/Routing/Api/` and `src/opnsense/mvc/app/models/OPNsense/Routing/`]

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- Ultimate context engine analysis completed - comprehensive developer guide created.
- 2026-06-14: `git rev-parse HEAD` captured baseline `fbaf085e8287ae4f00f786484cd1ab622d77716a`.
- 2026-06-14: Published OPNsense routing/routes API docs checked; no gateway-group endpoint listed.
- 2026-06-14: OPNsense `master` and `stable/26.1` `SettingsController.php` checked; both expose individual gateway endpoints only.
- 2026-06-14: OPNsense `master` `GatewayGroups.php` and `GatewayGroups.xml` checked; model evidence exists, but no API controller was found at likely controller paths.
- 2026-06-14: OPNsense `stable/26.1` `GatewayGroups.php` and `GatewayGroups.xml` checked paths returned 404.
- 2026-06-14: Existing provider coverage checked in `internal/service/system` for gateways and static routes.
- 2026-06-14: Rechecked published OPNsense routing/routes API docs; docs list individual gateway/static route endpoints only and no gateway-group endpoint.
- 2026-06-14: Rechecked OPNsense `master` and `stable/26.1` routing `SettingsController.php`; both expose individual gateway CRUD/search/reconfigure only.
- 2026-06-14: Rechecked likely `master` API controller paths for gateway groups (`GatewayGroupsController.php`, `GatewayGroupController.php`, `GroupsController.php`, `GroupController.php`); all returned 404.
- 2026-06-14: Regenerated provider docs with `docker run --rm -v "$PWD:/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go generate ./tools` after updating `templates/index.md.tmpl`.
- 2026-06-14: `make check` passed: lint, format, test, security, scan, and docs.
- 2026-06-14: Code review patch follow-ups applied; `make check` passed again.

### Completion Notes List

- Created a revalidation-gate story for gateway groups that prevents unsafe implementation while the API remains unavailable in the target release.
- Captured current upstream nuance: `master` has gateway-group model evidence, but no published API endpoint or controller evidence was found, and target `stable/26.1` did not expose the checked model paths.
- Preserved existing provider boundaries: individual gateways/static routes are already supported; gateway groups remain a separate upstream-blocked domain.
- Revalidation outcome: still upstream-blocked for the provider target (`stable/26.1`) and no `opnsense_system_gateway_group` implementation was created.
- Public and planning docs now classify gateway groups as blocked in the target release while tracking `master` model-only evidence as future recheck context.
- Existing provider coverage remains distinct and non-duplicated: `opnsense_system_gateway`, `opnsense_system_route`, dynamic routing resources, and firewall NAT/routing resources are already supported through their existing packages.
- Addressed review finding: captured `GatewayGroups` model field semantics, defaults, constraints, tier/member representation, and missing API wrapper/reconfigure evidence in the story record.
- Addressed review finding: aligned `core-config-gap-analysis.md` verified provider-owned resource candidate count with the two named candidates.
- Validation complete: `make check` passed before moving this story to review.

### File List

- `_bmad-output/implementation-artifacts/5-6-gateway-group-resource.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `docs/index.md`
- `docs/migration-import.md`
- `docs/upstream-blocked.md`
- `templates/index.md.tmpl`

## Change Log

- 2026-06-14: Revalidated gateway-group API status and refreshed blocked-domain documentation; no provider implementation created.
