# 31.0v Kea DHCPv4 settings lookup

Status: review

## Outcome

Add a read-only `opnsense_kea_dhcpv4_settings` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not implement `kea_dhcpv6_settings`.

## Reconciliation

Fetched `main` is `44cbb79`, the merge of PR 95 and published v0.22.0. Open PRs 90, 91, 92, 96, 97, and 98 do not implement this lookup and are not a base for this change. `trust_cert` remains owned by 31.0f. No existing `kea_dhcpv4_settings` data source is registered. The resource story that owns the singleton remains unchanged.

## Source contract

Latest stable tag used for this story is `26.7.9`. Its annotation is `stable release`. The peeled commit is `228ab68d54513046f79bc36d7f583936922d1951`. The previous stable tag is `26.7.8`, annotation `stable release`, peeled commit `96978001a8343669154e4492172ce0a07dac7e4b`.

These blobs are identical between those tags:

- `src/opnsense/mvc/app/controllers/OPNsense/Kea/Api/Dhcpv4Controller.php` (`40fd71fd3e50fde3fd7abe4e524ca632e889e247`)
- `src/opnsense/mvc/app/controllers/OPNsense/Kea/forms/generalSettings4.xml` (`35db043da089bd545a92d904164f5994d51741d1`)

`KeaDhcpv4.xml` changed (`2c2825f53d0489704817f1765dda1822d487944f` to `6ff24f63ec8fe3dbd756c4a138b0d35b66f4f634`). The patch adds subnet match fields only. The general node is unchanged.

`getAction` calls the mutable-model GET, then returns `general`, `lexpire`, and `ha`. Identity is the synthetic singleton `kea-dhcpv4`, not a UUID. The public allowlist is the resource refresh contract: required `enabled` (`BooleanField`, default `0`), optional `manual_config` (`BooleanField`, no default), optional multiple `interfaces`, required `valid_lifetime` (`IntegerField`, default `4000`), required `fwrules` (`BooleanField`, default `1`), and required `dhcp_socket_type` (`OptionField`, `raw` or `udp`, default `raw`). Socket retry, decline probation, compatibility, lease expiry, and HA are returned by GET and are not lookup state. The released general model has no password or write-only secret.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/kea/dhcpv4/get` through the existing singleton client.
- [x] Publish an explicit public allowlist equal to the resource refresh contract.
- [x] Keep identity `kea-dhcpv4` and reject any other ID before a read.
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
- `kea_dhcpv6_settings` remains a separate lookup gap.
- Fields outside the resource refresh contract remain unread even though the stable GET returns them.
- Merging the new `0.23.0` changelog heading authorizes the automatic publication path after successful main CI. That merge is not authorized here.
