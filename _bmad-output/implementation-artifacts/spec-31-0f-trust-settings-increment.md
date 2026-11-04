---
title: 'Story 31.0f: Trust settings increment'
type: 'feature'
created: '2026-11-03'
status: 'done'
baseline_revision: 'c69824405e777bafa8c434378e77f72f7b14e4be'
review_loop_iteration: 0
followup_review_recommended: true
context:
  - 'AGENTS.md'
  - '_bmad-output/implementation-artifacts/epic-31-context.md'
warnings: [oversized]
deferred:
  - summary: >-
      Shared HTTP clients have no explicit request deadline when callers supply none.
    evidence: |-
      Existing pkg/opnsense/client.go leaves http.Client.Timeout zero; the settings transport inherits it. Caller cancellation is honored and tested. A provider-wide timeout policy needs a separate transport contract, not a Trust-specific arbitrary deadline.
    location: >-
      pkg/opnsense/client.go
    severity: medium
---

<intent-contract>

## Intent

**Problem:** The existing Trust owners have no supported singleton for the six audited OpenSSL/CRL settings. The broader31.0f lifecycle story is too broad to equate a settings implementation with full Trust parity.

**Approach:** Add `opnsense_trust_settings` resource and lookup as the first reviewable31.0f increment, preserving imported and unrelated policy. Keep CA/certificate actions and full-story completion explicitly outstanding.

## Boundaries & Constraints

**Always:** Use released26.7.8 `Trust/General.xml` and `SettingsController.php`: GET/POST `/api/trust/settings/get|set|reconfigure`, `trust` envelope, five bool wire strings0/1 (`store_intermediate_certs`, `install_crls`, `fetch_crls`, `enable_legacy_sect`, `enable_config_constraints`), and Terraform `cipher_string` set(string) -> API `CipherString` selected choices. Fixed ID `trust_settings`; import-first Create, Update set/read/reconfigure with affirmative status=ok, state-only Delete. Read authoritative values; omitted config preserves prior/current policy; explicit false and empty set are real clears. Global mutation lock spans baseline, save and recovery/activation. Durable private pending state prevents repeated set after saved/uncertain outcome; retain confirmed identity/state on failures. Validate malformed responses, unselected choices and unknown elements without accidental clear. Partial updates must not write five unaudited TLS fields (`Ciphersuites`, `SignatureAlgorithms`, `groups`, `MinProtocol`, `MinProtocol_DTLS`). Mark output/code/testing evidence accurately, never appliance- or release-proven without execution.

**Never:** Reset settings on Create/Delete, introduce static optional defaults, copy operator topology or credentials, manage extra TLS fields, mutate a shared/production appliance, change CA/cert behavior, claim parent31.0f done, or promote full Trust audit parity from protocol-only evidence.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|---------------|---------------------------|----------------|
| Adoption | Exact fixed import ID then GET | Six authoritative values; subsequent omitted no-change config preserves them | Reject wrong ID and malformed/missing trust object; never erase state |
| Save | One bool false or explicit empty cipher set, other fields omitted | Only configured fields written, unrelated six/extra5 remain unchanged; one reconfigure | Unknown/configured unresolved set members reject before any write |
| Decode | Strict0/1 booleans; selected CipherString maps, CSV/string choices if source supported | Exact set of selected choice keys, including empty selection | Reject malformed bools/selectors/duplicates or missing material fields instead of treating as clear |
| Failure/retry | Save confirmed/uncertain, GET or activation fails | Persist identity/request intent and pending flag; subsequent refresh observes then activates without repeating set | Body-validation failures remain failures; uncertain save with conflicting read requires fresh plan, no blind replay |
| Removal | Resource Delete with unavailable client/pending marker | State removed with zero HTTP calls | Normal planning refresh documented separately from Delete |

</intent-contract>

## Code Map

- `internal/service/trust/exports.go` -- currently2resources/1lookup, register one settings owner and lookup.
- `internal/service/kea/ddns_settings_{resource,lifecycle,schema,model,data_source}.go` -- proven partial-config, durable recovery and strict decode patterns; reuse design without DDNS capability checks/enable semantics.
- `internal/service/wireguard/general_{resource,model,schema,data_source}.go` and `internal/service/unbound/general_resource.go` -- import-first singleton examples (inspect actual filenames before reuse).
- `pkg/opnsense/{crud,reconfigure,reqopts,types}.go` -- supported singleton helpers/strict reconfigure; do not weaken shared contracts.
- Source: `https://github.com/opnsense/core/blob/26.7.8/src/opnsense/mvc/app/models/OPNsense/Trust/General.xml` and matching `controllers/OPNsense/Trust/Api/SettingsController.php`; read-only source oracle. The full model has11fields; this increment owns audited6 only.
- `internal/coverageaudit/audit_test.go` and planning `provider-ecosystem-coverage-audit.json` -- increment exact local inventory113resources/99datasources; retain missing-buildable/needs-field-audit status and followup for unexecuted lifecycle, do not mark whole Trust covered.
- `templates/{resources,data-sources}/trust_settings.md.tmpl`, matching `examples/` and generated `docs/` -- HCL/import/omission/clear/state-only destroy and test-scope notes.
- `_bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md` and `sprint-status.yaml` -- record settings increment in progress, four remaining Trust action/lookup steps and known source risks; reconcile only definitively released31.0n/o tracking.

