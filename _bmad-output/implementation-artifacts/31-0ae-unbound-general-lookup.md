# 31.0ae Unbound general lookup

Status: review

Priority: P1

## Outcome

Add a read-only `opnsense_unbound_general` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not reimplement sibling lookups already shipped on main, including `opnsense_unbound_settings` and `opnsense_unbound_dnsbl`. Do not modify release automation.

## Reconciliation

Fetched `main` is `929c60cc732dd807cdd08cc853854639de0ca051`, the merge of pull request 119. No open pull request implements `unbound_general`. The resource exists. `trust_cert` remains owned by 31.0f. The settings lookup remains the aggregate view and is not this change. The resource story that owns the singleton remains 22.1 and is unchanged.

## Source contract

Latest stable core tag used for this story is `26.7.10`. Its annotation is `stable release`. The peeled commit is `4092e306ab6cfa9748875053fbb41c4c7cd5af2d`. The previous stable tag is `26.7.9`, annotation `stable release`, peeled commit `228ab68d54513046f79bc36d7f583936922d1951`. The letter-suffixed development tag is not the contract.

`Unbound.xml` is identical between those tags (`e81b0746f78b6bb0d290b15fd42f46a7d193a988`). `SettingsController.php` and `ApiMutableModelControllerBase.php` are identical. `Unbound.php` differs only by a comparison-operator edit in DNSBL `source_nets` generation. That sibling edit does not change the general, advanced, ACL-default, or forwarding GET.

`getAction` returns `getNodes()` for the whole `unbound` model. Identity is the synthetic singleton `unbound-general`, not a UUID. The stable read is the existing settings GET. It does not call the compatibility probe or the legacy flatten path.

Required fields with model defaults stay required. A missing or blank value fails closed and is not replaced with that default. Optional booleans and numeric fields with no Required and no Default stay null when blank. Literal zero is kept. The allowlist excludes `safesearch`, `belownxdomain`, ACL items, and sibling nodes. The nodes have no password or write-only secret.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read the singleton through the existing settings GET.
- [x] Publish an explicit public allowlist equal to the resource refresh fields.
- [x] Keep identity `unbound-general` and reject any other ID before a read.
- [x] Fail closed when a required field is missing or blank. Do not copy model defaults.
- [x] Keep blank optional booleans and integers null. Keep an empty optional string empty.
- [x] Cover schema allowlist, model read, framework read, wrong ID, empty object, malformed object, missing field, disabled flag, padded token, and plugin-not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, and coverage inventory. Do not add a versioned changelog heading.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

Given a registered general resource and the stable settings GET, when the lookup reads identity `unbound-general`, then it publishes only the public allowlist and does not write.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- `opnsense_unbound_settings` remains the aggregate view.
- `opnsense_unbound_dnsbl` remains the owner of `safesearch`.
- `belownxdomain` remains unread because the resource never refreshes it.
- A missing required field is an error. That is not the resource schema default.
- A compact empty settings object, including one whose only difference is whitespace, is the settings reader's `resource not found` because the decoder trims and unmarshals it as an empty object.
- Invalid JSON is malformed, not not-found.
- Current main uses Release Please. This feature merge does not select a version or publish. A later maintainer-reviewed release PR owns the changelog and authorizes publication after its own main CI. That release merge is not authorized here.
