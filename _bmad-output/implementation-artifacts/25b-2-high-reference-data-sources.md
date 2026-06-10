---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 25B.2: High-Reference Data Sources

Status: done

## Story

As a provider maintainer,
I want read-only data sources for the highest-reference supported resources,
so that operators can compose Terraform configurations and brownfield imports without hardcoding OPNsense UUIDs.

## Acceptance Criteria

1. Data sources are implemented and registered for all Batch 1 candidates from `_bmad-output/planning-artifacts/data-source-parity-plan.md`: `haproxy_server`, `haproxy_backend`, `haproxy_frontend`, `haproxy_acl`, `haproxy_healthcheck`, `firewall_category`, `firewall_filter_rule`, `firewall_nat_port_forward`, `firewall_nat_outbound`, `system_vlan`, `system_vip`, `system_route`, `system_gateway`, `wireguard_server`, `wireguard_peer`, `openvpn_instance`, `openvpn_client_overwrite`, `ipsec_connection`, `ipsec_child`, `ipsec_local`, and `ipsec_remote`.
2. Each new data source follows the provider's existing lookup contract: `id` is the required UUID lookup key and all resource fields returned by OPNsense are computed read-only outputs.
3. Data-source implementation reuses existing resource models, `fromAPI` conversions, request options, and OPNsense `Get` calls; it does not duplicate API model logic, add fuzzy/name search semantics, or change resource behavior.
4. Sensitive or write-only resource fields are not exposed as misleading computed values. Any field omitted from a data source because the API cannot return it is documented in the data-source schema/docs.
5. New data sources are included in each affected service module's `DataSources()` registration and provider aggregation picks them up through existing module registration.
6. Terraform Registry docs are generated for every new data source, with example HCL under `examples/data-sources/opnsense_<type>/data-source.tf` and generated docs under `docs/data-sources/<type>.md`.
7. Tests cover the new data-source implementation pattern for each affected family or special case, including at least one representative per domain and direct coverage for any sensitive/write-only omission behavior.
8. `_bmad-output/planning-artifacts/data-source-parity-plan.md` is updated to mark Batch 1 complete or otherwise remove/move completed Batch 1 rows without changing Batch 2, Batch 3, or Batch 4 scope.
9. No singleton/special-case Batch 4 data sources are implemented in this story.
10. `make check` passes.

## Tasks / Subtasks

- [x] Task 1: Reconfirm Batch 1 inventory before implementation (AC: 1, 8, 9)
  - [x] Verify that the 21 Batch 1 data sources listed in AC1 still have matching resources and no existing data-source docs.
  - [x] Confirm `system_info` remains standalone and is not counted as a `system_*` resource data source.
  - [x] Record any discovered mismatch in the Dev Agent Record before changing code.
- [x] Task 2: Add HAProxy data sources (AC: 1, 2, 3, 5, 6, 7)
  - [x] Add `server_data_source.go`, `backend_data_source.go`, `frontend_data_source.go`, `acl_data_source.go`, and `healthcheck_data_source.go` in `internal/service/haproxy/`.
  - [x] Reuse `ServerResourceModel`, `BackendResourceModel`, `FrontendResourceModel`, `ACLResourceModel`, `HealthcheckResourceModel`, the matching `fromAPI` methods, and existing request options.
  - [x] Register constructors in `internal/service/haproxy/exports.go`.
  - [x] Add docs templates/examples and representative acceptance coverage.
- [x] Task 3: Add firewall data sources (AC: 1, 2, 3, 5, 6, 7)
  - [x] Add `category_data_source.go`, `filter_rule_data_source.go`, `nat_port_forward_data_source.go`, and `nat_outbound_data_source.go` in `internal/service/firewall/`.
  - [x] Reuse existing models/conversions and preserve firewall filter savepoint behavior by keeping these read-only lookups as plain `Get` calls only.
  - [x] Register constructors in `internal/service/firewall/exports.go` without disturbing existing `alias` and `nat_one_to_one` data sources.
  - [x] Add docs templates/examples and representative acceptance coverage.
