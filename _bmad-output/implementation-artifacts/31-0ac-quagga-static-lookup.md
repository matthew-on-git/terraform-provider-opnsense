# 31.0ac Quagga static lookup

Status: review

Priority: P1

## Outcome

Add a read-only `opnsense_quagga_static` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not reimplement sibling lookups already shipped on main, including `opnsense_quagga_static_route`. Do not modify release automation.

## Reconciliation

Fetched `main` is `fbb93a78e58c299527ea9320131772a758a05e11`, the merge of pull request 107. No open pull request implements `quagga_static`. `trust_cert` remains owned by 31.0f. No existing static general data source is registered. The route lookup is already shipped and is not this change. The resource story that owns the singleton remains 19.7 and is unchanged.

## Source contract

Latest stable plugins tag used for this story is `26.7.9`. Its annotation is `stable release`. The peeled commit is `a01a2e8cef738b3762830a7bbfc6e857384328f1`. The previous stable tag is `26.7.8`, annotation `stable release`, peeled commit `2fcfc29295e298827bda4236fd6574248e428968`. The letter-suffixed development tag is not the contract.

These blobs are identical between those tags:

- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/STATICd.xml` (`bbe622f85a3ab67abf4e64f9bc8415786c65f893`)
- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/STATICd.php` (`5d9213b7382565c318c48db2be818350eb87b27f`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/Api/StaticController.php` (`9a616cdb5f8785091c8775f0f550d15ae9daabea`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/forms/static.xml` (`b5d4690ccc3a6ca42939f817ccd0e36251878e7e`)

Core `26.7.9` is also a stable release, peeled commit `228ab68d54513046f79bc36d7f583936922d1951`. Previous stable core `26.7.8` peels to `96978001a8343669154e4492172ce0a07dac7e4b`. `ApiMutableModelControllerBase.php` is blob `c67953d211b5a3ac6975e51e90b337dc9410f1fd` on both tags. `BooleanField.php` is blob `6e78238a75e99e58f17d3fd37603c587c0e838ed` on both tags. `getAction` returns `getModel()->getNodes()` under the internal model name `staticd` on GET. The plugin controller does not override that action. Identity is the synthetic singleton `static`, not a UUID. The endpoint is `/api/quagga/static/get`.

The staticd node requires `enabled`. It is a BooleanField with default `0`. A missing or blank required field fails closed and is not replaced with that default. A disabled flag stays false. A token other than `0` or `1` fails closed. The token is trimmed once. `routes` is a child array returned by the whole-model GET. The resource never refreshes it, so it is not lookup state. Individual routes remain owned by `opnsense_quagga_static_route`. The node has no password or write-only secret.

The lookup uses the resource's existing `GetSingleton` reader and `staticd` monad. That helper classifies a compact empty object as not found. It does not collapse internal whitespace, so a whitespace-padded empty object unmarshals and fails closed as a missing required field. That is not not-found.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/quagga/static/get` through the existing singleton client.
- [x] Publish an explicit public allowlist equal to the resource refresh fields the stable staticd node returns.
- [x] Keep identity `static` and reject any other ID before a read.
- [x] Keep disabled as false. Do not publish the routes child array or other non-allowlisted response fields.
- [x] Fail closed when enabled is missing, blank, or not `0`/`1`. Do not invent the model default.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, blank required, disabled flag, padded token, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, and coverage inventory. Do not add a versioned changelog heading.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

Given a registered static routing resource and the stable staticd GET, when the lookup reads identity `static`, then it publishes only the public allowlist and does not write.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- The routes child array remains unread because the resource never refreshes it.
- A missing required field is an error. That is not the resource schema default.
- Compact `{}` is the pinned not-found. A whitespace-padded empty object is malformed, not not-found, because the resource reader does not collapse internal space.
- A null scalar fails closed as malformed, not not-found.
- The enabled token is trimmed once.
- Current main uses Release Please. This feature merge does not select a version or publish. A later maintainer-reviewed release PR owns the changelog and authorizes publication after its own main CI. That release merge is not authorized here.
