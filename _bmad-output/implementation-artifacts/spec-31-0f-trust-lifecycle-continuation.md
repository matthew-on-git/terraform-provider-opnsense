---
title: 'Story 31.0f: Remaining Trust lifecycle and read parity'
type: 'feature'
created: '2026-11-03'
status: 'done'
baseline_revision: '4bd107fc769586d5904919c8e69562897a87e84d'
review_loop_iteration: 0
followup_review_recommended: false
context:
  - 'AGENTS.md'
  - 'DEVELOPMENT.md'
  - '_bmad-output/implementation-artifacts/epic-31-context.md'
warnings: [oversized]
deferred:
  - summary: >-
      Historical unavailable-live-proof gate is resolved by the recorded fresh guarded matrix.
    evidence: |-
      Initially no attested target was available. That historical blocker was subsequently resolved; candidate35799c0 passed the full fresh26.7.8_2 guarded run with26apply/no-change pairs, final digest gates, Terraform deletion, exact restoration and independent retirement. Historical failed receipts remain retained.
    location: >-
      test/scripts/run-story-31-0f-acceptance.sh
    resolved: true
    severity: medium
  - summary: >-
      Keyless CA import/update remains unsupported by the provider despite the upstream patched-package correction.
    evidence: |-
      Original26.7.8 source required exactly one private-key section. Authentic26.7.8_2 corrects that condition, but this bounded provider increment still conservatively requires a matching key. Enabling keyless support requires a separate provider implementation/proof increment; no fake key or XML workaround is permitted.
    location: >-
      _bmad-output/implementation-artifacts/31-0f-trust-source-contract.md
    severity: medium
---

<intent-contract>

## Intent

**Problem:** Existing Trust CA/certificate owners cover only self-signed creation and certificate import, misinterpret volatile editor values, and omit certificate lookup and full released action/read semantics. Settings shipped separately and must not be reimplemented.

**Approach:** Extend the existing owners additively with released action-specific contracts, safe metadata updates, authoritative public/sensitive outputs, one-time key retention and certificate lookup. Validate every buildable remaining lifecycle; explicitly retain upstream and unavailable-live-proof gates rather than claiming full parent completion.

## Boundaries & Constraints

**Always:** Preserve resource names, UUID import, legacy attribute types and equivalent configured PEM input text; add optional fields without destructive state migration. CA omission retains legacy internal-create behavior; certificate omission retains legacy import behavior, inferred at creation rather than copied from GET. Implement CA existing/import, internal root/intermediate and OCSP; cert internal, external CSR, import, sign_csr, import_csr, reissue and safe manual behavior using released26.7.8 semantics. Separate creation/operation intent from API-inferred action, reconstructed metadata and template defaults. Preserve configured lifetime instead of overwriting with editor825/397; retain legacy CA lifetime type and document its semantics honestly. Subject/SAN/digest/key type/issuer/cert_type and operational fields must use exact released wire names; SANs are newline-delimited; EKU purpose is derived metadata. All private-key inputs and outputs are Sensitive. Legacy certificate/private_key inputs stay text-preserved; distinct returned outputs may report normalized API payloads. Capture local-only internal key from top-level add response into a retained Sensitive output before follow-up reads; do not retransmit it or pretend lookup/import can recover it. Serialize mutation/readback boundaries with existing global lock, use authenticated single-attempt POST and redirect rejection, preserve read policy. Retain confirmed identity and available one-time key when post-add read fails; do not leak raw payloads/errors. Ordinary description edits never generate, sign or reissue: cert uses manual; CA uses existing plus current certificate/key. Action transitions and explicit repeat-operation intent must be distinguishable from metadata edits; local-only generation through set must be rejected because set cannot return the key. Preserve merged settings and recovery tests unchanged.

