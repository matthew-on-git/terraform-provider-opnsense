# 31.0y Quagga general lookup

Status: review

## Outcome

Add a read-only `opnsense_quagga_general` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not reimplement sibling lookups already shipped on main.

## Reconciliation

Fetched `main` is `f995a24`, the merge of PR 102 and published v0.25.0. Open PRs 90, 91, 92, 96, 97, and 98 do not implement this lookup and are not a base for this change. `trust_cert` remains owned by 31.0f. No existing `quagga_general` data source is registered. The resource story that owns the singleton remains 19.1 and is unchanged.

## Source contract

Latest stable plugins tag used for this story is `26.7.9`. Its annotation is `stable release`. The peeled commit is `a01a2e8cef738b3762830a7bbfc6e857384328f1`. The previous stable tag is `26.7.8`, annotation `stable release`, peeled commit `2fcfc29295e298827bda4236fd6574248e428968`. The letter-suffixed `27.1.d` tag is development and is not the contract.

These blobs are identical between those tags:

- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/General.xml` (`32d7e35c3f92f722ced17d318a5f0bc06b75f3ce`)
- `net/frr/src/opnsense/mvc/app/models/OPNsense/Quagga/General.php` (`c4cfafd66ba8f911475006bf1dbd2fd6117cb44c`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/Api/GeneralController.php` (`d8e615d6a063879ae01ad8fabfe1899ff3d54cef`)
- `net/frr/src/opnsense/mvc/app/controllers/OPNsense/Quagga/forms/general.xml` (`c1b22a8bc4adbeb6cb9bf6841efd6ea283b34a63`)

Core `26.7.9` is also a stable release, peeled commit `228ab68d54513046f79bc36d7f583936922d1951`. `ApiMutableModelControllerBase.php` is blob `c67953d211b5a3ac6975e51e90b337dc9410f1fd` on both `26.7.9` and `26.7.8`. `getAction` returns `getModel()->getNodes()` under the internal model name `general` on GET. The plugin controller does not override that action. Identity is the synthetic singleton `general`, not a UUID. The endpoint is `/api/quagga/general/get`.

The general node requires `enabled`, `profile`, `enablecarp`, `enablesyslog`, `enablesnmp`, `sysloglevel`, and `fwrules`. Each has a model default. A missing or blank required field fails closed and is not replaced with that default. A disabled flag stays false, including syslog and firewall rules whose model default is enabled. `profile` is `traditional` or `datacenter`. `sysloglevel` is one of the eight released option values. Volatile `daemons` and required `manual_config` are returned by the whole-model GET and are not resource refresh state, so they are not lookup state. The general node has no password or write-only secret.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/quagga/general/get` through the existing singleton client.
- [x] Publish an explicit public allowlist equal to the resource refresh fields the stable general node returns.
- [x] Keep identity `general` and reject any other ID before a read.
- [x] Keep disabled as false. Do not publish non-allowlisted response fields.
- [x] Fail closed when a required field is missing or not a documented value. Do not invent model defaults.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, blank required, disabled flags, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, changelog, and coverage inventory.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- Volatile daemons and manual configuration remain unread even though the stable whole-model GET returns them.
- A missing required field is an error. That is not the resource schema default.
- Compact `{}` is the pinned not-found. A whitespace-padded empty object is malformed, not not-found.
- A padded option scalar fails closed through the existing selector decoder. Boolean tokens are trimmed once.
- Merging the new `0.26.0` changelog heading authorizes the automatic publication path after successful main CI. That merge is not authorized here.
