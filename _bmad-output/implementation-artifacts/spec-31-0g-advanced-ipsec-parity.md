---
title: 'Story 31.0g: Advanced IPsec field parity'
type: 'feature'
created: '2026-11-03'
status: 'done'
baseline_revision: '4bd107fc769586d5904919c8e69562897a87e84d'
review_loop_iteration: 0
followup_review_recommended: true
context: ['AGENTS.md', 'DEVELOPMENT.md']
warnings: [oversized]
deferred:
  - summary: >-
      Selected-map reads cannot establish IKE/ESP proposal preference order.
    evidence: |-
      Existing SelectedMapList already sorted generated selected keys; API choice-map order is not persisted preference order. New behavior preserves prior spelling only for equal membership. Documented limitation; independent ordered-read evidence is needed before promising preference fidelity.
    location: >-
      internal/service/ipsec/advanced.go
    severity: medium
  - summary: >-
      Missing legacy policy fields may still decode as empty values.
    evidence: |-
      Pre-existing owners accepted absent legacy fields and replacement zero values. This additive strict decoder validates all new advanced fields but does not worsen or close legacy response completeness; a legacy-response validation follow-up needs source-shaped malformed old-field controls.
    location: >-
      internal/service/ipsec/connection_model.go and child_model.go
    severity: medium
  - summary: >-
      Legacy optional defaults retain their previous ownership behavior.
    evidence: |-
      Existing enabled/remote_addrs/proposals/version/unique and child legacy defaults remain unchanged for compatibility. New omission-safe serialization applies to additive advanced arguments; broad legacy ownership migration requires separate compatibility design and state/Core tests.
    location: >-
      internal/service/ipsec/connection_schema.go and child_schema.go
    severity: medium
---

<intent-contract>

## Intent

**Problem:** Existing connection/child owners omit 15 connection and six child controls from the pinned field audit; child multi-proposal reads lose values. Imported advanced policy cannot be represented faithfully.

**Approach:** Extend the existing resources and their read-only lookups against released OPNsense26.7.8, retaining existing names/types/UUID ownership. Deliver the entire audited advanced field surface, source-grounded protocol/Core tests, documentation and an executable guarded disposable acceptance path. Keep live proof explicitly outstanding if no isolated target can be attested.

## Boundaries & Constraints

**Always:** Source oracle is opnsense/core26.7.8 commit96978001a8343669154e4492172ce0a07dac7e4b, Swanctl.xml and ConnectionsController.php. Preserve existing `/api/ipsec/connections/{add,get,set,del}_{connection,child}`, monads and `/api/ipsec/service/reconfigure`; story's swanctl-controller URL is stale. Add connection TF→API fields: aggressive→aggressive, dpd_delay→dpd_delay, dpd_timeout→dpd_timeout, ike_lifetime→over_time, ip_pools→pools, keying_tries→keyingtries, local_addresses→local_addrs, local_port→local_port, mobike→mobike, reauthentication_time→reauth_time, rekey_time→rekey_time, remote_port→remote_port, send_certificate→send_cert, send_certificate_request→send_certreq, udp_encapsulation→encap. Keep existing remote_addrs (audit remote_addresses alias), proposals and unique. Child additions: close_action, dpd_action, install_policies→policies, request_id→reqid, sha256_96, rekey_time. All new attributes Optional+Computed without static defaults; boolean flags are Bool, decimal values are strings to preserve blank versus zero, remaining fields strings/CSV matching current public conventions. Resource writes include new fields only when explicitly configured, preserving omitted imported values and replay_window/auth children via API omission. Explicit false, zero and empty string are real values, not omission. Reject unresolved explicitly configured values before mutation. GET populates authoritative state; computed lookup schema matches model. Preserve CSV spelling/order when membership matches configuration/prior state to prevent perpetual string diffs. Do not change legacy public field types. Pin numeric limits0..500000, keying_tries0..1000, request_id blank or1..65535; child rekey_time cannot be blank and server default3600 applies on omission. Ports only blank(daemon500) or4500; send_certificate blank/ifasked/never/always; close_action none/trap/start; dpd_action clear/trap/start; start_action also accepts trap|start. Document ike_lifetime as overtime grace, NOT total IKE lifetime. Mutation rejection must not activate; saved mutations with reconfigure/read failures must retain known identity/state so retries do not duplicate creation. Strict read errors must not silently erase advanced policy. Full container gate and exact evidence required.

