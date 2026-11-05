---
title: '31.0f bounded live acceptance firmware repair'
type: 'bugfix'
created: '2026-11-04'
status: 'done'
baseline_revision: 'cec52d80454cbcbee23cc4604d383d63f2f1770e'
review_loop_iteration: 0
followup_review_recommended: true
context: []
warnings: [oversized]
deferred: []
---

<intent-contract>

## Intent

**Problem:** The guarded Trust runner pins unpatched OPNsense 26.7.8, but the actual official signed package and release-revert snapshot now provide 26.7.8_2. The disposable guest is real; its package must not be relabelled or silently accepted as equivalent because two released Trust controller conditions differ.

**Approach:** Pin live acceptance explicitly to the actual tested stable package 26.7.8_2, retaining exact version rejection and documenting the prior source contract separately. Execute the full existing real action/settings/adoption/restoration matrix on that exact firmware and preserve failed attempts. This is acceptance repair in existing PR81, not another capability or PR.

## Boundaries & Constraints

**Always:** Preserve settings recovery, exact digest binding, runtime/owner/endpoint/TLS identity and every cleanup/restoration gate. Distinguish original 26.7.8 source review from actual 26.7.8_2 execution. Public evidence is consumer-neutral and sanitized; raw credentials/configuration remain private. Use pinned dev-toolchain:1.12.24 for checks. No commits/pushes by implementer; parent reviews and publishes.

**Never:** Fake firmware, weaken to prefix matching/arbitrary caller firmware, remove a required test/gate, implement keyless CA support in this repair, claim unexecuted live proof, change provider lifecycle algorithms without a concrete preserved live failure and a spec amendment, target a shared/production guest, or publish private topology/local paths/credentials.

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Exact live package | Authentic bound helper reports 26.7.8_2 | Firmware gate permits existing matrix | All other helper gates still enforced |
| Earlier unpatched package | Authentic helper reports 26.7.8 | Reject before snapshot/provider mutation | Retain sanitized gate failure |
| Different package | Helper reports 26.7.8_1, 26.7.8_3, 26.7.9 or malformed | Reject exactly | Never prefix-normalize |
| Unknown trust identity | Correct version but owner/TLS/guest or isolation mismatch | Reject | No API mutation |

</intent-contract>

## Code Map

- `test/scripts/trust_lifecycle_acceptance.py` -- `helper` currently exact 26.7.8; main computes source hash and runs full Core fixture matrix. Version pin plus bounded fixture/harness repairs only; no provider algorithm changes without concrete failure.
- `test/scripts/trust_acceptance_guard_test.py` -- offline synthetic helper orchestration, unsafe-attestation and cancellation controls. Update valid fixture version and cover exact package reject rows.
- `_bmad-output/implementation-artifacts/31-0f-trust-source-contract.md` -- original reviewed core revision 96978001a8343669154e4492172ce0a07dac7e4b. Append actual package distinction: CaController line124 and CertController line61 change private-key section count from !=1 to >1; seven other audited files byte-identical. Signed release-revert also serves _2. Do not rewrite historical source evidence as _2 proof.
- `_bmad-output/implementation-artifacts/31-0f-trust-lifecycle-and-settings-parity.md` -- owning acceptance contract has exact firmware field and still-open live proof.
- `_bmad-output/implementation-artifacts/spec-31-0f-trust-lifecycle-continuation.md` -- prior implementation spec remains historical, parent open, no new full implementation workflow.

## Tasks & Acceptance

**Execution:**

- `test/scripts/trust_lifecycle_acceptance.py` -- pin exact 26.7.8_2 without loosening helper identity.
- `test/scripts/trust_acceptance_guard_test.py` -- execute each matrix version/control row with real helper orchestration; preserve all existing guard/cancellation tests.
- `test/scripts/run-story-31-0f-acceptance.sh` -- preserve Git source inventory in isolated worktree containers by read-only mounting the actual common Git metadata at its absolute path; forward explicit SSL_CERT_FILE for the private trusted CA without mounting SSH keys. Support ordinary checkout and isolated worktree. Validate TLS CA file lies in already mounted private runtime root, no fallback to insecure TLS. Actual original container source inventory failed because worktree .git pointer points outside /workspace; private launcher corrected mount then preserved real version gate failure.
- Owning source-contract and parent-story artifacts -- record actual package/source discrepancy and actual executed evidence only. Keep keyless import blocked by provider behavior despite official controller repair. Source inspection is not live proof.
- This spec -- record verified checks/review triage; leave actual live proof incomplete until parent executed it.