- [x] Task 4: Add system data sources (AC: 1, 2, 3, 5, 6, 7)
  - [x] Add `vlan_data_source.go`, `vip_data_source.go`, `route_data_source.go`, and `gateway_data_source.go` in `internal/service/system/`.
  - [x] Reuse existing models/conversions and preserve `system_info` as a separate singleton data source.
  - [x] Treat `system_vip.password` carefully: do not expose a computed secret if OPNsense does not return it reliably.
  - [x] Register constructors in `internal/service/system/exports.go`.
  - [x] Add docs templates/examples and representative acceptance coverage.
- [x] Task 5: Add WireGuard and OpenVPN data sources (AC: 1, 2, 3, 4, 5, 6, 7)
  - [x] Add `server_data_source.go` and `peer_data_source.go` in `internal/service/wireguard/`.
  - [x] Add `instance_data_source.go` and `client_overwrite_data_source.go` in `internal/service/openvpn/`.
  - [x] Omit or document write-only/sensitive fields such as WireGuard `private_key` and any OpenVPN key/certificate fields that are not returned by the API.
  - [x] Register constructors in `wireguard/exports.go` and `openvpn/exports.go`.
  - [x] Add docs templates/examples and targeted sensitive-field coverage.
- [x] Task 6: Add core IPsec data sources (AC: 1, 2, 3, 5, 6, 7)
  - [x] Add `connection_data_source.go`, `child_data_source.go`, `local_data_source.go`, and `remote_data_source.go` in `internal/service/ipsec/`.
  - [x] Reuse existing hand-written IPsec models/conversions and do not implement Batch 4 `ipsec_psk` or `ipsec_key_pair` in this story.
  - [x] Register constructors in `internal/service/ipsec/exports.go` without disturbing existing `manual_spd`, `pool`, and `vti` generated data sources.
  - [x] Add docs templates/examples and representative acceptance coverage.
- [x] Task 7: Update docs, generated docs, and parity plan (AC: 6, 8)
  - [x] Add `templates/data-sources/<type>.md.tmpl` and `examples/data-sources/opnsense_<type>/data-source.tf` for each new data source.
  - [x] Run the repository docs generation path used by `make check` if needed so `docs/data-sources/<type>.md` files exist.
  - [x] Update `_bmad-output/planning-artifacts/data-source-parity-plan.md` to reflect Batch 1 completion while preserving unresolved Batch 2/3/4 items.
- [x] Task 8: Validate (AC: 10)
  - [x] Run targeted searches for Batch 1 registration/docs/examples.
  - [x] Run `make check` and fix all failures without suppressing checks.

### Review Findings

