# Story 31.0f: Trust Lifecycle and Settings Parity

## Status

in-progress — merged settings plus lifecycle continuation; live proof keeps the umbrella story open.

## Priority

P0 — complete before Stories 31.1–31.3.

## Evidence and Scope

Extend the existing CA/certificate owners for import, internal generation, intermediate/OCSP, external CSR/sign/import/reissue actions, subject/SAN/EKU, digest/key type/lifetime, and operational metadata. Add the certificate data source and settings singleton resource/data source. Stable settings evidence is `GET /api/trust/settings/get`, `POST /api/trust/settings/set`, and `POST /api/trust/settings/reconfigure`; existing trust CA/cert endpoints own object lifecycles. Private-key values are sensitive and local import inputs are currently write-only.

## Acceptance Criteria

- **Given** each supported CA/certificate action, **when** create/import/read/update/delete or state-only lifecycle applies, **then** the action-specific contract is validated and no secret is exposed without `Sensitive`.
- **Given** a certificate UUID, **when** the data source reads it, **then** public metadata and API-returned sensitive fields match documented state semantics.
- **Given** imported trust settings, **when** updated, **then** all six audited settings round-trip and reconfigure; destroy removes state only.

## Verification

Trust unit/schema tests, action-matrix disposable-appliance acceptance, docs generation, and `make check`.

## Settings increment checkpoint (2026-11-03)

Implemented the six audited settings resource/lookup with fixed import identity,
partial omission-preserving updates, strict reads, durable private recovery and
state-only destruction. The released 26.7.8 XML/controller are the source oracle.
Local inventory is 113 resources / 99 data sources; the 16 lookup gaps remain 16.
HTTP-backed protocol evidence must not be promoted to live or release proof.

Remaining resumable steps:

1. Certificate/CA authoritative read parity and certificate lookup.
2. CA import/internal/intermediate/OCSP actions with legacy compatibility.
3. Certificate CSR/generation/sign/import/reissue, including one-time keys.
4. Disposable action-matrix and settings lifecycle acceptance/restoration proof.

Source risks: volatile editor action/lifetime fields and possible CA cert-only
import defect need action-specific investigation; do not add naive writable fields.
Five other upstream TLS settings remain outside this increment. No shared or
production appliance was mutated and no appliance lifecycle was executed.

### Released source oracle

OPNsense tag 26.7.8 resolves to commit `96978001a8343669154e4492172ce0a07dac7e4b`:
[General.xml](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/Trust/General.xml)
and [SettingsController.php](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/controllers/OPNsense/Trust/Api/SettingsController.php).
The model mount is `//OPNsense/trust/general`; the controller exposes the model
as `trust`, not a second `general` API nesting. Reconfigure returns `status=ok`.
The baseline audit classification stays `missing-buildable` until its full
lifecycle follow-up is satisfied, despite the bounded local implementation.

## Remaining lifecycle continuation (2026-11-03)

The settings increment is merged at
`4bd107fc769586d5904919c8e69562897a87e84d`; the v0.12.0 automatic release
completed with all 16 expected assets. Registry availability is a separate gate.
The continuation is specified in
[the lifecycle spec](spec-31-0f-trust-lifecycle-continuation.md), grounded in
[the released source contract](31-0f-trust-source-contract.md).

The full parent remains **in-progress**. Buildable scope includes CA root,
intermediate, OCSP and matching-key import; certificate internal/local-only,
CSR/sign/import/reissue/manual lifecycles; safe metadata updates, separate
returned public/sensitive values and certificate lookup. OPNsense 26.7.8's
certificate-only CA import restriction remains a provider limitation. The later patched package fixes the upstream condition, but this provider capability has not been enabled.

Disposable action/settings proof remains **unexecuted**: no newly reserved,
identity-bound runtime has been established. The acceptance runner refuses to
use an ordinary configured endpoint without disposable authority; there have
been zero appliance mutations in this continuation. A draft PR is an
implementation/evidence checkpoint, not full story completion or release proof.

### Disposable runner contract

Run `test/scripts/run-story-31-0f-acceptance.sh` only with an external private
`OPNSENSE_TRUST_RUNTIME` JSON file and ephemeral API environment. The wrapper
uses the pinned toolchain and mounts the repository, private runtime root and
actual common Git metadata read-only at its absolute path (including isolated
worktree references). An explicit `SSL_CERT_FILE` is forwarded only when its
resolved file is inside the already-mounted private runtime root; no SSH keys
are mounted and insecure TLS has no fallback. Required JSON fields are `owner` (fresh reservation UUID), `guest_uuid`,
`endpoint` (verified HTTPS), `tls_sha256`, `source_sha256`, `guard`, `oracle`,
`guard_sha256` and `oracle_sha256`. Helpers must be executable, digest-bound,
not writable by others, and inside that private root. The source digest hashes
sorted Git tracked/unignored files as path, NUL, bytes, NUL; it must match the
reviewed frozen candidate before building and after execution.