## Tasks & Acceptance

**Execution:**
- `internal/service/trust/settings_*.go` -- implement typed singleton, lookup, schema, strict model and recovery; add HTTP-backed framework/protocol tests, not only helper mirrors.
- `templates/`, `examples/`, `docs/`, `CHANGELOG.md` -- document partial import-first ownership; next version0.12.0(Unreleased), since0.11.0 is already tagged; generate docs.
- `_bmad-output/planning-artifacts/provider-ecosystem-coverage-audit.json` and generated report -- update exact registration inventory and bounded implementation evidence without asserting live proof.
- `31-0f-trust-lifecycle-and-settings-parity.md`, `sprint-status.yaml` -- record resumable next pieces and stale shipped labels; do not modify unrelated historical done statuses without evidence.

**Acceptance Criteria:**
- Given imported settings, when Terraform protocol plan/update/read runs, then omission is no-change, explicitfalse/clear persists, fixed identity is stable and unowned fields are absent from requests.
- Given partialsave/activation failures, when framework private state is retained and refreshed, then recovery retains identity and never repeats the set; validation and malformed responses fail closed.
- Given published v0.11.0, when docs and audit are reviewed, then new capability is unreleased, local counts are exact, existing16lookup gaps remain16, and CA/cert plus live acceptance remain outstanding.

## Spec Change Log

## Review Triage Log

### 2026-11-03 — Review pass 1
- Execution: four independent layers completed; the four-slot platform limit required intent review to start after the verification reviewer finished. No layer was skipped.
- verdicts: 23 findings — high 1, medium 13, low 1, false 8, maybe-false 0
- findings:
  - `[medium]` `[patch]` Blind1 stale omitted plan — verified known planned fields could differ at apply. Fresh baseline comparison now rejects before saving; real protocol plan/apply race tests cover bool and set.
  - `[high]` `[patch]` Blind2 activation redirects — shared reconfigure followed redirects through authenticated transport. Activation now uses redirect-disabled single-attempt transport; target receives zero requests.
  - `[medium]` `[patch]` Blind3 activation retries — shared retry policy could replay POST. Explicit one-attempt activation covers500 and dropped response with retained intent.
  - `[medium]` `[patch]` Blind4 private-state preflight — deleting a key did not exercise nonempty storage. Valid pending intent is stored before save; injected storage failure stops before POST; deterministic rejection clears it.
  - `[false]` `[reject]` Blind5 absent version gate — docs identify the source contract, not an enforced minimum. Complete six-field strict baseline reads gate writes; no unsupported version with a harmful compatible response was identified. Older-core proof remains explicitly unexecuted.
  - `[medium]` `[patch]` Blind6 ambiguous activation outcome — strict object decoding now rejects duplicate, aliased and contradictory outcomes without dropping pending state.
  - `[low]` `[reject]` Blind7 invalid cipher detected at apply — static invalid values already fail before any write. Additional schema-validator branches are not needed for the bounded runtime contract; upstream choice membership remains authoritative.
  - `[medium]` `[patch]` Blind8 generic field/lookup descriptions — source-backed policy descriptions and read-only lookup wording regenerated.
  - `[medium]` `[patch]` Blind9 cancellation/nil-client Delete tests — cancel-after-save retains state/private intent; direct canceled-context nil-client Delete and pending zero-API Delete now run.
  - `[medium]` `[patch]` Blind10 private integrity/concurrency tests — corrupt canonical/identity markers remain unchanged and make zero HTTP; channel-backed resource interleavings cover all four locked stages.
  - `[medium]` `[patch]` Edge1 duplicate save result — both result orderings and aliases now classify uncertain, retaining recovery instead of deterministic validation.
  - `[medium]` `[patch]` Edge2 contradictory activation — same verified outcome-decoder cause as Blind6, same repair/tests.
  - `[medium]` `[defer]` Edge3 missing timeout — verified inherited shared-client behavior, not introduced by this owner. Caller cancellation is tested; provider-wide deadline policy recorded in deferred work.
  - `[medium]` `[patch]` Gap1 lookup-value assertion — six fields plus fixed identity and subsequent remote change now asserted through datasource protocol.
  - `[medium]` `[patch]` Gap2 pending Update — fresh-provider Update receives saved and uncertain private bytes, never repeats set and requests a fresh plan.
  - `[false]` `[reject]` Intent1 partial umbrella scope — user authorized working through stories; this is an explicitly bounded first increment and31.0f remains in progress, not asserted complete.
  - `[false]` `[reject]` Intent2 invocation not established by diff — parent executed installed bmad-build-auto renderer, compiled context and followed isolated implementation/review; diff is not expected to prove tool execution.
  - `[false]` `[reject]` Intent3 no CLI/live proof — expressly documented as unexecuted; parent story and audit follow-up retain that gate. No appliance claim is made by this increment.
  - `[false]` `[reject]` Intent4 manually constructed update plan — added actual PlanResourceChange→Apply regression with omission races and strict-response cases; former helper calls remain narrower fixtures.
  - `[false]` `[reject]` Intent5 direct Delete versus destroy — docs explicitly distinguish state-only Delete from normal Core planning refresh and pending activation.
  - `[false]` `[reject]` Intent6 CA/cert and extra TLS omitted — this coherent settings increment leaves those action/read scopes and five unaudited fields explicitly outstanding.
  - `[false]` `[reject]` Intent7 context/tracking changes — user asked remaining-story audit and BMAD progression; dated released-state reconciliation is authorized and does not change runtime claims.
  - `[medium]` `[patch]` Intent8 lookup resource-lifecycle wording — lookup audit/schema/templates now describe read-only lifecycle separately.

