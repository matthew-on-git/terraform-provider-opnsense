---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 25B.3: Routing, DNS, DHCP, ACME, and Trust Data Sources

Status: done

## Story

As a provider maintainer,
I want read-only UUID lookup data sources for the remaining straightforward routing, DNS, DHCP, ACME, Dynamic DNS, and Trust resources,
so that operators can reference existing OPNsense configuration during brownfield migration and Terraform composition without hardcoding unmanaged UUIDs.

## Acceptance Criteria

1. Data sources are implemented and registered for all Batch 2 candidates from `_bmad-output/planning-artifacts/data-source-parity-plan.md`: `quagga_bgp_neighbor`, `quagga_prefix_list`, `quagga_route_map`, `quagga_bgp_aspath`, `quagga_bgp_communitylist`, `quagga_bgp_peergroup`, `quagga_bgp_redistribution`, and `quagga_static_route`.
2. Data sources are implemented and registered for all Batch 3 candidates from `_bmad-output/planning-artifacts/data-source-parity-plan.md`: `unbound_host_override`, `unbound_host_alias`, `unbound_domain_override`, `unbound_acl`, `dhcpv4_subnet`, `dhcpv4_reservation`, `kea_dhcpv6_subnet`, `kea_dhcpv6_reservation`, `ddclient_account`, `acme_account`, `acme_certificate`, `acme_challenge`, and `trust_ca`.
3. Each new data source follows the provider's existing lookup contract: required UUID `id`, all returned resource attributes computed, no name/fuzzy/search-by-field lookup behavior.
4. Implementations reuse existing resource models, `fromAPI` conversions, request options, and `opnsense.Get` behavior where safe. Do not duplicate API model logic, inline conversion logic, or change resource behavior.
5. Sensitive/write-only fields are not exposed as misleading computed values. `ddclient_account.password` must be omitted from the data-source schema/docs unless live/API evidence proves it safely round-trips, and any omission must be documented.
6. New constructors are registered in the affected service module `DataSources()` functions and provider aggregation picks them up through existing module registration. Do not manually wire individual data sources in `internal/provider/` unless existing aggregation has changed.
7. Terraform Registry docs are generated for every new data source, with `templates/data-sources/<type>.md.tmpl`, `examples/data-sources/opnsense_<type>/data-source.tf`, and generated `docs/data-sources/<type>.md` files.
8. Tests cover constructor counts, required `id` schemas, representative runtime `Read` behavior for each affected family, and sensitive/write-only omission behavior for `ddclient_account.password`.
9. `_bmad-output/planning-artifacts/data-source-parity-plan.md`, `_bmad-output/planning-artifacts/support-matrix.md`, `templates/index.md.tmpl`, and generated `docs/index.md` are updated to reflect Batch 2/3 completion. Expected post-story counts are 76 data sources/docs and 15 remaining resource-matching data-source gaps if all Batch 2/3 items are completed.
10. Batch 4 singleton/sensitive special cases are not implemented in this story: `dnsmasq_settings`, `kea_ctrl_agent`, `kea_dhcpv6_settings`, `quagga_general`, `quagga_bgp_global`, `quagga_ospf_general`, `quagga_ospf6_general`, `quagga_rip`, `quagga_static`, `unbound_general`, `unbound_dnsbl`, `ipsec_psk`, `ipsec_key_pair`, `openvpn_static_key`, and `trust_cert`.
11. Any Batch 2/3 candidate that cannot be implemented safely because a read endpoint is absent, returns misleading data, or has unclear lookup semantics is documented in the parity plan with a concrete evidence-backed deferral reason.
12. `make check` passes.

## Tasks / Subtasks

- [x] Task 1: Reconfirm Batch 2/3 inventory before implementation (AC: 1, 2, 9, 10, 11)
  - [x] Verify the 21 Batch 2/3 candidates still have matching resources and lack existing data-source docs before adding files.
  - [x] Confirm current baseline remains 55 data sources/docs and 36 resource-matching gaps before implementation.
  - [x] Confirm Batch 4 candidates remain outside this story even if they are nearby in `exports.go` or resource files.
  - [x] Record any mismatch or deferral evidence in the Dev Agent Record before changing code.