Each helper receives `action`, private attempt directory and reservation owner.
Every reply must affirm `ok`, `isolated`, `non_production`,
`reservation_current`, exact core version `26.7.8_2`, and match owner, guest UUID,
endpoint and TLS fingerprint. The guard must establish actual VM identity,
current reservation and endpoint/TLS binding from authoritative live evidence,
not echo caller assertions. Helpers must never expose credentials publicly.

The oracle handles `snapshot`, `observe`, `before-import`, `verify-import`, `before-destroy`, `cleanup`, `verify-zero`, `restore`
and `verify-restored`. Snapshot persists a private exact baseline and returns
SHA256 hashes for config/settings/routes/runtime plus `cipher_choice`, a safe
available CipherString choice read from the live API catalog, not an invented
caller value. Observe must independently
verify marker-owned API changes and isolation. Before-destroy requires a nonempty owned set and captures current settings; the pre-rollback verify-zero receipt must affirm `settings_destroy_equal=true`. Cleanup removes only objects
owned by this reservation; verify-zero returns `owned_count=0` before rollback.
Restore reinstates the baseline; verify-restored returns `exact_restored=true`
and identical four hashes. A failed Terraform deletion is retained as failure,
not hidden by oracle cleanup. Signal/error paths attempt every cleanup phase.
Before-import starts an independent request/config observation window;
verify-import must return `cryptographic_writes=0` and `config_equal=true`,
using bound request instrumentation rather than inferring no writes from equal
certificate fingerprints alone. The adoption workspace imports both UUIDs with
description-only configuration, checks unchanged plans and removes only its
local state; the main workspace retains deletion ownership.

The oracle must protect against a changed guest/endpoint and abort unsafe writes.
Every command is time-bounded; timeout/cancellation terminates and reaps only its
owned process group before cleanup. Malformed/non-object helper replies are
failures but do not prevent attempts at later restoration phases.

The CLI path creates root/intermediate/OCSP/import CAs, internal/firewall and
local-only certificates, external CSR and imported certificate; signs and
imports the signed result back into the same CSR holder (verifying UUID, stored
key and CSR/certificate public-key correspondence), reissues, verifies metadata
fingerprint stability and one-time key retention, reads CA/cert/settings lookups,
imports both object UUIDs and settings, exercises all five settings booleans as
true/false/restored and CipherString selection/empty-clear/restored,
and destroys before exact restoration. These are executable test intentions,
**not executed appliance results**. Offline guard-negative tests are separate.
Before calling this complete, establish the runtime, execute this frozen path,
retain successful no-change/import/action/restore receipts and fill any remaining
matrix gaps (including older-firmware/OpenTofu boundaries) honestly.

### Bounded live acceptance repair (2026-11-04)

