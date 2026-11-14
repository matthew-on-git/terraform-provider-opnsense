# 31.0x Quagga BGP global lookup

Status: review

## Outcome

Add a read-only `opnsense_quagga_bgp_global` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not reimplement sibling lookups already shipped on main.

## Reconciliation

Fetched `main` is `40b775e`, the merge of PR 100 and published v0.24.0. Open PRs 90, 91, 92, 96, 97, and 98 do not implement this lookup and are not a base for this change. `trust_cert` remains owned by 31.0f. No existing `quagga_bgp_global` data source is registered. The resource story that owns the singleton remains unchanged.

## Source contract

Latest stable plugins tag used for this story is `26.7.9`. Its annotation is `stable release`. The peeled commit is `a01a2e8cef738b3762830a7bbfc6e857384328f1`. The previous stable tag is `26.7.8`, annotation `stable release`, peeled commit `2fcfc29295e298827bda4236fd6574248e428968`. The letter-suffixed `27.1.d` tag is development and is not the contract.

These blobs are identical between those tags:

- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/Api/BgpController.php` (`3a4dba797c06b9fb6397769a86035387031ce9c4`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/forms/bgp.xml` (`5be7fb2e2e405a8424c05f001ff722ed4aa5f940`)
- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/BGP.php` (`6af510a0c24e8ee30b0d41693ebbda1baaa339f9`)

`BGP.xml` changed (`4f762a2bf0c098902f71591258a1aae450334a48` to `967b6e6f128f2799e2d6ac2be5573a167be73243`). The patch adds a neighbor `localas` constraint and `localasoptions`. It does not touch the general node.

Core `26.7.9` is also a stable release, peeled commit `228ab68d54513046f79bc36d7f583936922d1951`. `ApiMutableModelControllerBase::getAction` returns `getModel()->getNodes()` under the internal model name `bgp` on GET. The plugin controller does not override that action. Identity is the synthetic singleton `bgp`, not a UUID. The endpoint is `/api/quagga/bgp/get`.

The general node has required `enabled`, `asnumber`, `graceful`, `networkimportcheck`, `enforce_first_as`, and `logneighborchanges`. `asnumber` has model default `65551`. Optional fields with no default are `routerid`, `distance`, `networks`, `maximumpaths`, and `maximumpathsibgp`. A present blank optional stays null. `bestpath` and child arrays are returned by the whole-model GET and are not resource refresh state, so they are not lookup state. The general node has no password or write-only secret. Neighbor password material is not part of this allowlist.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/quagga/bgp/get` through the existing singleton client.
- [x] Publish an explicit public allowlist equal to the resource refresh fields the stable general node returns.
- [x] Keep identity `bgp` and reject any other ID before a read.
- [x] Keep disabled as false. Keep a present blank optional null. Do not publish non-allowlisted response fields.
- [x] Fail closed when a required field is missing or not a documented value. Do not invent model defaults.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, blank optional, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, changelog, and coverage inventory.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- `bestpath` and child arrays remain unread even though the stable whole-model GET returns them.
- A present blank router ID, distance, or maximum-paths value stays null. That is not the resource schema default of an empty string or zero.
- Independent review passed with no security or logic findings. A whitespace-padded empty object is not a separate not-found case; compact `{}` is the pinned not-found. Zero is rejected for distance and both maximum-path fields by the same range check; the distance case is the dedicated test.
- Merging the new `0.25.0` changelog heading authorizes the automatic publication path after successful main CI. That merge is not authorized here.