**Never:** Expose write-only/private material through unsafely public outputs; use a GET editor action as original creation provenance; claim editor lifetime is actual validity; create fake private keys or raw-XML workarounds; replay non-idempotent actions automatically; introduce competing owners; mutate shared/production appliances; claim live, merged or released proof from local tests. CA certificate-only import is upstream-broken on this pinned release: reject before POST with focused diagnosis, including metadata update of an existing keyless CA, never regenerate it. Public changes must be consumer-neutral. Keep full parent31.0f open until all lifecycle/live gates have evidence; any PR without executed live matrix stays draft.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|---|---|---|---|
| Legacy compatibility | Existing CA internal config and cert/key import config or imported state | Existing names/types/configured PEM survive; optional new intent absent on import stays absent rather than fake default | No forced migration/replacement solely from added attributes; configured unknown intent must not become empty/default writes |
| CA action matrix | Root, issuer-linked intermediate, OCSP, matching-cert/key existing import | Exact released payload including subject/key/digest/lifetime/OCSP, authoritative read and ID | Missing required action inputs, mismatched key and cert-only imports fail before writes; no secret-bearing diagnostics |
| Cert action matrix | Internal/firewall, internal/local, external CSR, import, sign_csr, import_csr, reissue/manual | Exact action payload, cert_type/SAN wire mapping and action-specific prerequisites; meaningful transitions use explicit intent changes | Invalid/unknown fields and unsupported local-only set generation rejected without write; no automatic replay |
| Metadata and refresh | Nondefault generation inputs plus GET inferred action/template lifetime; description edit | No regenerate/sign/reissue; preserve configured intent/PEM; strict public read outputs and numeric/string timestamps | Malformed/missing monad and transport errors retain managed state; only affirmative missing-item representation removes it |
| Lookup and sensitivity | CA/cert UUID, public PEM/CSR/subject/validity/purpose plus API-returned private material | Read-only lookup with authoritative metadata, Sensitive private fields, null when absent; no synthetic generation provenance or one-time key | No API writes, malformed values diagnose instead of empty successful state |
| One-time/partial failure | Successful add returns UUID+private_key then GET fails or later omits key | Persist UUID and Sensitive one-time output before failure; retain across refresh and never send key back | Lost response cannot be falsely recovered; do not replay add or emit response payload; document irrecoverable boundary |
| Acceptance gate | No newly reserved identity-bound disposable runtime supplied | Guarded executable acceptance entry point fails closed before any API access; explicit unexecuted live-matrix disposition | Runtime must attest isolation/identity/version/endpoint, own exact restore and cleanup, and reject unsafe/missing guard inputs |

</intent-contract>

## Code Map

- `internal/service/trust/{ca,cert}_{model,schema,resource}.go`: current owners; CA update currently replays internal generation, cert generic Add loses one-time key. Extend these, do not duplicate settings.
- `internal/service/trust/ca_data_source.go`, new `cert_data_source.go`, `exports.go`: separate read models from new resource-only intent fields; add cert registration and accurate sensitivity.
- `internal/service/trust/settings_transport.go`: existing authenticated GET-policy/single-attempt POST/redirect-disabled requester available; preserve settings contracts. `pkg/opnsense/client.go` provides global mutex/read limits; generic CRUD is unsuitable when top-level add key must survive.
- Released source oracle: OPNsense26.7.8 commit `96978001a8343669154e4492172ce0a07dac7e4b`, `src/opnsense/mvc/app/controllers/OPNsense/Trust/Api/{Ca,Cert}Controller.php`; models `src/opnsense/mvc/app/models/OPNsense/Trust/{Ca,Cert}.xml`, `FieldTypes/{CAs,Certificates}Field.php`; library `src/opnsense/mvc/app/library/OPNsense/Trust/Store.php`. CA controller55–155 owns generation/import;117–129 proves keyless-import defect. Cert82–190 owns actions;215–225 returns one-time key only from add. Field loaders infer action; Store457–458 emits numeric timestamps and never reconstructs lifetime. CSR parser reconstructs DN only, so retain configured SAN/type/key/digest inputs.
- `internal/service/trust/{ca,cert}_resource_test.go`, `settings_protocol_test.go`, `data_source_schema_test.go`: legacy acceptance/schema and protocol patterns. Add actual framework/protocol contract tests, not only serializers.
- `internal/coverageaudit/`, planning `provider-ecosystem-coverage-audit.{json,md}`: current113resource/99lookup baseline; new cert lookup increments lookup only; remaining gaps and partial/live-blocked Trust classification must be accurate.
- `templates/{resources,data-sources}/trust_*.md.tmpl`, `examples/`, generated `docs/`, `CHANGELOG.md`: public contracts and next unreleased version beyond published/tagged0.12.0. No local paths or private topology in public files.
- `test/scripts/run-story-31-0o-acceptance.sh` and corresponding Core tests: safety pattern for external attested oracle/cleanup, not reusable credentials or authority to resurrect old guests. No current newly reserved target/runtime was found by parent; implement an executable guarded path, leave live proof explicitly unexecuted if safety cannot be established.

## Tasks & Acceptance

