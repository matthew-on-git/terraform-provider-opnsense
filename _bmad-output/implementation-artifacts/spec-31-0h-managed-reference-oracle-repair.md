---
title: 'Story31.0h: managed policy-reference acceptance-oracle repair'
type: 'test-repair'
created: '2026-11-04'
status: 'done'
baseline_revision: 'b0cada56890dd8afeb1d61c2372f03d12e738d1e'
review_loop_iteration: 0
followup_review_recommended: true
context: ['_bmad-output/implementation-artifacts/spec-31-0h-advanced-frr-parity.md', '_bmad-output/implementation-artifacts/31-0h-advanced-live-evidence.md']
---

<intent-contract>

## Intent — immutable

Repair one acceptance-oracle category error: a managed resource may correctly retain a configured policy-entry UUID while the appliance returns another canonical sibling of the same authoritative logical group. Step12 must prove managed and raw policy selections independently belong to the intended group, rather than require managed state to equal a fresh-import canonical representation.

This is a bounded test-only repair, not a provider behavior change. Preserve exact canonical equality for genuinely fresh imports/lookups. Preserve the existing full17-step matrix and all numeric-policy, current-route, identity, authentication, restoration and cleanup gates. A repaired oracle or passing focused tests are not live acceptance.

### Always

- Require an identified managed peer-group owner, a present nonempty configured relation, and a present nonempty fresh raw relation for the same owner.
- Validate both selections independently through authoritative per-UUID reads; every selected UUID must exist and resolve to the exact fixture route-map group. Never infer equivalence from UUID order, resemblance, state alone, or merely nonempty values.
- Validate any inspected flatmap set count against its actual members. Reject missing/malformed/negative counts, count/member disagreement and empty member UUIDs. This strengthens a helper-input boundary; no reachable Terraform state-corruption or provider cardinality bug has been established.
- Keep `frrCanonicalReferences` strict for genuinely fresh import/data-source state. Same-group but noncanonical fresh state must still fail.
- Retain password-free lookup checks and bidirectional comparison of all noncanonicalized public fields, including explicit empty values and ordinary set counts/members.
- Leave numeric declarations, active-session withdrawal from current prefixes, forbidden-default history, stable route-map UUIDs, subsequent restored-positive observations, advanced checks, actual Terraform deletion, exact restoration and zero-leftover checks intact.

### Never

- Relax production `normalizeBGPRead`, policy membership/equivalence, payloads, schemas, resource/data-source behavior or configuration to satisfy a test.
- Globally disable canonical-reference verification, accept the wrong group, accept missing/stale selections, replace current prefixes with historical counters, or remove a failing check without replacing its intended proof.
- Substitute the existing data-source snapshot for managed state at Step12 and assume it is post-update: its unchanged-ID dependency can allow its read before route-map updates.
- Log raw configuration, authentication, private inventory, endpoints or local private paths. Merge, release, deploy, touch production, or promote parent Story31.0h to complete.

</intent-contract>

## Baseline and executed context

- Baseline full head: `b0cada56890dd8afeb1d61c2372f03d12e738d1e`.
- Failed attempt label: `matrix04`; frozen candidate source digest: `d027c2626e5fbd681be39152e2a68eff8959ae6a08d01c93206ae6d0a2299df7`.
- The real non-skipped run failed Step12/17 with `fresh import/data-source reference linked_routemap_out is not raw canonical selection`. It remains a failed attempt, not full acceptance.
- Retained independent observation contains `permit 30` and `deny 5`, an active collector and no approved prefix in current routes. The run failed before reaching the mandatory withdrawal loop, so that observation does not establish a passing Step12.
- Parent reported completed seven-namespace Terraform deletion, unchanged frozen source and six-predicate fallback restoration. These restoration results do not convert the failed matrix into PASS.
- Public checkpoint remains partial. Old-firmware/OpenTofu execution, numeric-clear interface design and blanket advanced-neighbor session proof are outside this repair.

## Source-backed diagnosis and code map

At baseline, `internal/service/quagga/frr_compatibility_acceptance_test.go:471` passes the managed `opnsense_quagga_bgp_peergroup.policy` state to `frrCheckImportedPolicy`. That helper invokes `frrCanonicalReferences`, which correctly requires exact canonical selection for fresh state, but is the wrong contract for this managed state.

- Fixture `linked_routemap_out` intentionally references route-map sibling B (`frr_compatibility_acceptance_test.go`, `frrCompatibilityConfig`).
- `internal/service/quagga/compatibility_lifecycle.go`, `normalizeBGPRead`, preserves prior configured policy UUIDs when complete authoritative membership proves group equivalence.
- `internal/service/quagga/compatibility_policy.go`, `equivalentPolicies`/`policyGroups`, establishes equivalence by authoritative group identity, never UUID ordering.
- `internal/service/quagga/bgp_peergroup_data_source.go` reads fresh raw state using a config model, unlike managed resource state. The existing lookup fixture depends only on the peer-group ID; it is not automatically a post-reordering observation.
- `internal/service/quagga/compatibility_review_test.go`, `TestReviewOraclesRejectWrongSiblingNarrowingAndMisplacedAF`, already requires strict rejection of a fresh wrong canonical sibling. Keep that regression.

