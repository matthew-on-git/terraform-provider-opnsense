# 31.0s IPsec key pair lookup

Status: review

## Outcome

Add a read-only `opnsense_ipsec_key_pair` lookup for an existing key pair. Do not duplicate trust certificate work owned by 31.0f. Do not implement a second key-pair owner.

## Source contract

Latest stable tag is `26.7.9` (`stable release`), peeled commit `228ab68d54513046f79bc36d7f583936922d1951`. Previous stable tag `26.7.8` peels to `96978001a8343669154e4492172ce0a07dac7e4b`. Development tags `27.1.a` and `27.1.b` are not the contract.

Identical between `26.7.8` and `26.7.9`:

- `src/opnsense/mvc/app/models/OPNsense/IPsec/IPsec.xml` (`9fc546a5e7652f91df7b763685abca820fbb7ad7`)
- `src/opnsense/mvc/app/controllers/OPNsense/IPsec/forms/dialogKeyPair.xml` (`0456bcea3c255f7238d8751120d3dc5c910c3648`)

`KeyPairsController.php` differs only in `genKeyPairAction` switch syntax (`e33fad5702076cfb005b80758d7098f068a78035` to `5a79cb2462255981e18326a7ceb4cbfaf68b49fd`). `getItemAction($uuid)` is unchanged and calls `getBase("keyPair", "keyPairs.keyPair", $uuid)`. Identity is the UUID.

The model node is `name`, `keyType` (required, default `rsa`), `publicKey` (required stored PEM), optional `privateKey`, and derived `keySize` / `keyFingerprint`. The dialog says the private key stays with the peer when only a public key is stored. The existing resource refreshes only name, key type, size, and fingerprint because the API re-encodes PEM payloads. The lookup uses that same public allowlist and does not copy either PEM into state.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/ipsec/key_pairs/getItem/{uuid}` through the existing reader.
- [x] Publish an explicit public allowlist equal to the resource refresh contract.
- [x] Keep identity as the UUID and reject any other ID before a read.
- [x] Keep empty derived size and fingerprint as empty strings. Do not publish public or private key PEM payloads.
- [x] Fail closed on a missing name or key type and on an absent UUID. Do not invent a secret or a key type.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, changelog, and coverage inventory.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- `ipsec_psk` remains a separate resource-only gap.