- [x] Task 2: Add Quagga/FRR Batch 2 routing data sources (AC: 1, 3, 4, 6, 7, 8)
  - [x] Add data sources beside existing resources in `internal/service/quagga/` for `bgp_neighbor`, `prefix_list`, `route_map`, `bgp_aspath`, `bgp_communitylist`, `bgp_peergroup`, `bgp_redistribution`, and `static_route`.
  - [x] Reuse the corresponding `*ResourceModel`, API response type, request options, and `fromAPI` conversion for each data source.
  - [x] Register the new constructors in `internal/service/quagga/exports.go` without disturbing existing OSPF/OSPFv3 generated data sources.
  - [x] Add `internal/service/quagga/data_source_schema_test.go` or extend an existing equivalent with constructor count, required `id`, and at least one local HTTP-backed `Read` test for a representative routing data source.
- [x] Task 3: Add Unbound DNS data sources (AC: 2, 3, 4, 6, 7, 8, 10)
  - [x] Add `host_override_data_source.go`, `host_alias_data_source.go`, `domain_override_data_source.go`, and `acl_data_source.go` in `internal/service/unbound/`.
  - [x] Do not implement `unbound_general` or `unbound_dnsbl`; those singleton/settings resources are Batch 4.
  - [x] Register constructors in `internal/service/unbound/exports.go`.
  - [x] Add schema and representative local HTTP-backed `Read` coverage for Unbound.
- [x] Task 4: Add DHCPv4 and Kea DHCPv6 UUID data sources (AC: 2, 3, 4, 6, 7, 8, 10)
  - [x] Add `subnet_data_source.go` and `reservation_data_source.go` in `internal/service/dhcp/` for `dhcpv4_subnet` and `dhcpv4_reservation`.
  - [x] Add `dhcpv6_subnet_data_source.go` and `dhcpv6_reservation_data_source.go` in `internal/service/kea/`.
  - [x] Preserve the existing generated `kea_ha_peer` data source and do not implement `kea_ctrl_agent` or `kea_dhcpv6_settings` in this story.
  - [x] Register constructors in `internal/service/dhcp/exports.go` and `internal/service/kea/exports.go`.
  - [x] Add schema and representative local HTTP-backed `Read` coverage for DHCPv4 and Kea.
- [x] Task 5: Add Dynamic DNS account data source with password omission (AC: 2, 3, 4, 5, 6, 7, 8)
  - [x] Add `account_data_source.go` in `internal/service/ddclient/` for `ddclient_account`.
  - [x] Use a dedicated data-source model if necessary so `password` is omitted without producing schema-incompatible state.
  - [x] Document the omitted `password` field in schema Markdown and `templates/data-sources/ddclient_account.md.tmpl`.
  - [x] Register the constructor in `internal/service/ddclient/exports.go`.
  - [x] Add schema tests proving `password` is not exposed and a representative runtime `Read` test.
- [x] Task 6: Add ACME data sources (AC: 2, 3, 4, 6, 7, 8)
  - [x] Add `account_data_source.go`, `certificate_data_source.go`, and `challenge_data_source.go` in `internal/service/acme/`.
  - [x] Reuse existing ACME models/conversions and only expose attributes returned by the API.
  - [x] Register constructors in `internal/service/acme/exports.go`.
  - [x] Add schema and representative local HTTP-backed `Read` coverage for ACME.
- [x] Task 7: Add Trust CA data source only (AC: 2, 3, 4, 6, 7, 8, 10)
  - [x] Add `ca_data_source.go` in `internal/service/trust/` for `trust_ca`.
  - [x] Do not implement `trust_cert`; it is Batch 4 because private key/certificate material has write-only/sensitive behavior.
  - [x] Register the constructor in `internal/service/trust/exports.go`.
  - [x] Add schema and representative local HTTP-backed `Read` coverage for Trust CA.
- [x] Task 8: Add docs templates, examples, and generated docs (AC: 7)
  - [x] Add a UUID lookup example under `examples/data-sources/opnsense_<type>/data-source.tf` for each new data source.
  - [x] Add a template under `templates/data-sources/<type>.md.tmpl` for each new data source.
  - [x] Regenerate docs using the repository docs generation path. If Docker leaves generated docs owned by root, restore ownership before continuing.