**Acceptance Criteria:**

- Given original candidate and actual official _2 runtime, when parent attempts frozen runner, then its real exact-version rejection is preserved as a failed attempt, not a stale-network blocker.
- Given corrected exact _2 pin and independently guarded disposable runtime, when parent executes full runner, then only actual non-skipped pass plus zero leftovers and exact restored four hashes can count as matrix proof. Failure remains incomplete with private logs and next action.
- Given final candidate, when full container checks and independent review run, then no suppression or fabricated evidence is introduced and existing PR81 carries resulting fixes/evidence; no duplicate PR.

## Spec Change Log

- 2026-11-04 actual matrix03 passed all crypto/adoption and four boolean
  matrices, then authentic firmware rejected enable_config_constraints=true
  when every constraint was empty. General.php::performValidation lines43-61
  requires one of CipherString/Ciphersuites/groups/MinProtocol/MinProtocol_DTLS/
  SignatureAlgorithms nonempty. Preserve this negative live evidence. Failed
  attempt cleanup zero and all four original baseline hashes restored exactly.
  Amend execution: obtain/validate actual oracle cipher choice before booleans;
  enable constraints with that supported nonempty cipher selection, then disable
  and restore original cipher before the subsequent cipher select/clear/restore
  matrix. Do not skip a boolean or relax any comparison. Add an executed
  regression proving prerequisite selection and invalid-choice rejection;
  no provider algorithm change is justified by this correct API validation.
  Preserve a valid initial constraints=true/nonempty-cipher baseline too: clear
  requires paired disable; final restore reinstates original boolean+cipher.
  Current actual lab baseline is false/empty; synthetic regression must cover
  both source-valid initial states without skipping empty-clear.

- 2026-11-04 actual matrix attempt: all initial actions, CSR transition,
  metadata and adoption instrumentation passed, then settings comparison
  failed because the independent lookup was read BEFORE the singleton update.
  Saved real state has managed store_intermediate_certs=true and lookup=false;
  subsequent no-change plan re-read the lookup but does not persist refreshed
  state for terraform show. This is a fixture dependency defect, not proof of
  provider failure. Amend execution task: in resources(), give the settings
  lookup an explicit dependency on its managed singleton, and test actual
  Terraform Core ordering/validated fixture so post-apply observations agree.
  Keep all strict comparisons and no-change gates; never ignore mismatch.
  Preserve this failed attempt and its restoration outcome separately.

- 2026-11-04: Applied bounded runner pin, executable offline version/identity
  controls, absolute read-only common-Git mount and private-root-confined CA
  forwarding. No provider lifecycle changes, commits or pushes.

## Review Triage Log

