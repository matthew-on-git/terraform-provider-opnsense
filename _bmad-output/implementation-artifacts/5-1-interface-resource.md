---
baseline_commit: fbaf085e8287ae4f00f786484cd1ab622d77716a
---

# Story 5.1: Interface Resource API Revalidation Gate

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a provider maintainer,
I want to revalidate whether OPNsense now exposes a stable API for base interface assignment, interface IP configuration, and PPPoE,
so that we either unblock a safe future `opnsense_system_interface` implementation or keep the upstream-blocked register honest with current evidence.

## Acceptance Criteria

1. **Given** the current target OPNsense release and upstream source state
   **When** the developer investigates base interface assignment / IP config / PPPoE support
   **Then** the story records whether a stable upstream API exists, including endpoint paths, model/controller names, wrapper keys, required fields, defaults, reconfigure/apply behavior, and release/version evidence

2. **And** if OPNsense PR #8436 is still unmerged or not present in the target release, the developer must **not** implement `opnsense_system_interface`; instead they update the relevant planning/public docs with the latest blocked evidence

3. **And** if a stable upstream API has shipped in the target release, the developer creates a follow-up implementation story or updates this story with exact implementation scope before writing provider code

4. **And** the investigation explicitly compares existing provider coverage so no duplicate resource is created for already-supported interface types: `opnsense_system_vlan`, `opnsense_system_vip`, `opnsense_interface_bridge`, `opnsense_interface_gre`, `opnsense_interface_gif`, `opnsense_interface_vxlan`, `opnsense_interface_loopback`, and `opnsense_interface_neighbor`

5. **And** all touched docs/status files remain aligned: `docs/upstream-blocked.md`, `docs/migration-import.md`, `docs/index.md` / `templates/index.md.tmpl`, `_bmad-output/planning-artifacts/support-matrix.md`, `_bmad-output/planning-artifacts/core-config-gap-analysis.md`, `_bmad-output/planning-artifacts/feature-complete-roadmap.md`, `_bmad-output/planning-artifacts/prd.md`, and `_bmad-output/implementation-artifacts/sprint-status.yaml`

6. **And** `make check` passes before completion

## Tasks / Subtasks

