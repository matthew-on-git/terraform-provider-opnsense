# 31.0w Kea DHCPv6 settings lookup

Status: review

## Outcome

Add a read-only `opnsense_kea_dhcpv6_settings` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not reimplement `kea_dhcpv4_settings`.

## Reconciliation

Fetched `main` is `31d0cd4`, the merge of PR 99 and published v0.23.0. Open PRs 90, 91, 92, 96, 97, and 98 do not implement this lookup and are not a base for this change. `trust_cert` remains owned by 31.0f. No existing `kea_dhcpv6_settings` data source is registered. The resource story that owns the singleton remains unchanged.

## Source contract

Latest stable tag used for this story is `26.7.9`. Its annotation is `stable release`. The peeled commit is `228ab68d54513046f79bc36d7f583936922d1951`. The previous stable tag is `26.7.8`, annotation `stable release`, peeled commit `96978001a8343669154e4492172ce0a07dac7e4b`.

These blobs are identical between those tags:

- `src/opnsense/mvc/app/controllers/OPNsense/Kea/Api/Dhcpv6Controller.php` (`962bb0c0c6c3af7a56ac5ae2bd67d51f16e9ac43`)
- `src/opnsense/mvc/app/controllers/OPNsense/Kea/forms/generalSettings6.xml` (`07924207f948f035fefc2e97d12094cfa760e213`)
- `src/opnsense/mvc/app/models/OPNsense/Kea/KeaDhcpv6.xml` (`fffb04afba064e285ccefe31c810208755932b85`)

`getAction` calls the mutable-model GET, then returns `general`, `lexpire`, and `ha`. Identity is the synthetic singleton `kea-dhcpv6`, not a UUID. The released general node has required `enabled` (`BooleanField`, default `0`), optional `manual_config` (`BooleanField`, no default), optional multiple `interfaces`, required `valid_lifetime` (`IntegerField`, default `4000`), and required `fwrules` (`BooleanField`, default `1`). It also returns socket retry, decline probation, and required multiple `mac_sources`. Those extra GET fields are not resource refresh state. The provider resource schema copies `dhcp_socket_type` from DHCPv4, but that field is absent from the stable general model and form. It is not lookup state and is not replaced with the generator default `raw`. The general node has no password or write-only secret. Subnet TSIG material is not part of this GET.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/kea/dhcpv6/get` through the existing singleton client.
- [x] Publish an explicit public allowlist equal to the resource refresh fields the stable general model returns.
- [x] Keep identity `kea-dhcpv6` and reject any other ID before a read.
- [x] Keep disabled as false. Keep a blank optional manual configuration null. Do not publish non-allowlisted response fields.
- [x] Fail closed when a required field is missing or not a documented value. Do not invent model defaults.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, blank optional, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, changelog, and coverage inventory.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- `dhcp_socket_type` remains on the resource schema but is not lookup state, because stable 26.7.8 and 26.7.9 do not return it.
- Fields outside the resource refresh contract remain unread even though the stable GET returns them.
- Merging the new `0.24.0` changelog heading authorizes the automatic publication path after successful main CI. That merge is not authorized here.