**Execution:**
- `internal/service/trust/` -- implement all buildable action/read/lookup scope and legacy-safe explicit operation semantics; source-check each action rather than invent payloads.
- `internal/service/trust/*test.go` -- cover every matrix row via executed model/schema and HTTP-backed framework/protocol tests, failure counterexamples, no unintended action replay and secret retention.
- `test/scripts/` and Trust acceptance tests -- supply executable fail-closed disposable action/settings matrix entry point with owned cleanup/restoration hooks; never execute against unbound endpoints. Document live requirements honestly.
- `templates/`, `examples/`, `docs/`, `CHANGELOG.md` -- document all new fields, sensitive state, import/operation/irrecoverable one-time boundaries and unsupported CA cert-only behavior; regenerate deterministically.
- Owning `31-0f-trust-lifecycle-and-settings-parity.md` and audit -- record precise implementation versus unit/protocol/live gates; avoid unrelated story rewrites.

**Acceptance Criteria:**
- Given legacy configurations and imported UUID state, when real framework read/plan/update runs, then compatible state and PEM are retained without false generation-provenance defaults or metadata-driven regeneration.
- Given every supported Trust action and each invalid prerequisite, when protocol create/update runs, then exact released requests execute once or fail before mutation, and confirmed saved identity/available one-time key survive read failure.
- Given lookup reads, when public and private fields are returned or omitted, then metadata is authoritative and every returned private-key value is Sensitive, with absent one-time values never fabricated.
- Given unavailable safe disposable authority or upstream keyless-CA defect, when evidence is published, then exact blocker and follow-up are explicit, parent story remains open and PR stays draft; executed offline proof is not mislabeled live acceptance.

## Spec Change Log

## Review Triage Log

### 2026-11-03 — Review pass 1

- Execution: four independent native review layers completed. Runtime rejected simultaneous gap/intent launches at its thread cap; both subsequently ran fresh as slots became available. No layer was skipped; triage began only after all four results.
- verdicts: 21 findings — high 4, medium 13, low 1, false 3, maybe-false 0
- findings:
  - `[medium]` `[patch]` Blind1: CA legacy certificate freezes after regeneration/external replacement — Verified fromRecord only refreshes unknown/null, unlike original computed API observation; restore authoritative legacy refresh with regression.
  - `[medium]` `[patch]` Blind2: CA legacy common_name/country freeze after first observation — Same known-state guard loses legacy computed drift; refresh source-backed observations while keeping separate volatile intent semantics.
  - `[low]` `[patch]` Blind3: Unconfigured legacy CA lifetime can report editor default — Configured lifetime is preserved and lookup already labels editor default, so no fabricated validity is exposed there; clarify resource description directly without changing compatible legacy integer behavior.
  - `[high]` `[patch]` Blind4: Removing action/token can trigger regeneration — Verified null/changed intent enters explicit branch then internal/import fallback; require affirmative nonempty operation intent and removal controls.
  - `[false]` `[reject]` Blind5: Import into configuration with explicit operation intent requests an operation — This is the documented action/token contract, not accidental legacy behavior. Legacy configurations have neither new field; the existing Core import path passes. Description-only no-change adoption is explicitly documented; silently consuming a new explicit operation is not requested.
  - `[medium]` `[patch]` Blind6: Already-absent or uncertain delete lacks affirmative absence success — A refresh/delete race or committed response loss reaches strict deleted-only response check; confirm absence with bounded GET and retain all ambiguous errors.
  - `[medium]` `[patch]` Blind7: New transition/local-key/error persistence lacks Core coverage — Verified only legacy lifecycle/import uses resource.UnitTest; extend bounded offline Core coverage without claiming appliance proof.
  - `[medium]` `[patch]` Blind8: Runner imports signed CSR result into a fresh object — Verified csr_import new resource misses existing holder key/UUID transition; use the existing external holder with detached CSR to avoid cycles.
  - `[medium]` `[patch]` Blind9: Runner omits CA/cert UUID-import no-change adoption — Only settings import exists; add private adoption workspace with description-only configs and remove adoption state only.
  - `[false]` `[reject]` Blind10: Unreleased version heading must be disabled until live proof — Repository AGENTS explicitly requires versioned heading and explanation before a code PR. Overnight authority requires draft PR when live proof is absent and forbids auto-merge. No automatic publication happens from this draft branch; publication still requires maintainer merge and successful main CI.
  - `[high]` `[patch]` Edge1: Null action removal selects internal generation — Same demonstrated cause as Blind4; affirmative-intent guard and removal regression cover it.
  - `[high]` `[patch]` Edge2: Non-object helper JSON aborts cleanup with AttributeError — Verified value.get before type validation and finally catches only named errors; reject nonobjects as AcceptanceError so every remaining restore gate is attempted.
  - `[high]` `[patch]` Edge3: Unbounded Terraform child can prevent cleanup — Verified subprocess.run has no timeout and asynchronous cancellation can wait on live children; bound, terminate and reap only owned process group before restoration.
  - `[medium]` `[patch]` Gap1: Intermediate issuer linkage has no observable regression assertion — Pre-verified gap: fixture ignores caref and no returned_ca_ref assertion distinguishes root/intermediate; persist and assert issuer in existing protocol matrix.
  - `[medium]` `[patch]` Gap2: Cleanup orchestration has no valid-runtime failure test — Pre-verified gap: all main tests exit before finally; inject offline fake-runtime command/cleanup failures and mismatch to verify phase continuation and failure propagation.
  - `[medium]` `[defer]` Intent1: Full operational completion remains unproven — Real external prerequisite existed before this change: no newly reserved attested runtime is available. Overnight instructions explicitly permit implemented/tested draft with live proof blocked and continued queue. Parent story stays open; zero appliance writes.
  - `[medium]` `[patch]` Intent2: Protocol action matrix is wider than actual Core coverage — Same concrete coverage risk as Blind7; add explicit/error/local-key Core controls, keep remaining protocol versus live distinctions explicit.
  - `[medium]` `[defer]` Intent3: Restoration attestation contract is not executed restoration proof — Same unavailable disposable prerequisite as Intent1, not a new claim of proof; no helpers/receipts are fabricated and successful live evidence remains a named future gate.
  - `[medium]` `[patch]` Intent4: Dormant runner exercises only one of six audited settings — Verified only enable_config_constraints toggles; extend five bool false/restore and authoritative cipher selection/clear/restore fixture. Protocol fault cases remain separately proven, not live.
  - `[medium]` `[defer]` Intent5: Unqualified CA import has a released keyless exception — Pinned controller unconditionally requires one private-key section. This predates provider changes; matching-key import is implemented, cert-only import is diagnosed, no fabricated key/XML workaround or full parent completion is claimed.
  - `[false]` `[reject]` Intent6: Configured intent is not the same as every live drift observation — User explicitly requires preserving volatile action/lifetime intent; separate returned observations satisfy that semantic distinction. Concrete stale legacy computed attributes are separately accepted under Blind1/2, not excused by this rejection.