**Never:** Add parallel resource owners, move credentials/auth/key/PSK into these models/lookups, mutate shared/production appliances, fabricate live proof, claim parent31.0g complete without lifecycle evidence, rewrite unrelated story statuses, copy private hostnames/paths into public artifacts, install tools on host, or mark pending PR capabilities released. No automatic merges.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|---------------------------|----------------|
| Import/lookup/no-op | Source-shaped all-field responses, multiple selected proposals/addresses/pools, optional blanks and zero | Every field survives framework Read/lookup/import and actual protocol plan; CSV equivalent membership gives no diff; no writes | Malformed advanced field shape/value reports error without replacing prior state |
| Partial update/clear | Imported advanced state; new fields omitted vs explicit false/empty/zero; unrelated replay/auth and second object | Exact configured wire keys only, correct target UUID/envelope; omitted fields untouched; explicit clears accepted only where model permits; one activation after save | Unknown configured values and invalid enums/ranges fail before mutation |
| Create/drift/delete | Disabled connection and child, all advanced fields; drift then read/plan/repair, later delete child then parent | Authoritative readback and stable identity; repair only owned object; child deletion proven independently; no lookup mutation | Missing object and canceled context handled; no accidental object recreation on read |
| Failure/recovery | Validation rejection, malformed GET, save confirmed then reconfigure or readback fails | Error accurately distinguished; known new UUID and prior confirmed state retained; no duplicate add on retry; no secret payload diagnostic | Canceled operations return diagnostics; failed save never reconfigures |
| Compatibility | Legacy model/state and CSV strings; child multi-proposal and combined trap/start | Existing field names/types intact; all selected proposals retained; matching read-only lookup parity; secrets absent | No broad shared SelectedMap behavior change |

</intent-contract>

## Code Map

- `internal/service/ipsec/{connection,child}_{model,schema,resource,data_source}.go` -- existing handwritten owners; expand additive fields and conditional request serialization. Existing endpoints are correct. Child esp_proposals currently SelectedMap loses multiple choices; use SelectedMapList/CSVSelectedMapList and stable membership-aware CSV readback. Version is positional option array.
- `internal/service/ipsec/data_source_schema_test.go` -- existing HTTP-backed framework schema/state helpers. Add independent protocol plan/apply tests and request-count assertions, not only conversion mirrors.
- `pkg/opnsense/{types,crud,errors}.go` -- SelectedMapList, mutation wrappers, MutationReconfigureError; reuse without weakening shared contracts.
- Upstream `src/opnsense/mvc/app/models/OPNsense/IPsec/Swanctl.xml` lines6..107/228..321, fieldtypes IKEAddressField/PoolsField and `Swanctl.php`211..316; `controllers/OPNsense/IPsec/Api/ConnectionsController.php`79..178/247..277 -- immutable source oracle. BaseField.php810..830 proves omitted/non-null key semantics; setBase is upsert. Use only public immutable source citations in committed artifacts.
- `templates/{resources,data-sources}/ipsec_{connection,child}.md.tmpl`, `examples/`, generated `docs/` -- explain mapping, clears, imported omission, limits, reconfigure scope, secret boundaries and source/live evidence limits.
- `_bmad-output/implementation-artifacts/31-0g-ipsec-field-parity.md`, planning `provider-ecosystem-coverage-audit.json` and generated report -- localized status/evidence, retain live-proof gate.
- `CHANGELOG.md` -- current tag0.12.0, next additive version0.13.0(Unreleased) unless refreshed source tags dictate otherwise; explain auto-release effect.

## Tasks & Acceptance