### 2026-11-04 — Four-layer review pass
- verdicts: 15 findings — high 1, medium 6, low 2, false 6, maybe-false 0
- findings:
  - `[medium]` `[patch]` Writable ordinary-checkout metadata alias — confirmed source mount exposed /workspace/.git writable despite absolute metadata mount; source now read-only, with retained ordinary/worktree mount assertions.
  - `[high]` `[patch]` Canonical runtime may overlap a repository ancestor — resolving a runtime-file symlink to an ancestor could mount unrelated private files; canonical overlap is now rejected both directions before Docker, with equal/ancestor/descendant/symlink controls.
  - `[medium]` `[patch]` No committed wrapper boundary regressions — executable subprocess tests now create real Git checkout/worktree metadata and assert read-only mounts, CA forwarding and pre-Docker rejection.
  - `[false]` `[reject]` Runner must independently validate cipher catalog membership — actual catalog membership is the private oracle's explicit responsibility. The sealed live helper reads actual API choice keys and selects from that set, independently pinned in runtime. The runner checks malformed reply shape, not a second invented API/catalog protocol. The actual selected cipher was observed and executed; no unsupported well-formed choice can arise from the verified helper.
  - `[low]` `[patch]` Independent booleans only restored false — prerequisite regression now crosses true/false independent initial states with both constraint baselines and checks final omitted intent.
  - `[medium]` `[patch]` Core fixture omitted intent untested — actual fixture HCL Core regression now adds omitted apply and empty plan, retains true owner/lookup and requires exactly four saves/activations.
  - `[medium]` `[patch]` No reproducible published executable binding — final sanitized proof publishes an executable-only digest plus exact inventory/hash procedure; parent compares preserved tested executable bytes with final publication, while whole-tree digests remain labelled pre-evidence candidates.
  - `[low]` `[reject]` Historical implementation checkpoints sound pending — changing this build's spec is rejected under workflow triage; normal finalization adds an explicit current result while retaining dated earlier failures/checkpoints. Owning source/story sections explicitly supersede the old discovery blocker.
  - `[medium]` `[patch]` Gap: TLS CA confinement/forwarding lacked retained wrapper execution — inside/outside/symlink cases now invoke actual shell wrapper with a controlled Docker observer; accepted environment and rejection-before-Docker are asserted.
  - `[medium]` `[patch]` Gap: worktree Git metadata mount lacked retained verification — real temporary checkout/worktree subprocess cases assert exact absolute read-only metadata and source mounts. The independently executed actual-container inventory/TLS smoke remains separately recorded.
  - `[false]` `[reject]` Intent: synthetic helper tests do not authenticate firmware — they are deliberately offline negatives, not the authenticity proof. Parent independently executed sealed live qm/guest/TLS guards on the newly reserved disposable and retained authentic package/version receipts.
  - `[false]` `[reject]` Intent: permitted synthetic path is not full live matrix — separate real non-skipped runner passed every action/settings/adoption/deletion/restoration gate; no synthetic path is counted as live.
  - `[false]` `[reject]` Intent: local settings rule model is not released controller — model is only regression control; actual released General.php validation rejection was preserved, followed by real successful supported-prerequisite settings execution.
  - `[false]` `[reject]` Intent: Core ordering test is not appliance execution — parent distinctly records Core loopback proof and the separate actual appliance matrix; both ran.
  - `[false]` `[reject]` Intent: published narrative is not private raw proof — publication intentionally excludes credentials, topology and private state. Parent read actual receipts: nine objects before real destroy, zero before rollback, seven adoption requests/four object GETs/zero writes, and identical four baseline/restored hashes; sanitized reproducible executable binding is published separately.

All four layers completed. Native fourth simultaneous launch hit the active
thread limit; intent review launched immediately after a slot became available.
No layer was skipped. Six root-cause patch groups: high 1, medium 4, low 1.
Post-patch read-only wrapper changes require a new full real frozen run before
final publication; earlier success remains preserved rather than retroactively
assigned to the repaired executable.


- Source-backed constraint prerequisite: validate the authoritative oracle
  choice before all settings intents. Exercise constraints true/false with
  that nonempty cipher and restore the original boolean/cipher together.
  The separate cipher empty-clear step explicitly disables constraints in
  the same update, then restores both original fields, including a valid
  initially-enabled/nonempty-cipher baseline. Strict comparisons remain.
- Executed runner regression models the released nonempty-constraint rule,
  checks all five true/false/restored booleans plus cipher select/empty/restore,
  and requires invalid oracle choices to fail before any settings apply or
  config edit. The staged pre-fix fixture fails this regression at the exact
  missing-prerequisite condition; this is not provider or live-success proof.

- Proven live fixture dependency defect: patched the generated settings lookup
  with an explicit managed-singleton dependency. Strict persisted owner/lookup
  comparisons and all no-change gates remain unchanged; no provider changes.
- Added real Terraform Core regression using HCL settings blocks parsed from the
  actual Python fixture, with import-first singleton and loopback HTTP API.
  Four alternating updates must persist matching lookup state, each followed by
  an empty plan. A read-only fixture overlay removing only the dependency
  reproduces the stale persisted lookup failure; corrected fixture passes.

- Implementation self-check: exact equality still rejects base/other/malformed
  package strings; executable controls stop after guard without snapshot.
- Wrapper self-check: runtime/CA canonicalization rejects an outside-root CA
  and a symlink escape before Docker. Common metadata is mounted read-only;
  source inventory is still the actual Git tracked/unignored inventory.