- [x] Task 9: Update parity/support docs and provider index counts (AC: 9, 11)
  - [x] Update `_bmad-output/planning-artifacts/data-source-parity-plan.md` to mark Batch 2/3 complete or record evidence-backed deferrals.
  - [x] Update `_bmad-output/planning-artifacts/support-matrix.md` supported data-source domains and counts.
  - [x] Update `templates/index.md.tmpl` and generated `docs/index.md` counts and remaining-gap language.
- [x] Task 10: Validate (AC: 8, 12)
  - [x] Run targeted searches for Batch 2/3 registration, docs, and examples.
  - [x] Run targeted service tests for all affected packages.
  - [x] Run `make check` and fix all failures without suppressing checks.

### Review Findings

- [x] [Review][Patch] Correct import guidance for singleton/settings resource IDs [templates/index.md.tmpl:103, docs/index.md:135]
- [x] [Review][Patch] Reuse existing ddclient account `fromAPI` conversion in the data source [internal/service/ddclient/account_data_source.go:75]

## Dev Notes

### Scope Boundary

This story implements Batch 2 and Batch 3 from `_bmad-output/planning-artifacts/data-source-parity-plan.md`. It must not implement Batch 4 singleton/sensitive special cases. Batch 4 exists because singleton IDs and write-only secret behavior need explicit design decisions before implementation.

Do not add name-based, fuzzy, or search-by-field lookup. Current provider data sources use required UUID `id`; changing lookup semantics is a separate product/design decision.

Do not migrate hand-written resources to the generator just to add data sources. Prefer the smallest correct change: add hand-written data-source files beside existing hand-written resource files unless a target resource is already generated and the generator can be used without broad refactoring.

### Batch 2 Source of Truth

| Data source | Files to inspect | Risk note |
|---|---|---|
| `quagga_bgp_neighbor` | `internal/service/quagga/bgp_neighbor_*`, `internal/service/quagga/exports.go` | High brownfield BGP migration value. |
| `quagga_prefix_list` | `internal/service/quagga/prefix_list_*` | Referenced by route policy. |
| `quagga_route_map` | `internal/service/quagga/route_map_*` | Referenced by BGP/routing policy. |
| `quagga_bgp_aspath` | `internal/service/quagga/bgp_aspath_*` | Policy lookup. |
| `quagga_bgp_communitylist` | `internal/service/quagga/bgp_communitylist_*` | Policy lookup. |
| `quagga_bgp_peergroup` | `internal/service/quagga/bgp_peergroup_*` | Neighbor composition value. |
| `quagga_bgp_redistribution` | `internal/service/quagga/bgp_redistribution_*` | Routing migration. |
| `quagga_static_route` | `internal/service/quagga/static_route_*` | Static routing migration. |

Existing Quagga data sources are generated OSPF/OSPFv3 item data sources only. Preserve those registrations and add the Batch 2 constructors alongside them.

### Batch 3 Source of Truth

| Data source | Files to inspect | Risk note |
|---|---|---|
| `unbound_host_override` | `internal/service/unbound/host_override_*` | DNS migration. |
| `unbound_host_alias` | `internal/service/unbound/host_alias_*` | Host override linkage. |
| `unbound_domain_override` | `internal/service/unbound/domain_override_*` | Forward/domain lookup. |
| `unbound_acl` | `internal/service/unbound/acl_*` | ACL migration. |
| `dhcpv4_subnet` | `internal/service/dhcp/subnet_*` | DHCP migration. |
| `dhcpv4_reservation` | `internal/service/dhcp/reservation_*` | Static mapping migration. |
| `kea_dhcpv6_subnet` | `internal/service/kea/dhcpv6_subnet_*` | DHCPv6 migration. |
| `kea_dhcpv6_reservation` | `internal/service/kea/dhcpv6_reservation_*` | DHCPv6 migration. |
| `ddclient_account` | `internal/service/ddclient/account_*` | `password` is sensitive/write-only and should be omitted/documented. |
| `acme_account` | `internal/service/acme/account_*` | Account registration fields. |
| `acme_certificate` | `internal/service/acme/certificate_*` | Certificate status/issuance fields. |
| `acme_challenge` | `internal/service/acme/challenge_*` | Provider-specific fields; expose returned model fields only. |
| `trust_ca` | `internal/service/trust/ca_*` | Certificate material is returned; `trust_cert` remains out of scope. |