**Execution:**
- `internal/service/ipsec/` -- implement all21additions, read-only schema parity, deterministic selection semantics and omission-safe serialization; regression tests for every matrix row including protocol planner.
- `internal/service/ipsec/*_test.go` and appropriate acceptance script/example -- executable disabled disposable connection/child create/import/no-change/drift/clear/delete proof, identity/reservation/isolation gate and owned cleanup; do not run without attestation. Existing infrastructure is not reachable; no credentials fallback.
- `templates/`, `docs/`, `examples/`, `CHANGELOG.md` -- source-accurate public behavior and unreleased status; generation deterministic after staging intended outputs.
- Story/audit files above -- record actual tests versus unexecuted live lifecycle, no unrelated count/status edits.

**Acceptance Criteria:**
- Given imported advanced objects, when actual framework protocol refresh/plan and read-only lookup run, then audited values are faithfully represented and omission causes no policy reset or perpetual diff.
- Given a configured advanced change or drift repair, when update runs, then only target-owned values change, unrelated values survive, and service reconfigure runs after successful persistence.
- Given invalid input, cancellation or partial failure, when resource operations execute, then diagnostics and recoverable state accurately reflect confirmed effects without exposing secrets or duplicating creation.
- Given legacy schema/state, when additive models and docs are consumed, then existing names/types remain valid, full child proposal selections survive, and auth/PSK/key owners remain separate.
- Given no positively attested isolated runtime, when delivery is reviewed, then live lifecycle remains unexecuted/draft with zero mutations and a concrete guarded proof path, not a claim of completed parent story.

## Spec Change Log

## Review Triage Log

### 2026-11-04 — Review pass 1

- Execution: blind, edge, verification-gap and intent layers all completed. Native four-active-thread ceiling refused the fourth simultaneous launch; intent ran immediately after one reviewer returned. No layer skipped.
- verdicts: 24 findings — high 3, medium 16, low 0, false 5, maybe-false 0
- findings:
  - `[medium]` `[defer]` Blind1 proposal preference ordering — pre-existing selected-map decoding already sorted selections; upstream option-choice order cannot prove persisted preference. Record inherited API limitation and clarify docs, without inventing an ordered read.
  - `[medium]` `[patch]` Blind2 recovery before config validation — confirmed Update invoked activation before unknown/invalid config rejection. Move request validation first; pending-marker regression requires zero mutations.
  - `[medium]` `[defer]` Blind3 absent legacy response fields — pre-existing legacy zero-value decoding exists independently of new fields. Record explicit legacy completeness follow-up, not a new advanced-parity claim.
  - `[medium]` `[patch]` Blind4 selection validation — new selection validator accepted duplicate/empty entries and arbitrary pool identifiers. Reject malformed lists and non-UUID/non-radius pools before save, preserving upstream address validation authority.
  - `[medium]` `[patch]` Blind5 imported update test — verified refreshed imported result was unused. Drive omission plan/update from that imported state.
  - `[medium]` `[patch]` Blind6 Core merge gap — manually supplied ProposedNewState proves protocol behavior but not Core merge/no-op. Add actual offline CLI-backed import/omission/CSV plan regression and keep evidence labels separate.
  - `[medium]` `[patch]` Blind7 delete recovery gap — add confirmed delete/activation failure and repeated-retry/absence/marker-consumption controls.
  - `[medium]` `[patch]` Blind8 fresh-provider recovery — instantiate a fresh provider from returned public/private state, assert retry counts and a further refresh has no activation.
  - `[medium]` `[patch]` Blind9 live omission path absent — executable acceptance must include imported omission-preserving update and independent raw before/after; real execution remains unperformed.
  - `[high]` `[patch]` Blind10 expiring live reservation — preflight-only lease allowed subsequent mutations after expiry. Require execution+cleanup lease runway and guard every mutation while preserving owned cleanup ability.
  - `[medium]` `[patch]` Gap1 populated pools masked by fixture — pre-verified gap: oracle seeded desired pools before request overlay. Assert independent exact advanced wire map and changed-populated pool update; omission counterfactual must fail.
  - `[medium]` `[patch]` Gap2 create activation marker assertion — pre-verified gap: removing Create mark call still passed. Count retry/zero repeated save and second refresh without activation.
  - `[medium]` `[patch]` Gap3 general acceptance workflow regression — TF_ACC=1 general workflow has no dedicated receipt. Dedicated guarded test skips without its separate acknowledgment, but explicit opt-in with bad/missing receipt fails closed.
  - `[high]` `[patch]` Edge1 permissive activation success — both request opts inherited non-strict reconfigure, so empty/unknown body could clear recovery. Require affirmative response and retain marker on malformed outcomes.
  - `[medium]` `[patch]` Edge2 duplicate selection collapse — same demonstrated serialization/readback defect as Blind4; reject duplicate normalized new selections before mutation.
  - `[medium]` `[patch]` Edge3 whitespace empty array — exact byte comparison rejected valid JSON [ ]. Decode structural empty arrays, reject nonempty arrays, cover resource and lookup.
  - `[medium]` `[patch]` Edge4 unknown config pending recovery — same ordering defect as Blind2, same validation-first repair and zero-mutation control.
  - `[high]` `[patch]` Edge5 lease expiry — same unsafe reservation runway as Blind10, same guarded acceptance repair. No live execution performed.
  - `[false]` `[reject]` Intent1 tagged-source versus fixture oracle — source verified against immutable released model/controller; tests accurately labeled offline interpretation, not executed PHP/appliance proof. No unsupported live claim exists.
  - `[false]` `[reject]` Intent2 actual appliance lifecycle not executed — authorized overnight job explicitly permits draft implementation with honest live blocker when no attested target. Story remains open, live evidence unexecuted and zero writes; not presented as completed lifecycle parity.
  - `[false]` `[reject]` Intent3 service-wide ownership proof — unchanged fixture objects establish only protocol targeting; documentation expressly says reconfigure is service-wide and real unrelated runtime proof remains gated. No object-scoped runtime claim made.
  - `[false]` `[reject]` Intent4 secret-owner lifecycle untested — these changes do not alter secret owners; schema exclusion, redaction and no auth wire writes are verified. Actual secret lifecycle is not claimed by this additive field story.
  - `[medium]` `[defer]` Intent5 broad legacy omission — inherited legacy defaults/serialization predate story and remain compatibility behavior; advanced omission is implemented and docs bound the guarantee. Record separate legacy ownership compatibility follow-up.
  - `[false]` `[reject]` Intent6 epic context rewrite — installed mandatory BMAD step01 compiled stale cached context from current planning files. This generated planning input is required workflow context, not unrelated runtime scope or fabricated project status.