## Design Notes

Action is operation intent, not durable appliance metadata. Preserve explicit configured values through reads; separate GET editor observations. Changes requesting a real action must be deliberate and testable, while description changes are safe in-place metadata edits. Local-only generated key can be irretrievable after response loss; retaining a known key/identity is required, claiming recovery without evidence is forbidden.

## Verification

- All Go tests, generators and linters run inside `ghcr.io/devrail-dev/dev-toolchain:1.12.24`; no host tool installation. Parent owns final whole-tree gates; implementation runs focused tests.
- Container `go test -count=1 -race -v ./internal/service/trust/... ./internal/coverageaudit/...` -- every offline matrix covering test passes without skipped proof rows; existing basic TF_ACC tests remain distinct unexecuted live tests.
- Container generation twice with intended outputs staged -- deterministic, no unstaged generated changes.
- `DEVRAIL_IMAGE=ghcr.io/devrail-dev/dev-toolchain:1.12.24 make check` -- six gates pass; no suppression or skipped required check.
- Guard-negative acceptance test -- missing/unattested authority refuses before API access. Actual disposable action/settings matrix only if newly reserved/identity-bound endpoint can be established, otherwise persist exact live blocker/draft status.
- `git diff --check` plus whole publication diff privacy audit -- clean; no internal paths, hosts, network allocations, credentials or raw private-key payloads.

## Parent matrix audit

The parent independently ran container `go test -count=1 -race -v
./internal/service/trust/... ./internal/coverageaudit/...` successfully. Matrix
coverage is bound to these executed tests, not implementation-agent claims:

- Legacy compatibility: `TestTrustLifecycleCoreLegacyCreateRefreshMetadataImport`,
  `TestTrustLifecycleProtocolLegacyReplacementCompatibility`,
  `TestTrustLifecycleProtocolLegacyDescriptionOnlyCA` and lookup/import/legacy.
- CA/certificate actions: all 12 subcases of
  `TestTrustLifecycleProtocolActionMatrix`, 16 invalid-prerequisite subcases,
  explicit transitions/local-set rejection and stored-key/CSR hydration with
  mismatched-signed-CSR rejection.
- Metadata/read: the action matrix checks safe metadata writes and unchanged
  plans; partial-failure/strict-read tests cover malformed truth and numeric
  timestamp decoding; uncertain-update tests retain consumed action/token and
  unresolved status across refresh/metadata without replay.