### Current Registration State

| Module | Current data-source registrations | Expected additions in this story |
|---|---|---|
| `quagga` | 11 generated OSPF/OSPFv3 item data sources | 8 Batch 2 routing data sources |
| `unbound` | none | 4 UUID-backed DNS data sources |
| `dhcp` | none | 2 DHCPv4 data sources |
| `kea` | `newHAPeerDataSource` | 2 DHCPv6 data sources |
| `ddclient` | none | 1 Dynamic DNS account data source |
| `acme` | none | 3 ACME data sources |
| `trust` | none | 1 Trust CA data source |

### Existing Implementation Patterns to Reuse

Hand-written UUID lookup pattern: `internal/service/haproxy/server_data_source.go` and `internal/service/firewall/alias_data_source.go` require `id`, read config into the resource model, call `opnsense.Get[...]` with existing request options, call `fromAPI`, and set state.

Generated UUID lookup pattern: `internal/service/quagga/ospf_network_data_source.gen.go` and `internal/service/kea/ha_peer_data_source.gen.go` use the same contract with required `id` and computed outputs.

Sensitive omission pattern: `internal/service/system/vip_data_source.go` and `internal/service/wireguard/server_data_source.go` use dedicated data-source models when resource models contain fields that should not be exposed by the data source.

Runtime read test pattern: Story 25B.2 added local HTTP-backed tests in `internal/service/*/data_source_schema_test.go`. Reuse that style to exercise real `Read` methods without requiring live OPNsense acceptance infrastructure.

### Sensitive and Write-Only Field Guardrails

`ddclient_account.password` is sensitive in the resource schema and `AccountResourceModel.fromAPI` explicitly does not populate it. Do not expose a fake computed `password` in the data source. Use a dedicated data-source model if the resource model would otherwise create schema/state incompatibility.

`trust_ca.certificate` is certificate material returned by the CA read model and is in scope because `trust_ca` is Batch 3. Do not infer that `trust_cert` is also in scope; `trust_cert` is Batch 4 because private key/certificate write-only behavior needs separate handling.

ACME, Quagga, Unbound, DHCPv4, and Kea DHCPv6 target schemas have no known secret fields in the current models, but implementation must still expose only attributes returned by the API response model.

### Architecture and Quality Guardrails

- Use Go and Terraform Plugin Framework v6 with current project imports. Do not add dependencies.
- Keep data source files beside matching resources under `internal/service/<module>/`.
- File names must be lowercase underscore-separated: `<resource>_data_source.go`.
- Constructors must follow existing naming, e.g. `newBGPNeighborDataSource()`, `newHostOverrideDataSource()`, `newAccountDataSource()` where package-local names do not conflict.
- `Metadata` type names must be `req.ProviderTypeName + "_<module>_<resource>"`, matching resource/doc basenames.
- `Configure` must reuse provider data as `*opnsense.Client`; never create a client in a data source.
- `Read` must use `opnsense.Get[...]` with existing request options and must surface errors through diagnostics.
- `fromAPI` receives an unwrapped API struct. Do not unwrap OPNsense monads in data-source code.
- All Terraform attributes must be snake_case and should match resource attribute names except intentional omissions for unavailable sensitive/write-only fields.
- Avoid helper extraction unless it is genuinely shared and clearer than local code.

### Documentation Requirements

Use the Batch 1 data-source docs pattern:

- Template: `templates/data-sources/haproxy_server.md.tmpl`
- Example: `examples/data-sources/opnsense_haproxy_server/data-source.tf`
- Generated doc: `docs/data-sources/haproxy_server.md`

Every new example must use UUID lookup via `id`. Examples may reference an existing managed resource's `.id` when that makes composition clearer, but must not imply name-based lookup.

When `ddclient_account.password` is omitted, document that the field is unavailable because it is sensitive/write-only and not populated from OPNsense read responses.

### Testing Requirements

Required final command: `make check`.

Recommended targeted tests before `make check`:

```bash
go test ./internal/service/quagga ./internal/service/unbound ./internal/service/dhcp ./internal/service/kea ./internal/service/ddclient ./internal/service/acme ./internal/service/trust
```