## Verification

- Container `go test -count=1 -race -v ./internal/service/ipsec/... ./internal/coverageaudit/...` -- all named offline matrix tests execute/pass; acceptance skips clearly distinguished.
- Container `go generate ./...` twice after staging intended outputs -- deterministic, no unstaged generated differences.
- `make check DEVRAIL_IMAGE=ghcr.io/devrail-dev/dev-toolchain:1.12.24` -- all6gates pass, no suppressions.
- `git diff --check` and complete public privacy diff inspection -- clean; no internal data.
- Guarded acceptance only on newly reserved and attested isolated guest; otherwise record zero writes/live proof blocked and keep PR draft.


### Independent repaired-boundary recheck

- verdicts: 1 finding — high 1, medium 0, low 0, false 0, maybe-false 0
- `[high]` `[patch]` Pending-activation/expired-window cleanup interaction — Delete attempted superseded activation before owned removal, and Core destroy refresh could fail first. Delete now activates only its final deleted state; a separate guarded bounded-context fallback deletes only recorded UUIDs child-before-parent, requires independent exact absence and preserves failure/uncertain ownership. Offline interaction regression covers expiry, pending marker, changed identity, failed absence and unrelated object survival.

## Auto Run Result

Implemented all 15 connection and six child audited controls on existing resources/lookups, omission-preserving requests, strict advanced decode and activation success, complete multi-proposal membership, compatible legacy field types and recoverable known UUIDs. The parent story remains in progress: live lifecycle is unexecuted and the PR must stay draft.

Files: `internal/service/ipsec/advanced.go` owns additive schemas/serialization/decode/recovery; existing connection/child models/resources/lookups integrate it. Protocol/Core/guard tests cover matrix rows and repairs; examples/templates/generated docs describe truthful limits; owning story and audit record offline-only evidence; changelog adds0.13.0 and reconciles released0.12.0.