Exact compared managed/raw reference UUIDs were not retained in the bounded evidence reviewed here. Do not claim that reversal itself necessarily changed the canonical sibling. The source-backed category error is sufficient and independently testable.

## Implementation tasks

1. Add a narrowly named managed-policy-reference oracle in the acceptance test support code. Keep its pure selection/cardinality/group validation separable from HTTP observation so negative controls can run without an appliance.
2. At Step12 replace only the erroneous managed-state invocation with the managed-reference oracle. Read the same owner through the raw API, parse its canonical selection with existing strict decoding, and independently resolve both managed and raw UUID sets to the exact expected fixture group. Require at least one member on each side. Keep managed and raw UUID equality optional only when this authoritative group proof succeeds.
3. Preserve strict `frrCanonicalReferences` calls for actual import/lookup verification. Share cardinality validation where useful without broadening which relation fields are ignored or weakening existing semantics. Avoid a global boolean that accidentally makes every fresh check permissive.
4. Add focused regression/negative controls in `frr_compatibility_acceptance_test.go` and/or `compatibility_review_test.go`. Reuse existing HTTP fixture patterns; no mock response may be presented as live appliance proof.
5. Document the bounded repair and failed attempt truthfully in the existing advanced evidence checkpoint after verification. Do not overwrite historical failed-run evidence or claim the full matrix passed before a fresh full execution.

## Acceptance criteria — Given / When / Then

### Managed representation

- Given managed selection B and raw canonical selection A, when independent authoritative reads prove both existing UUIDs belong to the exact expected route-map group, then the managed-reference oracle accepts them without modifying managed state or provider behavior.
- Given either selection contains a UUID in another group, a stale/unreadable/missing UUID, an empty member, an absent relation or an empty raw selection, when the managed-reference oracle runs, then it fails even if the other selection is correct.
- Given a missing/noninteger/negative flatmap count or a count inconsistent with enumerated members, when selection validation runs, then it fails. Correctly formed one-member and supported multi-member same-group representations remain accepted; do not invent a one-UUID-only policy contrary to existing grouped semantics.
- Given an owner-ID mismatch or absent managed owner, when Step12 attempts authoritative validation, then it fails rather than validating another object.

### Fresh canonical representation and public fields

- Given fresh import/lookup state B and raw canonical selection A, when strict canonical validation runs, then it rejects B even if both UUIDs belong to the same group.
- Given a fresh state matching canonical selection and correct shape, when strict validation runs, then it passes.
- Given lookup password exposure, missing/extra/changed actual public fields, explicit-empty loss, or ordinary set count/member mismatch, when existing parity checks run, then they still fail.

### Numeric policy and lifecycle

- Given unchanged route-map UUIDs with exact declared permit30/deny5 values in released lexical rendering, when Step12 runs, then it still requires active-session withdrawal of the approved prefix from current routes and rejects any forbidden default observation before logging effective numeric order.
- Given the original configuration is restored, when the next positive check runs, then it must observe the approved current prefix on an active session and retain stable identities and authoritative policy binding.
- Given any full-run failure, when cleanup executes, then seven-namespace Terraform deletion and exact restoration remain independently checked and the failure remains visible.

## Verification matrix and execution

Run focused verification only in the existing pinned DevRail container `ghcr.io/devrail-dev/dev-toolchain:1.12.24`, as specified by `Makefile` and CI. The repository Go requirement is `1.26.10`; record the actual toolchain version. Use the parent-approved existing container execution command and do not install or substitute a host toolchain.

Required focused coverage:

- Managed same-group/different-canonical positive, exact canonical positive, both-sided wrong-group negatives, stale/unreadable UUID, missing/empty selection and malformed-cardinality negatives.
- Existing fresh wrong-canonical negative and canonical positive.
- Existing advanced/generic public-flatmap and released lexical-sequence regressions.
- Focused affected quagga package suite and race run in the approved container where supported; existing repository checks required by the bounded review workflow. Do not run live acceptance from the implementer or review fixtures.

After focused checks and the parent-directed four-layer bounded review complete, parent owns a **new full frozen17-step non-skipped acceptance run** on a newly attested disposable target, with exact source/head/digest binding. Do not resume at Step12 or count prior partial observations as completed steps. Preserve all earlier failed attempts, the final actual deletion/restoration evidence and explicit unverified platform rows.

## Completion boundary

