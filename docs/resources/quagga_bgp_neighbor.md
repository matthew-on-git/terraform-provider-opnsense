---
page_title: "opnsense_quagga_bgp_neighbor Resource - opnsense"
subcategory: "BGP"
description: |-
  Manages a BGP neighbor on OPNsense.
---

# opnsense_quagga_bgp_neighbor (Resource)

Manages a BGP neighbor peering session on OPNsense via the FRR/Quagga plugin. Requires `os-frr`.

`family` is a typed set of lowercase `ipv4` / `ipv6`. Existing boolean-only
`multi_protocol` configuration and schema-version-0 state upgrade in place:
`true` selects both families; `false` selects the transport family (including
IPv6-only transport). Explicit family may differ from transport on FRR 2.0;
legacy plugins reject that single cross-family intent. An explicitly configured
legacy alias must agree with explicit family. Omitted aliases are not conflicts,
and unknown configured intent is never treated as false or empty. Import reads
the actual API family. With both aliases omitted, changing transport between IPv4
and IPv6 updates a previously transport-derived single family; imported cross-family
API truth and dual-stack state remain unchanged. Unknown transport-dependent alias
comparisons are deferred until apply, where unresolved/conflicting intent rejects writes.

## Plugin compatibility and activation boundary

Numeric 2.0 package revisions are eligible for capability validation, not blanket
live-version certification. The existing live proof covers 2.0_1 only; 2.0_2 and
other numeric revisions have fixture/protocol regression coverage. New plugin
releases such as 2.1 and prerelease/malformed versions remain rejected.

BGP owners discover the **installed `os-frr` package** with read-only
`/api/core/firmware/info` and cross-check the neighbor model. The core firmware
version does not select the wire format. The compatibility gate accepts stable
legacy 1.46–1.55 package metadata or 2.0 with numeric package revisions (such as 2.0_2); an absent plugin, unknown future
plugin release, malformed revision, contradictory model, permission failure or unresolved configured
intent rejects writes. Grant read access to firmware package information as well
as the BGP API. Legacy live acceptance remains unverified; mocks are not appliance
version proof. Owner-lock waiting respects caller cancellation/deadlines and has no
separate fixed timeout. After acquisition, each read-only metadata/model request
has its own 30-second limit, including retries/read-slot waiting, capped by any
earlier caller deadline. A timeout rejects writes; check the appliance
firmware/package/configd backend before retrying.

On FRR 2.0, linked policies select **one logical group**, not an arbitrary entry
from multiple differently named groups. Existing entry UUID resources are retained.
A configured same-group sibling UUID is kept in state only after authoritative
full-membership verification; fresh imports/data sources expose the API's canonical
representative. Stale UUIDs and selected mixed-family prefix groups diagnose before
writes. Route-map `order` remains the sequence number, distinct from the UUID.

FRR service reconfigure is a **soft reload**, not a daemon-set restart or a PF
filter reload. On a fresh disabled appliance, configure BGP before enabling FRR;
changing the enabled daemon set on an already running appliance can require an
explicit reviewed service restart. Automatic BGP firewall rules are broad inet46
TCP/179 to/from self when general `enabled` + `firewall_rules` and BGP are enabled;
they are independent of neighbor family. A separate filter activation is required
for effective PF changes. The disposable acceptance runner records its explicit
lab filter activation separately and never claims it was performed by the provider.

If a confirmed save succeeds but activation/read-back fails, the UUID and private
pending-activation state are retained. Refresh retries activation without creating
another object. Resolve unsupported configuration/capabilities before retrying;
reconcile an ambiguous create with no uniquely proven identity rather than blindly
re-applying. The provider does not install or upgrade plugins, make production adoption or full appliance-proof claims.

## Advanced neighbor ownership

The candidate advanced surface requires released FRR 1.55 or 2.0 capabilities;
older supported base plugins require authoritative field-presence proof before
advanced writes. New public properties are Optional+Computed without defaults:
omission adopts remote values, explicit false, weight zero, empty strings and
empty capability sets are configured intent. Null numeric values remain unset.
`remote_as` is optional: `remote_as_mode` (`internal` / `external`) takes
precedence; a peer group may supply the AS. `attribute_unchanged` is one scalar
option, not a set. `remove_private_as` uses source values such as
`remove-private-AS all replace-AS`, not XML element labels. `allow_as_in` accepts
empty, `1` through `10`, or `origin`. Capabilities are a set of
`extended-nexthop`, `dynamic`, `fqdn`, and `link-local`.

`next_hop_self_all` requires `next_hop_self`; `bfd_strict` requires `bfd`.
Transport-wide flags and per-family policies follow the installed plugin template.
`link_local_interface` is meaningful for IPv6 transports; update-source rendering
is transport-dependent. Saved settings alone do not prove an established session.

