# 31.0u Kea control agent lookup

Status: review

## Outcome

Add a read-only `opnsense_kea_ctrl_agent` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not implement `kea_dhcpv4_settings` or `kea_dhcpv6_settings`.

## Reconciliation

Fetched `main` is `c4b728b`, the merge of PR 94 and published v0.21.0. Open PRs 90–92 are unrelated firewall and Unbound fixes and are not a base for this change. `trust_cert` remains owned by 31.0f. The resource story 21.5 is done. No existing `kea_ctrl_agent` data source is registered.

## Source contract

Latest stable tag used for this story is `26.7.9`. Its annotation is `stable release`. The peeled commit is `228ab68d54513046f79bc36d7f583936922d1951`. The previous stable tag is `26.7.8`, annotation `stable release`, peeled commit `96978001a8343669154e4492172ce0a07dac7e4b`.

These blobs are identical between those tags:

- `src/opnsense/mvc/app/controllers/OPNsense/Kea/Api/CtrlAgentController.php` (`02ee37254dccd6da334f86868ece77ba0b4179e6`)
- `src/opnsense/mvc/app/models/OPNsense/Kea/KeaCtrlAgent.xml` (`876d5e2c8c7a0e8da9fcf7b09378df2b3ac2225d`)

`getAction` calls the mutable-model GET, then returns only `ctrlagent.general`. The general node is required `enabled` (`BooleanField`, default `0`), required `http_host` (`NetworkField`, IPv4, default `127.0.0.1`), and required `http_port` (`PortField`, default `8000`). There is no password or write-only secret. Identity is the synthetic singleton `kea-ctrl-agent`, not a UUID.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/kea/ctrl_agent/get` through the existing singleton client.
- [x] Publish an explicit public allowlist equal to the resource refresh contract.
- [x] Keep identity `kea-ctrl-agent` and reject any other ID before a read.
- [x] Keep disabled as false. Do not publish non-allowlisted response fields.
- [x] Fail closed when enabled is not `0` or `1`, or when the listen address or port is blank. Do not invent model defaults.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, changelog, and coverage inventory.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
- `kea_dhcpv4_settings` and `kea_dhcpv6_settings` remain separate lookup gaps.
- Merging the new `0.22.0` changelog heading authorizes the automatic publication path after successful main CI. That merge is not authorized here.