- [x] Task 1: Revalidate upstream API availability (AC: #1, #2, #3)
  - [x] 1.1 Check OPNsense PR #8436 status and current maintainer comments
  - [x] 1.2 Check current OPNsense published API docs/source for interface assignment, IP config, and PPPoE endpoints
  - [x] 1.3 If endpoints appear, capture controller/model names, endpoint paths, wrapper keys, payload shape, reconfigure/apply behavior, and target release availability
  - [x] 1.4 If endpoints are absent or PR remains unmerged, record the blocker and stop before provider implementation
- [x] Task 2: Compare existing provider interface coverage to prevent duplicate work (AC: #4)
  - [x] 2.1 Confirm `internal/service/system` already owns VLAN/VIP and that `internal/service/iface` owns bridge/GRE/GIF/VXLAN/loopback/neighbor
  - [x] 2.2 Confirm this story is about base assignment/IP config/PPPoE only, not interface type resources already supported
- [x] Task 3: Update public/planning docs based on evidence (AC: #2, #5)
  - [x] 3.1 If still blocked, refresh blocked-register/support docs with current PR #8436 status and date
  - [x] 3.2 If unblocked, update docs to move the domain from Upstream-blocked to Coming and create precise implementation handoff notes
- [x] Task 4: Update sprint/planning status consistently (AC: #5)
  - [x] 4.1 If still blocked, mark this story done as an upstream-blocked revalidation record and leave implementation work uncreated
  - [x] 4.2 If unblocked, create or update a follow-up implementation story before marking this story complete
- [x] Task 5: Run `make check` (AC: #6)

### Review Findings

- [x] [Review][Patch] PRD still promises blocked interface management unconditionally [_bmad-output/planning-artifacts/prd.md:494]
- [x] [Review][Patch] Missing defaults in upstream API evidence [_bmad-output/implementation-artifacts/5-1-interface-resource.md:147]
- [x] [Review][Defer] ACME issuance marked done despite contradictory sprint-status note [_bmad-output/implementation-artifacts/sprint-status.yaml:280] — deferred, pre-existing

## Dev Notes

### Current Classification

This story is a **revalidation gate**, not a build-now resource story. The original Epic 5 requirement asked for `opnsense_system_interface`, but current planning artifacts classify base interface assignment / IP configuration / PPPoE as **Upstream-blocked** because the target OPNsense release has no stable supported API.

Do **not** implement provider code against fork-only, local-plugin, private, or unmerged endpoints. Terraform cannot safely manage durable appliance interface assignment without upstream-supported API semantics.

### Upstream Evidence to Check

- OPNsense PR #8436: `https://github.com/opnsense/core/pull/8436`
- Current fetched status during story creation: PR #8436 is open, titled "Add interface assignments to API".
- Maintainer evidence from PR #8436:
  - 2025-03-14: OPNsense maintainer stated the PR is unlikely to mature soon and requires broader upstream work.
  - 2025-03-21: OPNsense maintainer stated the exact endpoints are not used by UI, which is a requirement, and assignment refactor is not yet planned.
  - 2026 page evidence still shows the PR open, not merged.

Treat this web evidence as a starting point. Re-check live/current source during implementation because upstream status can change.

### Existing Provider Coverage to Preserve

Already-supported interface-adjacent resources:

| Domain | Provider resource(s) | Location |
|---|---|---|
| VLAN | `opnsense_system_vlan` | `internal/service/system/*vlan*` |
| Virtual IP | `opnsense_system_vip` | `internal/service/system/*vip*` |
| Bridge | `opnsense_interface_bridge` + data source | `internal/service/iface/*bridge*` |
| GRE | `opnsense_interface_gre` + data source | `internal/service/iface/*gre*` |
| GIF | `opnsense_interface_gif` + data source | `internal/service/iface/*gif*` |
| VXLAN | `opnsense_interface_vxlan` + data source | `internal/service/iface/*vxlan*` |
| Loopback | `opnsense_interface_loopback` + data source | `internal/service/iface/*loopback*` |
| Neighbor / static ARP/NDP | `opnsense_interface_neighbor` + data source | `internal/service/iface/*neighbor*` |

Do not create alternate resources for these. If docs mention "interfaces" generically, keep the distinction clear: type resources are supported; base assignment/IP config/PPPoE remain blocked unless upstream changes.

### If Upstream Has Shipped the API

Before any provider implementation, capture all of the following in this story or a new implementation story:

- Resource name decision: likely `opnsense_system_interface` only if it maps to base assignment/IP config/PPPoE and does not conflict with `internal/service/iface` interface-type resources.
- Endpoint table: add/get/set/delete/search or singleton get/set paths, HTTP methods, whether UUID-backed or assignment-name-backed, and reconfigure/apply endpoint.
- Monad/wrapper keys and model root path.
- Required fields, defaults, and API enum values.
- Import semantics: UUID, interface name, assignment key, or synthetic ID.
- Delete semantics: whether deletion is safe, resets assignment, disables interface, or must be no-op/state removal only.
- Live validation plan on disposable OPNsense only; interface assignment changes can lock out management access.
- Safety notes for WAN/LAN and management-interface lockout avoidance.

### What NOT to Build

- Do not implement `opnsense_system_interface` while PR #8436 remains unmerged or absent from the target OPNsense release.
- Do not use legacy PHP pages or scrape UI state.
- Do not model runtime-only or local-plugin endpoints as supported provider resources.
- Do not create duplicates of existing VLAN/VIP/interface-type resources.
- Do not add backward-compatibility shims for speculative API shapes.

### Testing / Validation Requirements

- If still blocked: documentation/status-only changes plus `make check` are sufficient.
- If unblocked and implementation is explicitly authorized by captured endpoint evidence: follow standard provider resource requirements: unit tests for model/API conversion, resource lifecycle tests, import behavior, examples/templates/docs, `go generate ./tools`, and `make check`.
- Acceptance testing for any future implementation must avoid mutating the active management interface and should run only on disposable test appliances.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md` Story 5.1]
- [Source: `_bmad-output/planning-artifacts/core-config-gap-analysis.md` Interfaces status]
- [Source: `_bmad-output/planning-artifacts/support-matrix.md` Upstream-Blocked]
- [Source: `_bmad-output/planning-artifacts/feature-complete-roadmap.md` Two-front upstream contribution track]
- [Source: `_bmad-output/planning-artifacts/prd.md` Tier 3 and upstream-collaboration notes]
- [Source: `docs/upstream-blocked.md` blocked domains register]
- [Source: `_bmad-output/implementation-artifacts/28-1-upstream-blocked-register-and-maintenance.md` maintenance workflow]
- [Source: OPNsense PR #8436]

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- Ultimate context engine analysis completed - comprehensive developer guide created.
- 2026-06-12: `git rev-parse HEAD` captured baseline `fbaf085e8287ae4f00f786484cd1ab622d77716a`.
- 2026-06-12: Rechecked OPNsense PR #8436 page; PR remains open and maintainer comments still indicate upstream assignment refactor/API maturity is incomplete.
- 2026-06-12: Rechecked published OPNsense interface API docs; docs list bridge/GIF/GRE/LAGG/loopback/neighbor/overview/settings/VIP/VLAN, but not assignment/IP config/PPPoE CRUD.
- 2026-06-12: Checked upstream `master` source and found emerging `OPNsense/Interfaces/Api/AssignmentController.php` and `NetworkInterface` memory model.
- 2026-06-12: Checked target `stable/26.1`; `AssignmentController.php` returns 404 and is not present in the target release branch.
- 2026-06-12: Checked interface ACL/menu source; `master` menu exposes `/ui/interfaces/assignment`, but ACL coverage does not include `api/interfaces/assignment/*`.
- 2026-06-12: Compared provider coverage via `internal/service/system` and `internal/service/iface` for VLAN/VIP and interface-type resources.
- 2026-06-12: Regenerated provider docs with `docker run --rm -v "$PWD:/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go generate ./tools` after updating `templates/index.md.tmpl`.
- 2026-06-12: `make check` passed: lint, format, test, security, scan, and docs.

### Completion Notes List

- Revalidation outcome: still upstream-blocked for the provider target (`stable/26.1`) and no `opnsense_system_interface` implementation was created.
- Upstream nuance captured: OPNsense `master` now contains an emerging `interfaces/assignment` controller backed by `NetworkInterface`; endpoints are item-style `search_item`, `add_item`, `get_item`, `set_item`, `del_item`, and `reconfigure` with wrapper/model root `interface`.
- Emerging `master` endpoint details captured for future handoff only: `GET,POST /api/interfaces/assignment/search_item`, `POST /api/interfaces/assignment/add_item`, `GET /api/interfaces/assignment/get_item/$ifname`, `POST /api/interfaces/assignment/set_item/$ifname`, `POST /api/interfaces/assignment/del_item/$ifnames`, and `POST /api/interfaces/assignment/reconfigure`.
- Emerging `master` payload/model details captured for future handoff only: model `OPNsense\Interfaces\NetworkInterface`, root/wrapper `interface`, fields `descr`, `identifier`, volatile read-only `icon`/`optgroup`, and required unique device field `if` using `DeviceField`; no stable defaults were identified for the emerging assignment API beyond model-defined field behavior.
- Emerging `master` apply/delete behavior captured for future handoff only: reconfigure runs configd `interface apply`, then processes `/tmp/.interfaces.todo` pending delete/relink actions, removes rules for deleted interfaces, saves config, flushes the todo file, and runs `filter reload skip_alias`; delete refuses non-POST requests, in-use interfaces, and locked interfaces.
- Target-release blocker captured: `AssignmentController.php` is absent from `stable/26.1`, published interface API docs do not list it, ACL coverage does not include `api/interfaces/assignment/*`, and the model only covers assignment `descr`/`identifier`/device `if`, not IP configuration or PPPoE.
- Existing provider coverage remains distinct and non-duplicated: `opnsense_system_vlan`, `opnsense_system_vip`, `opnsense_interface_bridge`, `opnsense_interface_gre`, `opnsense_interface_gif`, `opnsense_interface_vxlan`, `opnsense_interface_loopback`, and `opnsense_interface_neighbor` are already supported through their existing packages.
- Public and planning docs now classify interface assignment as blocked in target release while tracking emerging `master` API evidence as research for a future story.
- Validation complete: `make check` passed before moving this story to review.

### File List

- `_bmad-output/implementation-artifacts/5-1-interface-resource.md`
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

- 2026-06-11: Created upstream-blocked revalidation story for base interface assignment/IP config/PPPoE.
- 2026-06-12: Revalidated upstream/target-release interface assignment API status and refreshed blocked-domain documentation; no provider implementation created.