The original official-package attempt reached a real disposable guest and
failed the frozen runner's exact `26.7.8` gate because the authentic signed
package is `26.7.8_2`. Earlier statements above about unavailable reservation
and zero appliance mutations describe the original continuation checkpoint,
not current discovery or provisioning. The failure remains private evidence;
it is not live matrix success. See the [source distinction](31-0f-trust-source-contract.md#actual-acceptance-package-distinction-2026-11-04).

The corrected runner requires exactly `_2` and preserves all owner, guest,
endpoint, TLS, digest, isolation, cancellation and recovery gates. No provider
lifecycle algorithm or keyless import support changes in this repair. Live
matrix execution remains incomplete pending the parent's actual run, independent
cleanup/zero-leftover oracle and four-hash exact restoration receipts. Offline
synthetic helpers validate orchestration only and never establish appliance proof.


### Executed live matrix supersedes discovery blocker (2026-11-04)

The bounded repair's real frozen runner passed on a newly reserved isolated
OPNsense 26.7.8_2 guest. The executed source digest is
`91aebd345968990c4b8b84ae3a80eede0b8a3f00d1ff003d07a42dfd67eee26d`;
see the [repair execution checkpoint](spec-31-0f-live-acceptance-repair.md#parent-real-execution-checkpoint-2026-11-04)
for actual action/settings/adoption/destroy/restoration coverage and failed-run
history. The previous unavailable-lab/live-unexecuted statements are historical,
not a current infrastructure blocker. Nine owned objects were deleted before
rollback, singleton deletion was state-only, and all four baseline hashes
restored exactly. Keyless CA provider support and broader firmware/OpenTofu
validation remain deferred; this does not mark the whole parent complete,
merge the draft PR or establish a published release.

### Current post-review proof and reproducible code binding (2026-11-04)

The post-review read-only harness rerun passed the complete non-skipped matrix.
The [sanitized proof](31-0f-live-proof.json) records the exact tested working-tree
digest, executable digest, authentic firmware, Terraform/toolchain, coverage,
request counters, baseline/restoration hashes and retirement result. The newly
reserved disposable guest, disk, owner marker and owned listeners were retired
after restoration. Private raw baselines and failed attempts remain retained.

Evidence-only documentation changes make the final whole-tree digest different.
Verify final executable equality from the repository root with this inventory
procedure (no private files or credentials are read):

```python
import hashlib
import subprocess
from pathlib import Path

names = subprocess.check_output(
    ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"]
).split(b"\0")
digest = hashlib.sha256()
for name in sorted(set(names) - {b""}):
    text = name.decode()
    path = Path(text)
    if path.is_file() and (
        text.endswith(".go") or text in ("go.mod", "go.sum")
        or text.startswith("test/scripts/")
    ):
        digest.update(name + b"\0" + path.read_bytes() + b"\0")
print(digest.hexdigest())
```

Expected executable SHA256:
`66bbfa18dccf14768d733a120e4a7a211efcb6943e8102adf1b6f9e104d4122b`.
This binds the final public executable bytes to the actual tested candidate
without publishing private guest identity, raw state or keys.

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
tree. [Pinned earlier sanitized proof](https://github.com/matthew-on-git/terraform-provider-opnsense/blob/17cb7b0064a3beb217e9f011be3048d6103d760b/_bmad-output/implementation-artifacts/31-0f-live-proof.json) and its pinned historical
reference retain that distinction. All6 local gates,16 offline guard methods,
Trust/coverage race suites and3 real Core create-recovery controls pass. The
old production snapshot fails all3 controls; repaired creation emits warnings
after acknowledged identity instead of tainting and automatically recreating.

Four independent review layers produced13 individually triaged findings; the
follow-up bounded production-diff review found no concrete blocker. The CA PEM
normalization claim was rejected on actual released-source evidence, not masked
by retaining stale observations. Keyless CA remains explicitly unsupported,
older-firmware/OpenTofu execution is unverified and the parent story stays open.


### Final signed-date executable proof — 2026-11-04

This supersedes the earlier merge-candidate proof for current executable bytes.
Candidate `d87b4aa141384c3356dbff93eb2f03029a2a9941` passed the **complete**
identity/TLS-bound guarded matrix on authentic OPNsense26.7.8_2, exit0 with final
source/runtime digest assertions and success marker: **26 apply/no-change pairs**
including19 settings cases. A real1969 CA/certificate retained exact signed dates
`-31536000`/`-86400` in both imported resources and both authoritative lookups.
Signing, CSR import, usable one-time key, stable metadata, reissue and new-token/
unchanged-token controls passed; independent cryptographic-write counts were
**1/1/0**. Terraform destroyed all9 owned objects, leaving0 before rollback and
at final inspection; settings destroy changed no policy. All four original
configuration/settings/routes/runtime hashes restored exactly. The exact owned
guest, disk, marker, container and local/remote listeners were retired; absence
was independently read back.

Executed whole-tree SHA256:
`d34bdd2b0c249283c369a3f63be4bdce43ec9ce7340dc42719c9bcb7e6bc95e7`.
Executable SHA256 (753 files):
`287556a87130595829bd216cf425e4f81cf85faf9cfeee6aef023edd60d4362d`.
The evidence-only publication is not called the executed whole tree. Its executable
inventory must remain equal. [Sanitized current proof](31-0f-live-proof.json)
retains the earlier successful proof and failed signed-date receipt as distinct
history. The1661ccd failure was fixture-year interpretation, not a passing
acceptance run; fallback recovery does not prove Terraform deletion for it.

All6 candidate gates passed with0 skipped (32,568ms),16 offline guards passed,
and real pre-epoch protocol controls passed under the race detector. All8 hosted
checks passed at d87b4aa. Final evidence-head gates/CI are verified separately.
Keyless provider support, older firmware and OpenTofu execution remain outside
this proof; the parent story remains in progress. No production appliance was
targeted. Readiness, authorized merge and automatic release are separate events.