Recommended targeted searches:

```bash
rg "new.*DataSource" internal/service/quagga internal/service/unbound internal/service/dhcp internal/service/kea internal/service/ddclient internal/service/acme internal/service/trust
rg "^# opnsense_(quagga|unbound|dhcpv4|kea|ddclient|acme|trust)_" docs/data-sources
rg "data \"opnsense_(quagga|unbound|dhcpv4|kea|ddclient|acme|trust)_" examples/data-sources
```

Tests should include:

- Constructor count assertions per affected package.
- Required `id` schema assertions for all new data sources.
- Local HTTP-backed `Read` tests for representative resources in Quagga, Unbound, DHCPv4, Kea, Dynamic DNS, ACME, and Trust.
- Explicit schema omission test for `ddclient_account.password`.

### Previous Story Intelligence

Story 25B.1 corrected data-source parity math: `system_info` is standalone, so resource-matching gaps are resource docs minus data-source docs, not raw resource count minus data-source count.

Story 25B.2 completed Batch 1 and established the current baseline: 55 registered data sources, 55 data-source docs, 36 resource-matching gaps, and all Batch 1 docs/examples present. It also added the local HTTP-backed runtime `Read` test pattern and documented sensitive omissions for `wireguard_server.private_key` and `system_vip.password`.

Story 25B.2 review findings showed schema-only tests are not enough for new data-source families. Include runtime `Read` tests in this story from the start.

Recent release commits matter because the provider is already published:

- `b82ada3 chore(release): mark v0.1.0 as released and fix GPG signing config`
- `a2bc1d4 docs(changelog): add v0.1.0 provider CHANGELOG in Terraform format (Epic 12-5)`
- `b2ea1c6 ci(release): add release + acceptance workflows, structure validator (Epic 12-4)`

### References

- `_bmad-output/planning-artifacts/data-source-parity-plan.md` — Batch 2/3 source of truth, Batch 4 boundary, and verification commands.
- `_bmad-output/planning-artifacts/post-release-epics.md` — Epic 25B sequence and post-release positioning.
- `_bmad-output/planning-artifacts/support-matrix.md` — current public support matrix and count baseline.
- `_bmad-output/planning-artifacts/prd.md` — provider data-source naming contract, UUID IDs, write-only field constraints, migration/import goals.
- `_bmad-output/planning-artifacts/architecture.md` — service-module structure, data-source naming, API client boundaries, documentation patterns, testing patterns, anti-patterns.
- `_bmad-output/implementation-artifacts/25b-2-high-reference-data-sources.md` — previous story implementation and review learnings.
- `internal/service/haproxy/server_data_source.go` — hand-written UUID data-source pattern.
- `internal/service/quagga/ospf_network_data_source.gen.go` — generated Quagga UUID data-source pattern.
- `internal/service/system/vip_data_source.go` — data-source model omission pattern.
- `internal/service/haproxy/data_source_schema_test.go` — local HTTP-backed `Read` test pattern.

## Project Structure Notes

Expected implementation file locations:

- `internal/service/quagga/*_data_source.go`, `internal/service/quagga/exports.go`, and `internal/service/quagga/data_source_schema_test.go`.
- `internal/service/unbound/*_data_source.go`, `internal/service/unbound/exports.go`, and `internal/service/unbound/data_source_schema_test.go`.
- `internal/service/dhcp/*_data_source.go`, `internal/service/dhcp/exports.go`, and `internal/service/dhcp/data_source_schema_test.go`.
- `internal/service/kea/*_data_source.go`, `internal/service/kea/exports.go`, and `internal/service/kea/data_source_schema_test.go`.
- `internal/service/ddclient/account_data_source.go`, `internal/service/ddclient/exports.go`, and `internal/service/ddclient/data_source_schema_test.go`.
- `internal/service/acme/*_data_source.go`, `internal/service/acme/exports.go`, and `internal/service/acme/data_source_schema_test.go`.
- `internal/service/trust/ca_data_source.go`, `internal/service/trust/exports.go`, and `internal/service/trust/data_source_schema_test.go`.
- `templates/data-sources/<type>.md.tmpl`, `examples/data-sources/opnsense_<type>/data-source.tf`, and generated `docs/data-sources/<type>.md` for all 21 new data sources.

