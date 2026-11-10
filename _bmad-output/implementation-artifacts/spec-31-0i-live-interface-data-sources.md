---
title: 'Story31.0i: read-only live interface data sources'
type: 'feature'
created: '2026-11-04'
status: 'done'
route: 'dispatch'
review_loop_iteration: 0
baseline_commit: 'b090de1cf22e467c32e7711d2b09a6593f5d2bf1'
baseline_revision: 'b090de1cf22e467c32e7711d2b09a6593f5d2bf1'
story_key: '31-0i-live-interface-data-sources'
context: ['AGENTS.md', '_bmad-output/implementation-artifacts/epic-31-context.md']
---

<frozen-after-approval reason="human-owned intent">

## Intent

**Problem:** Operators cannot inspect complete runtime interface diagnostics and overview through the four audited read-only surfaces.

**Approach:** Add `opnsense_interface`, `opnsense_interface_all`, `opnsense_interfaces_overview`, and `opnsense_interfaces_overview_all`, preserving the pinned public field/type contract while handling real released API shapes safely. Singles select exact required `device`; aggregates return computed `interfaces` lists. Assignment/IP management remains the existing resource's responsibility.

## Boundaries & Constraints

**Always:** Use the explicit field allowlist in `31-0i-interface-api-contract.md`. Expose absent optional scalars/objects as null, not fabricated zero values; preserve explicit false/zero/empty strings. Missing address/set collections become known empty collections. Sort aggregate rows by device and address rows by complete public tuple without claiming first is primary. Preserve collection members verbatim, including blank and multi-address DHCP strings. Use the existing authenticated TLS/retry client and exactly one bounded read semaphore acquisition per GET. Validate HTTP before JSON; use actionable credential/ACL/endpoint guidance without raw response bodies. Reject malformed envelopes/types, null backend results, identity mismatch/duplicates, and incomplete overview counts. Ignore unknown fields but never expose them. Keep tests, generated docs, coverage and source evidence distinct from actual appliance proof.

**Never:** POST, reconfigure, apply, save configuration, acquire mutation ownership, change assignment/CRUD/generated interface resources, expose overview config/wireless/raw JSON or secrets, inherit the assignment-only26.7.6 requirement, invent identifier aliases or mutate a production appliance. No unsupported older-firmware/OpenTofu/hardware claim. Sensitive-state privacy and consumer-neutral publication boundaries apply. Do not duplicate released retry/Trust/IPsec/FRR increments or close their unfinished parents.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Diagnostic read | Device-keyed object, exact selector | Full public row; diagnostic MTU int64 | Unknown device, key mismatch, wrong types fail without fabricated state |
| Empty inventory | Diagnostics `[]`/`{}` or complete overview zero rows | Aggregate empty list; single not-found | Null/malformed backend is failure, not empty success |
| Overview | Assigned/unassigned rows, optional VLAN/LAGG/CARP | Public typed values; absent/null retained; overview MTU string | Duplicate/missing identity or incomplete total/rowCount/current fails |
| Determinism | Reordered map/addresses, repeated refresh | Identical public values/order for unchanged observations | Unknown extras ignored and not logged |
| Failures | Auth/ACL/404/5xx, cancellation, nil client | Actionable bounded diagnostic; no writes | No raw secret-bearing body or stale success |

</frozen-after-approval>

## Code Map