- Lookup/sensitivity: `TestTrustLifecycleProtocolLookupSensitivityImportAndLegacy`
  covers observed PEM values, sensitive schema, null missing keys, no writes and
  absent operation provenance.
- One-time/partial: partial-failure/key-retention and CA partial-create/delete
  tests preserve known identity; single-attempt/redirect rejection covers
  server failure, validation, malformed/duplicate and lost-identity replies.
- Acceptance gate: ten `trust_acceptance_guard_test.py` methods passed,
  including eight unsafe-attestation variants, executable synthetic-runtime
  failure/cleanup/hash-veto paths and actual owned-subprocess cancellation.
  TERM-resistant descendant controls cover PIPE and redirected-log modes.
  These helpers are test doubles, not appliance attestations. Shellcheck `-x`,
  shfmt, Ruff and offline validation of the full runner HCL with the built
  provider passed inside the pinned toolchain.

Existing `TestAccTrustCA_basic` and `TestAccTrustCert_basic` were skipped because
live acceptance was not enabled; they are not counted as passed matrix proof.
The new offline Core tests actually executed and passed. No live matrix,
OpenTofu execution or appliance restoration is claimed. Generation succeeded
and a further staged rerun produced no unstaged changes.

### 2026-11-04 — Bounded repair rechecks (not a fresh four-layer pass)

- verdicts: 2 findings — high 1, medium 0, low 0, false 1, maybe-false 0
- findings:
  - `[false]` `[reject]` Configured CA PEM will be normalized by the pinned appliance after authoritative refresh — released CaController existing action stores `base64_encode(crt_payload)` and CAsField returns `base64_decode`; text is preserved on this supported import path. Unlike certificate inputs, CA certificate was historically computed-only. A proposed private provenance/plan-modifier extension for hypothetical normalization was removed rather than guarding an undemonstrated released state. Returned observations stay authoritative. A synthetic normalization experiment is not source evidence.
  - `[high]` `[patch]` Redirected-log parent exits on TERM while an owned descendant ignores TERM — independently reproduced. Parent completion no longer skips unconditional owned-group SIGKILL escalation before restoration. Both PIPE and actual redirected-file controls launch a TERM-resistant descendant and require it to be stopped. The bounded edge recheck found no remaining flaw in that corrected path.

All twelve first-pass patch groups were applied: three high, eight medium and
one low. The redirected-log correction completes the existing cancellation
root cause, not another public capability. Explicit removal-only intent never
requests crypto; old computed CA certificate/CN/country refresh authoritatively;
uncertain deletes require affirmative bounded absence; issuer linkage and real
offline Core explicit/local/uncertain behavior are asserted. Runner fixtures
now sign/import into the same CSR holder, adopt both UUIDs with zero-crypto
receipts and exercise all six settings. Live proof remains unavailable.

## Auto Run Result

Implemented the full independently buildable remaining Trust continuation:
released CA/certificate actions, safe metadata boundaries, explicit consumed
operation intent, strict authoritative reads, Sensitive returned/private and
one-time-key state, legacy adoption and certificate lookup. The merged settings
implementation and recovery remain unchanged. Parent story status remains
in-progress because the live matrix/restoration and upstream keyless-CA gate
are not resolved by offline code.

### Reviewed files changed