## Design Notes

Settings is the first coherent increment, not closure of umbrella31.0f. Subsequent work: certificate/CA read parity; CA import/internal/intermediate/OCSP with legacy compatibility; certificate CSR/generation/sign/import/reissue including one-time keys; disposable action-matrix proof. Source discovered volatile editor action/lifetime fields and possible CA cert-only import defect; do not implement those by adding naive writable fields.

## Verification

- Container `go test -count=1 -race -v ./internal/service/trust/... ./internal/coverageaudit/...` -- all matrix tests execute/pass (name mapping in result).
- Container `go generate ./...` twice -- deterministic generated docs/report, no unstaged differences after expected outputs staged.
- `git diff --check` -- clean.
- `make check` -- clean command exit and all six gates passed; tools remain container-only.
- Review source XML/controller against exact wire requests; document live acceptance not executed if no attested disposable target exists. This does not close umbrella story.

## Auto Run Result

- Implemented the first Trust settings resource/lookup increment across typed schema/model, transport, lifecycle recovery, registration, tests, templates/generated docs/examples and exact inventory. CA/certificate parity and live acceptance remain open in31.0f.
- Review repaired10 root-cause groups (1high,9medium); one pre-existing shared timeout policy deferred, one low enhancement rejected and8false findings refuted individually above.
- Follow-up review recommended: true. Exact risk: interactions among stale-plan rejection, pre-save private journaling, strict outcomes and cancellation have focused regression proof but no second independent review after these patches or live-appliance proof.
- Parent reran the full Trust and coverage-audit race suites after repair: all passed, including every matrix row. Initial six-gate check passed; final repaired-tree gate and deterministic generation are required before publication.
- Residual gates: TerraformCLI/Core against disposable26.7.8, exact settings restore/cleanup, CA/certificate reads/actions and one-time keys; OpenTofu/older-core execution and publication are not asserted.

### Matrix test audit

- Adoption: `TestSettingsProtocolImportPlanPartialClearLookupAndDestroy` and `TestSettingsProtocolStaleOmittedPlanBeforePartialSave`.
- Save/decode: `TestSettingsUnknownMembersAndBooleanIntent`, `TestSettingsStrictMaterialFieldsAndSelections`, partial-clear protocol and lookup value/drift assertions.
- Failure/retry: `TestSettingsProtocolDurableRecoveryNoRepeatedSet`, `TestSettingsProtocolStrictSaveOutcomeAndPendingUpdate`, `TestSettingsProtocolActivationSingleAttemptStrictAndPending`, `TestSettingsProtocolCanceledAfterSaveRetainsPrivate`, storage/corrupt-marker/concurrency regressions.
- Removal: import/clear/destroy protocol, pending destroy, `TestSettingsDeleteNilClientCanceledContext` and corrupt-marker Delete.
- Parent executed all named tests via container `go test -count=1 -race -v ./internal/service/trust/... ./internal/coverageaudit/...`; none of these matrix tests was skipped.

### Final verification

Final repaired-tree `make check` exited0: all6gates passed, none skipped (99020ms). Two deterministic generation runs and all matrix/race tests passed. Only this settings implementation increment is complete; parent31.0f and live-proof gates remain open. Publication and merge remain separate.
