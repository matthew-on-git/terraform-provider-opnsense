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
certificate-only CA import defect remains an explicit release limitation.

Disposable action/settings proof remains **unexecuted**: no newly reserved,
identity-bound runtime has been established. The acceptance runner refuses to
use an ordinary configured endpoint without disposable authority; there have
been zero appliance mutations in this continuation. A draft PR is an
implementation/evidence checkpoint, not full story completion or release proof.

### Disposable runner contract

Run `test/scripts/run-story-31-0f-acceptance.sh` only with an external private
`OPNSENSE_TRUST_RUNTIME` JSON file and ephemeral API environment. The wrapper
uses the pinned toolchain and mounts only the repository and private runtime
root. Required JSON fields are `owner` (fresh reservation UUID), `guest_uuid`,
`endpoint` (verified HTTPS), `tls_sha256`, `source_sha256`, `guard`, `oracle`,
`guard_sha256` and `oracle_sha256`. Helpers must be executable, digest-bound,
not writable by others, and inside that private root. The source digest hashes
sorted Git tracked/unignored files as path, NUL, bytes, NUL; it must match the
reviewed frozen candidate before building and after execution.

Each helper receives `action`, private attempt directory and reservation owner.
Every reply must affirm `ok`, `isolated`, `non_production`,
`reservation_current`, exact core version `26.7.8`, and match owner, guest UUID,
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