- `CHANGELOG.md` — Next unreleased version and automatic publication effect.
- `_bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md` — Open parent status and guarded disposable proof contract.
- `_bmad-output/implementation-artifacts/31-0f-trust-source-contract.md` — Pinned released-source contracts and limitations.
- `_bmad-output/implementation-artifacts/epic-31-context.md` — Regenerated owning epic context after the merged settings baseline.
- `_bmad-output/implementation-artifacts/spec-31-0f-trust-lifecycle-continuation.md` — Frozen intent, acceptance matrix, individual review triage and run result.
- `_bmad-output/planning-artifacts/provider-ecosystem-coverage-audit.json` — Additive lookup inventory and explicitly incomplete Trust lifecycle evidence.
- `_bmad-output/planning-artifacts/provider-ecosystem-coverage-audit.md` — Additive lookup inventory and explicitly incomplete Trust lifecycle evidence.
- `_bmad-output/planning-artifacts/support-matrix.md` — Additive lookup inventory and explicitly incomplete Trust lifecycle evidence.
- `docs/data-sources/trust_ca.md` — Generated documentation or source template for the public lifecycle/read contract.
- `docs/data-sources/trust_cert.md` — Generated documentation or source template for the public lifecycle/read contract.
- `docs/index.md` — Generated documentation or source template for the public lifecycle/read contract.
- `docs/resources/trust_ca.md` — Generated documentation or source template for the public lifecycle/read contract.
- `docs/resources/trust_cert.md` — Generated documentation or source template for the public lifecycle/read contract.
- `examples/data-sources/opnsense_trust_cert/data-source.tf` — Consumer-neutral operation or UUID-import example.
- `examples/resources/opnsense_trust_ca/import.sh` — Consumer-neutral operation or UUID-import example.
- `examples/resources/opnsense_trust_ca/resource.tf` — Consumer-neutral operation or UUID-import example.
- `examples/resources/opnsense_trust_cert/import.sh` — Consumer-neutral operation or UUID-import example.
- `examples/resources/opnsense_trust_cert/resource.tf` — Consumer-neutral operation or UUID-import example.
- `internal/coverageaudit/audit_test.go` — Executed offline schema, protocol, Terraform Core or guarded-orchestration regressions.
- `internal/service/trust/ca_data_source.go` — Read-only authoritative observations with Sensitive key output and no fake provenance.
- `internal/service/trust/ca_model.go` — Configured intent versus authoritative observation mapping and safe payload construction.
- `internal/service/trust/ca_resource.go` — Existing owner lifecycle, serialization, partial identity retention and non-replay semantics.
- `internal/service/trust/ca_schema.go` — Additive operation and returned observation schema with preserved legacy types.
- `internal/service/trust/cert_data_source.go` — Read-only authoritative observations with Sensitive key output and no fake provenance.
- `internal/service/trust/cert_model.go` — Configured intent versus authoritative observation mapping and safe payload construction.
- `internal/service/trust/cert_resource.go` — Existing owner lifecycle, serialization, partial identity retention and non-replay semantics.
- `internal/service/trust/cert_schema.go` — Additive operation and returned observation schema with preserved legacy types.
- `internal/service/trust/data_source_schema_test.go` — Executed offline schema, protocol, Terraform Core or guarded-orchestration regressions.
- `internal/service/trust/exports.go` — Register the additive certificate lookup.
- `internal/service/trust/legacy_plan_modifier.go` — Conditional legacy replacement compatibility without new explicit-operation replacement.
- `internal/service/trust/lifecycle_config.go` — Reject unknown configured intent before mutation.
- `internal/service/trust/lifecycle_core_test.go` — Executed offline schema, protocol, Terraform Core or guarded-orchestration regressions.
- `internal/service/trust/lifecycle_pem.go` — Certificate/key/CSR prerequisites and matching checks without secret diagnostics.
- `internal/service/trust/lifecycle_protocol_test.go` — Executed offline schema, protocol, Terraform Core or guarded-orchestration regressions.
- `internal/service/trust/lifecycle_transport.go` — Authenticated single-attempt write, strict response and bounded authoritative read policy.
- `internal/service/trust/lookup_model.go` — Read-only authoritative observations with Sensitive key output and no fake provenance.
- `templates/data-sources/trust_ca.md.tmpl` — Generated documentation or source template for the public lifecycle/read contract.
- `templates/data-sources/trust_cert.md.tmpl` — Generated documentation or source template for the public lifecycle/read contract.
- `templates/index.md.tmpl` — Generated documentation or source template for the public lifecycle/read contract.
- `templates/resources/trust_ca.md.tmpl` — Generated documentation or source template for the public lifecycle/read contract.
- `templates/resources/trust_cert.md.tmpl` — Generated documentation or source template for the public lifecycle/read contract.
- `test/scripts/run-story-31-0f-acceptance.sh` — Pinned-container wrapper requiring external attested disposable authority.
- `test/scripts/trust_acceptance_guard_test.py` — Executed offline schema, protocol, Terraform Core or guarded-orchestration regressions.
- `test/scripts/trust_lifecycle_acceptance.py` — Dormant action/settings/UUID-adoption runner with bounded owned-command cleanup and exact restore gates.

### Review result

All 21 first-pass findings were individually triaged before grouping. Twelve
bounded patch groups (high 3, medium 8, low 1) were applied, including the later
high redirected-log escalation correction. Three findings were deferred into
two external prerequisite entries: unavailable live execution/restoration and
the pre-existing released keyless-CA defect. Three first-pass false findings
were rejected: newly declared explicit operation intent after import is an
intentional request; the versioned unreleased heading is required and does not
publish a draft branch; volatile intent is deliberately distinct from observed
metadata. A bounded fourth false claim about configured CA text normalization
was disproven on the pinned base64 encode/decode import path and its speculative
provenance extension was removed. No finding was silently dropped.

