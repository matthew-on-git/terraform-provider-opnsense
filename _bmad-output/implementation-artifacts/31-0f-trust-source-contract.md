# Trust lifecycle source contract

## Released oracle

OPNsense 26.7.8 resolves to `96978001a8343669154e4492172ce0a07dac7e4b`.

- [CA controller](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/controllers/OPNsense/Trust/Api/CaController.php): `internal` generates a root or issuer-linked intermediate, `ocsp` generates an OCSP signer, `existing` imports. CA add/set/delete already invoke trust configuration.
- [Certificate controller](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/controllers/OPNsense/Trust/Api/CertController.php): `internal`, `external`, `import`, `sign_csr`, `import_csr`, `reissue`, `manual`. Reissue does not restart consumers. `manual` avoids generation on ordinary description edits.
- [CA model](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/Trust/Ca.xml) and [certificate model](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/Trust/Cert.xml): action, generation inputs and payloads are volatile. Subject, SAN, key type, digest and certificate type use their existing editor fields; `rfc3280_purpose` is derived output, not arbitrary EKU input.
- [CA loader](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/Trust/FieldTypes/CAsField.php) and [certificate loader](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/Trust/FieldTypes/CertificatesField.php): GET action is inferred from stored certificate/CSR/key, not original request provenance. SAN inputs are newline-delimited strings.
- [Store](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/library/OPNsense/Trust/Store.php): certificate validity is numeric Unix time; lifetime is not reconstructed. CSR parsing reconstructs subject/name only. Editor lifetime defaults must not overwrite configured generation intent or be called actual validity.

## Historical unpatched keyless-CA limitation

CA controller lines 117–129 check a nonempty key for parse errors, then unconditionally require exactly one private-key section. An empty key therefore fails with zero sections even for a valid certificate. This also affects metadata updates of an existing keyless CA through `existing`. The provider must diagnose this before mutation, not regenerate the CA or invent a private key. The later patched-package correction is documented below; the provider still requires a separate implementation and proof increment before enabling keyless import.

## One-time keys

Only `internal` certificate creation with `private_key_location=local` returns a nonpersisted key, as top-level `private_key` in the add response. GET and set cannot recover that response-only value. Capture known UUID/key before follow-up reads; preserve that Sensitive state across blank refreshes. Do not send the retained output back to the appliance. A lost add response can irretrievably lose this key; automatic replay is not recovery.

## Evidence boundaries

Source inspection is not appliance proof. This continuation requires offline action/protocol tests, deterministic generated documentation and the full quality gate. At the original continuation checkpoint the disposable matrix was unexecuted; the later actual-package execution below supersedes that discovery blocker. Newly reserved, identity/version/endpoint-bound authority, exact restoration and zero-owned-leftover proof remain mandatory. Historical targets must not be reused without fresh reservation and attestation. The parent Trust story remains open for the explicit remaining capability/compatibility boundaries; a live matrix pass is not a merge or release.

## Actual acceptance package distinction (2026-11-04)

The authentic official signed package and signed release-revert snapshot now
serve `26.7.8_2`, not the original unpatched `26.7.8` source oracle above.
The actual-package comparison changes the private-key section count condition
from `!= 1` to `> 1` at CaController line 124 and CertController line 61;
the other seven audited Trust files are byte-identical. This is a real released
behavior difference, not a version-label equivalence. The historical source
review at `96978001a8343669154e4492172ce0a07dac7e4b` remains evidence of the
original `26.7.8` contract only.

The guarded live runner therefore pins exactly `26.7.8_2`. It rejects the
original `26.7.8`, other suffixes and malformed versions before snapshot or
provider mutation. The prior frozen-runner attempt against the actual package
failed its exact-version gate; that failure is retained privately, not replaced
with a network-discovery blocker or counted as successful matrix evidence.
The provider still rejects keyless CA import/update before mutation: the
official controller correction does not enable this deferred provider feature.
That actual non-skipped run subsequently passed; see the
[execution checkpoint](spec-31-0f-live-acceptance-repair.md#parent-real-execution-checkpoint-2026-11-04)
for its source digest, covered actions and exact zero/restoration evidence.
Original source/package inspection remains distinct from this live execution.

The actual released General.php::performValidation requires at least one
nonempty constraint before enabling config constraints. The fixture therefore
uses an actual available CipherString choice, disables enforcement with an
empty clear, and restores the original boolean/cipher pair. The authentic
earlier validation rejection is retained as negative evidence, not suppressed.
