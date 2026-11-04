# Trust lifecycle source contract

## Released oracle

OPNsense 26.7.8 resolves to `96978001a8343669154e4492172ce0a07dac7e4b`.

- [CA controller](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/controllers/OPNsense/Trust/Api/CaController.php): `internal` generates a root or issuer-linked intermediate, `ocsp` generates an OCSP signer, `existing` imports. CA add/set/delete already invoke trust configuration.
- [Certificate controller](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/controllers/OPNsense/Trust/Api/CertController.php): `internal`, `external`, `import`, `sign_csr`, `import_csr`, `reissue`, `manual`. Reissue does not restart consumers. `manual` avoids generation on ordinary description edits.
- [CA model](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/Trust/Ca.xml) and [certificate model](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/Trust/Cert.xml): action, generation inputs and payloads are volatile. Subject, SAN, key type, digest and certificate type use their existing editor fields; `rfc3280_purpose` is derived output, not arbitrary EKU input.
- [CA loader](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/Trust/FieldTypes/CAsField.php) and [certificate loader](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/models/OPNsense/Trust/FieldTypes/CertificatesField.php): GET action is inferred from stored certificate/CSR/key, not original request provenance. SAN inputs are newline-delimited strings.
- [Store](https://github.com/opnsense/core/blob/96978001a8343669154e4492172ce0a07dac7e4b/src/opnsense/mvc/app/library/OPNsense/Trust/Store.php): certificate validity is numeric Unix time; lifetime is not reconstructed. CSR parsing reconstructs subject/name only. Editor lifetime defaults must not overwrite configured generation intent or be called actual validity.

## Upstream keyless-CA limitation

CA controller lines 117–129 check a nonempty key for parse errors, then unconditionally require exactly one private-key section. An empty key therefore fails with zero sections even for a valid certificate. This also affects metadata updates of an existing keyless CA through `existing`. The provider must diagnose this before mutation, not regenerate the CA or invent a private key. A future released-source correction and disposable proof are needed before enabling keyless import.

## One-time keys

Only `internal` certificate creation with `private_key_location=local` returns a nonpersisted key, as top-level `private_key` in the add response. GET and set cannot recover that response-only value. Capture known UUID/key before follow-up reads; preserve that Sensitive state across blank refreshes. Do not send the retained output back to the appliance. A lost add response can irretrievably lose this key; automatic replay is not recovery.

## Evidence boundaries

Source inspection is not appliance proof. This continuation requires offline action/protocol tests, deterministic generated documentation and the full quality gate. The disposable action/settings matrix is still unexecuted until a newly reserved, identity/version/endpoint-bound target is available, with exact restoration and zero-owned-leftover evidence. Existing historical targets must not be reused without fresh reservation and attestation. The parent Trust story remains open, and the continuation PR remains draft while live proof is absent.
