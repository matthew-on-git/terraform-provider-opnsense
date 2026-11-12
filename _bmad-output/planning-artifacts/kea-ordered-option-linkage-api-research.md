# Kea ordered parent-option linkage — released stable API research

## Verdict

Released stable OPNsense core 26.7.8 does not expose an authoritative
ordered read of generic-option links on DHCPv4 subnets or reservations.
Parent option links are not implemented. Raw configuration XML,
selected-map key or label order, and paginated grid-search CSV are not
a supported linkage contract.

This is a source-only gate. It does not prove appliance behavior, packet
delivery, or OpenTofu lifecycle.

## Reconciled baselines

- Provider `main` `b901ddd3543567075bc282573f0c62905bb28df2` (merge of
  PR 85). GitHub release `v0.17.0` was published at
  2026-11-11T01:37:38Z. No open pull requests existed before this
  research branch. Shipped generic-option and DDNS owners in v0.10.0
  are reused, not reimplemented. v0.17.0 is live interface data sources
  and does not add option linkage.
- Released stable core tag `26.7.8`, annotation `stable release`, commit
  `96978001a8343669154e4492172ce0a07dac7e4b`, committer date
  2026-10-29T08:42:16Z.
- Tag `27.1.a` is annotated `27.1-ALPHA`. It is not released stable and
  is not authority for this gate.
- Branch `stable/26.7` tip `5a45bc9c6e6713d1f74b87a573fa329eacb3f980`
  (2026-11-11T03:40:37Z) is after the tag. It is not a released stable
  source. Its Kea model/generator delta adds shared-network generation.
  It does not add `<Sorted>Y</Sorted>`, a new controller action, or an
  ordered option field. The two `option` relation blocks match the tag.

## Source identity

Kea DHCPv4 model, generator, and controller blobs are identical on tags
26.7.7 and 26.7.8:

| Path | Git blob |
|---|---|
| `src/opnsense/mvc/app/models/OPNsense/Kea/KeaDhcpv4.xml` | `2c2825f53d0489704817f1765dda1822d487944f` |
| `src/opnsense/mvc/app/models/OPNsense/Kea/KeaDhcpv4.php` | `874914885ae5925f31a8f3d488b88ed1d2d6e860` |
| `src/opnsense/mvc/app/controllers/OPNsense/Kea/Api/Dhcpv4Controller.php` | `40fd71fd3e50fde3fd7abe4e524ca632e889e247` |
| `src/opnsense/mvc/app/controllers/OPNsense/Kea/Api/ServiceController.php` | `06984a9ef7a05f9630dfe44288ce0cec86d50d70` |

`ModelRelationField.php` changed between those tags
(`642aca7631b6d68c323715b02c2a72a673c58e90` to
`e076ceb2149c5def736cc33c180233fd4cf47036`). The 26.7.8 diff only
null-guards `mdlStructure` when `setModel()` was not executed. Sort
behavior is unchanged. `stable/26.7` has that same 26.7.8 blob.

These serialization files have the same blob on 26.7.7, 26.7.8, and
`stable/26.7`:

| Path | Git blob |
|---|---|
| `ApiMutableModelControllerBase.php` | `c67953d211b5a3ac6975e51e90b337dc9410f1fd` |
| `BaseListField.php` | `5d42e552243f3304b0ca2905cb94f2060be6b918` |
| `ArrayField.php` | `80e515894dca511ca809d5c462aa34612c3d3df1` |

SHA256 of the 26.7.8 file bytes used for this reading:

| File | SHA256 |
|---|---|
| `KeaDhcpv4.xml` | `ca3fc35c0338418148efd1a5b40f436e64288c75600d83b590df1e00bd73942c` |
| `KeaDhcpv4.php` | `54e3ccf1339fb567fcb6dce2374ebaad0e526455ad9e692e2c0b3066872342e8` |
| `Dhcpv4Controller.php` | `9fa1ca628fadcba5b844e536a9ed082da59bb2fa1fbfffdace5610ae73614669` |
| `ModelRelationField.php` | `55b779dca03c4f86ac6dd412c6c766d3a5e585c1a672483cccdd9656bd6c2d2f` |
| `BaseListField.php` | `681f43624d7d9ac7210cc9a5738544f8403e05b94ca94ebf033dcff7ea941f24` |
| `ApiMutableModelControllerBase.php` | `f9d39ec4987b83d17edb2946b8e247b83373a95c6d537c3ff531a9fdd3dd57f1` |
| `UIModelGrid.php` | `aaa054fd047d14a5c31d9d8859c174eb471ca81a1fd9efe1792162dea587ec69` |
| `BaseField.php` | `249370673cfca7e1351377cd237b1add3081f407d823307a39d9fbada848d62d` |