Do not update `_bmad-output/implementation-artifacts/sprint-status.yaml` for this post-release story unless a matching `25b-3-*` key is added separately. The current sprint-status file does not track post-release `25B.*` stories.

## Dev Agent Record

### Agent Model Used

gpt-5.5

### Debug Log References

- Batch 2/3 inventory check: all 21 target resources had resource docs and no pre-existing data-source docs.
- Baseline before implementation: 55 registered data sources, 55 data-source docs, 36 resource-matching data-source gaps.
- Post-implementation counts: 76 registered data sources, 76 data-source docs, 15 remaining resource-matching gaps.
- Targeted service tests passed: `go test ./internal/service/quagga ./internal/service/unbound ./internal/service/dhcp ./internal/service/kea ./internal/service/ddclient ./internal/service/acme ./internal/service/trust`.
- Final validation: `make check` passed.

### Completion Notes List

- Implemented and registered all 21 Batch 2/3 UUID lookup data sources across Quagga/FRR, Unbound, DHCPv4, Kea DHCPv6, Dynamic DNS, ACME, and Trust CA.
- Preserved Batch 4 boundaries: no singleton/settings data sources or `trust_cert` were implemented.
- Omitted `ddclient_account.password` from the data source and documented the sensitive/write-only behavior.
- Added schema and local HTTP-backed runtime `Read` tests for all affected service packages.
- Added templates, examples, and generated docs for all 21 new data sources.
- Updated parity/support docs and provider index counts to 76 data sources/docs and 15 remaining data-source gaps.

### Change Log

- 2026-06-02: Implemented Batch 2/3 routing, DNS, DHCP, ACME, Dynamic DNS, and Trust CA data sources; story moved to review after `make check` passed.

### File List

