# 31.0ab Quagga RIP lookup

Status: review

Priority: P1

## Outcome

Add a read-only `opnsense_quagga_rip` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not reimplement sibling lookups already shipped on main. Do not modify open os-frr 2.1 compatibility or release-automation pull requests.

## Reconciliation

Fetched `main` is `a5030bf67c7959191eb19b456da9062b1a4fdc03`, the merge of PR 106 and published v0.28.0. No open pull request implements `quagga_rip`. Open PR 107 is os-frr 2.1 compatibility, not a RIP lookup. Open PR 108 is release automation. `trust_cert` remains owned by 31.0f. No existing RIP data source is registered. The resource story that owns the singleton remains 19.6 and is unchanged.

## Source contract

Latest stable plugins tag used for this story is `26.7.9`. Its annotation is `stable release`. The peeled commit is `a01a2e8cef738b3762830a7bbfc6e857384328f1`. The previous stable tag is `26.7.8`, annotation `stable release`, peeled commit `2fcfc29295e298827bda4236fd6574248e428968`. The letter-suffixed development tag is not the contract.

These blobs are identical between those tags:

- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/RIP.xml` (`054564ec72574fc562569c4214bfbff268c6f6e5`)
- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/RIP.php` (`7a9838a28bd85ddde617a6ef4eef297f6e190e0a`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/Api/RipController.php` (`c738188bc9b0ba8ec7e2f1c564ab201259484bbb`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/Api/QuaggaControllerBase.php` (`50d80f40eab2a9a0d5898df9bc5072e0de46ab33`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/forms/rip.xml` (`640f1b3ef8b84220b92abe91d652c88ecabe1344`)

Core `26.7.9` is also a stable release, peeled commit `228ab68d54513046f79bc36d7f583936922d1951`. Previous stable core `26.7.8` peels to `96978001a8343669154e4492172ce0a07dac7e4b`. `ApiMutableModelControllerBase.php` is blob `c67953d211b5a3ac6975e51e90b337dc9410f1fd` on both tags. `getAction` returns `getModel()->getNodes()` under the internal model name `rip` on GET. Neither the plugin controller nor `QuaggaControllerBase` overrides that action. Identity is the synthetic singleton `rip`, not a UUID. The endpoint is `/api/quagga/rip/get`.

`CSVListField.php`, `BooleanField.php`, `IntegerField.php`, `BaseListField.php`, `BaseField.php`, and `BaseSetField.php` are identical between those core tags. `OptionField.php` differs (`22c2b343e498be862abecdcc3ce0760225c3fcff` versus `77c6a6cb6ea783cb1105ac46e1256d70be3ed5c2`). The patch only removes a static option-list cache from `setOptionValues`. `getNodeData` remains on the unchanged `BaseListField`, so the redistribute GET contract is unchanged.

The RIP node requires `enabled` and `version`. `enabled` is a BooleanField with default `0`. `version` is an IntegerField with minimum 1, maximum 2, and default `2`. A missing or blank required field fails closed and is not replaced with that default. A disabled flag stays false. Literal version 0 fails closed. `networks` is an optional CSVListField with no default. The stable GET returns a selected map of stored CIDR keys, not stored order. An empty stored list is the empty-key selected map and stays an empty set. A selected key that is not a four-octet IPv4 prefix from 0 through 32 fails closed. Leading zeros fail closed. `redistribute` is an optional multiple OptionField with no default. Selected keys must be `bgp`, `connected`, `kernel`, `ospf`, or `static`. An unknown or padded selected token fails closed. Unselected options stay out of state. `defaultmetric` is an optional IntegerField with minimum 1, maximum 16, and no default. A blank or omitted metric stays null. Literal 0 is not the resource unset sentinel and fails closed. `passiveinterfaces` is returned by the whole-model GET and is not resource refresh state, so it is not lookup state. The node has no password or write-only secret.

The lookup uses the resource's existing `GetSingleton` reader. That helper classifies a compact empty object as not found. It does not collapse internal whitespace, so a whitespace-padded empty object unmarshals and fails closed as a missing required field. That is not not-found.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/quagga/rip/get` through the existing singleton client.
- [x] Publish an explicit public allowlist equal to the resource refresh fields the stable RIP node returns.
- [x] Keep identity `rip` and reject any other ID before a read.
- [x] Keep disabled as false. Do not publish non-allowlisted response fields.
- [x] Fail closed when a required field is missing or not a documented value. Do not invent model defaults.
- [x] Keep a blank optional metric null. Fail closed on literal version or metric 0. Keep version 1.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, blank required, disabled flags, both integer ranges, network shape, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, and coverage inventory. Do not add a versioned changelog heading.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

Given a registered RIP resource and the stable RIP GET, when the lookup reads identity `rip`, then it publishes only the public allowlist and does not write.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- `passiveinterfaces` remains unread because the resource never refreshes it.
- A missing required field is an error. That is not the resource schema default.
- Compact `{}` is the pinned not-found. A whitespace-padded empty object is malformed, not not-found, because the resource reader does not collapse internal space.
- A null scalar fails closed as malformed, not not-found.
- Boolean and integer tokens are trimmed once. Network and redistribute keys are not trimmed; a padded key fails closed.
- Selected-map key order is not stored order. The lookup publishes a set.
- Current main uses Release Please. This feature merge does not select a version or publish. A later maintainer-reviewed release PR owns the changelog and authorizes publication after its own main CI. That release merge is not authorized here.
