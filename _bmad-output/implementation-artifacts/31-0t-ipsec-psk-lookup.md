# 31.0t IPsec pre-shared key lookup

Status: review

## Outcome

Add a read-only `opnsense_ipsec_psk` lookup for an existing pre-shared key. Do not duplicate trust certificate work owned by 31.0f. Do not duplicate the key-pair lookup owned by 31.0s.

## Source contract

Latest stable tag used for this story is `26.7.9`. `PreSharedKeysController.php` `getItemAction($uuid)` calls `getBase("preSharedKey", "preSharedKeys.preSharedKey", $uuid)`. Identity is the UUID.

The model node is required `ident`, optional `remote_ident`, required `keyType` (`PSK` default, `EAP`), `Key`, and `description`. The existing resource refreshes identity, remote identity, key type, and description because the secret is write-only. The lookup uses that same public allowlist and does not copy `Key` into state.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/ipsec/pre_shared_keys/get_item/{uuid}` through the existing reader.
- [x] Publish an explicit public allowlist equal to the resource refresh contract.
- [x] Keep identity as the UUID and reject any other ID before a read.
- [x] Keep empty remote identity and description as empty strings. Do not publish the pre-shared key value.
- [x] Fail closed on a missing identity or key type and on an absent UUID. Do not invent a secret or a key type.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, changelog, and coverage inventory.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- `ipsec_key_pair` remains owned by 31.0s and is not part of this change.
