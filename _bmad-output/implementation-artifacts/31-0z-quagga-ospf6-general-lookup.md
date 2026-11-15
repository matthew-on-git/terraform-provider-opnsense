# 31.0z Quagga OSPFv3 general lookup

Status: review

## Outcome

Add a read-only `opnsense_quagga_ospf6_general` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not reimplement sibling lookups already shipped on main.

## Reconciliation

Fetched `main` is `b9ff88d`, the merge of PR 103 and published v0.26.0. No open pull request implements `quagga_ospf6_general`. `trust_cert` remains owned by 31.0f. No existing OSPFv3 general data source is registered. The resource story that owns the singleton remains 19.5 and is unchanged.

## Source contract

Latest stable plugins tag used for this story is `26.7.9`. Its annotation is `stable release`. The peeled commit is `a01a2e8cef738b3762830a7bbfc6e857384328f1`. The previous stable tag is `26.7.8`, annotation `stable release`, peeled commit `2fcfc29295e298827bda4236fd6574248e428968`. The letter-suffixed `27.1.d` tag is development and is not the contract.

These blobs are identical between those tags:

- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/OSPF6.xml` (`7493656b34a0ad7227ed30ff3f3ca68d1216caa5`)
- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/OSPF6.php` (`96742aaf4569e1e7f95e6f797a3002720f3c114a`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/Api/Ospf6settingsController.php` (`96c02afb54137de3163caa652ec48378d8c964f1`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/Api/QuaggaControllerBase.php` (`50d80f40eab2a9a0d5898df9bc5072e0de46ab33`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/forms/ospf6.xml` (`8b4b014559adf36932414e4b3501145acd7e84d2`)

Core `26.7.9` is also a stable release, peeled commit `228ab68d54513046f79bc36d7f583936922d1951`. Previous stable core `26.7.8` peels to `96978001a8343669154e4492172ce0a07dac7e4b`. `ApiMutableModelControllerBase.php` is blob `c67953d211b5a3ac6975e51e90b337dc9410f1fd` on both tags. `getAction` returns `getModel()->getNodes()` under the internal model name `ospf6` on GET. Neither the plugin controller nor `QuaggaControllerBase` overrides that action. Identity is the synthetic singleton `ospf6`, not a UUID. The endpoint is `/api/quagga/ospf6settings/get`.

The OSPFv3 node requires `enabled`, `carp_demote`, `originate`, and `originatealways`. Each is a BooleanField with default `0`. A missing or blank required field fails closed and is not replaced with that default. A disabled flag stays false. `routerid` is an optional IPv4 NetworkField with no mask and no default. A blank or omitted router ID stays null. A present value that is not an IPv4 identifier without a mask fails closed. `originatemetric` is an optional IntegerField with minimum 0, maximum 16777214, and no default. A blank or omitted metric stays null. Literal 0 is a real value, not the resource sentinel. Child arrays `areas`, `interfaces`, `prefixlists`, `routemaps`, and `redistributions` are returned by the whole-model GET and are not resource refresh state, so they are not lookup state. The node has no password or write-only secret.

The lookup uses the resource's existing `GetSingleton` reader. That helper classifies a compact empty object as not found. It does not collapse internal whitespace, so a whitespace-padded empty object unmarshals and fails closed as a missing required field. That is not not-found.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/quagga/ospf6settings/get` through the existing singleton client.
- [x] Publish an explicit public allowlist equal to the resource refresh fields the stable OSPFv3 node returns.
- [x] Keep identity `ospf6` and reject any other ID before a read.
- [x] Keep disabled as false. Do not publish non-allowlisted response fields.
- [x] Fail closed when a required field is missing or not a documented value. Do not invent model defaults.
- [x] Keep a blank optional router ID or metric null. Keep literal metric 0.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, blank required, disabled flags, zero metric, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, changelog, and coverage inventory.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- Child arrays remain unread even though the stable whole-model GET returns them.
- A missing required field is an error. That is not the resource schema default.
- Compact `{}` is the pinned not-found. A whitespace-padded empty object is malformed, not not-found, because the resource reader does not collapse internal space.
- A null scalar fails closed as malformed, not not-found.
- Boolean tokens are trimmed once. Optional router ID and metric use that same trimmed token.
- Merging the new `0.27.0` changelog heading authorizes the automatic publication path after successful main CI. That merge is not authorized here.
