# Advanced FRR acceptance checkpoint

Status: **in progress — the full17-step matrix has not passed**. This is partial
executed evidence and a resumable checkpoint, not a release or completed parent
Story31.0h. Candidate implementation is in
[PR83](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/83).

## Executed candidate and runtime

- Tested head: `82de6314ab78e7e9e93e5ad2722f4b0f11f9de45`.
- Frozen source digest: `7efa6f7be577cd0722a74de4a5f16b72072e014ac89f2c096c68719d5624e7de`.
- Actual disposable versions: OPNsense26.7.8_2, os-frr2.0_4, bgpd10.7.1.
- Installed BGP model/controller and both routing/authentication templates exactly
  match released pin `5ac2959c8514b4f60e94d65439be5f629f10521f`; source hashes are
  in `31-0h-advanced-source-evidence.json`.
- New isolated target identity, endpoint/TLS, bridge isolation and versions were
  independently attested; private inventory and raw configuration are withheld.

## Actually observed

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
  prefix withdrawal before its effective-order marker. That corrected complete
  matrix has not yet run.
- Earlier two attempts failed synthetic flatmap-count assertions after adding
  resource-only authentication. Both comparator sites are now covered by focused
  missing/extra/value/set/password negative controls. Failed attempts remain
  failures; none is counted as a full matrix PASS.

## Restoration and cleanup evidence

On each failed attempt, actual Terraform deletion reported zero objects in all
seven FRR namespaces before rollback. Exact API settings, generated FRR,
effective/rendered PF and routing restoration passed. Independent full-config
fallback restoration also matched all six captured predicates, including the
configuration hash. The owned disposable guest/disk and collector, tunnel, oracle listeners were then retired; independent checks found no owned listener or acceptance container remaining. These cleanup results do not turn a failed matrix into PASS.

## Resume and limits

Rerun the entire frozen17-step matrix on a newly attested disposable target after
the test-only ordering correction. Retain source/head binding, current-prefix
positive/withdrawal/restoration observations, UUID/group checks, all advanced
flags, actual Terraform deletion, exact restoration and zero leftovers. Do not
replace these gates with historical counters or a mocked response.

Later drift/repair, firewall-toggle/restored-positive steps and the full final
candidate remain unverified. Legacy appliance/OpenTofu execution and the numeric
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
