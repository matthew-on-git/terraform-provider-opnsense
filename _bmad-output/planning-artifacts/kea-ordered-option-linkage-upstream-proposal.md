# Upstream proposal draft — ordered Kea option-link reads

Status: draft for provider tracking. Not submitted upstream.

## Problem

OPNsense core 26.7.8 stores generic DHCPv4 option links on subnet and
reservation `ModelRelationField` values as a comma-separated UUID
string. `KeaDhcpv4.php` walks that stored order when emitting
flex-option actions, and same-code actions are last-match-wins.

`GET /api/kea/dhcpv4/get_subnet/{uuid}` and
`GET /api/kea/dhcpv4/get_reservation/{uuid}` return `getNodes()`.
Those relations do not set `<Sorted>Y</Sorted>`, so
`ModelRelationField` natcasesorts choices by description before
building the selected map. Clients cannot recover stored order from
that map. JSON object key order is not a stable contract.

Paginated `search_subnet` / `search_reservation` rows include the raw
stored string via `getValue()`, but that grid is not a documented
parent-scoped ordered read. Configuration XML is not an acceptable
management API for this link.

The same unsorted relation shape exists on DHCPv6 subnet and
reservation in the same tag.

## Requested contract

On the next stable core release, please expose stored link order
through the existing parent GET actions:

1. Set `<Sorted>Y</Sorted>` on the subnet and reservation `option`
   relations so `getNodeData()` keeps saved UUID order. This uses the
   existing `ModelRelationField::setSorted()` behavior.
2. Also return an explicit ordered array, for example `option_uuids`,
   from `get_subnet` and `get_reservation`. Do not require clients to
   infer order from JSON object keys, labels, or description text.
3. Document that array order is significant because flex-option
   generation is last-match-wins for the same code, with subnet actions
   emitted before reservation actions.
4. Keep the write as an ordered UUID list or the existing
   comma-separated scalar. Rejecting a selected-map array in
   `setNodes()` can remain, as long as the ordered read and the write
   use the same sequence.
5. Apply the same ordered read to the DHCPv6 subnet and reservation
   relations if they keep this field type.

## Non-goals

- Do not require raw XML access.
- Do not change generic option CRUD, `option_data`, or service
  activation semantics as part of this read contract.
- This draft does not ask the provider to implement links before the
  released read exists.
