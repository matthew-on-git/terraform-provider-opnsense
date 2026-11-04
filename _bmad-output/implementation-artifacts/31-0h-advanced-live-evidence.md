# Advanced FRR acceptance checkpoint

Status: **in progress - the complete17-step Go matrix passed, but the guarded
public runner was interrupted during its final digest gate**. This is a
resumable validation checkpoint, not end-to-end guarded-runner success, a release
or completed parent Story31.0h. Historical failures below remain failures; the
latest attempt and exact orchestration boundary are recorded at the end. Candidate implementation is in
[PR83](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/83).

## Historical executed candidate and runtime - first six attempts

- Latest tested head: `56253b1643deefb6cdaadcca0badc89f24550010`.
- Latest frozen source digest: `2e26cb01403f0f3927d77313412ac9d39331493294b4e798e1c2fb642feeff44`.
- Actual disposable versions: OPNsense26.7.8_2, os-frr2.0_4, bgpd10.7.1.
- Installed BGP model/controller and both routing/authentication templates exactly
  match released pin `5ac2959c8514b4f60e94d65439be5f629f10521f`; source hashes are
  in `31-0h-advanced-source-evidence.json`.
- New isolated target identity, endpoint/TLS, bridge isolation and versions were
  independently attested; private inventory and raw configuration are withheld.

## Historical observations - first three failed attempts

The third real non-skipped attempt reached Step12/17 after initial creation,
no-change planning and nine resource imports. Actual legacy-provider schema-v0
import/upgrade/no-change/deletion ran on this modern appliance; it is not older
firmware acceptance.

- All19 public advanced fields, password-free lookup, zero weight, generated
  transport/IPv6 flags, all four capabilities and two correctly owned listen
  ranges passed their live assertions.
- The active independent collector received the approved documentation prefix;
  no forbidden default was observed. Per-family policy and effective PF checks
  passed. Filter activation was an explicit lab action, not provider behavior.
- Updating route-map sequences produced `permit30` and `deny5` declarations.
  The retained independent observation showed an active collector with the
  approved prefix withdrawn from current routes and retained only in history.
- **The attempt still failed:** its oracle incorrectly required numeric line
  positions in generated text. The released template sorts string IDs, rendering
  `30` before `5`; FRR's numeric evaluation is a different surface. The corrected
  test requires exact numeric declarations plus mandatory active-session current
  prefix withdrawal before its effective-order marker. At that checkpoint the corrected complete
  matrix had not yet run; matrix08 below later passed its Go matrix.
- Earlier two attempts failed synthetic flatmap-count assertions after adding
  resource-only authentication. Both comparator sites are now covered by focused
  missing/extra/value/set/password negative controls. Failed attempts remain
  failures; none is counted as a full matrix PASS.

## Historical restoration and cleanup - first three failed attempts

On the first three failed attempts, actual Terraform deletion reported zero objects in all
seven FRR namespaces before rollback. Exact API settings, generated FRR,
effective/rendered PF and routing restoration passed. For those attempts, independent full-config
fallback restoration also matched all six captured predicates, including the
configuration hash. The owned disposable guest/disk and collector, tunnel, oracle listeners were then retired; independent checks found no owned listener or acceptance container remaining. These cleanup results do not turn a failed matrix into PASS.

## Historical resume instructions - first three failed attempts

Rerun the entire frozen17-step matrix on a newly attested disposable target after
the test-only ordering correction. Retain source/head binding, current-prefix
positive/withdrawal/restoration observations, UUID/group checks, all advanced
flags, actual Terraform deletion, exact restoration and zero leftovers. Do not
replace these gates with historical counters or a mocked response.

At that historical checkpoint, later drift/repair and firewall-toggle/restored-
positive steps remained unverified; matrix08 below subsequently passed those Go
assertions. Its final public-runner digest gate was interrupted. Legacy appliance/OpenTofu execution and the numeric
clear-interface design remain parent-story work. No production writes, merge,
release or blanket advanced-neighbor session proof is claimed.