`password` is an Optional Sensitive, **configuration-owned input**. Omitted or null
passwords are never read from the appliance, sent as clears, or adopted on import.
An explicit empty string clears authentication. Refresh retains prior configured
input, and lookups have no password property. A configured nonempty password
requires a host `local_ip`, as required by the released model. Sensitive marks
state for display redaction; it is **not encryption**. Protect Terraform state and
backend access. Remote password drift cannot be proven. A failed ambiguous create
with authentication cannot recover identity from remote secret comparison.

Candidate protocol/Core checks and disposable proof are separate from release
publication. Legacy appliance, OpenTofu and route-collector coverage remain
unverified unless separately evidenced; parent Story 31.0h remains open.

## Example Usage

```terraform
# MetalLB BGP peering session
resource "opnsense_quagga_bgp_neighbor" "metallb" {
  address             = "192.0.2.2"
  remote_as           = 65001
  family              = ["ipv4"]
  next_hop_self       = true
  weight              = 0
  connect_timer       = 30
  attribute_unchanged = "as-path"
  capabilities        = ["dynamic"]
  # Authentication is config-owned Sensitive input; omission never adopts it.
  description = "MetalLB speaker"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `address` (String) Neighbor IP address.

### Optional

- `allow_as_in` (String) Source-supported allow_as_in; omitted values adopt appliance state.
- `as_override` (Boolean) Source-supported as_override; omitted values adopt appliance state.
- `attribute_unchanged` (String) Source-supported attribute_unchanged; omitted values adopt appliance state.
- `bfd` (Boolean) Source-supported bfd; omitted values adopt appliance state.
- `bfd_strict` (Boolean) Source-supported bfd_strict; omitted values adopt appliance state.
- `capabilities` (Set of String) Source-supported capabilities; omitted values adopt appliance state.
- `connect_timer` (Number) Source-supported connect_timer; omitted values adopt appliance state.
- `default_originate` (Boolean) Source-supported default_originate; omitted values adopt appliance state.
- `description` (String) Description.
- `disable_connected_check` (Boolean) Source-supported disable_connected_check; omitted values adopt appliance state.
- `enabled` (Boolean) Whether this neighbor is enabled.
- `family` (Set of String) Activated address families: lowercase `ipv4`, `ipv6`, or both. Omission preserves legacy multi_protocol/transport behavior. Explicit family may differ from transport only on os-frr 2.0.
- `holddown` (Number) Hold-down timer in seconds (3-3000).
- `keepalive` (Number) Keepalive interval in seconds (1-1000).
- `link_local_interface` (String) Source-supported link_local_interface; omitted values adopt appliance state.
- `linked_prefixlist_in` (Set of String) Inbound prefix list UUIDs.
- `linked_prefixlist_out` (Set of String) Outbound prefix list UUIDs.
- `linked_routemap_in` (Set of String) Inbound route map UUIDs.
- `linked_routemap_out` (Set of String) Outbound route map UUIDs.
- `local_as` (Number) Source-supported local_as; omitted values adopt appliance state.
- `local_ip` (String) Source-supported local_ip; omitted values adopt appliance state.
- `multi_protocol` (Boolean) Legacy alias: true activates ipv4 and ipv6; false selects the transport family. Prefer family.
- `multihop` (Boolean) Source-supported multihop; omitted values adopt appliance state.
- `next_hop_self` (Boolean) Set next-hop to self for advertised routes.
- `next_hop_self_all` (Boolean) Source-supported next_hop_self_all; omitted values adopt appliance state.
- `password` (String, Sensitive) Config-owned authentication input. Omitted values are never adopted or cleared. Sensitive state is not encrypted; remote password drift cannot be proven.
- `peer_group` (String) Existing BGP peer-group UUID. Its policies are activated per selected address family on FRR 2.0.
- `remote_as` (Number) Remote AS number; mode takes precedence and peer groups may supply it.
- `remote_as_mode` (String) Source-supported remote_as_mode; omitted values adopt appliance state.
- `remove_private_as` (String) Source-supported remove_private_as; omitted values adopt appliance state.
- `route_reflector_client` (Boolean) Source-supported route_reflector_client; omitted values adopt appliance state.
- `soft_reconfiguration_inbound` (Boolean) Source-supported soft_reconfiguration_inbound; omitted values adopt appliance state.
- `update_source` (String) Interface or IP to use as update source.
- `weight` (Number) Source-supported weight; omitted values adopt appliance state.

### Read-Only

- `id` (String) UUID.

## Import

```shell
terraform import opnsense_quagga_bgp_neighbor.example 12345678-1234-1234-1234-123456789012
```

### Remaining numeric-clear interface limitation

`local_as`, `connect_timer`, and `remote_as` are Optional+Computed numeric
properties. After adopting a non-null appliance value, omission or `null`
preserves that value; neither is a configured clear. There is currently no
configured numeric clear for these properties. No zero sentinel is introduced.
Designing an explicit clear interface remains an open parent Story 31.0h
follow-up; empty string/set clears described above apply only to those types.
