---
baseline_commit: fbaf085e8287ae4f00f786484cd1ab622d77716a
---

# Story 17.2: Interface LAGG Resource

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator,
I want to manage OPNsense LAGG interfaces through Terraform,
so that I can define link aggregation groups as code using the stable OPNsense `interfaces/lagg_settings` API.

## Acceptance Criteria

1. **Given** a Vagrant OPNsense test VM with selectable LAGG member interfaces
   **When** the developer runs the focused LAGG acceptance test
   **Then** `opnsense_interface_lagg` can create, read back, import, update, and delete a LAGG using `/api/interfaces/lagg_settings/*` endpoints

2. **And** the implementation uses the existing generated interface-resource pattern in `internal/generate/schemas/interfaces.yaml` and `internal/service/iface`, not hand-written duplicate CRUD code

3. **And** the resource models upstream fields safely: computed LAGG device name, required member interfaces, optional primary member, protocol, LACP fast timeout, flow ID behavior, LAGG hash options, LACP strict behavior, MTU, and description

4. **And** the Vagrant test environment is updated or documented so it provides enough free/unassigned VirtualBox NICs for LAGG live validation; the previous "needs hardware box" assumption is replaced by a Vagrant-first validation path

5. **And** generated docs, examples, provider index/support-matrix references, and migration guidance are aligned with the new supported resource after live validation

6. **And** `make check` passes before completion

## Tasks / Subtasks

