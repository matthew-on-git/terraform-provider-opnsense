# 31.0aa Quagga OSPF general lookup

Status: review

Priority: P1

## Outcome

Add a read-only `opnsense_quagga_ospf_general` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not reimplement sibling lookups already shipped on main.

## Reconciliation

Fetched `main` is `084b1745`, the merge of PR 105 and published v0.27.0. No open pull request implements `quagga_ospf_general`. `trust_cert` remains owned by 31.0f. No existing OSPF general data source is registered. The resource story that owns the singleton remains 19.4 and is unchanged.

## Source contract

Latest stable plugins tag used for this story is `26.7.9`. Its annotation is `stable release`. The peeled commit is `a01a2e8cef738b3762830a7bbfc6e857384328f1`. The previous stable tag is `26.7.8`, annotation `stable release`, peeled commit `2fcfc29295e298827bda4236fd6574248e428968`. The letter-suffixed `27.1.d` tag is development and is not the contract.

These blobs are identical between those tags:

- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/OSPF.xml` (`2bf4e9e41604081c97b9d7462aadd962726a8738`)
- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/OSPF.php` (`f370c4b651fe84c1444e8936d2f86e702c586124`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/Api/OspfsettingsController.php` (`898a53bb6b8d0b4186f47098c7e3f43be371932a`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/Api/QuaggaControllerBase.php` (`50d80f40eab2a9a0d5898df9bc5072e0de46ab33`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/forms/ospf.xml` (`f55610b145a57e1dc649fd47b1f95ced0cd4c0fe`)

Core `26.7.9` is also a stable release, peeled commit `228ab68d54513046f79bc36d7f583936922d1951`. Previous stable core `26.7.8` peels to `96978001a8343669154e4492172ce0a07dac7e4b`. `ApiMutableModelControllerBase.php` is blob `c67953d211b5a3ac6975e51e90b337dc9410f1fd` on both tags. `getAction` returns `getModel()->getNodes()` under the internal model name `ospf` on GET. Neither the plugin controller nor `QuaggaControllerBase` overrides that action. Identity is the synthetic singleton `ospf`, not a UUID. The endpoint is `/api/quagga/ospfsettings/get`.

The OSPF node requires `enabled`, `carp_demote`, `logadjacencychanges`, `originate`, and `originatealways`. Each is a BooleanField with default `0`. A missing or blank required field fails closed and is not replaced with that default. A disabled flag stays false. `routerid` is an optional IPv4 NetworkField with no mask and no default. A blank or omitted router ID stays null. A present value that is not an IPv4 identifier without a mask fails closed. `costreference` is an optional IntegerField with minimum 1, maximum 4294967, and no default. A blank or omitted cost stays null. Literal 0 is not the resource unset sentinel and fails closed. `originatemetric` is an optional IntegerField with minimum 0, maximum 16777214, and no default. A blank or omitted metric stays null. Literal 0 is a real value. Child arrays `areas`, `networks`, `neighbors`, `interfaces`, `prefixlists`, `routemaps`, and `redistributions` are returned by the whole-model GET and are not resource refresh state, so they are not lookup state. The resource schema includes `passive_interfaces`, but the released OSPF node has no such field, so it is not lookup state. The node has no password or write-only secret.

The lookup uses the resource's existing `GetSingleton` reader. That helper classifies a compact empty object as not found. It does not collapse internal whitespace, so a whitespace-padded empty object unmarshals and fails closed as a missing required field. That is not not-found.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/quagga/ospfsettings/get` through the existing singleton client.
- [x] Publish an explicit public allowlist equal to the resource refresh fields the stable OSPF node returns.
- [x] Keep identity `ospf` and reject any other ID before a read.
- [x] Keep disabled as false. Do not publish non-allowlisted response fields.
- [x] Fail closed when a required field is missing or not a documented value. Do not invent model defaults.
- [x] Keep a blank optional router ID, cost reference, or metric null. Fail closed on literal cost 0. Keep literal metric 0.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, blank required, disabled flags, both integer ranges, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, changelog, and coverage inventory.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

Given a registered OSPF general resource and the stable OSPF GET, when the lookup reads identity `ospf`, then it publishes only the public allowlist and does not write.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- Child arrays remain unread even though the stable whole-model GET returns them.
- `passive_interfaces` remains unread because the released OSPF node does not return it.
- A missing required field is an error. That is not the resource schema default.
- Compact `{}` is the pinned not-found. A whitespace-padded empty object is malformed, not not-found, because the resource reader does not collapse internal space.
- A null scalar fails closed as malformed, not not-found.
- Boolean tokens are trimmed once. Optional router ID, cost reference, and metric use that same trimmed token.
- Merging the new `0.28.0` changelog heading authorizes the automatic publication path after successful main CI. That merge is not authorized here.