Follow-up review recommendation: **true** (first-pass threshold: patched high
3, medium 8, low 1). Specific residual review risk is interaction among
removal-only operation intent, authoritative legacy CA refresh, uncertain
consumed-token state, and cancellation before external restoration. Focused
regressions and bounded rechecks passed, but they are not a fresh independent
four-layer post-patch pass or appliance proof.

### Final verification

Final parent container race run passed: Trust 19.243s and coverageaudit 1.350s;
actual offline Core legacy create/refresh/metadata/import plus explicit operation,
local-key and uncertain consumed-intent controls executed. The only skipped
Trust tests were the two existing explicitly live basic acceptance tests, not
counted as proof. `make check` exited 0 with all six gates passed, none skipped,
in 96,909ms (lint, format, test, security, scan, docs). Generation was rerun
twice after repairs, with the staged second pass producing no diff. Whitespace
checks and the complete publication added-line privacy audit passed (zero
prohibited local/internal identifiers). Frontmatter parsed as one deferred
list with exactly two entries. Ten guard/orchestration test methods passed; Ruff,
Shellcheck with source following and shfmt passed. Full fixture HCL validates
offline against the actual built provider. No production/disposable appliance
was mutated; no live or OpenTofu execution is claimed.

### Later acceptance repair checkpoint (2026-11-04)

This implementation spec remains historical. The parent story and live proof
gate remain open. The [bounded live repair spec](spec-31-0f-live-acceptance-repair.md)
owns the exact `26.7.8_2` runner pin and worktree/TLS harness repairs; it does
not start a new capability workflow or enable keyless CA imports. Original
`26.7.8` source review is not `_2` execution evidence. The actual signed `_2`
package caused a preserved real exact-version rejection of the original runner;
full non-skipped matrix and exact restoration are still pending parent execution.

