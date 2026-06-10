---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 25B.1: Data Source Parity Inventory and Implementation Batches

Status: done

## Story

As a provider maintainer,
I want a current data-source parity inventory grouped into implementation batches,
so that the remaining data-source work can be implemented predictably after v0.1.0 without duplicating work or misrepresenting support status.

## Acceptance Criteria

1. Inventory compares registered resources against registered data sources and documents every missing resource-matching data source.
2. Inventory is generated from implementation truth, not from stale roadmap counts alone.
3. Missing data sources are grouped into implementation batches by domain, reference value, and implementation pattern.
4. Each batch includes scope, expected file patterns, likely source files to inspect, documentation/example requirements, and verification commands.
5. Unsupported or intentionally deferred data-source candidates are explicitly classified with rationale instead of silently omitted.
6. A new planning artifact is created at `_bmad-output/planning-artifacts/data-source-parity-plan.md`.
7. `_bmad-output/planning-artifacts/support-matrix.md` links to the parity plan and retains accurate Supported / Coming / Upstream-blocked counts.
8. No provider code, generated code, resource behavior, or Registry docs are changed except where needed to link the new planning artifact.
9. `make check` passes.

## Tasks / Subtasks

- [x] Task 1: Generate current implementation inventory (AC: 1, 2)
  - [x] Count registered resource constructors from `internal/service/*/exports.go`.
  - [x] Count registered data-source constructors from `internal/service/*/exports.go`.
  - [x] Cross-check counts against `docs/resources/*.md` and `docs/data-sources/*.md`.
  - [x] Record mismatches between registration and generated docs, if any.
- [x] Task 2: Classify missing data sources (AC: 3, 5)
  - [x] Identify missing resource-matching data sources by comparing resource doc basenames to data-source doc basenames.
  - [x] Separate normal UUID-backed resources from singleton resources and special/manual resources.
  - [x] Mark any candidate as deferred only with a concrete reason.
- [x] Task 3: Create implementation batches (AC: 3, 4)
  - [x] Batch high-reference resources first: HAProxy, firewall, system, WireGuard/OpenVPN, and core IPsec resources.
  - [x] Batch routing resources next: BGP, prefix list, route map, RIP, static, and singleton FRR resources.
  - [x] Batch DNS/DHCP/ACME/Trust resources next.
  - [x] Batch remaining singleton or special-case resources separately.
- [x] Task 4: Write the parity plan artifact (AC: 6)
  - [x] Create `_bmad-output/planning-artifacts/data-source-parity-plan.md`.
  - [x] Include summary counts, full missing list, batch tables, implementation notes, and verification commands.
  - [x] Include a “do not implement in this story” boundary.
- [x] Task 5: Update support matrix link/counts (AC: 7)
  - [x] Add a link from `_bmad-output/planning-artifacts/support-matrix.md` to the parity plan.
  - [x] Update counts after verifying that the current inventory is 90 resources / 34 data sources / 57 resource-matching data-source gaps.
- [x] Task 6: Validate (AC: 8, 9)
  - [x] Confirm no unintended code/resource/doc generation changes were made.
  - [x] Run `make check`.

## Dev Notes

### Current Baseline

The planning baseline as of 2026-06-02 is 90 supported resources, 34 supported data sources, and matching documentation counts. Treat these as expected values, but verify from the implementation during this story. The resource-matching data-source gap is 57 because `system_info` is a standalone data source without a matching resource.

Primary sources:

- `_bmad-output/planning-artifacts/post-release-epics.md`: Epic 25B says the goal is closing the gap between 90 supported resources and 34 supported data sources.
- `_bmad-output/planning-artifacts/support-matrix.md`: current release positioning and Supported / Coming / Upstream-blocked matrix.
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`: domain-by-domain supported vs coming vs upstream-blocked status.
- `_bmad-output/planning-artifacts/prd.md`: FR60 says every resource should eventually have a corresponding read-only data source; current baseline is 34 data sources and parity remains planned work.
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`: Story 25B.1 is the first recommended post-release step.

### Scope Boundary

This is an inventory and planning story only.

Do not implement data sources in this story. Do not add generated `.gen.go` files, new Terraform schema code, provider registrations, examples, templates, or generated Registry docs except for the planning artifact and support-matrix link. Data-source implementation belongs to Stories 25B.2 and 25B.3.

### Implementation Truth Sources

Use implementation truth in this order:

1. `internal/service/*/exports.go` for registered resources/data sources.
2. `docs/resources/*.md` and `docs/data-sources/*.md` for generated Registry docs coverage.
3. Existing data-source implementation files to classify pattern type.
4. Planning docs only for product intent and grouping, not as the source of exact registration counts.

