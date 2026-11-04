# Story 31.0g: IPsec Field Parity

## Status

in-progress — implementation and offline verification; isolated live lifecycle
proof is outstanding. Keep the PR draft. Parent story 31.0g is **not complete**.

## Priority

P1 — complete before Stories 31.1–31.3.

## Evidence and Scope

Extend only the existing connection and child owners and their UUID lookups.
All 15 connection and six child audited controls are Optional+Computed without
provider defaults. Flags are Bool; numeric values remain decimal strings. The
existing `remote_addrs`, proposal strings, `unique`, resource identities and
legacy field types/defaults remain intact. Child multi-proposal reads preserve
all selections; matching CSV membership retains prior/configured spelling/order.

The released source oracle is OPNsense core **26.7.8**, immutable commit
[`96978001a8343669154e4492172ce0a07dac7e4b`](https://github.com/opnsense/core/tree/96978001a8343669154e4492172ce0a07dac7e4b):

- [Swanctl.xml connection and child fields](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/IPsec/Swanctl.xml)
- [ConnectionsController.php CRUD/cascading parent deletion](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/controllers/OPNsense/IPsec/Api/ConnectionsController.php)
- [BaseField.php omitted/non-null key semantics](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/Base/FieldTypes/BaseField.php#L810-L830)
- [IKEAddressField.php selected addresses](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/IPsec/FieldTypes/IKEAddressField.php)
- [PoolsField.php UUID/radius selections](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/IPsec/FieldTypes/PoolsField.php)
- [Swanctl.php saved-to-daemon conversion](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/IPsec/Swanctl.php#L211-L316)

The earlier `/api/ipsec/swanctl/*` reference was stale. Retain
`/api/ipsec/connections/{add,get,set,del}_{connection,child}`, `connection`/`child`
envelopes and `/api/ipsec/service/reconfigure`. Successful saves activate once;
rejected saves never activate. Saved create failures retain known UUID/state;
failed activation has UUID-bound private recovery and refresh retries only
activation, without repeating the object save. Read-only lookups never activate.
Malformed/missing advanced field responses diagnose errors without erasing policy.

Omitted advanced configuration does not serialize wire keys, preserving imported
policy, replay-window and independently owned auth. Explicit false, zero and
permitted blanks are real writes. Unknown explicit values fail before mutation.
Ports accept blank (daemon 500) or 4500 only. Numeric bounds follow the pinned
model; child rekey time cannot be blank and its omitted server default is 3600.
`ike_lifetime` is **overtime grace**, not total IKE lifetime. Authentication,
certificates, keys and PSKs remain outside these models/lookups.

## Acceptance Criteria

- **Given** imported advanced connection/child objects, **when** read and actual
  framework protocol plan/lookup run, **then** every audited field round-trips
  without loss or perpetual diff and lookup causes no writes.
- **Given** an advanced update, omission, clear or drift, **when** apply runs,
  **then** exact configured wire keys target only the owned UUID, unowned policy
  survives, and activation follows successful persistence.
- **Given** invalid/unknown values, cancellation, malformed GET or confirmed-save
  partial failure, **when** lifecycle operations run, **then** diagnostics retain
  recoverable identity/prior confirmed state without duplicate creation.
- **Given** legacy models and secret/auth child owners, **when** schemas/readback
  are reviewed, **then** names/types remain compatible, child proposals and
  `trap|start` are faithful, and secret ownership remains separate.
- **Given** no positively attested isolated runtime, **when** delivery is
  reviewed, **then** live proof stays unexecuted, with zero appliance mutations
  and a guarded executable acceptance path rather than a completed-story claim.

## Verification

The offline test suite includes independent source-shaped HTTP fixtures and
actual `PlanResourceChange`/`ApplyResourceChange`, not only conversion helpers:

- `TestIPsecAdvancedProtocolLifecycle`: both owners, all-field create/read,
  no-change plan, UUID lookup/import, omission-safe update, explicit clears,
  drift/repair, independently addressed deletion, mutation/activation counts, and unchanged
  second-object/auth policy snapshots.
- `TestIPsecAdvancedProtocolFailureRecovery` and
  `TestIPsecAdvancedProtocolUpdateFailures`: rejected save, failed activation,
  failed authoritative readback, retained identity/prior state, refresh recovery.
- `TestIPsecAdvancedProtocolStrictRead` and
  `TestIPsecAdvancedProtocolMalformedEveryField`: cancellation, missing object,
  all 21 malformed/missing advanced fields, unchanged prior state and zero writes.
- `TestIPsecAdvancedProtocolValidation`,
  `TestIPsecAdvancedProtocolUnknownBeforeMutation` and
  `TestIPsecAdvancedProtocolOmittedCreateReadFailure`: numeric/enumeration
  boundaries, unknown rejection and usable saved identity with omitted fields.
- `TestIPsecAdvancedSchemaCompatibility`: all 21 additions, lookup type parity,
  unchanged legacy string representation and absence of secret attributes.
- `TestIPsecDisposableIsolationGuard`: exact target, reservation, new/disposable
  isolation, production-network exclusion, bounded lease, independent identity
  fingerprint and explicit service-wide activation authorization.

Initial implementation verification (superseded by final parent recheck below):

- Container `go test -count=1 -race -v ./internal/service/ipsec/... ./internal/coverageaudit/...`: **PASS**. All named offline matrix tests ran; `TestAccIPsecAdvancedDisposable` was explicitly **SKIPPED**, not counted as live proof.
- Container `go generate ./...`, twice after staging intended outputs: **PASS**, zero unstaged generated differences after each run.
- `make check DEVRAIL_IMAGE=ghcr.io/devrail-dev/dev-toolchain:1.12.24`: **PASS**, all six gates, no suppressions or skipped gates. Final aggregate: 82,781 ms; lint 29,214 ms, format 240 ms, test 43,933 ms, security 1,735 ms, scan 6,878 ms, docs 486 ms. The isolated worktree invocation additionally mounted its Git metadata read-only so history scanning could resolve the worktree; no repository/tool configuration was changed.
- `git diff --check` and staged diff check: **PASS**. Complete authored/generated public diff reviewed for private hostnames, inventory, credentials and operator-specific paths: **clean**.
- The first aggregate caught formatting differences in the new Go files; container `gofumpt` corrected them before the passing final check. No failing check was suppressed.
- API diagnostics classify validation rejection versus confirmed saved activation failure without rendering untrusted server validation/result payloads; a payload-marker regression assertion passes.

## Live gate and next action

`TestAccIPsecAdvancedDisposable` is executable but **unexecuted/skipped** without
`TF_ACC=1`, explicit disposable acknowledgment and a fresh private isolation
receipt. No existing/shared target or credential fallback is permitted. Follow
[the guarded proof procedure](../../docs/ipsec-advanced-acceptance.md): newly
reserve/attest an isolated 26.7.8 guest, run disabled connection/child
create/import/no-change/drift/clear/delete, verify independent raw saved fields
and independent child-before-parent absence, and retain sanitized receipts and
cleanup evidence. A reservation service/signature is not verified by the test;
independent infrastructure attestation remains an operator responsibility.

No target was reserved or positively attested during implementation. **Zero live
writes; live lifecycle remains blocked.** Audit classification intentionally stays
`missing-buildable`; no counts or unrelated story statuses change. The proposed
0.13.0 unreleased changelog heading is an automatic release input after merge and
successful main CI, not evidence of merge, release or Registry publication.


## Final parent recheck after review repairs

All four BMAD review layers completed and each finding received a verdict. Bounded
independent review found one pending-activation/expired-lease cleanup interaction;
that was repaired and independently rechecked. Parent executed all20named offline
IPsec controls under the race detector, including actual Terraform Core import,
omitted advanced no-change/update and configured CSV no-change plans. Final
IPsec9.019s, coverageaudit1.422s, Core4.67s: PASS. Exact wire/activation-marker
counterfactuals failed as intended and original bytes were restored. Final
post-repair generation twice was deterministic; public/whitespace audits clean.
Final required six-gate make check PASS,58979ms, none skipped or suppressed.

Activation now requires affirmative status=ok; invalid pending updates never
activate first. Dedicated live opt-in is separate from ordinary TF_ACC; every
live mutation/retry is origin/identity/lease guarded. Independent bounded cleanup
retains only recorded owned IDs, deletes child-before-parent, proves absence and
preserves failed Terraform outcomes/uncertain ownership. These are executed
**offline guard controls**, not a real appliance result. Live proof remains
unexecuted and the parent story/PR stay in progress/draft. Legacy omission/read
completeness and unobservable proposal preference order remain explicit deferred
limitations in the build-auto spec. Published0.12.0 is historical; proposed0.13.0
is an unreleased automatic release input, to refresh after other pending merges.