- Independent final review and actual live matrix remain parent-owned and
  incomplete here. No synthetic helper result is appliance evidence.

## Verification

**Commands:**

- Pinned container `python3 test/scripts/trust_acceptance_guard_test.py` -- all guard/orchestration/cancellation and exact-version tests pass.
- Shell wrapper `--help`, shellcheck/shfmt and ordinary/worktree Git inventory/container TLS smoke -- succeed using actual source; no SSH keys mounted.
- Pinned container Ruff and shell/static runner checks -- pass without suppression.
- `make check` with dev-toolchain:1.12.24 -- all six gates pass, none skipped.
- `git diff --check` and publication added-line privacy audit -- clean.
- Parent private `launch_acceptance.py` -- actual full runner, independent oracle cleanup/zero/restore receipts; never count offline synthetic helpers as live.

### Initial firmware/harness implementation checks (2026-11-04)

- Pinned container guard suite: **12 methods pass**, including seven executable
  version rows, eight executable identity/control rows and all prior cleanup,
  hash-mismatch, malformed helper and owned-process cancellation controls.
- Pinned Ruff lint/format, Shellcheck `-x`, shfmt and wrapper `--help`: pass.
- Actual pinned-container wrapper smoke: ordinary checkout, temporary isolated
  worktree and actual candidate all produce the identical source digest to host
  Git inventory and complete a hostname-verified local TLS handshake using an ephemeral
  private CA forwarded through `SSL_CERT_FILE`.
  Docker arguments include the read-only absolute common metadata mount and
  exclude SSH keys. Both outside-root CA and symlink escape reject before Docker.
- `DEVRAIL_IMAGE=ghcr.io/devrail-dev/dev-toolchain:1.12.24 make check`:
  **pass**, 88,258ms; lint, format, test, security, scan and docs all pass,
  no skipped gates or suppression.
- `git diff --check` and publication added-line privacy audit (including this
  previously untracked repair spec): pass; no private identifiers or topology.
- Independent review and real non-skipped live matrix/cleanup/zero/four-hash
  restoration remain pending parent execution. No appliance commands were run
  by this implementation agent and no live success is claimed.

### Proven settings-fixture follow-up checks (2026-11-04)

- Pinned `go test -count=1 -race -v ./internal/service/trust -run
  TestTrustSettingsAcceptanceFixtureCoreLookupAfterUpdate`: pass (Core import,
  four actual loopback settings writes/reconfigures, matching persisted lookup,
  and four empty plans). The generated full fixture is HCL-parsed, with only
  unrelated PKI blocks excluded from this focused Core execution.
- Negative control: mounting a read-only copy of the real fixture without its
  settings dependency fails the same Core test on step 1 with opposite owner
  and lookup boolean values. No repository fixture or guest was mutated by this
  control. This proves the regression detects the original ordering defect.
- Pinned guard suite: 12 methods pass; Ruff lint/format and focused
  `golangci-lint run ./internal/service/trust/...`: pass, zero issues.
  `git diff --check` and new-added-line privacy audit: pass.
  Parent owns the new full gate, independent review and live matrix retry.
  The earlier full six-gate pass above predates this dependency follow-up.

### Source-backed constraints prerequisite follow-up (2026-11-04)

- Pinned guard suite: **13 methods pass**, including 12 prerequisite/choice
  subcases (two valid initial constraints baselines across six choice forms).
  Both successful cases execute all 19 settings applies; the valid initially
  enabled baseline restores its original cipher and enforcement exactly.
  Invalid choices produce zero applies and unchanged fixture text.
- Read-only overlay of the staged prior actual fixture fails the new regression
  with `constraints enabled without prerequisite`, confirming detection of
  the preserved firmware validation failure rather than a string assertion.
- Pinned Ruff lint/format, added-line privacy audit and `git diff --check`: pass.
  Existing actual-fixture Core dependency regression rerun with race checking:
  pass (3.329s), four loopback updates and unchanged plans.
- Parent owns final full gate, independent review and live matrix retry.
  No guest or acceptance commands, provider lifecycle changes, commits or
  pushes were performed for this follow-up. The preserved failed matrix and
  its exact restoration remain distinct from pending successful live proof.


## Parent real execution checkpoint (2026-11-04)