The later [real execution checkpoint](spec-31-0f-live-acceptance-repair.md#parent-real-execution-checkpoint-2026-11-04) supersedes that pending-live status: the full frozen matrix passed on actual 26.7.8_2 with zero owned leftovers and four exact restored hashes. Historical source/failed attempts remain distinct; broader firmware/OpenTofu and keyless provider gates remain open.

## 2026-11-04 — Maintainer-authorized sequential merge review

Integrated main after retry PR84 without rewriting existing commits. Trust now
targets the next unpublished minor v0.14.0; published v0.13.0 retry and v0.12.0
settings history are preserved. This is a bounded supported-lifecycle increment,
not completion of keyless CA support or the broader compatibility parent.

### Review Triage Log — integration pass

Thirteen findings from four BMAD layers: high1, medium9, low1, false2. All
layers completed; no ambiguous product decision was needed. The standing
maintainer instruction to validate and merge sequentially authorizes applying
review repairs without another confirmation.

- [medium][patch] Verification gap: cert lookup CSR was not asserted; add exact populated and absent-null protocol reads with zero-write counts.
- [false][reject] Edge CA input normalization: carried prior rejection; released controller base64-encodes exact configured PEM and loader decodes it. Authoritative subject refresh is intended legacy behavior. Synthetic normalization is not an appliance contract.
- [high][patch] Blind1 confirmed Create/readback errors could taint and replace an acknowledged object. Use warning-only recovery after confirmed UUID; real Terraform Core tests cover CA, cert and missing local key, unchanged second apply and no-change plan, one add only. Old resource snapshot fails all three controls; fixed Core cases pass.
- [false][reject] Blind2 CA normalization repeats the same disproven claim; preserve authoritative refresh, not hidden configured-value caching.
- [medium][patch] Blind3 JSON null silently decodes as an empty Go string. Reject null known scalar observations before changing state.
- [medium][patch] Blind4 truncated editor with refid alone could erase observations. Require released description/certificate/key fields, distinguishing missing from explicit empty strings.
- [medium][patch] Blind5 delimited substring membership accepted compound options. Enforce exact membership with compound-value negative protocol controls.
- [medium][patch] Blind6 caller or returned IDs could alter URL paths. Accept UUID and legacy identifier characters only; reject unsafe import/lookup before API calls and malformed add identity before accepting it.
- [low][patch] Blind7 CSR labels were unchecked despite valid DER. Require CERTIFICATE REQUEST or NEW CERTIFICATE REQUEST PEM labels.
- [medium][patch] Blind8 live reissue checked only post-action existence. Assert changed certificate, same UUID/key and exactly one independently logged cryptographic write.
- [medium][patch] Blind9 one-time key marker did not prove usability. Parse it with OpenSSL and compare public keys against the returned certificate; assert no persisted appliance key.
- [medium][patch] Blind10 live matrix omitted repeat-token semantics. Exercise same action/new token exactly once, then unchanged token with zero cryptographic writes.
- [medium][patch] Acceptance auditor: string validity timestamps bypassed numeric validation. Validate both JSON numeric and string representations; reject malformed/null timestamps while preserving the full signed-int64 Unix range. A final source recheck corrected an unintended nonnegative bound; parseable pre-1970 legacy certificates are accepted upstream.

The stronger live runner needs new before-operation/verify-operation oracle
checkpoints exposing cryptographic_writes from access-log cursors; snapshot,
identity, exact restoration and all prior gates remain required. Original live
receipts remain historical and are not retroactively attributed to this repaired
source. Fresh identity-bound proof must be recorded separately.

### Repaired merge-candidate proof — 2026-11-04

Candidate `35799c04ea213613d07e9cc987a8a7a9763edf60` passed a fresh complete
guarded matrix on authentic OPNsense26.7.8_2: **26 apply/no-change pairs**,
including all19 settings cases, real certificate-lookup CSR equality, OpenSSL
one-time-key/certificate matching, and reissue preserving UUID/key while changing
the certificate. Independent access-log counts were1 for reissue,1 for the
same action with a new token, then0 for an unchanged token.

Terraform destroyed all9 owned objects; immediate and final counts were0, with
settings unchanged by destruction. Original configuration/settings/routes/runtime
hashes restored exactly. The full wrapper exited0 after its final source/runtime
digest assertions and success marker. The exact owned guest, disk, marker and
listeners were retired and their absence independently read back. This is not
just passing inner steps or fallback recovery.

Executed whole-tree SHA256:
`fef4ec6366728aac69b8546c869ab5832c3cd34790431870af5e8a26298c776d`.
Reproducible executable SHA256 (752 files):
`5c2be5cb1176c49d05559d01c5b0521bccc29e9a409ee9c98f4feb41eccfb9d4`.
The evidence-only publication is not retroactively called the executed whole
tree. [Current sanitized proof](31-0f-live-proof.json) and its pinned historical
reference retain that distinction. All6 local gates,16 offline guard methods,
Trust/coverage race suites and3 real Core create-recovery controls pass. The
old production snapshot fails all3 controls; repaired creation emits warnings
after acknowledged identity instead of tainting and automatically recreating.

Four independent review layers produced13 individually triaged findings; the
follow-up bounded production-diff review found no concrete blocker. The CA PEM
normalization claim was rejected on actual released-source evidence, not masked
by retaining stale observations. Keyless CA remains explicitly unsupported,
older-firmware/OpenTofu execution is unverified and the parent story stays open.


### Signed validity compatibility correction

Released `Store.php` copies parsed X.509 Unix dates directly; CA/certificate
imports do not exclude expired or pre-1970 certificates. The stricter editor
reader therefore must accept negative signed-int64 timestamps in numeric and
string form. Null, malformed and overflow values remain rejected. The real
RSA2048 certificate control uses 1969 validity and verifies signature,
CA/cert resource refresh and authoritative lookups with zero writes. The live
matrix now imports a genuine pre-epoch CA/certificate and checks exact signed
timestamps in both resources and both lookups. Earlier proof remains bound to
`35799c0`; it is not retroactively evidence for this executable correction.


The first signed-date live attempt at `1661ccd` failed the exact-date assertion:
OpenSSL correctly encoded 1965/1966, but the appliance PHP parser exposed
2065/2066 (`2997993600`/`3029529600`). A direct read-only PHP parse reproduced
this upstream year-pivot behavior. The failed attempt is retained separately;
owned cleanup reached zero and all four original hashes restored exactly. It is
not full acceptance or Terraform-destroy proof. A genuine 1969 certificate was
independently parsed by that same appliance as `-31536000`/`-86400` before retry.
The fixture now uses that unambiguous pre-epoch interval; negative-date assertions
and malformed-value rejection remain strict. This is a fixture correction, not
provider-side date rewriting or assertion relaxation.

Exact dependency source confirms the observation:
[PHP8.5.12 UTCTIME conversion](https://github.com/php/php-src/blob/php-8.5.12/ext/openssl/openssl_backend_common.c#L177)
adds100 for two-digit years below68; the
[validity parser](https://github.com/php/php-src/blob/php-8.5.12/ext/openssl/openssl.c#L1083)
uses that helper. The provider preserves authoritative appliance observations.