Important: generated docs can lag implementation if docs were not regenerated. If registration and docs disagree, document the discrepancy and prefer registered implementation for support status.

### Existing Data-Source Patterns to Reuse Later

This story should document which pattern each batch should use. Existing examples:

- Hand-written UUID lookup: `internal/service/firewall/alias_data_source.go` uses `id` as required config, calls `opnsense.Get[...]`, calls `fromAPI`, then sets state.
- Generated UUID lookup: `internal/service/iface/bridge_data_source.gen.go` and `internal/service/trafficshaper/pipe_data_source.gen.go` follow the same pattern with generated schema and `newXDataSource` constructor.
- Singleton/special source: `internal/service/system/system_info_data_source.go` does not mirror a resource; it calls `/api/core/firmware/info` directly and returns `id = system_info`.

The parity plan should call out that singleton resources may need a different data-source pattern or may be lower priority if a resource itself already reads the single configuration object.

### Inventory Commands

The dev agent may use these commands, or equivalent scripts, to produce auditable counts:

```bash
rg "\t\tnew[A-Za-z0-9]+Resource," "internal/service" | wc -l
rg "\t\tnew[A-Za-z0-9]+DataSource," "internal/service" | wc -l
rg "^# opnsense_" "docs/resources" | wc -l
rg "^# opnsense_" "docs/data-sources" | wc -l
comm -23 <(printf '%s\n' docs/resources/*.md | xargs -n1 basename | sort) <(printf '%s\n' docs/data-sources/*.md | xargs -n1 basename | sort)
```

When writing `data-source-parity-plan.md`, include the exact command outputs or a summarized table with the command used.

### Known Missing Data-Source Docs at Story Creation

Verify before relying on this list. It was generated by comparing `docs/resources/*.md` to `docs/data-sources/*.md` during planning.

```text
acme_account, acme_certificate, acme_challenge, ddclient_account,
dhcpv4_reservation, dhcpv4_subnet, dnsmasq_settings, firewall_category,
firewall_filter_rule, firewall_nat_outbound, firewall_nat_port_forward,
haproxy_acl, haproxy_backend, haproxy_frontend, haproxy_healthcheck,
haproxy_server, ipsec_child, ipsec_connection, ipsec_key_pair, ipsec_local,
ipsec_psk, ipsec_remote, kea_ctrl_agent, kea_dhcpv6_reservation,
kea_dhcpv6_settings, kea_dhcpv6_subnet, openvpn_client_overwrite,
openvpn_instance, openvpn_static_key, quagga_bgp_aspath,
quagga_bgp_communitylist, quagga_bgp_global, quagga_bgp_neighbor,
quagga_bgp_peergroup, quagga_bgp_redistribution, quagga_general,
quagga_ospf6_general, quagga_ospf_general, quagga_prefix_list,
quagga_rip, quagga_route_map, quagga_static, quagga_static_route,
system_gateway, system_route, system_vip, system_vlan, trust_ca,
trust_cert, unbound_acl, unbound_dnsbl, unbound_domain_override,
unbound_general, unbound_host_alias, unbound_host_override,
wireguard_peer, wireguard_server
```

### Required Batch Plan Structure

`_bmad-output/planning-artifacts/data-source-parity-plan.md` should include:

- Frontmatter with `title`, `date`, `author`, `status`, and `inputs`.
- Summary counts table.
- Full missing data-source inventory table with columns: `Resource`, `Domain`, `Current data source?`, `Pattern`, `Priority`, `Batch`, `Notes`.
- Batch sections with columns: `Data source`, `Files to inspect`, `Likely pattern`, `Docs/examples needed`, `Risks`.
- Deferred/exception section for any singleton or API-limited candidates.
- Verification section with commands and expected checks.
- Next-story handoff for `25B.2` and `25B.3`.

### Suggested Batches

Use this as a starting point; adjust after inventory if facts changed.

| Batch | Focus | Rationale |
|---|---|---|
| Batch 1 | HAProxy, firewall, system, WireGuard/OpenVPN, high-reference IPsec | Highest composition and migration value; many resources are referenced by other resources. |
| Batch 2 | Quagga/FRR BGP, prefix list, route map, RIP, static, singleton general resources | Routing users need lookups for brownfield migration and route policy composition. |
| Batch 3 | DNS, DHCP/Kea, Dynamic DNS, ACME, Trust | Important for full-appliance import, but generally less interdependent than HAProxy/routing. |
| Batch 4 | Singleton/special cases and deferred candidates | Prevents singleton ambiguity from blocking straightforward UUID-backed data sources. |