- `_bmad-output/implementation-artifacts/25b-3-routing-dns-dhcp-acme-trust-data-sources.md`
- `_bmad-output/planning-artifacts/data-source-parity-plan.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `docs/index.md`
- `templates/index.md.tmpl`
- `internal/service/acme/account_data_source.go`
- `internal/service/acme/certificate_data_source.go`
- `internal/service/acme/challenge_data_source.go`
- `internal/service/acme/data_source_schema_test.go`
- `internal/service/acme/exports.go`
- `internal/service/ddclient/account_data_source.go`
- `internal/service/ddclient/data_source_schema_test.go`
- `internal/service/ddclient/exports.go`
- `internal/service/dhcp/data_source_schema_test.go`
- `internal/service/dhcp/exports.go`
- `internal/service/dhcp/reservation_data_source.go`
- `internal/service/dhcp/subnet_data_source.go`
- `internal/service/kea/data_source_schema_test.go`
- `internal/service/kea/dhcpv6_reservation_data_source.go`
- `internal/service/kea/dhcpv6_subnet_data_source.go`
- `internal/service/kea/exports.go`
- `internal/service/quagga/bgp_aspath_data_source.go`
- `internal/service/quagga/bgp_communitylist_data_source.go`
- `internal/service/quagga/bgp_neighbor_data_source.go`
- `internal/service/quagga/bgp_peergroup_data_source.go`
- `internal/service/quagga/bgp_redistribution_data_source.go`
- `internal/service/quagga/data_source_helpers.go`
- `internal/service/quagga/data_source_schema_test.go`
- `internal/service/quagga/exports.go`
- `internal/service/quagga/prefix_list_data_source.go`
- `internal/service/quagga/route_map_data_source.go`
- `internal/service/quagga/static_route_data_source.go`
- `internal/service/trust/ca_data_source.go`
- `internal/service/trust/data_source_schema_test.go`
- `internal/service/trust/exports.go`
- `internal/service/unbound/acl_data_source.go`
- `internal/service/unbound/data_source_schema_test.go`
- `internal/service/unbound/domain_override_data_source.go`
- `internal/service/unbound/exports.go`
- `internal/service/unbound/host_alias_data_source.go`
- `internal/service/unbound/host_override_data_source.go`
- `templates/data-sources/acme_account.md.tmpl`
- `templates/data-sources/acme_certificate.md.tmpl`
- `templates/data-sources/acme_challenge.md.tmpl`
- `templates/data-sources/ddclient_account.md.tmpl`
- `templates/data-sources/dhcpv4_reservation.md.tmpl`
- `templates/data-sources/dhcpv4_subnet.md.tmpl`
- `templates/data-sources/kea_dhcpv6_reservation.md.tmpl`
- `templates/data-sources/kea_dhcpv6_subnet.md.tmpl`
- `templates/data-sources/quagga_bgp_aspath.md.tmpl`
- `templates/data-sources/quagga_bgp_communitylist.md.tmpl`
- `templates/data-sources/quagga_bgp_neighbor.md.tmpl`
- `templates/data-sources/quagga_bgp_peergroup.md.tmpl`
- `templates/data-sources/quagga_bgp_redistribution.md.tmpl`
- `templates/data-sources/quagga_prefix_list.md.tmpl`
- `templates/data-sources/quagga_route_map.md.tmpl`
- `templates/data-sources/quagga_static_route.md.tmpl`
- `templates/data-sources/trust_ca.md.tmpl`
- `templates/data-sources/unbound_acl.md.tmpl`
- `templates/data-sources/unbound_domain_override.md.tmpl`
- `templates/data-sources/unbound_host_alias.md.tmpl`
- `templates/data-sources/unbound_host_override.md.tmpl`
- `examples/data-sources/opnsense_acme_account/data-source.tf`
- `examples/data-sources/opnsense_acme_certificate/data-source.tf`
- `examples/data-sources/opnsense_acme_challenge/data-source.tf`
- `examples/data-sources/opnsense_ddclient_account/data-source.tf`
- `examples/data-sources/opnsense_dhcpv4_reservation/data-source.tf`
- `examples/data-sources/opnsense_dhcpv4_subnet/data-source.tf`
- `examples/data-sources/opnsense_kea_dhcpv6_reservation/data-source.tf`
- `examples/data-sources/opnsense_kea_dhcpv6_subnet/data-source.tf`
- `examples/data-sources/opnsense_quagga_bgp_aspath/data-source.tf`
- `examples/data-sources/opnsense_quagga_bgp_communitylist/data-source.tf`
- `examples/data-sources/opnsense_quagga_bgp_neighbor/data-source.tf`
- `examples/data-sources/opnsense_quagga_bgp_peergroup/data-source.tf`
- `examples/data-sources/opnsense_quagga_bgp_redistribution/data-source.tf`
- `examples/data-sources/opnsense_quagga_prefix_list/data-source.tf`
- `examples/data-sources/opnsense_quagga_route_map/data-source.tf`
- `examples/data-sources/opnsense_quagga_static_route/data-source.tf`
- `examples/data-sources/opnsense_trust_ca/data-source.tf`
- `examples/data-sources/opnsense_unbound_acl/data-source.tf`
- `examples/data-sources/opnsense_unbound_domain_override/data-source.tf`
- `examples/data-sources/opnsense_unbound_host_alias/data-source.tf`
- `examples/data-sources/opnsense_unbound_host_override/data-source.tf`
- `docs/data-sources/acme_account.md`
- `docs/data-sources/acme_certificate.md`
- `docs/data-sources/acme_challenge.md`
- `docs/data-sources/ddclient_account.md`
- `docs/data-sources/dhcpv4_reservation.md`
- `docs/data-sources/dhcpv4_subnet.md`
- `docs/data-sources/kea_dhcpv6_reservation.md`
- `docs/data-sources/kea_dhcpv6_subnet.md`
- `docs/data-sources/quagga_bgp_aspath.md`
- `docs/data-sources/quagga_bgp_communitylist.md`
- `docs/data-sources/quagga_bgp_neighbor.md`
- `docs/data-sources/quagga_bgp_peergroup.md`
- `docs/data-sources/quagga_bgp_redistribution.md`
- `docs/data-sources/quagga_prefix_list.md`
- `docs/data-sources/quagga_route_map.md`
- `docs/data-sources/quagga_static_route.md`
- `docs/data-sources/trust_ca.md`
- `docs/data-sources/unbound_acl.md`
- `docs/data-sources/unbound_domain_override.md`
- `docs/data-sources/unbound_host_alias.md`
- `docs/data-sources/unbound_host_override.md`
