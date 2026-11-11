# Live interface API contract

Source authorities: pinned public schema revision [`33c2a03`](https://github.com/browningluke/terraform-provider-opnsense/tree/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/internal/service), client revision [`9cd7726`](https://github.com/browningluke/opnsense-go/tree/9cd77265c81e329ba6d4399defcbf884c61ac75f/pkg), released core26.7.8 commit [`96978001a8343669154e4492172ce0a07dac7e4b`](https://github.com/opnsense/core/tree/96978001a8343669154e4492172ce0a07dac7e4b). No independent26.7.8_2 Git tag/source equality is inferred.

## Request and envelope

Only GET `/api/diagnostics/interface/getInterfaceConfig` and GET `/api/interfaces/overview/interfaces_info` are used, without JSON request bodies. Diagnostics returns a map keyed by kernel device, or legitimate empty `[]`/`{}`. Each nonempty row must have a matching nonempty device. Null is a failed/malformed backend, not an empty inventory.

Overview returns `{total,rowCount,current,rows}` with integer counts, page1 and at most9999 rows by default. The base controller reads pagination from POST, not GET query parameters. Require total and rowCount equal actual rows and current1; reject incomplete results rather than silently omitting devices. Missing/duplicate row devices fail closed. Pinned single lookup is by exact device, not identifier/description. Aggregates expose the same row attributes under computed `interfaces`, with computed nested device.

## Allowlisted typed fields

Below, String/Bool/Int64 are Terraform scalar types, Set means set(string), and List/Object shapes are defined by their corresponding nested model. All fields computed except single-source required `device`.

### Diagnostics

- interfaceDataSourceModel: `device` String, `media` String, `media_raw` String, `macaddr` String, `is_physical` Bool, `mtu` Int64, `status` String, `flags` Set, `capabilities` Set, `options` Set, `supported_media` Set, `groups` Set, `ipv4` List, `ipv6` List.
- ipv4Model: `ipaddr` String, `subnetbits` Int64, `tunnel` Bool.
- ipv6Model: `ipaddr` String, `subnetbits` Int64, `tunnel` Bool, `autoconf` Bool, `deprecated` Bool, `link_local` Bool, `tentative` Bool.

### Overview

`link_type` is the IPv4 configuration mode, not Ethernet media. Upstream IPv6
mode `link_typev6` is intentionally outside this pinned public allowlist; this is
not a complete assignment-configuration view. Overview omits upstream `pfsync0`
while diagnostics can include it. Unassigned `description` is localized and is
not a stable identity or suitable `for_each` key; use `device`.

- overviewInterfaceModel: `identifier` String, `description` String, `device` String, `enabled` Bool, `link_type` String, `addr4` String, `addr6` String, `macaddr` String, `macaddr_hw` String, `mtu` String, `status` String, `media` String, `media_raw` String, `is_physical` Bool, `vlan_tag` String, `lagg_proto` String, `lagg_hash` String, `lagg_options` Object, `lagg_statistics` Object, `flags` Set, `capabilities` Set, `options` Set, `supported_media` Set, `groups` Set, `gateways` Set, `routes` Set, `nd6_flags` Set, `ifctl_nameserver` Set, `ifctl_router` Set, `ifctl_prefix` Set, `ifctl_searchdomain` Set, `ipv4` List, `ipv6` List, `vlan` Object.
- overviewIPv4Model: `ipaddr` String, `vhid` String, `carp_status` String, `adv_base` String, `adv_skew` String, `peer` String, `peer6` String.
- overviewIPv6Model: `ipaddr` String, `vhid` String, `carp_status` String, `adv_base` String, `adv_skew` String, `peer` String, `peer6` String.
- overviewVLANModel: `tag` String, `proto` String, `pcp` String, `parent` String.
- overviewLaggOptionsModel: `flags` Set, `flowid_shift` String.
- overviewLaggStatisticsModel: `active_ports` String, `flapping` String.

## Nonidentity wire mappings

These are explicit source-derived translations, not generic underscore replacement:

| Public field | Wire location |
| --- | --- |
| Diagnostic IPv6 `link_local` | `link-local` |
| Overview `nd6_flags` | nested `nd6.flags` |
| Overview `ifctl_nameserver/router/prefix/searchdomain` | corresponding dotted `ifctl.nameserver/router/prefix/searchdomain` |
| Overview `lagg_proto`, `lagg_hash` | `laggproto`, `lagghash` |
| Overview `lagg_options`, `lagg_statistics` | `laggoptions`, `laggstatistics` |
| LAGG statistics `active_ports` | `active ports` |
| Overview IPv4/IPv6 `carp_status` | `status` |
| Overview IPv4/IPv6 `adv_base`, `adv_skew` | `advbase`, `advskew` |

Test fixtures must use actual wire names; public-name fixtures alone cannot prove these translations.

## Null, ordering and operational semantics

Preserve optional absent/null scalar/object values as typed null, explicit false/zero/empty strings as supplied, and missing collection fields as known empty collections. Diagnostics MTU is parsed from a wire decimal string to Int64; overview MTU stays String. Invalid present values fail rather than silently becoming null/zero. Sets collapse exact duplicate strings because the public type is a set, while preserving every distinct member exactly; DHCP file-derived strings may contain whitespace, multiple addresses or an empty string and must not be split/trimmed/filtered.

Sort aggregate rows by device and address lists by complete public tuple; do not claim element0 is primary. Overview addr4/addr6 are explicit primary observations. Assigned address entries contain CIDR/CARP data; unassigned entries can retain bare ipaddr with extra subnetbits that the public overview model does not expose. Missing CARP metadata is null. Unassigned interfaces have identifier empty and may omit enabled/link_type/addr4/addr6/vlan_tag/gateways; assigned non-VLAN vlan_tag can explicitly be null. Optional VLAN/LAGG/ND6 objects depend on hardware. No synthetic objects or assignment aliases.

Unknown extra backend keys are ignored. NEVER expose config, wireless, arbitrary backend values or raw JSON/body diagnostics; those can contain authentication material. Public outputs are observational and may contain addresses; treat Terraform state accordingly. These source paths read runtime/config/autoconf data, not save/apply/reconfigure configuration. This is not proof that incidental runtime caches never change.

## Evidence boundary

Pinned/core source review establishes contract only, not live appliance coverage. Unit/Core fixture checks and actual supported-appliance read-only observations must be reported separately. No universal older-firmware, OpenTofu, CARP/VLAN/LAGG hardware or runtime traffic proof is inferred.