## Fourth failed attempt and bounded managed-reference repair

Attempt `matrix04`, at head `b0cada56890dd8afeb1d61c2372f03d12e738d1e`
and frozen source digest
`d027c2626e5fbd681be39152e2a68eff8959ae6a08d01c93206ae6d0a2299df7`,
failed Step12/17: the managed peer-group was incorrectly checked using the strict
fresh import/data-source canonical-selection oracle. Managed state can retain a
configured sibling UUID when authoritative membership proves the same logical
route-map group. The exact compared UUIDs were not retained in the reviewed
bounded evidence, so this does not claim reversal changed the canonical sibling.

The retained independent observation contained `permit30`, `deny5`, an active
collector and no approved prefix in current routes. Failure preceded the mandatory
withdrawal loop: that observation is not passing Step12 evidence. Parent reported
seven-namespace Terraform deletion, unchanged frozen source and six-predicate
fallback restoration. This remains a failed attempt, not full acceptance.

The test-only repair now reads the same identified managed owner freshly, validates
its configured flatmap cardinality and nonempty selection, decodes the raw
selection strictly and independently resolves every selected UUID on both sides
to the exact fixture route-map group. Missing owners, owner-ID mismatch, missing
or empty selections, malformed counts, stale/unreadable members and wrong groups
fail closed. Correct same-group single/multiple sibling representations pass.
Strict canonical equality remains mandatory for genuinely fresh imports/lookups;
no provider normalization, schema, configuration or lifecycle gate changed.

Focused controls and the full affected quagga package suite passed with the race
detector in pinned `ghcr.io/devrail-dev/dev-toolchain:1.12.24`, using actual
`go1.26.10 linux/arm64`. Controls include managed same-group/different-canonical
and exact-canonical positives, both-sided group/stale negatives, cardinality and
owner identity failures, strict fresh canonical rejection/acceptance, HTTP fixture
observation wiring, public-field parity and released lexical-rendering regressions.
Fixtures and offline passing tests are not live appliance acceptance.

A new frozen full17-step non-skipped run on a newly attested disposable target
remains mandatory after bounded review and repository checks. Do not resume at
Step12 or count prior partial observations as completed steps. All numeric-policy,
current-prefix withdrawal/restored-positive, identity, authentication, actual
Terraform deletion, exact restoration and zero-leftover requirements remain.
Parent Story31.0h and the separately unverified platform/interface rows stay open.


## Fifth failed attempt — optional None placeholder

Attempt `matrix05` tested head `e1de88f0d4c530cbe39a2f4937d000985cef5093`
with frozen source digest
`acda084d6f1ad7496ece492803e561be0b26b1fa586839b5902b82c13103de76`.
It failed Step1/17 after the advanced-field checks because the shared test decoder
incorrectly rejected the released API's exact empty-key optional None choice.
That same defect interrupted captured BGP settings restoration at `networks`;
the test-level exact-restoration gate consequently failed. Actual Terraform
deletion still reported zero in all7namespaces. The independent full-config
fallback then passed all6exact predicates with frozen source unchanged. This is
another retained failed attempt, not appliance acceptance.

Released core BaseListField optional selectors and the existing production
decoder establish that exact empty key as a non-member placeholder. The narrow
test-only correction skips only that key after validating its selected flag;
whitespace/padded member IDs remain invalid. Focused controls retain required
managed nonempty selection, strict fresh canonical equality, cardinality and
all authoritative group/current-prefix/lifecycle checks. Independent bounded
source recheck passed. Full corrected17-step execution remains pending.


## Sixth failed attempt — actual firmware metadata stall

Attempt `matrix06` tested head
`56253b1643deefb6cdaadcca0badc89f24550010` and frozen source digest
`2e26cb01403f0f3927d77313412ac9d39331493294b4e798e1c2fb642feeff44`.
It executed initial creation/no-change and nine imports. The corrected Step12
check passed exact numeric sequence declarations, stable route-map identities,
independent managed/raw authoritative group reads and active-session withdrawal
from current routes. Its effective-numeric-order marker was actually emitted.
This does not establish successful completion of Step12's later refresh or the
remaining matrix.