The controller SHA256 matches the earlier generic-option validation
record. That record did not establish ordered parent linkage.

## Model and generator

`KeaDhcpv4.xml` subnet and reservation nodes each have
`<option type="ModelRelationField">` with `Multiple=Y`, source
`OPNsense.Kea.KeaDhcpv4`, items `options.option`, display
`description`. Neither sets `Sorted`. `option_data` is a different
field and is already owned by the subnet and reservation resources.

`KeaDhcpv4.php` emits flex-option actions by walking
`$subnet->option->getValues()` and then
`$reservation->option->getValues()`. The generator comment states that
same-code supersede actions are order-sensitive and that the last match
wins, with subnet actions before reservation actions. `getValues()`
explodes the stored comma-separated internal value, so stored order
matters. That does not mean any public read returns that order.

## Endpoints on released 26.7.8

`Dhcpv4Controller.php` exposes no ordered-link action. Relevant routes:

| Route | Read shape | Ordered parent read? |
|---|---|---|
| `GET /api/kea/dhcpv4/get` | `general`, `lexpire`, and `ha` only | No |
| `GET /api/kea/dhcpv4/get_subnet/{uuid}` | `getBase` → `getNodes()` | No |
| `GET /api/kea/dhcpv4/get_reservation/{uuid}` | `getBase` → `getNodes()` | No |
| `POST /api/kea/dhcpv4/search_subnet` | paginated grid row | No |
| `POST /api/kea/dhcpv4/search_reservation` | paginated grid row | No |
| `GET /api/kea/dhcpv4/download_reservations` | CSV file export | No |
| option `get`/`search` | one generic option record | No |

`getBase` returns `$node->getNodes()`. For a list field that calls
`getNodeData()`. `ModelRelationField` loads choices and, unless
`internalIsSorted` is set, `natcasesort`s them by display text.
`getNodeData()` keeps saved comma-separated order only when sorted mode
is on. These relations do not enable it, so the selected map follows
description order. JSON object key order is not an order contract and
must not be parsed as linkage order.

`setNodes()` rejects an array for a non-container field. A selected map
is therefore not a valid write of the relation. The persisted form is a
scalar comma-separated UUID string. A successful write of that string
still has no authoritative read-back on `get_subnet` or
`get_reservation`.

`searchBase` builds rows through `UIModelGrid`, which stores
`getValue()` (the raw internal string) and a separate description.
That string is the stored CSV, but the endpoint is a paginated,
sortable grid (`rowCount`, `current`, `sort`, `searchPhrase`). A page
is not a parent-scoped ordered relation. The provider already uses
`search_subnet` and `search_reservation` as inventory lookups. Refresh
of existing parent owners uses `get_subnet` and `get_reservation`.
Search CSV is not a second supported order API.

`download_reservations` stringifies fields through `asRecordSet` and
excludes the subnet column. It is a file export, not an option-link
read. Reservation upload is a write. Neither is an ordered parent read.

Raw configuration XML contains the stored CSV. Reading or writing it is
not a supported API linkage and is not authorized.

## DHCPv6 boundary

Released `KeaDhcpv6.xml` has the same unsorted multi `option` relation
on subnet and reservation, and `Dhcpv6Controller.php` has the same
get/search actions with no ordered-link action. This story does not
implement v6 links. A later v6 story cannot treat that shape as an
ordered read without a new released-source gate.

## Provider implementation

`opnsense_kea_dhcpv4_subnet` and `opnsense_kea_dhcpv4_reservation` do
not expose the generic `option` relation. Their refresh endpoints are
`/api/kea/dhcpv4/get_subnet` and `/api/kea/dhcpv4/get_reservation`.
`option_data` remains the existing PXE/DNS/NTP field and must not be
overloaded as generic-option UUID order.

No parent-link attributes are added by this research.

## Proof boundary

No appliance was contacted for this gate. The earlier disabled-service
observation that reversing a stored UUID sequence changed generated
flex-option order remains historical evidence in
`kea-option-ddns-api-validation.md`. It is not repeated here, and it
does not prove that GET or search returns that order.

## Resume gate

Implement parent links only after a released stable tag documents an
order-preserving parent read that is not JSON object key order and not
raw XML, and that read is proven on that released firmware. Then extend
the existing subnet and reservation owners with typed ordered UUID
references. Preserve omitted fields and existing `option_data`/PXE
ownership. Do not create a competing parent owner. Packet delivery and
OpenTofu lifecycle remain separate gates.

The unsubmitted upstream proposal is
`kea-ordered-option-linkage-upstream-proposal.md`.