Review: all24first-pass findings classified individually above (3high,16medium,5false); 2high and10medium root-cause repair groups,3inherited limitations deferred,5false findings refuted. One additional high cleanup interaction from independent repair recheck was fixed. No layer skipped. Follow-up review recommended:true due repaired private-marker/lease/cleanup interactions not yet executed on an attested appliance, plus inherited legacy omission/response and proposal-order limits. None of these is claimed live-proven.

Parent independently executed the entire offline IPsec/coverage-audit race matrix and real Terraform Core import/omission/CSV no-change tests. All matrix rows have executed controls: lifecycle/import/lookup/no-op: `TestIPsecAdvancedProtocolLifecycle`, `TestIPsecCoreImportedOmissionAndCSVNoChange`; partial update/clear: lifecycle, `TestIPsecPopulatedPoolUpdateWire`, `TestIPsecPendingInvalidUpdateNeverActivates`, selection validation; create/drift/delete: lifecycle and `TestIPsecPendingDeleteSupersedesActivation`; failure/recovery: create/update failure tests, `TestIPsecStrictActivationRecoveryAndDelete`, `TestIPsecExpiredPendingRefreshIndependentCleanup`; compatibility: schema test and full selected-membership assertions. Malformed every-field and whitespace-array tests cover strict decode; live guard test is offline transport orchestration, not an appliance result.

Counterfactuals executed by implementer and inspected by parent: removing only pools serialization makes exact populated-pool wire test fail; removing Create marker calls makes strict recovery tests fail. Exact original file bytes restored before final tests. General TF_ACC without separate acknowledgment skips only this dedicated test; explicit opt-in without receipt fails before network. Source pin independently checked against GitHub tag/peeled commit and model SHA256.

Final verification outcomes are appended below after completion. Residual gates: new reservation/identity/isolation, real26.7.8 lifecycle/service activation, precise owned cleanup/restore, older-core and OpenTofu lifecycle; no target reserved, no appliance mutation performed. Proposed0.13.0 heading auto-releases after merge+mainCI and requires refresh if another pending release merges first. No merge/release claimed.


### Final parent verification

- Final repaired-tree container race suite PASS: IPsec9.019s and coverageaudit1.422s. All20named offline IPsec tests executed, including actual Terraform Core4.67s; dedicated live test skipped honestly.
- Two final post-cleanup-repair generation runs exited0 and left zero unstaged generated diff after each; whitespace/privacy audits clean. Frontmatter parsed in container: one deferred list containing3items.
- Required final `make check` exited0: all6gates passed, none skipped,58979ms. Parent's earlier nested-container invocation failed because docker was unavailable inside that image; corrected host make with container DOCKER_RUN and read-only Git metadata, no configuration changes/suppression. Final lint8009ms/format313ms/test41549ms/security1816ms/scan6975ms/docs311ms.
- Independent bounded recheck confirmed its cleanup interaction repaired, with no remaining concrete flaw found in that fix. Read-only source review is distinct from the executed tests.
- Build-auto implementation/review artifact is done; owning31.0gstory remains in progress pending real lifecycle proof. No appliance resources reserved/mutated. Task-only development containers/volumes retired; isolated source worktree intentionally retained for draft PR review. No automatic merge.


## Sequential merge review — 2026-11-04

Current baseline is released v0.14.0; this increment now targets v0.15.0.
The maintainer explicitly authorized unattended sequential validation/merging;
this selects applying all unambiguous review patches without another approval.
The full31-file/4,422-line diff was reviewed, rather than silently chunked.
All four independent BMAD layers completed. The first fourth launch hit the
native concurrency ceiling; it ran after a slot opened, before any verdicts.
No layer failed or was skipped. All6 integration gates pass (67,996ms); this
is not live proof or final repaired-head verification.

### Review Triage Log — integration pass

15 raw findings individually classified before grouping:5high,8medium,1low,
1false. Two duplicate root causes are grouped only after these verdicts;
12 patch groups,0 decision-needed,0 new defer,1 rejected. Existing inherited
legacy-ownership/read-completeness and proposal-order limits remain separate.

