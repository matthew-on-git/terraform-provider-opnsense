# Kea generic-option and DDNS stable API validation — 2026-10-25

## Outcome and boundary

Tagged core 26.7.7 source and non-skipped direct-API probes establish buildable
contracts for separate generic DHCPv4 option and DDNS settings owners. They do not
implement provider resources or prove Terraform import/no-change/drift/restart/
state-only destroy. Follow-ups are Stories 31.0n and 31.0o. Completed subnet/PXE,
DHCPv6 peer and delegated-prefix work is not repeated. No real leases, DNS updates,
production DHCP or private operator integration were enabled.

## Generic option contract

[Dhcpv4Controller](https://github.com/opnsense/core/blob/26.7.7/src/opnsense/mvc/app/controllers/OPNsense/Kea/Api/Dhcpv4Controller.php)
uses `options.option` storage and wrapper `option` for `get_option`, `add_option`,
`set_option`, `del_option`, `search_option`. UUIDs identify options independently
from subnet `option_data`.

Observed fields: `code`, `encoding`, `data`, `force`, `match_code`,
`match_encoding`, `match_data`, `description`. Option and encoding selectors are
selected maps on GET, scalar keys on write. **Description is required**, not an
optional clearable comment. Code/encoding combinations have semantic validation:
option 15 with `string` was rejected and explicitly required `fqdn` or `hex`.

The passing fixture used a reserved example domain and `fqdn` encoding. Create/get,
UUID-preserving partial data/force update, omitted-field preservation and explicit
force true-to-false clear passed. An empty description was rejected and the prior
record/data remained. Delete/search left zero generic options. Service reconfigure
returned `status=ok`; all DHCP services remained disabled. Wrong encoding and
required-description failures were retained, cleaned up and not counted as passes.
Matching-option clear/ordering and active subnet linkage still need additional
coverage in the implementation acceptance matrix; no proof is claimed for them.

## DDNS contract

[DdnsController](https://github.com/opnsense/core/blob/26.7.7/src/opnsense/mvc/app/controllers/OPNsense/Kea/Api/DdnsController.php)
inherits mutable model set/save behavior while returning only `ddns.general`.
The tagged [KeaDdns model](https://github.com/opnsense/core/blob/26.7.7/src/opnsense/mvc/app/models/OPNsense/Kea/KeaDdns.xml)
contains that general subtree: `enabled` (default false), `manual_config`,
`server_ip` (IPv4, default loopback) and `server_port` (default 53001).

POST `/api/kea/ddns/set` with only a changed `ddns.general.server_port` preserved
all other fields. Setting `manual_config` true then explicitly false preserved
port/address/disabled state. `/api/kea/service/reconfigure` returned `status=ok`.
Final full DHCPv4 GET and DDNS GET matched their baselines exactly and option
search was globally empty. Services remained inactive throughout. A future owner
must write only this subtree, never use this filtered GET as ownership of unrelated
Kea settings, and treat destroy as state-only.

## Integration is separate

The lab package was Kea 3.0.4; host/lease command libraries and an inactive Control
Agent were present. File presence does not prove loaded hooks, permissions, a
reachable target Control Agent or Smart Proxy interoperability. These remain
operator integration gates outside the public provider feature contract.

Controller SHA256s: DHCPv4 `9fa1ca628fadcba5b844e536a9ed082da59bb2fa1fbfffdace5610ae73614669`;
DDNS `f645aa48e3a2dfcfb6b2a6b096ceac5d84eb95624c77194afb8b1b10b051d2d3`.
DDNS model SHA256: `a54b2b92ab3832d77855d51ea5a6cf7030794f8cb622ec24ddb801786a04ae84`.


## Subsequent generic-option contract calibration

A second guarded isolated appliance on the same exact core verified population
and explicit clearing of `match_code`, `match_encoding` and `match_data`.
An encoding-only value with empty match code/data is valid; an all-three-or-none
provider validator would incorrectly reject this upstream-supported state.
Two same-code options with distinct descriptions coexist. Description is required,
unique and mutable, not an immutable option name.

An official-API subnet reference caused option deletion to return HTTP500 while
retaining the tracked option UUID and complete saved values. Unlinking allowed
cleanup. Reversing the stored parent option UUID sequence changed the installed
model's generated flex-option order, consistent with last-matching-same-code
precedence. The observation invoked installed model generation to a private
throwaway file while all DHCP/DDNS daemons stayed inactive; ordinary service
reconfigure correctly skipped generation while disabled. It is not active
packet-delivery, normal disabled-service-generation or Terraform linkage proof.
The stored force flag changed on read-back but did not change that generated
fixture; tagged generator source does not consume it, so no packet-level force
semantics are established.

A first calibration attempt stopped because its observer tried to parse the
unchanged stock commented runtime file as strict JSON. Its matching/dependency
checks had passed, but the attempt is retained as failed overall; zero entries
and exact restoration passed. The corrected private-artifact observation passed
independently. Both cleanups verified zero tagged options/subnets/reservations,
exact full DHCPv4 GET, saved configuration and IPv4/IPv6 route restoration.
No active DHCP or DNS traffic, production configuration or private inventory was
used. These observations refine the implementation contract; real Terraform
owner/lookup proof is recorded separately for Story31.0n in
`../implementation-artifacts/31-0n-kea-option-live-evidence.md`; the API calibration alone does not establish it.

Existing subnet/reservation resources do not yet expose the generic option
relation. Released core 26.7.8 parent GET choice maps are description-sorted
and do not prove stored selection order. Ordered parent linkage is
upstream-blocked, not completion conferred by creating an unlinked generic
option. See `kea-ordered-option-linkage-api-research.md`.
