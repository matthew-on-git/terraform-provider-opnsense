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