Terraform's post-apply refresh then failed the bounded read-only firmware-info
discovery. Independent diagnostics found a high-CPU `pkg update -q` metadata
refresh holding the repository lock, queued metadata queries and firmware-tier
timeouts. The actual BGP API returned HTTP200 in0.027s while firmware info timed
out after8.007s. Installed source confirms info discovery synchronously invokes
metadata refresh, not package installation. Memory/disk were ample and no
upgrade/install process was running. The cause beyond the observed metadata
stall and its nonrecurrence remain unproven; provider timeouts were not weakened.

The failed run's Terraform cleanup stalled in the same discovery path. After
preserving that outcome, the exact isolated source/identity-bound acceptance
container was stopped. **Actual Terraform deletion was not verified for this
attempt.** The independent full-config fallback was attempted but returned a
guarded-oracle HTTP error: **exact restoration did not pass for matrix06.** Frozen
source remained unchanged. Earlier failed attempts' passing deletion/restoration
results do not substitute for these missing predicates.

The whole corrected17-step matrix, later drift/repair, firewall toggles and final
exact restoration therefore remain pending. Do not count historical or partial
observations as a full live PASS. A future fresh isolated target must first prove
firmware discovery health and retain cumulative collector evidence. Any recovery
reboot and its effectiveness remain unexecuted/unverified in this attempt.

Local validation of the current test-only repairs passed focused race1.036s,
full make check6/6with none skipped83633ms, deterministic generation and
whitespace/privacy checks. All8 hosted checks completed successfully on the exact
latest tested head above. These are offline/hosted gates, not full appliance
acceptance or merge authorization. Keep the PR draft and parent Story31.0h open.

After retaining the failed receipt, the exact owned disposable guest/disk and collector, tunnel and oracle were retired. Independent parent checks found the owned guest/config/disk absent, all owned listeners gone and no acceptance container remaining. Physical disposable retirement is cleanup evidence only; it does not turn unverified Terraform deletion or failed exact restoration into successful predicates.

Final evidence-only publication gate: make checkPASS6/6, none skipped81938ms; whitespace/privacy auditPASS. The interrupted acceptance left generated Go build temporaries and a plugin socket; these initially failed formatting and the transitive-source fixture copy. Only the stopped run's generated temporary tree was retained outside the checkout, with no source modification, deletion or check suppression, before the full successful rerun. This local cleanup does not change the failed live disposition.


## Seventh failed attempt — fresh health gates did not prevent recurrence

Attempt `matrix07` tested head
`74e13df33d54b725f86aa6fdbbb171209a3355ba` and the unchanged frozen
source digest
`2e26cb01403f0f3927d77313412ac9d39331493294b4e798e1c2fb642feeff44`.
A newly reserved isolated disposable independently passed exact identity,
endpoint/TLS, package and four released-source-file checks. Three serial real
firmware-info requests returned HTTP200 with the expected product version in
2.794s,2.746s and2.701s under the unchanged30-second endpoint deadline.
The actual BGP API returned HTTP200 in0.024s. Backend observations found no
stalled metadata queue at readiness. The first full-activation PF-order mismatch
was retained; a second exact six-predicate fallback selfcheck passed against the
actual post-activation baseline without normalization.

The complete, non-skipped matrix was started from the beginning. Initial
creation, legacy-provider schema-v0 import/upgrade/no-change/deletion, all19
advanced public fields, password-free lookup and an active approved current
prefix passed their assertions. **Step1/17 then failed its post-apply refresh**
on actual firmware-info discovery. No later steps or full matrix PASS are
claimed.

The retained independent diagnosis observed a high-CPU `pkg rquery` process
under firmware tiers, repository lock ownership, an upstream HTTPS socket and
subsequent lock-waiting metadata requests. The process outlived the configd
120-second timeout. Installed pkg was2.8.4_1. This is more specific than the
previous attempt's busy-update observation, but it does not establish whether
the low-level defect is network/TLS, repository refresh or query iteration.
No package installation, metadata bypass or timeout relaxation was used.