This document is ready for development; it contains no implementation. The repair is reviewable only after narrow test changes and focused controls pass. Full acceptance remains pending until the parent verifies all17 steps plus restoration/cleanup. Parent Story31.0h remains open for its separately unverified requirements.


## Review Triage Log

### 2026-11-04 — Four-layer bounded review

All four layers launched; fourth launch initially hit native thread limit and was retried after one completed. None skipped. Verdicts: 12 findings — high0, medium3, low5, false4, maybe-false0. Findings were assessed individually before grouping.

- [medium] [patch] Blind1: reused member-name decode accepts JSON null after an earlier valid group. Null leaves a prior string unchanged; use a fresh member-local name and negative controls.
- [medium] [patch] Blind2: scalar plus set attributes bypass set shape validation. Reject mixed representations before comparison; retain valid scalar and set controls.
- [false] [reject] Blind3: arbitrary member-key syntax and duplicate values are not demonstrated outputs of typed Terraform flatmaps. Each actual member is independently resolved and count mismatch is rejected; no caller can derive the proposed malformed keys from a typed set. Do not invent a provider cardinality defect or expand its schema.
- [low] [patch] Blind4: unequal-cardinality same-group positives missing. Add both directions without changing equivalence behavior.
- [low] [reject] Blind5: exhaustive HTTP malformed-envelope cases absent. Adapter already rejects absent/empty/malformed object envelopes; member null defect is separately repaired and covered. Full exhaustive adapter fixtures add complexity for malformed responses unlikely in ordinary use; retain focused actual HTTP wiring and pure negatives.
- [low] [reject] Blind6: completion-boundary planning prose is stale during implementation. Finding requires editing this build specification, rejected under workflow rule. Final result records implementation/live limits independently.
- [medium] [patch] Edge1: mixed scalar/set bypass is the same root cause as Blind2; retain its own finding row and apply the same narrow guard.
- [low] [patch] Gap1: newly added prefix-list cardinality guard has no caller-level negative. Add a minimal real-helper HTTP fixture for valid/missing/malformed/mismatched counts; this is a helper boundary, not reachable provider corruption.
- [false] [reject] Intent1: additional fail-closed shared helper validation is expressly included by repair intent, with no production behavior change.
- [false] [reject] Intent2: owner-ID helper control is an input boundary, not an assertion that raw payload returns an independent UUID. Actual binding uses exact requested owner URL plus expected owner name; no false independent-payload claim.
- [low] [patch] Intent3: unequal-cardinality coverage is the same root cause as Blind4; add both positive directions.
- [false] [reject] Intent4: live gates remain unexecuted by this diff, and documents explicitly say so. No passing live claim is inferred from fixtures or review.

Grouped repairs: member-local group decoding (medium), mixed representation rejection (medium), unequal-cardinality positive controls (low), and actual import-helper prefix-count coverage (low). Same implementation child re-engaged; parent verification and complete fresh live rerun remain required.


## Auto Run Result

This bounded implementation/review is done; live acceptance and parent Story31.0h are not. Added a managed-state-only authoritative reference oracle at Step12, retained exact fresh canonical checks and all lifecycle gates, and hardened cardinality/mixed representation inputs without any provider behavior change.

Files: acceptance test support and negative controls; this repair specification and complete individual review triage; separate advanced live-evidence checkpoint retaining matrix04 as failed. No provider implementation, public schema or resource configuration changed.

Review: all12 findings individually triaged above. Four grouped repairs applied:2medium and2low. Null/missing member names now decode into a fresh local variable; mixed scalar/set shapes fail; unequal-cardinality positives verify both independent selections; actual import-helper HTTP controls cover prefix-list cardinality. Rejected findings and their specific reasons remain in the log. Follow-up review recommendation:true under the first-pass2medium threshold; named risk is the repaired oracle's as-yet-unexecuted interaction with real canonical policy selection during the complete live matrix. Bounded independent source recheck and parent verification are recorded separately, never substituted for appliance proof.

Verification observed before final parent checks: pinned DevRail Go1.26.10 linux/arm64; full affected race suitePASS6.340s and non-racePASS2.271s; package lint0issues. After review patches, targeted race controlsPASS1.040s and whitespacePASS. Parent owns full repository gate, deterministic generation, privacy audit, exact-source commit and new whole17-step run. No merged/released/legacy-firmware/OpenTofu or production claim.

Parent final verification: all affected quagga tests and race controls executed and passed6.458s, including every managed/fresh/cardinality/HTTP/mixed/null/lexical row. Full make checkPASS6/6, none skipped84841ms. Staged deterministic generation produced zero unstaged diff; whitespace and added-line privacy audit passed. Independent bounded source recheck confirmed both medium defects repaired, strict fresh versus managed semantics and actual caller coverage retained, and no residual defect in that scope. Full live matrix remains pending.