### Architecture and Quality Guardrails

- Follow DevRail rules: run `make check`; do not suppress failing checks.
- Keep edits to planning/docs. Avoid provider implementation changes in this story.
- If using scripts to produce inventory, keep them temporary unless the story explicitly decides to add a reusable script. Do not add host-installed tooling.
- Do not rely on stale sprint status for counts; `sprint-status.yaml` is known to include historical statuses and blocked/cancelled items.
- If `make docs` does not regenerate tfplugindocs content, do not invent generated output. Record the behavior if relevant.

### Previous Work Intelligence

Recent release work matters here:

- `b82ada3 chore(release): mark v0.1.0 as released and fix GPG signing config`
- `a2bc1d4 docs(changelog): add v0.1.0 provider CHANGELOG in Terraform format (Epic 12-5)`
- `b2ea1c6 ci(release): add release + acceptance workflows, structure validator (Epic 12-4)`

The provider is published. Do not create release-readiness work as backlog. This story exists to plan post-release parity work.

### References

- `_bmad-output/planning-artifacts/post-release-epics.md` — post-v0.1.0 epic sequence and Story 25B.1 position.
- `_bmad-output/planning-artifacts/support-matrix.md` — current support matrix and counts.
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md` — current domain status and cross-cutting data-source parity gap.
- `_bmad-output/planning-artifacts/prd.md` — FR60, FR61, FR66-FR68, current implementation baseline, and user journeys for migration/import.
- `_bmad-output/planning-artifacts/architecture.md` — service module structure, data-source naming, documentation, testing, and anti-pattern rules.
- `internal/service/firewall/alias_data_source.go` — hand-written data-source example.
- `internal/service/iface/bridge_data_source.gen.go` — generated data-source example.
- `internal/service/system/system_info_data_source.go` — singleton/special data-source example.

## Project Structure Notes

Expected files touched by this story:

- `_bmad-output/planning-artifacts/data-source-parity-plan.md` — new.
- `_bmad-output/planning-artifacts/support-matrix.md` — update link/counts only.
- `_bmad-output/implementation-artifacts/25b-1-data-source-parity-inventory-and-batches.md` — update Dev Agent Record during implementation.

Files to inspect but not modify unless a discovered mismatch requires a planning-doc correction:

- `internal/service/*/exports.go`
- `docs/resources/*.md`
- `docs/data-sources/*.md`
- Existing `*_data_source*.go` files
- `templates/index.md.tmpl` and `docs/index.md`

## Testing Requirements

- Required final command: `make check`.
- This is a planning/docs story, so no acceptance tests are expected.
- If implementation changes accidentally appear in the diff, stop and either revert only your own unintended changes or explain why they are required. Do not modify unrelated existing worktree changes.

## Completion Criteria

The story is complete only when:

- The parity plan exists and is specific enough for 25B.2/25B.3 implementation.
- The support matrix links to the parity plan.
- Current counts are verified from implementation and docs.
- Any deviations from the expected 90 resources / 34 data sources baseline are explained.
- `make check` passes.

## Dev Agent Record

### Agent Model Used

gpt-5.5

### Debug Log References

- Inventory counts: 90 registered resources, 34 registered data sources, 90 resource docs, 34 data-source docs.
- Missing data-source inventory generated by comparing `docs/resources/*.md` and `docs/data-sources/*.md` basenames: 57 resource-matching gaps.
- Review correction: `system_info` is standalone, so the resource-matching backlog is 57 even though there are 34 total data sources.
- Validation: `make check` passed.

### Completion Notes List

- Created `_bmad-output/planning-artifacts/data-source-parity-plan.md` with verified counts, existing data sources, 57 missing data-source candidates, four implementation batches, exception rules, and next-story handoff.
- Updated `_bmad-output/planning-artifacts/support-matrix.md` with a link to the parity plan.
- Corrected related planning artifacts from 56 to 57 resource-matching data-source gaps after review found the standalone `system_info` data source.
- Confirmed no provider implementation files were changed for this planning-only story.

### Change Log

- 2026-06-02: Created data-source parity plan and linked support matrix; story marked ready for review after `make check` passed.
- 2026-06-02: Review follow-up corrected resource-matching gap count to 57 and clarified Batch 4 handoff/verification guidance.
- 2026-06-02: Review closure confirmed all findings resolved; story marked done.

### File List

- `_bmad-output/implementation-artifacts/25b-1-data-source-parity-inventory-and-batches.md`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/data-source-parity-plan.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