Actual Terraform destruction failed with firmware-discovery errors; five
namespaces still contained objects before raw rollback. The separately
registered raw cleanup then reported zero objects in all seven namespaces, and
exact settings/generated-FRR/effective-and-rendered-PF/routes restoration
passed. The independent full-config fallback also passed all six predicates,
including the original config hash. Frozen source remained unchanged.
**Raw cleanup and fallback success are not successful Terraform deletion.**

## Guarded diagnostic recovery — readiness only

After both failed-run and fallback receipts finished, an identity-guarded
reboot was actually executed on that same owned disposable. UUID, installed
versions and authority-file hashes remained unchanged. The first exact restore
after reboot failed5/6 solely because PF's generated hostid changed; that
failed receipt and exact diff remain retained. No baseline was recaptured.

Installed source confirms filter generation reuses PF's current hostid.
Reapplying the entire unchanged original canonical PF input with options-only
`pfctl -O`, then explicit lab filter regeneration, restored the original
transient hostid. Exact effective-filter, NAT, config and route observations
were identical before/after; the original full six-predicate fallback then
passed. This was lab-owned state restoration, not provider behavior or output
normalization. Intermediate private transport/request-preflight failures were
retained separately, not represented as successful appliance checks.

Subsequent real, serial auth-only firmware-info calls returned HTTP200 with
the expected core/plugin metadata in2.972s,2.825s and2.843s. Backend process
observations were idle; all seven namespaces were empty and the same DevRail
container passed endpoint/oracle smoke. Cumulative collector history and
forbidden-default evidence were preserved without restarting/resetting the
collector. This proves post-recovery readiness only, not a cure or
nonrecurrence. The next whole matrix result is recorded separately.


## Eighth attempt - complete Go matrix PASS, final runner gate interrupted

Attempt `matrix08` reran the entire unchanged candidate from the beginning after
the guarded recovery. It used head
`74e13df33d54b725f86aa6fdbbb171209a3355ba` and frozen source digest
`2e26cb01403f0f3927d77313412ac9d39331493294b4e798e1c2fb642feeff44`.
The actual non-skipped `TestAccFRRCompatibility` passed all17 steps in1478.94s.
This includes nine imports, the mandatory numeric-policy current-prefix
withdrawal and subsequent restored-positive observation, authoritative grouped
references and stable identities, advanced fields, drift/repair, explicit lab
firewall disable/restore and forbidden-default monitoring. Actual Terraform
destruction reported zero in all seven namespaces; both before-rollback and
final cleanup also reported seven zeros. Exact test-level settings, generated
FRR, effective/rendered PF and route restoration passed.

**The end-to-end guarded public runner did not succeed.** Its post-run probe
hashes matched, secret scan and required proof/zero checks completed, but the
private orchestration wrapper reached its1500-second limit during the final
public source-digest hashing pipeline. The final public digest-equality
assertion and `FRR_COMPATIBILITY_EVIDENCE` summary were not completed. Retained
termination diagnostics identify that precise phase; this is an orchestration
failure, not a fabricated successful runner result or a new firmware timeout.

The private wrapper independently recomputed the unchanged frozen source and
its independent full-config fallback restored all six exact predicates,
including the original configuration hash. Its emergency disposal exited zero.
Parent readback independently verified the owned guest, disk, collector,
tunnel, oracle listeners and acceptance container absent. No retired target
will be resurrected.

Keep the PR draft and this delivery checkpoint in progress. A fresh isolated
whole-run retry must budget orchestration, the complete matrix, post-run scan
and hashing, fallback and disposal together; the observed Go runtime alone was
within21.06s of the outer limit. Increasing the orchestration budget is not
permission to relax provider/API discovery deadlines, skip post-run gates or
resume at Step17. Preserve all failed and intermediate recovery receipts.
Legacy appliance/OpenTofu execution and the numeric-clear interface remain
separate open parent-story requirements. No merge or release is claimed.