This is the first passing pre-review candidate, retained separately from the
required post-review rerun. The non-skipped pinned-container runner passed against authentic OPNsense
26.7.8_2 with candidate source SHA256
`91aebd345968990c4b8b84ae3a80eede0b8a3f00d1ff003d07a42dfd67eee26d`.
This is the executed pre-evidence working-tree digest, not a claim that the
earlier baseline commit includes these repairs. Final documentation records
will change that whole-tree digest; executable code/harness equality is
independently checked against the preserved tested manifest.

Actual proof: root/intermediate/OCSP/import CA, internal/local/external/import
certificates, CSR sign/import retaining UUID/key, reissue, metadata stability,
one-time key retention, CA/cert UUID adoption and all six settings passed.
Adoption's bounded request window contained seven requests/four observed object
GETs, zero cryptographic writes, exact configuration equality.
All five boolean true/false/restore cases and cipher select/empty-clear/restore,
omission/no-change, actual Terraform destroy and state-only singleton deletion
passed. There were 24 successful apply/no-change pairs; the extra unchanged and
UUID-adoption plan checks passed too. Nine owned objects existed before destroy;
zero remained before rollback, and singleton settings were unchanged by destroy.
Independent config/settings/routes/runtime SHA256 values matched their baseline
exactly after restoration. No production appliance was targeted.

Preserved failures preceded success: authentic package rejected original base
version pin; initial oracle GET header caused HTTP400 before mutations; stale
fixture lookup exposed pre-update read ordering; exact restore caught FreeBSD
symlink modes and was repaired without ignored differences; upstream correctly
rejected constraints=true without a nonempty constraint, then a source-valid
fixture supplied the actual available cipher. No failed run is counted as
successful proof. Private raw baselines, requests and state stay unpublished.

Parent checks after prerequisite repair: pinned make check passed six of six,
none skipped (31,043ms); thirteen guard methods passed. Independent four-layer
review and final hosted CI remain outstanding at this checkpoint. The parent
Trust story remains open for explicitly deferred provider keyless-CA support
and unproven firmware/OpenTofu boundaries, not for this now-executed live matrix.


## Auto Run Result

**Current result:** bounded acceptance repair and post-review real matrix passed;
the parent Trust story and draft PR remain open, not merged or released.
The [sanitized final proof](31-0f-live-proof.json) supersedes earlier pending
execution checkpoints. Final executable SHA256 is
`66bbfa18dccf14768d733a120e4a7a211efcb6943e8102adf1b6f9e104d4122b`;
its reproducible procedure is in the owning story. Actual tested whole-tree
SHA256 is `eb971828b62183878c72f247aacb1138b2715c6a800605404ee13b3c72128e46`.
The final evidence documentation is not retroactively called that tested tree.

Changed files: runner pins/authentic firmware and valid settings sequence;
wrapper confines roots/CA and mounts source/common metadata read-only; guard
tests retain executable boundaries and initial-value restoration; actual Core
fixture test covers persisted lookup ordering and omitted intent; owning source,
story, support/audit/generated resource docs and sanitized proof carry verified
current status. Provider lifecycle algorithms/settings recovery are unchanged.

All fifteen findings are individually logged above. Six root-cause patch groups
were applied (high 1, medium 4, low 1); six false findings and the spec-only low
finding were rejected with reasons. No review item was silently dropped.
Follow-up review remains recommended for cross-platform canonical path/mount
alias behavior outside the actually tested environment; it is not a claim of
unexecuted proof for the tested environment.

Post-patch parent verification: fourteen executable guard methods passed,
pinned make check passed all six gates with none skipped (60,747ms), and the
actual post-review full runner passed with 24 apply/no-change pairs, authentic
request instrumentation, real Terraform deletion, zero before rollback and
four exact restored hashes. Owned guest/disk/marker/listeners were retired.
Intermediate owned oracle/SSH process loss was preserved as guard/pre-mutation
failure, not a provider failure or a passing matrix; parent re-established
identical sealed helpers/forwards and reran the entire unchanged candidate.
Final generation/whitespace/privacy/full gate and hosted exact-head CI are
recorded separately in publication receipts after this result is written.

Residual boundaries: keyless provider support, older firmware and OpenTofu are
not established by this matrix. No production write, auto-merge or release is
claimed.
