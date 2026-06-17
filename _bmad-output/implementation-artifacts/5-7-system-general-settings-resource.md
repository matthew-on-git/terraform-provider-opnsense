---
baseline_commit: fbaf085e8287ae4f00f786484cd1ab622d77716a
---

# Story 5.7: System General Settings Resource API Revalidation Gate

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a provider maintainer,
I want to revalidate whether OPNsense now exposes a stable API for system general settings,
so that we either unblock a safe future `opnsense_system_general` singleton implementation or keep the upstream-blocked register honest with current evidence.

## Acceptance Criteria

1. **Given** the current target OPNsense release and upstream source state
   **When** the developer investigates system general settings support
   **Then** the story records whether a stable upstream API exists, including endpoint paths, controller/model names, wrapper keys, durable fields, defaults, validation constraints, read/update/import semantics, reconfigure/apply behavior, and release/version evidence

2. **And** if no stable target-release system general settings API exists, the developer must **not** implement `opnsense_system_general`; instead they update the relevant planning/public docs with the latest blocked evidence

3. **And** if a stable target-release API has shipped, the developer creates a follow-up implementation story or updates this story with exact implementation scope before writing provider code

4. **And** the investigation explicitly rejects unsafe substitutes: direct `config.xml` writes, legacy UI forms, setup-wizard-only APIs, private helpers, system status/actions, and unrelated singleton resources such as Unbound, Dnsmasq, Kea, FRR, or tunables

5. **And** all touched docs/status files remain aligned: `docs/upstream-blocked.md`, `docs/migration-import.md`, `docs/index.md` / `templates/index.md.tmpl`, `_bmad-output/planning-artifacts/support-matrix.md`, `_bmad-output/planning-artifacts/core-config-gap-analysis.md`, `_bmad-output/planning-artifacts/feature-complete-roadmap.md`, `_bmad-output/planning-artifacts/prd.md`, and `_bmad-output/implementation-artifacts/sprint-status.yaml`

6. **And** `make check` passes before completion

## Tasks / Subtasks

