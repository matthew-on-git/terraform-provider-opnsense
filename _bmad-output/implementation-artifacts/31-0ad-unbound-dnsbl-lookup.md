# 31.0ad Unbound DNSBL lookup

Status: review

Priority: P1

## Outcome

Add a read-only `opnsense_unbound_dnsbl` lookup for the existing compatibility singleton. Do not duplicate trust certificate work owned by 31.0f. Do not reimplement sibling lookups already shipped on main, including `opnsense_unbound_settings`. Do not modify release automation.

## Reconciliation

Fetched `main` is `b3f4914dfef37ca0662a9de02fad808684e97cdf`, the merge of pull request 111 and tag `v0.30.0`. No open pull request implements `unbound_dnsbl`. `trust_cert` remains owned by 31.0f. No existing DNSBL data source is registered. The settings lookup already exposes every array-backed item and is not this change. The resource story that owns the singleton remains 22.4 and is unchanged.

## Source contract

Latest stable core tag used for this story is `26.7.9`. Its annotation is `stable release`. The peeled commit is `228ab68d54513046f79bc36d7f583936922d1951`. The previous stable tag is `26.7.8`, annotation `stable release`, peeled commit `96978001a8343669154e4492172ce0a07dac7e4b`. The letter-suffixed development tag is not the contract.

These blobs are identical between those tags:

- `src/opnsense/mvc/app/models/OPNsense/Unbound/Unbound.xml` (`e81b0746f78b6bb0d290b15fd42f46a7d193a988`)
- `src/opnsense/mvc/app/models/OPNsense/Unbound/Unbound.php` (`0877fff1011873b1eb6e92f9b71f33a5b7ad5101`)
- `src/opnsense/mvc/app/controllers/OPNsense/Unbound/Api/SettingsController.php` (`cd6b1e537147c09b6b1d711923674c261a3174ae`)
- `src/opnsense/mvc/app/controllers/OPNsense/Unbound/forms/dnsbl.xml` (`6bcc1d886b0c86d0257f4dd40d098c95d09f2116`)
- `src/opnsense/mvc/app/controllers/OPNsense/Base/ApiMutableModelControllerBase.php` (`c67953d211b5a3ac6975e51e90b337dc9410f1fd`)
- `src/opnsense/mvc/app/models/OPNsense/Base/FieldTypes/BooleanField.php` (`6e78238a75e99e58f17d3fd37603c587c0e838ed`)

`getAction` returns `getNodes()` for the whole `unbound` model. `getDnsblAction` calls `getBase('blocklist', 'dnsbl.blocklist', $uuid)`. A call without a UUID runs `Add()` and returns a new in-memory node. That is not a stored read. Identity is the synthetic singleton `unbound-dnsbl`, not a UUID. The stable read of that identity is the existing search reader plus `get_dnsbl` of the sole returned UUID, and the settings GET for `general.safesearch`.

The blocklist item requires `enabled`. It is a BooleanField with default `1`. A missing or blank required field fails closed and is not replaced with that default. A disabled flag stays false. A token other than `0` or `1` fails closed. The token is trimmed once. `nxdomain` and `general.safesearch` are BooleanFields with no Required and no Default. A blank or omitted value stays null. `address` is an optional IPv4 NetworkField with masks forbidden. An empty value stays empty and is not replaced with `0.0.0.0`. A non-empty value must be four decimal octets. `type` is a multiple OptionField. Unselected keys are not state, and a selected key outside the released option list fails closed. `source_nets`, `cache_ttl`, and `description` are returned by the item GET. `cache_ttl` and `description` are required in the model, but the resource never refreshes them, so they are not lookup state. The node has no password or write-only secret.

Zero search rows are not found. The pinned summary is `DNSBL compatibility item not found`. A search body missing `rows` is malformed, not not-found. More than one item fails closed. The lookup does not use the legacy `unbound.dnsbl` reader. That reader is not the stable array contract. A missing search endpoint fails closed with the helper's plugin error and does not probe for a legacy appliance.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read the sole stable blocklist item through the existing search and `get_dnsbl` client, plus `general.safesearch` from the settings GET.
- [x] Publish an explicit public allowlist equal to the resource refresh fields the stable item and general node return.
- [x] Keep identity `unbound-dnsbl` and reject any other ID before a read.
- [x] Keep disabled as false. Do not publish `source_nets`, `cache_ttl`, `description`, or other non-allowlisted response fields.
- [x] Fail closed when enabled is missing, blank, or not `0`/`1`. Do not invent the model default.
- [x] Keep blank optional booleans null. Keep an empty address empty.
- [x] Cover schema allowlist, model read, Core read, wrong ID, zero items, many items, missing field, blank required, disabled flag, padded token, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, and coverage inventory. Do not add a versioned changelog heading.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

Given a registered DNSBL resource and the stable blocklist GET, when the lookup reads identity `unbound-dnsbl` and exactly one item exists, then it publishes only the public allowlist and does not write.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- `opnsense_unbound_settings` remains the UUID-backed view of every blocklist item.
- `source_nets`, `cache_ttl`, and `description` remain unread because the resource never refreshes them.
- A missing required enabled flag is an error. That is not the resource schema default.
- Zero search rows are the pinned not-found. A search body missing `rows` is malformed, not not-found.
- A whitespace-padded empty item object is malformed, not not-found, because the item decoder requires the field keys.
- A compact empty settings object is the settings reader's `resource not found`, not the DNSBL zero-row not-found.
- More than one blocklist item fails closed. This lookup has no prior ownership.
- The legacy singleton wire is outside this source-backed lookup.
- The enabled, nxdomain, and safesearch tokens are trimmed once.
- Current main uses Release Please. This feature merge does not select a version or publish. A later maintainer-reviewed release PR owns the changelog and authorizes publication after its own main CI. That release merge is not authorized here.