- `internal/service/iface/live_interface_{api,model,schema}.go` — new handwritten shared GET decoder, strict source-aware normalization and typed schema/model mapping; generic monad CRUD does not fit these responses.
- `internal/service/iface/{interface,interface_all,interfaces_overview,interfaces_overview_all}_data_source.go` — new constructors, Configure/Schema/Read and exact-device selection. `exports.go` registers four constructors; provider already aggregates it.
- `pkg/opnsense/{client,errors}.go` — reuse semaphore/HTTP client/HTTP errors; core endpoint404 must not become missing-plugin advice. Do not change global transport or assignment behavior.
- `internal/service/iface/live_interface_*test.go` — fixture/protocol/actual Core tests; external `iface_test` avoids provider cycle. Reuse gateway-group Core loopback pattern and existing client fixtures.
- `examples/data-sources/opnsense_*/data-source.tf`, matching templates and generated docs — four new source examples/descriptions, null/order/read-only/state-privacy boundaries.
- Coverage audit JSON/report and `internal/coverageaudit/audit_test.go` — add four data-source-only registrations;113 resources remain,100 data sources become104. Reconcile only four surface rows and shipped retry tracking, preserving15 resource-matching lookup gaps and parent limitations.
- `CHANGELOG.md` — next feature0.17.0 Unreleased, preserving released0.16history. Owning story/sprint/support/build order — accurate bounded status and next entry point.

## Tasks & Acceptance

**Execution:**
- [x] Implement shared API/model/schema and four sources/registrations using the contract.
- [x] Cover every I/O matrix case, all public fields, collection ordering and GET-only routing with fixtures; verify actual Terraform Core state and refresh/no-change planning.
- [x] Add explicit live read-only acceptance path; independently compare actual API truth and zero mutation requests on a supported appliance, storing inventory privately and publishing only sanitized version/result proof.
- [x] Generate docs/audit deterministically; update release/story/sprint/queue evidence without broad completion claims.
- [x] Run pinned container focused race/Core tests and full six-gate check; collect all three reviews, repair and independently verify affected changes.
- Publication gate: one focused PR with eight successful hosted checks on its exact head; readiness is reported in the PR handoff only after those checks finish. No auto-merge.

**Acceptance Criteria:**
- Given supported API observations, when all four data sources read, then complete allowlisted values match the source contract at stable addresses and no configuration endpoint is called.
- Given assignment management and runtime lookups, when lookups refresh, then assignment ownership/configuration remains unchanged and runtime differences are observations, not managed drift.
- Given actual supported-appliance reads, when acceptance completes, then version, tested source, zero mutation requests and sanitized results are recorded separately from mocked coverage; unobserved hardware/versions remain explicit.

## Implementation Notes

Parent owns private live read-only validation setup/credentials; implementer handles public code/fixtures/acceptance harness and never accesses real appliances or private secrets. Use full pinned schema authority paths in the public API contract; a local source-only mirror for inspection exists outside the repo, but no extra dependency is required.

No unresolved intent questions remain; technical choices are explicit above. Independent review, PR publication and release remain separate gates.

## Spec Change Log

## Review Triage Log

The three independent layers reviewed the same frozen patch (SHA256
`d26983cd34c1d0ec419fab53479ef9e79ed207adcbbcf0594ab6c86f2032fc65`).
Every finding is classified separately before grouping. No spec/intent loopback
is required. The unavailable original implementer cannot be re-engaged; the
parent applies only the trivial patches below under step04's fallback.

