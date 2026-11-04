# Story 31.0f: Trust Lifecycle and Settings Parity

## Status

in-progress — settings increment only; umbrella story remains open.

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