- [x] [Review][Patch] Add runtime read coverage for new data sources [internal/service/*/data_source_schema_test.go] — Resolved with local HTTP-backed `Read` tests for HAProxy, firewall, system VIP, WireGuard, OpenVPN, and IPsec representative data sources.
- [x] [Review][Patch] Document omitted sensitive/write-only fields in data-source docs [templates/data-sources/wireguard_server.md.tmpl, templates/data-sources/system_vip.md.tmpl] — Resolved with template/generated docs notes for `wireguard_server.private_key` and `system_vip.password` omissions.
- [x] [Review][Patch] Update provider index support counts [templates/index.md.tmpl, docs/index.md] — Resolved with 55 data-source and 55 data-source docs counts plus 36 remaining parity gaps.

## Dev Notes

### Scope Boundary

This story implements Batch 1 only from `_bmad-output/planning-artifacts/data-source-parity-plan.md`. Do not implement Batch 2 routing data sources, Batch 3 DNS/DHCP/ACME/Trust data sources, or Batch 4 singleton/sensitive special cases.

Do not add name-based, fuzzy, or search-by-field lookup behavior. Current provider data sources use UUID lookup via required `id`; changing that contract would require a separate product/design decision.

Do not migrate hand-written resources to the generator just to add data sources. Prefer the smallest correct change: add hand-written data-source files beside existing hand-written resource files unless a resource is already generated and the generator can be reused without broad refactoring.

### Batch 1 Source of Truth

Story 25B.1 produced the current parity plan and corrected the resource-matching backlog to 57 because `system_info` is standalone. Batch 1 is the high-reference group with the highest migration/composition value.

Batch 1 rows to implement:

| Data source | Files to inspect | Risk note |
|---|---|---|
| `haproxy_server` | `internal/service/haproxy/server_*`, `internal/service/haproxy/exports.go` | Referenced by backend chains. |
| `haproxy_backend` | `internal/service/haproxy/backend_*` | Referenced by frontends. |
| `haproxy_frontend` | `internal/service/haproxy/frontend_*` | Linked ACL/backend fields. |
| `haproxy_acl` | `internal/service/haproxy/acl_*` | Frontend routing references. |
| `haproxy_healthcheck` | `internal/service/haproxy/healthcheck_*` | Backend health check references. |
| `firewall_category` | `internal/service/firewall/category_*` | Category references in firewall objects. |
| `firewall_filter_rule` | `internal/service/firewall/filter_rule_*` | Safety-critical domain; read-only lookup only. |
| `firewall_nat_port_forward` | `internal/service/firewall/nat_port_forward_*` | NAT migration. |
| `firewall_nat_outbound` | `internal/service/firewall/nat_outbound_*` | NAT migration. |
| `system_vlan` | `internal/service/system/vlan_*` | Interface naming may vary by appliance. |
| `system_vip` | `internal/service/system/vip_*` | CARP/password fields may require care. |
| `system_route` | `internal/service/system/route_*` | Gateway selected map conversion. |
| `system_gateway` | `internal/service/system/gateway_*` | Gateway selected map conversion. |
| `wireguard_server` | `internal/service/wireguard/server_*` | Private key is sensitive/write-only. |
| `wireguard_peer` | `internal/service/wireguard/peer_*` | Peer references. |
| `openvpn_instance` | `internal/service/openvpn/instance_*` | Key/cert fields may be write-only. |
| `openvpn_client_overwrite` | `internal/service/openvpn/client_overwrite_*` | Instance linkage. |
| `ipsec_connection` | `internal/service/ipsec/connection_*` | Parent of children. |
| `ipsec_child` | `internal/service/ipsec/child_*` | Connection linkage. |
| `ipsec_local` | `internal/service/ipsec/local_*` | Connection linkage. |
| `ipsec_remote` | `internal/service/ipsec/remote_*` | Connection linkage. |

### Existing Implementation Pattern to Reuse

Hand-written UUID lookup pattern: `internal/service/firewall/alias_data_source.go` implements `opnsense_firewall_alias` by requiring `id`, reading config into the existing resource model, calling `opnsense.Get[...]` with the resource request options, calling `fromAPI`, and setting state.

Generated UUID lookup pattern: `internal/service/iface/bridge_data_source.gen.go`, `internal/service/ipsec/pool_data_source.gen.go`, and `internal/service/firewall/nat_one_to_one_data_source.gen.go` use the same contract with `id` required and all other attributes computed.

Generator behavior: `internal/generate/main.go` only emits a data source for generated resources with `kind: item`. It does not emit generated data sources for singleton resources. `internal/generate/templates.go` reuses the resource model, request options, API response type, and `fromAPI` conversion.

Current generated-schema coverage relevant to Batch 1 is limited. `internal/generate/schemas/firewall.yaml` only covers `nat_one_to_one`, which is already implemented as a data source and is not in Batch 1. `internal/generate/schemas/ipsec.yaml` covers `pool`, `vti`, and `manual_spd`, which are already implemented as data sources and are not in Batch 1. Most Batch 1 resources are hand-written today.

### Current Registration State

Affected modules currently register these data sources:

| Module | Current data-source registrations | Expected addition |
|---|---|---|
| `haproxy` | none | 5 new data sources |
| `firewall` | `newAliasDataSource`, `newNatOneToOneDataSource` | 4 new data sources |
| `system` | `newSystemInfoDataSource` | 4 new data sources |
| `wireguard` | none | 2 new data sources |
| `openvpn` | none | 2 new data sources |
| `ipsec` | `newManualSPDDataSource`, `newPoolDataSource`, `newVTIDataSource` | 4 new data sources |

Provider aggregation is module-based; update the module `exports.go` files only. Do not manually register individual data sources in `internal/provider/` unless existing provider aggregation has changed.

### Sensitive and Write-Only Field Guardrails

Resources with known sensitive/write-only risk in this batch:

| Data source | Field risk | Required behavior |
|---|---|---|
| `wireguard_server` | `private_key` is documented as write-only/not returned by API | Do not expose a fake computed private key. Omit it or document why it is unavailable. |
| `system_vip` | CARP `password` may be sensitive and may not round-trip reliably | Verify actual read behavior before exposing; do not invent state. |
| `openvpn_instance` | certificate/key references or material may be sensitive/write-only depending on schema | Expose returned references; do not expose unavailable secrets. |

If omitting a resource attribute from a data-source schema, keep the omission intentional and document it in schema Markdown and the docs template where useful. Do not change the corresponding resource schema unless required to fix an existing bug discovered during implementation.

### Architecture and Quality Guardrails

- Use Go, Terraform Plugin Framework v6, and existing provider patterns. Do not add dependencies.
- Keep data source files beside their matching resource files under `internal/service/<module>/`.
- File names must be lowercase underscore-separated: `<resource>_data_source.go` for hand-written files.
- Constructors must follow existing naming: `newServerDataSource()`, `newFilterRuleDataSource()`, etc.
- `Metadata` type names must be `req.ProviderTypeName + "_<module>_<resource>"`, matching resource/doc basenames.
- `Configure` must reuse provider data as `*opnsense.Client`; never create a new client.
- `Read` must use `opnsense.Get[...]` with existing request options and must surface errors through diagnostics.
- `fromAPI` receives a clean unwrapped API struct; do not unwrap OPNsense monads in data-source code.
- All Terraform attributes must be snake_case and should match the resource attribute names except intentional omissions for unavailable sensitive/write-only fields.
- Avoid helper extraction unless genuinely shared across multiple new data sources and clearer than keeping the code local.

### Documentation Requirements

Existing data-source docs pattern:

- Template: `templates/data-sources/firewall_alias.md.tmpl`
- Example: `examples/data-sources/opnsense_firewall_alias/data-source.tf`
- Generated doc: `docs/data-sources/firewall_alias.md`

For each Batch 1 data source, add a template and example with UUID lookup by `id`. Examples should reference plausible UUID placeholders or show use with an existing resource when that is clearer. Do not claim name-based lookup.

### Testing Requirements

Required final command: `make check`.

Targeted checks before `make check`:

```bash
rg "new.*DataSource" internal/service/haproxy internal/service/firewall internal/service/system internal/service/wireguard internal/service/openvpn internal/service/ipsec
rg "^# opnsense_(haproxy|firewall|system|wireguard|openvpn|ipsec)_" docs/data-sources
rg "data \"opnsense_(haproxy|firewall|system|wireguard|openvpn|ipsec)_" examples/data-sources
```

Follow existing acceptance patterns from `internal/service/firewall/alias_data_source_test.go` and `internal/service/ipsec/pool_data_source_test.go`: create or reference the managed resource, add a data source with `id = <resource>.id`, and assert representative attributes match.

Do not add tests that require secrets to round-trip when the API does not return them. For sensitive/write-only cases, assert the documented omission or returned non-secret references.

### Previous Story Intelligence

Story 25B.1 created the parity plan and found that raw `90 - 34 = 56` was wrong for resource-matching parity because `system_info` is standalone. Treat 57 as the current backlog until implementation changes it.

Story 25B.1 also clarified that Batch 4 is intentionally outside Stories 25B.2 and 25B.3. If a Batch 1 candidate reveals singleton or misleading-secret behavior, document the reason and move it to Batch 4 rather than forcing an unsafe data source.

Recent release commits matter because the provider is already published:

- `b82ada3 chore(release): mark v0.1.0 as released and fix GPG signing config`
- `a2bc1d4 docs(changelog): add v0.1.0 provider CHANGELOG in Terraform format (Epic 12-5)`
- `b2ea1c6 ci(release): add release + acceptance workflows, structure validator (Epic 12-4)`

### References

- `_bmad-output/planning-artifacts/data-source-parity-plan.md` — Batch 1 source of truth, verification commands, and Batch 4 boundary.
- `_bmad-output/planning-artifacts/post-release-epics.md` — Epic 25B sequence and positioning.
- `_bmad-output/planning-artifacts/prd.md` — FR60 data-source parity, current 90/34/57 baseline, migration/import user journey.
- `_bmad-output/planning-artifacts/architecture.md` — service-module structure, data-source naming, API client boundaries, docs/testing standards, anti-patterns.
- `internal/generate/main.go` and `internal/generate/templates.go` — generated item data-source behavior.
- `internal/service/firewall/alias_data_source.go` — hand-written UUID lookup example.
- `internal/service/iface/bridge_data_source.gen.go` and `internal/service/ipsec/pool_data_source.gen.go` — generated UUID lookup examples.
- `internal/service/firewall/alias_data_source_test.go` and `internal/service/ipsec/pool_data_source_test.go` — acceptance test examples.
- `templates/data-sources/firewall_alias.md.tmpl` and `examples/data-sources/opnsense_firewall_alias/data-source.tf` — documentation/example pattern.

## Project Structure Notes

Expected implementation file locations:

- `internal/service/haproxy/*_data_source.go`, `internal/service/haproxy/exports.go`
- `internal/service/firewall/*_data_source.go`, `internal/service/firewall/exports.go`
- `internal/service/system/*_data_source.go`, `internal/service/system/exports.go`
- `internal/service/wireguard/*_data_source.go`, `internal/service/wireguard/exports.go`
- `internal/service/openvpn/*_data_source.go`, `internal/service/openvpn/exports.go`
- `internal/service/ipsec/*_data_source.go`, `internal/service/ipsec/exports.go`
- `templates/data-sources/*.md.tmpl`
- `examples/data-sources/opnsense_*/data-source.tf`
- `docs/data-sources/*.md`
- `_bmad-output/planning-artifacts/data-source-parity-plan.md`

Files to inspect before modifying:

- Existing matching `*_resource.go`, `*_schema.go`, `*_model.go`, and `*_resource_test.go` files for each Batch 1 candidate.
- Existing generated data sources listed above to keep behavior consistent.
- `internal/provider/provider.go` only to confirm module aggregation if registration behavior is unclear.

Potential conflicts or variances:

- Architecture docs still describe older counts in some historical sections; use the current parity plan/PRD/support matrix for counts.
- The `schema/` path in historical architecture has moved to `internal/generate/schemas/` in the actual repository.
- `LICENSE` is MIT while source files use MPL SPDX headers; do not change license headers as part of this story.
- `sprint-status.yaml` does not contain post-release `25B.*` keys, so create-story should not update sprint status for this story unless sprint planning is refreshed first.

## Dev Agent Record

### Agent Model Used

gpt-5.5

### Debug Log References

- Batch 1 inventory check: 21 resources present, 0 existing Batch 1 data-source docs, `system_info` standalone confirmed.
- Registered data-source count after implementation: 55.
- Docs parity after implementation: 90 resource docs, 55 data-source docs, 36 remaining resource-matching gaps, standalone data source `system_info`.
- Targeted service package tests passed: `go test ./internal/service/haproxy ./internal/service/firewall ./internal/service/system ./internal/service/wireguard ./internal/service/openvpn ./internal/service/ipsec`.
- Targeted registration/docs/examples searches passed for Batch 1 domains.
- Final validation: `make check` passed.
- Review patch validation: targeted service package tests passed after adding runtime `Read` coverage; `make check` passed with lint, format, test, security, scan, and docs.

### Completion Notes List

- Implemented and registered all 21 Batch 1 UUID lookup data sources across HAProxy, firewall, system, WireGuard, OpenVPN, and core IPsec.
- Reused existing resource models, request options, API response types, `fromAPI` conversions, and `opnsense.Get` behavior where safe.
- Used dedicated data-source models for `wireguard_server` and `system_vip` so write-only/sensitive fields are omitted without setting schema-incompatible state.
- Added schema unit tests for all affected service modules, including omission checks for `wireguard_server.private_key` and `system_vip.password`.
- Added data-source examples/templates and generated `docs/data-sources` pages for all Batch 1 data sources.
- Updated the parity plan and support matrix to 55 data sources and 36 remaining resource-matching data-source gaps.
- Resolved code-review findings by adding representative runtime `Read` tests, documenting omitted sensitive/write-only fields, and correcting provider index support counts.

### Change Log

- 2026-06-02: Implemented Batch 1 high-reference data sources; story moved to review after `make check` passed.
- 2026-06-02: Applied code-review patches for runtime read coverage, sensitive-field documentation, and provider index counts; `make check` passed.
- 2026-06-02: Review closure confirmed all findings resolved; story marked done.

### File List

- `_bmad-output/implementation-artifacts/25b-2-high-reference-data-sources.md`
- `_bmad-output/planning-artifacts/data-source-parity-plan.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `docs/index.md`
- `templates/index.md.tmpl`
- `docs/data-sources/firewall_category.md`
- `docs/data-sources/firewall_filter_rule.md`
- `docs/data-sources/firewall_nat_outbound.md`
- `docs/data-sources/firewall_nat_port_forward.md`
- `docs/data-sources/haproxy_acl.md`
- `docs/data-sources/haproxy_backend.md`
- `docs/data-sources/haproxy_frontend.md`
- `docs/data-sources/haproxy_healthcheck.md`
- `docs/data-sources/haproxy_server.md`
- `docs/data-sources/ipsec_child.md`
- `docs/data-sources/ipsec_connection.md`
- `docs/data-sources/ipsec_local.md`
- `docs/data-sources/ipsec_remote.md`
- `docs/data-sources/openvpn_client_overwrite.md`
- `docs/data-sources/openvpn_instance.md`
- `docs/data-sources/system_gateway.md`
- `docs/data-sources/system_route.md`
- `docs/data-sources/system_vip.md`
- `docs/data-sources/system_vlan.md`
- `docs/data-sources/wireguard_peer.md`
- `docs/data-sources/wireguard_server.md`
- `examples/data-sources/opnsense_firewall_category/data-source.tf`
- `examples/data-sources/opnsense_firewall_filter_rule/data-source.tf`
- `examples/data-sources/opnsense_firewall_nat_outbound/data-source.tf`
- `examples/data-sources/opnsense_firewall_nat_port_forward/data-source.tf`
- `examples/data-sources/opnsense_haproxy_acl/data-source.tf`
- `examples/data-sources/opnsense_haproxy_backend/data-source.tf`
- `examples/data-sources/opnsense_haproxy_frontend/data-source.tf`
- `examples/data-sources/opnsense_haproxy_healthcheck/data-source.tf`
- `examples/data-sources/opnsense_haproxy_server/data-source.tf`
- `examples/data-sources/opnsense_ipsec_child/data-source.tf`
- `examples/data-sources/opnsense_ipsec_connection/data-source.tf`
- `examples/data-sources/opnsense_ipsec_local/data-source.tf`
- `examples/data-sources/opnsense_ipsec_remote/data-source.tf`
- `examples/data-sources/opnsense_openvpn_client_overwrite/data-source.tf`
- `examples/data-sources/opnsense_openvpn_instance/data-source.tf`
- `examples/data-sources/opnsense_system_gateway/data-source.tf`
- `examples/data-sources/opnsense_system_route/data-source.tf`
- `examples/data-sources/opnsense_system_vip/data-source.tf`
- `examples/data-sources/opnsense_system_vlan/data-source.tf`
- `examples/data-sources/opnsense_wireguard_peer/data-source.tf`
- `examples/data-sources/opnsense_wireguard_server/data-source.tf`
- `internal/service/firewall/category_data_source.go`
- `internal/service/firewall/data_source_schema_test.go`
- `internal/service/firewall/exports.go`
- `internal/service/firewall/filter_rule_data_source.go`
- `internal/service/firewall/nat_outbound_data_source.go`
- `internal/service/firewall/nat_port_forward_data_source.go`
- `internal/service/haproxy/acl_data_source.go`
- `internal/service/haproxy/backend_data_source.go`
- `internal/service/haproxy/data_source_schema_test.go`
- `internal/service/haproxy/exports.go`
- `internal/service/haproxy/frontend_data_source.go`
- `internal/service/haproxy/healthcheck_data_source.go`
- `internal/service/haproxy/server_data_source.go`
- `internal/service/ipsec/child_data_source.go`
- `internal/service/ipsec/connection_data_source.go`
- `internal/service/ipsec/data_source_schema_test.go`
- `internal/service/ipsec/exports.go`
- `internal/service/ipsec/local_data_source.go`
- `internal/service/ipsec/remote_data_source.go`
- `internal/service/openvpn/client_overwrite_data_source.go`
- `internal/service/openvpn/data_source_schema_test.go`
- `internal/service/openvpn/exports.go`
- `internal/service/openvpn/instance_data_source.go`
- `internal/service/system/data_source_schema_test.go`
- `internal/service/system/exports.go`
- `internal/service/system/gateway_data_source.go`
- `internal/service/system/route_data_source.go`
- `internal/service/system/vip_data_source.go`
- `internal/service/system/vlan_data_source.go`
- `internal/service/wireguard/data_source_schema_test.go`
- `internal/service/wireguard/exports.go`
- `internal/service/wireguard/peer_data_source.go`
- `internal/service/wireguard/server_data_source.go`
- `templates/data-sources/firewall_category.md.tmpl`
- `templates/data-sources/firewall_filter_rule.md.tmpl`
- `templates/data-sources/firewall_nat_outbound.md.tmpl`
- `templates/data-sources/firewall_nat_port_forward.md.tmpl`
- `templates/data-sources/haproxy_acl.md.tmpl`
- `templates/data-sources/haproxy_backend.md.tmpl`
- `templates/data-sources/haproxy_frontend.md.tmpl`
- `templates/data-sources/haproxy_healthcheck.md.tmpl`
- `templates/data-sources/haproxy_server.md.tmpl`
- `templates/data-sources/ipsec_child.md.tmpl`
- `templates/data-sources/ipsec_connection.md.tmpl`
- `templates/data-sources/ipsec_local.md.tmpl`
- `templates/data-sources/ipsec_remote.md.tmpl`
- `templates/data-sources/openvpn_client_overwrite.md.tmpl`
- `templates/data-sources/openvpn_instance.md.tmpl`
- `templates/data-sources/system_gateway.md.tmpl`
- `templates/data-sources/system_route.md.tmpl`
- `templates/data-sources/system_vip.md.tmpl`
- `templates/data-sources/system_vlan.md.tmpl`
- `templates/data-sources/wireguard_peer.md.tmpl`
- `templates/data-sources/wireguard_server.md.tmpl`
