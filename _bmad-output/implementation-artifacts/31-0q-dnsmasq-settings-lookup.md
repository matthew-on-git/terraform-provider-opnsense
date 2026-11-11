# 31.0q Dnsmasq settings lookup

Status: review

## Outcome

Add a read-only `opnsense_dnsmasq_settings` lookup for the existing singleton. Do not duplicate trust certificate work owned by 31.0f. Do not implement a second settings owner.

## Source contract

Released tags `26.7.8` and `26.7.7` are annotated `stable release`. Peeled commits are `96978001a8343669154e4492172ce0a07dac7e4b` and `f1a83f72b144469fd97839aba2419ad130d19043`.

These blobs are identical between those tags:

- `src/opnsense/mvc/app/controllers/OPNsense/Dnsmasq/Api/SettingsController.php`
- `src/opnsense/mvc/app/models/OPNsense/Dnsmasq/Dnsmasq.xml` (`34820d3dc4a7d95b7897405084d177cdab92b377`)
- `src/opnsense/mvc/app/models/OPNsense/Dnsmasq/ACL/ACL.xml` (`3cc2cfdcf0e4d5d028f9b5a678745811fcc1f970`)

`getAction` calls the mutable-model GET, then injects `dhcp['%domain']` from the system domain. That value is response metadata, not stored Dnsmasq configuration. ACL pattern `api/dnsmasq/*` covers the settings GET. The general and DHCP models contain no password or write-only secret. `dns_port` is volatile. `expand_hosts` is a stored boolean the current resource does not own.

## Acceptance criteria

- [x] Reconcile registration, sibling item lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/dnsmasq/settings/get` through the existing singleton client.
- [x] Publish an explicit public allowlist equal to the resource contract.
- [x] Keep identity `dnsmasq`; reject any other ID before a read.
- [x] Keep empty DHCP integers null. Do not publish `%domain`, `expand_hosts`, or child arrays.
- [x] Fail closed when nested DHCP is missing. Do not invent empty state.
- [x] Cover schema allowlist, model read, Core read, wrong ID, and missing DHCP with offline fixtures.
- [x] Register the data source, example, template, generated page, changelog, and coverage inventory.
- [ ] Live appliance proof is not part of this draft.
- [ ] `expand_hosts` remains unowned until a resource field story adopts it.
- [ ] Hosted security on unchanged Go 1.26.10 still fails GO-2026-6859. Do not absorb the separate repair.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- Child UUID lookups stay on their existing data sources.
- Merge, release, and the Go security baseline remain separate human decisions.

## Protocol-test decoder repair

Decode Terraform Number state into `big.Float`, the supported `tftypes.Value.As`
target, rather than `int64`. Keep the exact value128 assertion and existing
read-only/wrong-ID/metadata assertions intact. This repairs test decoding, not
provider schema or appliance behavior. Re-run required checks and bind hosted
results to the corrected head; no live appliance or release claim is added.

Validation of decoder repair: `make check` ran all six stages in DevRail1.12.24.
Lint, format, test, scan and docs passed. Security failed on the existing Go
standard-library advisory; direct `govulncheck ./...` output was retained and
confirms the finding. The separate toolchain repair remains PR87; it is not
silently merged or absorbed here. Hosted checks on this corrected head must be
read independently; the story remains in review. Test-only decoder correction
does not change provider behavior, so no additional release heading is created.