- [ ] Task 1: Revalidate upstream API availability (AC: #1, #2, #3)
  - [ ] 1.1 Check published OPNsense API docs for any durable system general settings controller/endpoints under `core`, `system`, or related modules
  - [ ] 1.2 Check current OPNsense `master` and target `stable/26.1` source for system settings controllers, models, forms, ACL/menu entries, and configd actions
  - [ ] 1.3 If endpoints appear, capture HTTP methods, paths, singleton ID/import semantics, wrapper/monad keys, model root, request/response payload shape, validation constraints, defaults, and reconfigure/apply behavior
  - [ ] 1.4 If endpoints are absent, setup-wizard-only, action/status-only, or not in the target release, record the blocker and stop before provider implementation
- [ ] Task 2: Compare existing provider coverage to prevent duplicate work (AC: #4)
  - [ ] 2.1 Confirm existing singleton resources already own their service domains, including Unbound general, Dnsmasq settings, Kea control-agent/settings, FRR/Quagga general, RIP/static general, and tunables when it ships
  - [ ] 2.2 Confirm `opnsense_system_info` remains a read-only data source and is not a durable system general settings resource
  - [ ] 2.3 Confirm this story is about base system settings only: hostname, domain, DNS server list, DNS override behavior, language, timezone, and NTP if upstream groups it there
- [ ] Task 3: Update public/planning docs based on evidence (AC: #2, #5)
  - [ ] 3.1 If still blocked, refresh blocked-register/support docs with current target-release and upstream evidence
  - [ ] 3.2 If unblocked, update docs to move system general settings from Upstream-blocked to Coming and create precise implementation handoff notes
- [ ] Task 4: Update sprint/planning status consistently (AC: #5)
  - [ ] 4.1 If still blocked, mark this story done as an upstream-blocked revalidation record and leave implementation work uncreated
  - [ ] 4.2 If unblocked, create or update a follow-up implementation story before marking this story complete
- [ ] Task 5: Run `make check` (AC: #6)

## Dev Notes

### Current Classification

This story is a **revalidation gate**, not a build-now resource story. The original Epic 5 requirement asked for `opnsense_system_general` as a singleton resource for hostname, domain, DNS servers, and NTP-style base system configuration. Current planning artifacts classify this domain as **Upstream-blocked** because no stable target-release MVC/API contract for durable system general settings has been confirmed.

Do **not** implement provider code against legacy config XML, UI-only forms, setup-wizard-only APIs, private helpers, local patches, status/action endpoints, or speculative System Settings MVC work. Terraform cannot safely manage base system settings without upstream-supported singleton get/set semantics and clear read-back behavior.

### Current Upstream Evidence Snapshot

Evidence checked during story creation on 2026-06-14:

| Source | Evidence |
|---|---|
| Published API docs: `development/api/core/core.html` | Lists core controllers including `backup`, `dashboard`, `defaults`, `hasync`, `initial_setup`, `menu`, `service`, `snapshots`, `system`, and `tunables`. No durable system general settings controller is listed. `core/system` exposes only status/action endpoints: `dismiss_status`, `halt`, `reboot`, and `status`. |
| Published API docs: `development/api/core/system.html` | Checked and returned 404; there is no separate documented system settings API page. |
| OPNsense `master`: `OPNsense/Core/Api/SettingsController.php` | Checked and returned 404; no generic Core Settings API controller was found at this path. |
| OPNsense `stable/26.1`: `OPNsense/Core/Api/SettingsController.php` | Checked and returned 404; no target-release generic Core Settings API controller was found at this path. |
| OPNsense `master` and `stable/26.1`: `OPNsense/Core/Api/SystemController.php` | Exists, but exposes system actions/status only (`halt`, `reboot`, `status`, `dismissStatus`). It is not a durable configuration get/set API. |
| OPNsense `master` and `stable/26.1`: `OPNsense/Core/Api/InitialSetupController.php` | Exists and uses `OPNsense\Core\InitialSetup`, but it is setup-wizard scoped. `configureAction()` updates multiple domains and removes `trigger_initial_wizard`; it is not safe as an ongoing Terraform-managed system general singleton. |
| OPNsense `master` and `stable/26.1`: `OPNsense/Core/InitialSetup.xml` | Model is mounted at `:memory:` and contains wizard fields for hostname, domain, language, DNS servers, DNS override, Unbound, timezone, WAN/LAN, deployment type, and root password. This is not durable System Settings MVC model evidence. |
| OPNsense `master`: `OPNsense/Core/System.xml` and `OPNsense/Core/Settings.xml` | Checked and returned 404; no durable Core System/Settings model was found at these paths. |

Treat this evidence as a starting point. Re-check live/current source during implementation because upstream status can change.

### Initial Setup API Is Not a Safe Substitute

The initial setup wizard currently looks tempting because it reads and writes some general settings, but it must not be used for `opnsense_system_general`:

| Concern | Evidence / risk |
|---|---|
| Model persistence | `InitialSetup.xml` mounts `:memory:`, then `InitialSetup.php` manually flushes selected fields into multiple config domains. This is wizard orchestration, not a durable singleton model contract. |
| Side effects | `updateConfig()` flushes general settings, WAN, LAN, root password, deployment-type settings, tunables, Unbound/Dnsmasq interactions, gateway changes, and removes the initial wizard trigger. |
| Lifecycle semantics | `configureAction()` is intended to complete initial setup, not repeated day-2 Terraform updates. Import/read-only refresh semantics for an always-existing singleton are not established. |
| Safety | Replaying wizard behavior from Terraform could unexpectedly mutate interface configuration, root credentials, DNS resolver settings, gateway state, DHCP/Dnsmasq/Unbound interactions, or initial-wizard state. |

### If Upstream Has Shipped the API

Before any provider implementation, capture all of the following in this story or a new implementation story:

- Resource name decision: likely `opnsense_system_general` only if it maps to a durable singleton for base system settings and not wizard orchestration.
- Endpoint table: singleton get/set paths, HTTP methods, import identity, and any reconfigure/apply endpoint.
- Monad/wrapper keys and model root path.
- Field model: hostname, domain, DNS server list, DNS override behavior, language, timezone, and NTP servers only if upstream includes NTP in the same stable model.
- Required fields, defaults, validation constraints, and maximum list sizes.
- Read-back behavior after update, including whether blank values represent defaults or true empty configuration.
- Update semantics, side effects, and whether changing hostname/domain/DNS requires service reloads or reboot warnings.
- Import semantics: singleton ID string such as `system`, appliance UUID, or no-op import marker.
- Permission/ACL requirements.
- Live validation plan on disposable OPNsense only; base system settings can affect resolver behavior, service reloads, identity, certificates, and API reachability.

### Existing Provider Coverage to Preserve

Already-supported system-adjacent resources and data sources:

| Domain | Provider resource(s) | Notes |
|---|---|---|
| System information | `opnsense_system_info` data source | Read-only appliance/firmware information, not durable settings. |
| Routing basics | `opnsense_system_gateway`, `opnsense_system_route`, `opnsense_system_vlan`, `opnsense_system_vip` | Separate resource domains already implemented under `internal/service/system`. |
| DNS resolver/forwarder settings | `opnsense_unbound_general`, `opnsense_dnsmasq_settings` | Service-specific DNS configuration, not base system DNS server list. |
| Kea/FRR/Quagga singletons | `opnsense_kea_ctrl_agent`, `opnsense_quagga_general`, `opnsense_quagga_rip`, `opnsense_quagga_static`, and related singletons | Separate service configuration singletons. |
| Tunables/sysctl | Coming with safety/live-validation gate | Persistent kernel/network tunables, not base system general settings. |

Do not create alternate resources for these. Keep the blocked-domain boundary clear: service-specific settings are supported where stable APIs exist; base system general settings remain blocked unless upstream exposes a durable stable API.

### Architecture and Implementation Guardrails

If this story becomes unblocked and implementation is explicitly authorized by captured endpoint evidence, follow the established singleton resource pattern:

- Service package: `internal/service/system` unless upstream module evidence points elsewhere.
- Resource name: `opnsense_system_general` unless the implementation handoff records a better name.
- Files: `general_resource.go`, `general_schema.go`, `general_model.go`, `general_resource_test.go`, plus `exports.go` registration.
- API client: use existing `pkg/opnsense` singleton helpers and `ReqOpts`; do not create raw HTTP calls in resource code.
- Lifecycle: no Create or Delete for a singleton. Implement Read, Update, and ImportState with a stable singleton ID.
- State: always read back from API after update; never set state from plan.
- Reconfigure: route to the endpoint confirmed by upstream evidence; do not invent a reload path.
- Tests: include import, update, read-back, drift behavior, and validation/default handling. Acceptance tests must use disposable appliances or isolated values that do not break management access.
- Documentation: add templates, examples, generated docs, and migration guidance only when the resource actually ships.

### What NOT to Build

- Do not implement `opnsense_system_general` while system general settings API support remains wizard-only, UI-only, legacy-config-only, or absent from the target release.
- Do not parse or write `config.xml` directly.
- Do not call `core/initial_setup/configure` from Terraform for day-2 configuration.
- Do not treat `core/system/status`, `halt`, `reboot`, or `dismiss_status` as configuration management endpoints.
- Do not merge unrelated service settings into this resource just because the setup wizard touches them.
- Do not add backward-compatibility shims for speculative API shapes.

### Previous Story Intelligence

- Story 5.1 established the pattern for upstream-blocked revalidation when target-release API support is missing.
- Story 5.6 established the current Epic 5 revalidation-gate pattern: capture upstream evidence, preserve provider boundaries, update blocked-domain docs, and avoid unsafe implementation.
- Singleton support exists in the provider, but singleton infrastructure alone is not enough. This story needs a stable upstream singleton settings API before resource code is safe.

### Testing / Validation Requirements

- If still blocked: documentation/status-only changes plus `make check` are sufficient.
- If unblocked and implementation is explicitly authorized by captured endpoint evidence: follow standard singleton provider resource requirements: unit tests for model/API conversion, resource lifecycle tests, import behavior, examples/templates/docs, `go generate ./tools`, and `make check`.
- Acceptance testing for any future implementation must avoid breaking hostname/DNS/API reachability on a production appliance. Use disposable test appliances.

### References

- [Source: `_bmad-output/planning-artifacts/epics.md` Story 5.7]
- [Source: `_bmad-output/planning-artifacts/prd.md` FR46]
- [Source: `_bmad-output/planning-artifacts/core-config-gap-analysis.md` System, Access, and Trust]
- [Source: `_bmad-output/planning-artifacts/support-matrix.md` Upstream-Blocked]
- [Source: `_bmad-output/planning-artifacts/feature-complete-roadmap.md` Two-front upstream contribution track]
- [Source: `docs/upstream-blocked.md` blocked domains register]
- [Source: `docs/migration-import.md` upstream-blocked migration guidance]
- [Source: `_bmad-output/implementation-artifacts/5-1-interface-resource.md`]
- [Source: `_bmad-output/implementation-artifacts/5-6-gateway-group-resource.md`]
- [Source: OPNsense published API docs: `development/api/core/core.html`]
- [Source: OPNsense upstream source paths checked under `src/opnsense/mvc/app/controllers/OPNsense/Core/Api/` and `src/opnsense/mvc/app/models/OPNsense/Core/`]

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- 2026-06-14: `git rev-parse HEAD` captured baseline `fbaf085e8287ae4f00f786484cd1ab622d77716a`.
- 2026-06-14: Published OPNsense core API docs checked; no durable system general settings endpoint listed.
- 2026-06-14: Published `development/api/core/system.html` checked and returned 404.
- 2026-06-14: OPNsense `master` and `stable/26.1` `Core/Api/SettingsController.php` checked and returned 404.
- 2026-06-14: OPNsense `master` and `stable/26.1` `Core/Api/SystemController.php` checked; action/status-only controller, not durable settings API.
- 2026-06-14: OPNsense `master` and `stable/26.1` `Core/Api/InitialSetupController.php`, `InitialSetup.xml`, and `InitialSetup.php` checked; wizard-only, memory-mounted orchestration model with broad side effects.
- 2026-06-14: OPNsense `master` `Core/System.xml` and `Core/Settings.xml` checked and returned 404.

### Completion Notes List

- Created a revalidation-gate story for system general settings that prevents unsafe implementation while the stable target-release API remains unavailable.
- Captured why the initial setup wizard cannot be used as the provider's day-2 singleton system general settings API.
- Preserved existing provider boundaries for system info, routing, DNS service settings, service singletons, and tunables/sysctl.

### File List

- `_bmad-output/implementation-artifacts/5-7-system-general-settings-resource.md`
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

- 2026-06-14: Created system general settings API revalidation-gate story; no provider implementation created.