- [x] Task 1: Make Vagrant suitable for LAGG validation (AC: #1, #4)
  - [x] 1.1 Inspect the current Vagrant VM NIC layout and confirm which interfaces OPNsense exposes as assignable/selectable LAGG members
  - [x] 1.2 If the existing VM only has WAN + LAN, update `test/Vagrantfile` to add dedicated test NICs, likely `--nic3 intnet` and `--nic4 intnet`, on an isolated internal network such as `opnsense-lagg-test`
  - [x] 1.3 Document the LAGG-specific Vagrant expectation in `test/README.md`, including the SSH tunnel path if TLS forwarding on `10443` is unreliable
  - [x] 1.4 Prove member availability before resource work by querying the live API/model response or OPNsense interface inventory from the Vagrant appliance
- [x] Task 2: Add LAGG to the interface generator schema (AC: #2, #3)
  - [x] 2.1 Add a `lagg` resource entry to `internal/generate/schemas/interfaces.yaml`
  - [x] 2.2 Use package `iface`, type name `interface_lagg`, title `a LAGG interface`, kind `item`, monad `lagg`, and endpoints under `/api/interfaces/lagg_settings`
  - [x] 2.3 Model fields from upstream `Lagg.xml`: `device`/`laggif`, `members`, `primary_member`, `protocol`/`proto`, `lacp_fast_timeout`, `use_flowid`, `lagg_hash`/`lagghash`, `lacp_strict`, `mtu`, and `description`/`descr`
  - [x] 2.4 Use existing generator field types where possible: `selectmaplist` for multi-select member/hash fields, `selectmap` for option fields, `bool` for string booleans, `int` for MTU, and `string` for description/device
  - [x] 2.5 Add option validators in the YAML where supported by the generator: protocol values `none`, `lacp`, `failover`, `fec`, `loadbalance`, `roundrobin`; LAGG hash values `l2`, `l3`, `l4`; flow/strict values empty/default, `1`, `0`
- [x] Task 3: Generate and register provider code (AC: #1, #2, #3)
  - [x] 3.1 Run the project generator so it emits `lagg_model.gen.go`, `lagg_schema.gen.go`, `lagg_resource.gen.go`, and `lagg_data_source.gen.go` if data sources are generated for interface items
  - [x] 3.2 Register `newLaggResource` and `newLaggDataSource` in `internal/service/iface/exports.go`
  - [x] 3.3 Do not manually edit generated `.gen.go` files; fix generation inputs/templates instead
- [x] Task 4: Add focused tests and live validation (AC: #1, #4)
  - [x] 4.1 Add or verify a focused acceptance test for `opnsense_interface_lagg` using Vagrant member interfaces, not production appliance interfaces
  - [x] 4.2 Test create/read/import/update/delete and ensure delete is blocked or reported clearly if OPNsense says the LAGG is in use
  - [x] 4.3 Use serial acceptance execution (`-p 1`) because interface mutations affect global appliance config
  - [x] 4.4 Capture the exact focused test command and result in the story Dev Agent Record
- [x] Task 5: Update docs/examples/status (AC: #5)
  - [x] 5.1 Add `templates/resources/interface_lagg.md.tmpl` and `examples/resources/opnsense_interface_lagg/resource.tf`
  - [x] 5.2 Regenerate provider docs with `go generate ./tools` through the dev-toolchain container or Makefile path used by the repo
  - [x] 5.3 Update planning/support docs that currently classify LAGG as Coming/live-validation-gated to Supported after live validation succeeds
  - [x] 5.4 Update `_bmad-output/implementation-artifacts/sprint-status.yaml` from `ready-for-dev` to the next workflow status only through the normal dev/review flow
- [x] Task 6: Run `make check` (AC: #6)

### Review Findings

- [x] [Review][Patch] LAGG delete can silently succeed when OPNsense refuses an in-use delete [`pkg/opnsense/crud.go:141`]
- [x] [Review][Patch] Focused LAGG acceptance test does not exercise update lifecycle [`internal/service/iface/lagg_resource_gen_test.go:21`]
- [x] [Review][Patch] `lagg_hash` accepts invalid set elements despite the declared option constraint [`internal/service/iface/lagg_schema.gen.go:36`]
- [x] [Review][Patch] LAGG test README fallback command runs the HAProxy action test instead of the LAGG test [`test/README.md:94`]
- [x] [Review][Patch] Generated LAGG resource documentation has nested Terraform code fences [`docs/resources/interface_lagg.md:21`]
- [x] [Review][Patch] Support/count documentation is internally inconsistent after adding LAGG [`_bmad-output/planning-artifacts/prd.md:258`]
- [x] [Review][Patch] MTU documents range 576-65535 but the generated schema does not validate it [`internal/service/iface/lagg_schema.gen.go:38`]

## Dev Notes

### Current Classification

This is a **buildable implementation story with a Vagrant validation prerequisite**. LAGG is not upstream-blocked: published OPNsense API docs and current `master`/`stable/26.1` source expose `interfaces/lagg_settings` item CRUD/search/reconfigure endpoints and `Lagg.xml`.

The prior blocker in `sprint-status.yaml` was environmental: the tested VM had no free/unassigned NICs to form a LAGG. The user has clarified that Vagrant should be usable for testing, so the developer should make the local Vagrant environment LAGG-capable before declaring implementation blocked.

### Upstream API Evidence

Evidence checked during story creation on 2026-06-15:

| Source | Evidence |
|---|---|
| Published API docs: `development/api/core/interfaces.html` | Lists `LaggSettingsController.php` endpoints: `add_item`, `del_item`, `get`, `get_item`, `reconfigure`, `search_item`, `set`, and `set_item`; uses `Lagg.xml`. |
| OPNsense `master`: `OPNsense/Interfaces/Api/LaggSettingsController.php` | Standard `ApiMutableModelControllerBase` item controller with internal model name `lagg`, class `OPNsense\Interfaces\Lagg`, search by `descr`, auto-assigned `laggif`, set overlay preserving existing `laggif`, delete in-use checks, and reconfigure via `interface lagg configure`. |
| OPNsense `stable/26.1`: `OPNsense/Interfaces/Api/LaggSettingsController.php` | Same controller evidence found on target release branch. |
| OPNsense `master` and `stable/26.1`: `OPNsense/Interfaces/Lagg.xml` | Model mounts `/laggs`, item array `lagg`, required `members`, default protocol `lacp`, booleans/options/hash/MTU/description fields. |
| OPNsense `master` and `stable/26.1`: `OPNsense/Interfaces/Lagg.php` | Performs validation that primary member is included in member list and attempts to prevent member reuse across LAGGs. |

Endpoint contract to use:

| Operation | Endpoint |
|---|---|
| Add | `POST /api/interfaces/lagg_settings/add_item` |
| Get | `GET /api/interfaces/lagg_settings/get_item/{uuid?}` |
| Set | `POST /api/interfaces/lagg_settings/set_item/{uuid}` |
| Delete | `POST /api/interfaces/lagg_settings/del_item/{uuid}` |
| Search | `GET,POST /api/interfaces/lagg_settings/search_item` |
| Reconfigure | `POST /api/interfaces/lagg_settings/reconfigure` |

Use wrapper/monad key `lagg` unless live API evidence proves otherwise.

### Field Handoff

Map upstream fields as follows unless live validation shows a different response shape:

| Terraform field | OPNsense JSON/XML field | Suggested type | Notes |
|---|---|---|---|
| `id` | UUID | computed string | Standard item ID. |
| `device` | `laggif` | optional+computed string | OPNsense auto-assigns `lagg0`, `lagg1`, etc. `setItemAction` preserves existing `laggif`; do not require operators to set it. |
| `members` | `members` | required set of string | `LaggInterfaceField`, multiple, required. Live validation must confirm values exposed in Vagrant, likely physical interface names such as `vtnet2`/`em2` or OPNsense identifiers depending response shape. |
| `primary_member` | `primary_member` | optional+computed string | Must be one of `members` when set. For failover, document expected use after live validation. |
| `protocol` | `proto` | optional+computed string, default `lacp` | Valid values: `none`, `lacp`, `failover`, `fec`, `loadbalance`, `roundrobin`. |
| `lacp_fast_timeout` | `lacp_fast_timeout` | bool, default `false` | OPNsense stores as string boolean. |
| `use_flowid` | `use_flowid` | optional+computed string | Values: empty/default, `1`/yes, `0`/no. |
| `lagg_hash` | `lagghash` | optional+computed set of string | Multiple option field. Valid values: `l2`, `l3`, `l4`. Use `types.Set` to avoid ordering diffs. |
| `lacp_strict` | `lacp_strict` | optional+computed string | Values: empty/default, `1`/yes, `0`/no. |
| `mtu` | `mtu` | optional+computed int | OPNsense validates 576-65535. |
| `description` | `descr` | optional+computed string | Description field. |

### Existing Code Patterns to Reuse

- Add LAGG to `internal/generate/schemas/interfaces.yaml`; do not hand-write parallel CRUD resources.
- Generated interface resources live in `internal/service/iface` and are named like `bridge_resource.gen.go`, `gre_model.gen.go`, and `vxlan_schema.gen.go`.
- Existing interface resources use generic `opnsense.Add`, `Get`, `Update`, `Delete`, and `Search` with `ReqOpts` endpoints and monad keys.
- Register resources/data sources in `internal/service/iface/exports.go`.
- Existing bridge implementation is the closest field-shape reference because it already models interface-member sets with `selectmaplist` and a generated resource/data source pair.
- Use `tfconv.SetToCSV` and `tfconv.SliceToSet` via generator output for multi-select fields.

### Vagrant Validation Requirements

Current `test/Vagrantfile` evidence:

- Box: `puzzle/opnsense`, version `~> 25.7`.
- Port forwarding: guest `443` to host `127.0.0.1:10443`.
- VirtualBox currently adds only `--nic2 intnet` on `opnsense-lan` for LAN.
- `test/README.md` documents the reliable SSH tunnel path `localhost:10444` because TLS over NAT forwarding can hang.

The developer should update Vagrant before implementation if the current VM has no LAGG member candidates:

- Add at least two dedicated VirtualBox NICs for LAGG testing on an isolated internal network. The implemented Vagrant setup uses NIC 5 and NIC 6 because the base box may already assign NIC 3 and NIC 4 as WAN/OPT interfaces.
- Ensure the new NICs are not used as WAN/LAN management interfaces and do not break API reachability.
- Confirm OPNsense exposes those NICs as valid `LaggInterfaceField` choices before using them in acceptance tests.
- Keep tests disposable. Creating/deleting LAGGs can alter network interface state and should run only against the Vagrant appliance or another dedicated test box.

### Testing / Validation Requirements

- Run `make check` before completion.
- Run the focused LAGG acceptance test against Vagrant before marking the story review-ready. Expected shape:
  - `cd test && vagrant up`
  - Keep SSH tunnel open if needed: `sshpass -p opnsense ssh -p 2222 -o PreferredAuthentications=password -o PubkeyAuthentication=no -N -L 10444:127.0.0.1:443 root@127.0.0.1`
  - Export `OPNSENSE_URI=https://localhost:10444`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET`, and `OPNSENSE_ALLOW_INSECURE=true`
  - Run focused acceptance test serially, preferably through the dev-toolchain container with host networking if host Terraform auto-install fails
- Include import verification and a second apply/plan-equivalent no-diff behavior where feasible.
- If LAGG live validation still fails, do not fake completion. Capture the exact VM NIC inventory, API response, validation error, and next required environment change in the story.

### Documentation Requirements

- Add resource docs template and example HCL for `opnsense_interface_lagg`.
- Regenerate `docs/` from templates; do not hand-edit generated docs except as part of generation output.
- Update `support-matrix.md`, `core-config-gap-analysis.md`, and provider index text from Coming to Supported only after live validation succeeds.
- If Vagrant changes are required, update `test/README.md` with the LAGG-specific setup and focused test command.

### What NOT to Build

- Do not write direct `config.xml` manipulation.
- Do not implement a hand-written LAGG resource unless the generator cannot represent a confirmed live API shape; if that happens, document why.
- Do not use production WAN/LAN interfaces as LAGG members in acceptance tests.
- Do not mark LAGG supported from published docs alone; live validation with selectable member interfaces is the gate.
- Do not duplicate existing bridge/GIF/GRE/VXLAN/loopback/neighbor resources or move the interface package.

### Previous Story / Git Intelligence

- No prior `17-1` implementation story file exists in `_bmad-output/implementation-artifacts`; implementation evidence is in generated interface code and sprint status.
- Existing generated interface resources are already live-validated for bridge, GRE, GIF, VXLAN, loopback, and neighbor.
- Recent commit baseline includes post-release resource completion and v0.2.0 release prep (`fbaf085`, `0caa626`), so expect many interface/service resources to already follow the codegen pattern.
- The current worktree is dirty with unrelated user/agent changes; do not revert or modify unrelated files.

### References

- [Source: `_bmad-output/planning-artifacts/prd.md` FR108]
- [Source: `_bmad-output/planning-artifacts/feature-complete-roadmap.md` Epic 17 Interface Types]
- [Source: `_bmad-output/planning-artifacts/support-matrix.md` Coming: Interface LAGG]
- [Source: `_bmad-output/planning-artifacts/core-config-gap-analysis.md` Interfaces]
- [Source: `_bmad-output/planning-artifacts/resource-gap-verification.md` Interface LAGG]
- [Source: `internal/generate/schemas/interfaces.yaml`]
- [Source: `internal/service/iface/bridge_model.gen.go`]
- [Source: `internal/service/iface/bridge_resource.gen.go`]
- [Source: `internal/service/iface/bridge_schema.gen.go`]
- [Source: `internal/service/iface/exports.go`]
- [Source: `test/Vagrantfile`]
- [Source: `test/README.md`]
- [Source: OPNsense published API docs: `development/api/core/interfaces.html`]
- [Source: OPNsense upstream source: `OPNsense/Interfaces/Api/LaggSettingsController.php`, `OPNsense/Interfaces/Lagg.xml`, `OPNsense/Interfaces/Lagg.php`]

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- 2026-06-15: `git rev-parse HEAD` baseline from current repo context is `fbaf085e8287ae4f00f786484cd1ab622d77716a`.
- 2026-06-15: Published OPNsense interfaces API docs checked; LAGG item endpoints are documented.
- 2026-06-15: OPNsense `master` and `stable/26.1` LAGG controller/model source checked; target-release API exists.
- 2026-06-15: Existing `internal/service/iface` generated bridge resource inspected as implementation pattern.
- 2026-06-15: `test/Vagrantfile` inspected; current file only declares a second LAN NIC, so story requires LAGG-capable Vagrant member NIC validation.
- 2026-06-15: Live VM initially exposed `em2` and `em3` as assigned WAN/DMZ interfaces, so LAGG member candidates were not available.
- 2026-06-15: Added Vagrant NIC 5 and NIC 6 on isolated `opnsense-lagg-test`; after graceful VM restart, OPNsense exposed `em4` and `em5` as physical interfaces.
- 2026-06-15: Live API `GET /api/interfaces/lagg_settings/get_item` through SSH tunnel on `localhost:10445` returned selectable `members` and `primary_member` options for `em4` and `em5` under wrapper key `lagg`.
- 2026-06-15: Focused acceptance test passed: `docker run --rm --network host -e TF_ACC=1 -e OPNSENSE_URI=https://localhost:10445 -e OPNSENSE_API_KEY=... -e OPNSENSE_API_SECRET=... -e OPNSENSE_ALLOW_INSECURE=true -v "$PWD:/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go test -run TestAccLagg_basic -count=1 -p 1 -timeout=20m ./internal/service/iface`.
- 2026-06-15: `make check` passed.
- 2026-06-15: Code review found seven patch findings; all were fixed. Updated LAGG acceptance coverage includes create/update/import/delete, `lagg_hash` and MTU validation are generated, delete responses parse failed validations, docs fences/counts were corrected, and the LAGG README command was fixed.
- 2026-06-15: Updated focused acceptance test passed: `docker run --rm --network host -e TF_ACC=1 -e OPNSENSE_URI=https://localhost:10445 -e OPNSENSE_API_KEY=... -e OPNSENSE_API_SECRET=... -e OPNSENSE_ALLOW_INSECURE=true -v "$PWD:/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go test -run TestAccLagg_basic -count=1 -p 1 -timeout=20m ./internal/service/iface`.

### Completion Notes List

- Implemented generated `opnsense_interface_lagg` resource and data source through `internal/generate/schemas/interfaces.yaml` and the existing interface codegen pattern.
- Extended the generator so required set attributes remain valid for generated schemas and acceptance-test config.
- Updated Vagrant and test documentation to provide free, non-management `em4`/`em5` LAGG member interfaces.
- Live-validated create/read/import/delete through `TestAccLagg_basic` against the Vagrant OPNsense VM.
- Updated provider docs, examples, support matrix, roadmap, and resource-gap planning references from Coming/live-validation-gated to Supported.

### File List

- `_bmad-output/implementation-artifacts/17-2-interface-lagg-resource.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/resource-gap-verification.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `docs/data-sources/interface_lagg.md`
- `docs/index.md`
- `docs/resources/interface_lagg.md`
- `examples/resources/opnsense_interface_lagg/resource.tf`
- `internal/generate/main.go`
- `internal/generate/templates.go`
- `internal/generate/schemas/interfaces.yaml`
- `internal/service/iface/exports.go`
- `internal/service/iface/lagg_data_source.gen.go`
- `internal/service/iface/lagg_model.gen.go`
- `internal/service/iface/lagg_resource.gen.go`
- `internal/service/iface/lagg_resource_gen_test.go`
- `internal/service/iface/lagg_schema.gen.go`
- `pkg/opnsense/crud.go`
- `pkg/opnsense/crud_test.go`
- `pkg/opnsense/errors.go`
- `templates/index.md.tmpl`
- `templates/resources/interface_lagg.md.tmpl`
- `test/README.md`
- `test/Vagrantfile`

## Change Log

- 2026-06-15: Created implementation story for Vagrant-validated `opnsense_interface_lagg` resource.
- 2026-06-15: Implemented and live-validated generated `opnsense_interface_lagg`; updated support/docs/status to Supported.