1. [high][patch] Blind1: ReqOpts omit StrictDeleteResponse; loose parser accepts ambiguous success and state can be forgotten. Require affirmative deletion for both owners, with malformed response/no-activation/state-retention controls.
2. [medium][patch] Blind2: Selection decoder drops invalid scalar members and skips field-domain read validation. Validate scalar/mapped memberships before state replacement.
3. [high][patch] Blind3: CleanupAllowed bypasses actual ExpiresAt. Stop every mutation at hard expiry, including after identity-read latency; execution cutoff and hard expiry are distinct.
4. [medium][patch] Blind4: Owned UUIDs exist only in memory; failed/crashed cleanup leaves no usable promised receipt. Persist a private journal before dispatch and as confirmed outcomes change.
5. [high][patch] Blind5: Go test global timeout does not guarantee t.Cleanup or descendant quiescence. Require an external bounded process-group runner with independent finally cleanup/restoration and reserved time.
6. [medium][patch] Blind6: Live imports do not persist imported Core state; omission uses original created state. Persist/import both owners and test the real imported-state path.
7. [medium][patch] Blind7: Live data sources execute but their values are not compared. Assert all advanced lookup outputs against independently observed expected values.
8. [medium][patch] Blind8: Live ipsecAbsent matches textual [] despite existing structural helper. Reuse structural affirmative absence including wrapped/unwrapped whitespace variants.
9. [low][patch] Blind9: Current owning-story release instructions still point to the already published0.13.0 rather than0.15.0. Correct current instructions, preserving clearly labelled historical proposals.
10. [false][reject] Blind10: The spec explicitly calls its original build-auto implementation artifact done while the owning31.0g story and live gate remain in progress/draft. It does not claim completed parity; no contrary automated consumer was demonstrated. Editing the spec to remove the gate is not an implementation repair and is rejected by review rules.
11. [high][patch] Edge1: Same loose-delete defect/outcome as Blind1; independently verified and grouped after verdict.
12. [medium][patch] Edge2: Same malformed-selection read defect as Blind2; independently verified and grouped after verdict.
13. [medium][patch] Edge3: A dispatched add with lost/malformed/no-UUID response is not owned or marked uncertain; empty-map cleanup can falsely succeed. Journal unresolved dispatched adds and fail reconciliation closed, without inventory discovery/deletion.
14. [medium][patch] Verification gap: All aggressive/sha256_96 controls use false; hardcoded-false mapping survives. Add enabled resource/import/lookup and false-to-true drift controls; verify counterexamples.
15. [high][patch] Acceptance auditor: Confirmed Create writes UUID then AddError; Core taints and replaces it despite private activation recovery. Warning-only confirmed-create recovery plus actual Core failed-create/retry controls must prove one add/no replacement. Unknown/unconfirmed effects remain errors.

### Review Findings

- [ ] [Review][Patch] Affirmative deletion and ambiguous-response controls (findings1/11).
- [ ] [Review][Patch] Strict selection membership/domain reads (findings2/12).
- [ ] [Review][Patch] Hard lease expiry even for cleanup (finding3).
- [ ] [Review][Patch] Private durable ownership journal (finding4).
- [ ] [Review][Patch] External bounded guarded runner with quiescent finally cleanup (finding5).
- [ ] [Review][Patch] Persisted live imports before omission update (finding6).
- [ ] [Review][Patch] All-field authoritative live lookup assertions (finding7).
- [ ] [Review][Patch] Structural exact absence (finding8).
- [ ] [Review][Patch] Current release references and historical boundaries (finding9).
- [ ] [Review][Patch] Uncertain dispatched-add accounting (finding13).
- [ ] [Review][Patch] True boolean/read/import/lookup/drift counterexamples (finding14).
- [ ] [Review][Patch] Actual Core confirmed-create recovery without taint (finding15).

### Rejected

- false — finding10 conflates explicitly bounded implementation-artifact completion
  with the separate open parent/live gate; no completed-parent claim is present.
