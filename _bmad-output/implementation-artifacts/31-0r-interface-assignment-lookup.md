# 31.0r Interface assignment lookup

Status: review

## Outcome

Add a read-only `opnsense_interface_assignment` lookup for an existing base assignment. Do not duplicate trust certificate work owned by 31.0f. Do not implement a second assignment owner, and do not replace the live interface data sources.

## Source contract

Latest stable tag is `26.7.9` (`stable release`), peeled commit `228ab68d54513046f79bc36d7f583936922d1951`. Previous stable tag `26.7.8` peels to `96978001a8343669154e4492172ce0a07dac7e4b`. Development tags `27.1.a` and `27.1.b` are alpha/beta and are not the contract.

These blobs are identical between `26.7.8` and `26.7.9`:

- `src/opnsense/mvc/app/controllers/OPNsense/Interfaces/Api/AssignmentController.php` (`61193fc27a8c0fa77a0c77b5c708bb371d92c1ba`), also identical at `26.7.7`
- `src/opnsense/mvc/app/models/OPNsense/Interfaces/NetworkInterface.xml` (`85e25100479392d5d5af13b12eb72f5aa25d3290`)

`getItemAction($ifname)` calls mutable-model `getBase("interface", "interface", $ifname)`. Identity is the assignment name, not a UUID. Core ACL pattern `api/interfaces/assignment/*` covers the GET. The released model has no password or write-only secret. Tunnel modes remain outside the resource allowlist and fail closed.

The `26.7.7` model differs only by blank-description text and a required default on `dhcp6-ia-pd-len`. That field is not lookup state. The lookup reuses the resource compatibility gate of OPNsense 26.7.6 or newer.

## Acceptance criteria

- [x] Reconcile registration, sibling lookups, open PRs, and story ownership before editing.
- [x] Read only `/api/interfaces/assignment/get_item/{name}` through the existing compatible reader.
- [x] Publish an explicit public allowlist equal to the resource contract.
- [x] Keep identity as `wan`, `lan`, or `optN`; reject any other ID before a read.
- [x] Keep empty addresses as empty strings and null prefix lengths. Do not publish device, description, MTU, hardware, or response-only fields.
- [x] Fail closed on missing required fields, unknown assignments, and unsupported tunnel modes. Do not invent safe mode or secret values.
- [x] Cover schema allowlist, model read, Core read, wrong ID, missing field, unsupported mode, and not-found with offline fixtures.
- [x] Register the data source, example, template, generated page, changelog, and coverage inventory.
- [ ] Live appliance proof is not part of this draft.
- [ ] Merge and release remain separate human decisions.

## Remaining limits

- No appliance mutation and no live read were performed.
- OpenTofu execution is not claimed.
- Live interface diagnostics remain separate data sources.
- PPPoE and other tunnel modes remain unsupported; a lookup of those assignments fails closed.
- `trust_cert` remains owned by 31.0f and is not part of this change.