| ID | Verdict | Evidence and route |
| --- | --- | --- |
| B1 IPv6 mode omission | low | `overviewShape` deliberately matches the pinned public allowlist, which has `link_type` but not upstream `link_typev6`. A complete assignment view must not be inferred. Patch: document the exclusion beside the IPv4 mode; do not expand the approved public surface. |
| B2 address shape documentation | medium | Diagnostic address `ipaddr` is bare with separate `subnetbits`; assigned overview addresses carry a prefix and unassigned rows can be bare. Patch: schema/template descriptions plus literal fixture assertions. |
| B3 debug-string ordering | medium | Address sorting currently calls the framework's logging-only `ObjectValue.String`; using that unstable representation as a state-order key couples refresh determinism to a debug format. Patch: serialize an explicit alphabetically keyed tuple of public scalar values, with null distinct from empty; retain reorder tests. |
| B4 acceptance harness completeness | false | The harness explicitly requires a separate all-field comparison; the owning story and verification receipt record that independent oracle as actually executed after apply and refresh. The harness alone was never the claimed proof. Its narrow assertions do not invalidate the separately executed oracle. No new appliance run is needed merely to collect the review. |
| B5 field descriptions | medium | Computed schema attributes have no explanations, so generated docs omit renamed wire semantics, primary addresses and type/null distinctions. Patch: direct schema descriptions and matching templates, without public-surface changes. |
| B6 duplicate-set wording | low | The contract says preserve members exactly but omits the required exact-duplicate collapse already implemented and tested. Patch: correct that sentence; distinct blank/whitespace-bearing strings remain unchanged. |
| B7 whitespace selector | low | A whitespace-only selector is already rejected as not found after a bounded GET and never aliases another device. The distinct diagnostic is cosmetic, unlikely in everyday use; reject under step04's low-finding complexity rule rather than add a new guard. |
| B8 standing pagination rule | low | The epic-context rewrite removed the general pagination rule, while the new reader correctly fails closed and the contract already names the 9999-row/POST-only limit. Defer: restoring agent-context guidance belongs to the workflow's explicit context-file defer route; retain that follow-up without enabling POST or changing this reader's contract. |
| B9 inventory set difference | low | Released overview skips `pfsync0`; diagnostics does not. Patch: document upstream-defined inventories and add a literal fixture proving the provider preserves that difference rather than filtering diagnostics. |
| B10 localized description | low | Released overview localizes the unassigned label. Patch: describe locale dependence and prohibit treating descriptions as stable identity in usage guidance. |
| B11 semaphore decode lifetime | low | The bounded read slot is retained through decode (maximum 16 MiB). It still bounds work and is always released; no deadlock or incorrect state occurs. A large-response optimization adds release-path complexity and is unlikely in everyday use; reject under step04's low-finding rule. |
| V1 empty-set regression gap | medium | Pre-verified: current null/ordering test never asserts missing sets. Patch: assert all top-level and nested allowlisted sets are known empty for absent and explicit-null keys, plus Core empty-set cardinality. |
| V2 self-derived allowlist test | medium | Pre-verified: dropping `macaddr` or `routes` also shrinks the generated fixture and expected count. Patch: literal independent expected names for all top-level and nested shapes, checked before fixture generation. |
| V-other narrow opt-in harness | false | Verified separately: opt-in harness skips by default and tests only selectors/GETs, but explicitly requires the independently executed all-field oracle. This is a limitation already recorded, not a claim that default `make check` supplies live proof. |
| R1 collection description | low | Independent bounded repair review confirmed all functional repairs and unchanged allowlists, but found set-only wording attached to address lists. Patch: move the set-specific sentence to SetAttribute descriptions; no runtime change. |
| S1 historical scanner diagnostic | low | Full local scan identified an unchanged IPsec private-state marker name, not credential material. The repository `.gitleaksignore` explicitly prescribes exact false-positive fingerprints and has Kea/Unbound marker precedents. Patch: add only the exact historical commit/file/rule/line fingerprint; keep scanners and real secret rules enabled. |
| E no findings | false | Edge-case layer reports an empty finding list after tracing guards and callees. No repair implied; its no-toolchain/no-appliance limits remain explicit. |

Survivors have distinct root causes and remain separate patch entries. B8 is the
only deferred entry. Re-review the repaired artifact independently; do not count
author triage or local passing tests as independent review.

## Verification

All Go/tests/generation/checks run in `ghcr.io/devrail-dev/dev-toolchain:1.12.24`, no host installs. Run focused race and non-skipped actual Terraform Core controls; `go generate ./...` twice with deterministic output; full `make _check` inside the pinned container; `git diff --check`. Inspect the complete public diff for private consumer data, not just scanner results. Exact-head hosted CI is a separate required gate.

Implementation review complete: all three independent frozen-patch reviews collected and triaged; independent repair review verified affected changes and scanner classification. Final pinned local `make check` passed all six categories in 71.667 seconds with none skipped; generation twice and whitespace checks passed. Publication readiness remains separately gated by the exact-head PR/check record.
