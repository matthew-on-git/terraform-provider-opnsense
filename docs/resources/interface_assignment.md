---
page_title: "opnsense_interface_assignment Resource - opnsense"
subcategory: "Interfaces"
description: |-
  Adopts and manages safe IP settings for an existing OPNsense base interface assignment.
---

# opnsense_interface_assignment (Resource)

Adopts an existing `wan`, `lan`, or `optN` assignment and manages its enabled
state, description, MTU/MSS, private/bogon blocking, per-interface offload controls,
and safe IPv4, IPv6, DHCPv4, and DHCPv6 settings. This resource requires
OPNsense **26.7.6 or newer** with the complete expanded
`interfaces/assignment` API.

This resource does not create assignments. Apply calls
`assignment/set_item/{name}`, applies the assignment reconfigure operation, and
reads the result back from OPNsense. Fields outside the documented safe subset
are preserved. PPP/PPPoE, 6rd, 6to4, media selection, MAC spoofing and promiscuous
mode remain outside this resource.

Omitting the new controls retains their saved values. Set `mtu` or `mss` to
`0` to clear the saved override (the API stores this as an empty string);
explicit values must be between 576 and 65535. A missing API field remains
null and cannot be configured until the appliance exposes it. Offload flags
only take effect when `override_hardware_settings` is enabled.

`device` can rebind an existing assignment to a device advertised by its current
API option map. It never creates a new assignment or changes its immutable name.
Rebinding can immediately interrupt connectivity. `vlan_hardware_filtering`
accepts `enable`, `disable`, or `default`.

The DHCPv6 client controls use the `dhcp6_` prefix. `dhcp6_prefix_length` accepts
`48` through `64` as strings, or `none`; the provider maps these displayed values
to the appliance encoding (`56` is encoded as `8`). Configured non-default prefix requests require DHCPv6 mode, even when unchanged.
Newer appliance versions discard them in other modes on any update, so the
provider rejects such updates before writing, including unrelated changes. Prefix-only and send-hint
flags preserve explicit false. Prefix/interface IDs and association ID retain
empty values as explicit clears. Decimal association IDs must use canonical
spelling (no leading zeros) and fit 0 through 9223372036854775807. Prefix and
interface hexadecimal IDs must fit 0 through `0x7fffffffffffffff`: the appliance
serializes them through signed 64-bit integers even though its XML mask permits
wider strings. Equivalent hexadecimal spelling is retained across normalized
readback; bare `0x` has the appliance value zero. These settings do not implicitly change
`ipv6_mode`, configure DHCPv6 server services, or change child assignments.

> **Connectivity warning:** Interface changes can immediately sever network
> access. Test with a disposable non-management `optN` before adopting a live
> interface. Do not experiment on the interface used to reach the OPNsense API.

IPv4 modes are `none`, `staticv4`, and `dhcp`. IPv6 modes are `none`,
`staticv6`, `dhcp6`, `linklocal`, `slaac`, `idassoc6`, and legacy `track6`. Static modes require both the
matching address and prefix-length attributes. Static-only address, prefix, and
gateway values are rejected for non-static modes.

## IPv6 identity association

`idassoc6` and legacy `track6` use the same assignment fields in OPNsense.
`ipv6_track_interface` selects the upstream assignment; `ipv6_prefix_id`
selects the delegated subnet, and `ipv6_interface_id` configures the interface
identifier. `ipv6_identity_association_id` and `ipv6_prefix_range` expose the
association and prefix-range settings. `ipv6_allow_manual_dhcpv6` retains the
legacy manual DHCPv6/RA override flag; it does not manage DHCP/RA services.

The association ID and prefix range are decimal **strings** because OPNsense
stores empty/default values distinctly from explicit numbers, including zero.
Empty strings explicitly clear these settings. Hexadecimal prefix/interface IDs
retain their stored spelling. Omitted or unknown optional values preserve the
existing appliance settings during an update. Switching modes retains inactive
delegation fields; it does not implicitly migrate legacy settings or alter the
upstream DHCPv6 client's prefix request. Unsupported/partial delegation API
responses are rejected before writing.

For read-only adoption, configure the existing assignment name, import it, then
review its state and explicitly declare the fields you intend to own. A no-change
plan verifies read/import coverage but does not validate activation or packet
behavior. Test intentional changes on a disposable interface before managing a
production interface. The upstream appliance still validates parent eligibility
and the delegated subnet's range.

## Retained on destroy

Terraform destroy only removes this resource from Terraform state. It **does
not** disable or delete the OPNsense assignment and does not revert its current
settings. This avoids the destructive upstream delete behavior, which can also
remove firewall references. The retained assignment can be adopted again with
`terraform import`.

## Example Usage

```terraform
resource "opnsense_interface_assignment" "pxe" {
  # The assignment must already exist. Use only a non-management interface
  # whose addressing can safely be changed.
  name               = "opt3"
  enabled            = true
  ipv4_mode          = "staticv4"
  ipv4_address       = "192.0.2.1"
  ipv4_prefix_length = 24
  ipv4_gateway       = ""
  ipv6_mode          = "none"
}

# Existing downstream interface using a delegated IPv6 prefix. Import first;
# replace these values with the intended configuration only after reviewing state.
resource "opnsense_interface_assignment" "delegated" {
  name                         = "opt4"
  ipv6_mode                    = "idassoc6"
  ipv6_track_interface         = "wan"
  ipv6_prefix_id               = "0x10"
  ipv6_interface_id            = "0x1"
  ipv6_identity_association_id = "0"
  ipv6_prefix_range            = ""
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Existing immutable assignment name: `wan`, `lan`, or `optN`.

### Optional

- `block_bogons` (Boolean) Block bogon networks on this interface. Omit to retain the saved value.
- `block_private` (Boolean) Block private networks on this interface. Omit to retain the saved value.
- `description` (String) Interface description (letters, digits and underscores). Omit to retain the saved value.
- `device` (String) Existing device to bind to this assignment; must be advertised by the API before rebinding. Omit to retain the saved value.
- `dhcp6_identity_association_id` (String) DHCPv6 identity association ID as a canonical nonnegative signed-64-bit decimal string; empty selects the default. Omit to retain the saved value.
- `dhcp6_information_only` (Boolean) Use DHCPv6 information-only mode.
- `dhcp6_interface_id` (String) DHCPv6 hexadecimal interface ID (0x prefix), up to 0x7fffffffffffffff because the appliance serializes through a signed 64-bit integer. Empty clears the override. Omit to retain the saved value.
- `dhcp6_prefix_id` (String) DHCPv6 hexadecimal prefix ID (0x prefix), up to 0x7fffffffffffffff because the appliance serializes through a signed 64-bit integer. Empty clears the override. Omit to retain the saved value.
- `dhcp6_prefix_length` (String) Requested DHCPv6 prefix length: decimal 48 through 64, or none. Exposes displayed prefix length, not the API encoding. Omit to retain the saved value.
- `dhcp6_prefix_only` (Boolean) Request only an IPv6 prefix through DHCPv6. Omit to retain the saved value.
- `dhcp6_rapid_commit` (Boolean) Enable DHCPv6 rapid commit.
- `dhcp6_request_dns` (Boolean) Request DNS information through DHCPv6.
- `dhcp6_request_options` (String) OPNsense DHCPv6 request-options value.
- `dhcp6_send_options` (String) OPNsense DHCPv6 send-options value.
- `dhcp6_send_prefix_hint` (Boolean) Send the requested prefix length as a DHCPv6 prefix hint. Omit to retain the saved value.
- `dhcp6_vlan_priority` (String) DHCPv6 VLAN priority: empty (disabled) or `0` through `7`.
- `dhcp_honor_mtu` (Boolean) Honor the MTU supplied by DHCPv4.
- `dhcp_hostname` (String) Hostname sent by the DHCPv4 client.
- `dhcp_reject_from` (Set of String) IPv4 DHCP server addresses from which offers are rejected.
- `dhcp_request_options` (String) OPNsense DHCPv4 request-options value.
- `dhcp_send_options` (String) OPNsense DHCPv4 send-options value.
- `dhcp_vlan_priority` (String) DHCPv4 VLAN priority: empty (disabled) or `0` through `7`.
- `disable_checksum_offloading` (Boolean) Disable hardware checksum offloading when hardware settings are overridden. Omit to retain the saved value.
- `disable_large_receive_offloading` (Boolean) Disable large receive offloading when hardware settings are overridden. Omit to retain the saved value.
- `disable_segmentation_offloading` (Boolean) Disable TCP segmentation offloading when hardware settings are overridden. Omit to retain the saved value.
- `enabled` (Boolean) Whether the assigned interface is enabled.
- `ipv4_address` (String) Static IPv4 address without a prefix. Valid only with `ipv4_mode = "staticv4"`.
- `ipv4_gateway` (String) Existing OPNsense IPv4 gateway name. Valid only with `ipv4_mode = "staticv4"`.
- `ipv4_mode` (String) IPv4 mode: `none`, `staticv4`, or `dhcp`.
- `ipv4_prefix_length` (Number) Static IPv4 prefix length (0-32). Valid only with `ipv4_mode = "staticv4"`.
- `ipv6_address` (String) Static IPv6 address without a prefix. Valid only with `ipv6_mode = "staticv6"`.
- `ipv6_allow_manual_dhcpv6` (Boolean) Allow manual DHCPv6/router-advertisement configuration for legacy track-interface mode. Preserves the appliance's dhcpd6track6allowoverride field; does not manage DHCP or RA services.
- `ipv6_gateway` (String) Existing OPNsense IPv6 gateway name. Valid only with `ipv6_mode = "staticv6"`.
- `ipv6_identity_association_id` (String) Prefix-delegation identity association ID, as a nonnegative decimal string or empty for the appliance default. Empty and explicit zero are distinct saved values.
- `ipv6_interface_id` (String) IPv6 interface identifier in hexadecimal (for example `0x1`), or empty for the appliance default.
- `ipv6_mode` (String) IPv6 mode: `none`, `staticv6`, `dhcp6`, `linklocal`, `slaac`, `idassoc6`, or `track6`.
- `ipv6_prefix_id` (String) Delegated subnet prefix ID, in the appliance's hexadecimal encoding (for example `0x10`). Empty retains the appliance's default representation; omission preserves the saved value.
- `ipv6_prefix_length` (Number) Static IPv6 prefix length (0-128). Valid only with `ipv6_mode = "staticv6"`.
- `ipv6_prefix_range` (String) Identity-association prefix range, as a positive decimal string or empty for the appliance default. Strings preserve the distinction between an unset value and a number.
- `ipv6_track_interface` (String) Upstream assignment for `idassoc6` or legacy `track6`, such as `wan`. Omission preserves the existing parent; these modes require a nonempty parent distinct from this assignment.
- `mss` (Number) MSS in bytes (576-65535); zero clears the override and uses the system default. Omit to retain the saved value.
- `mtu` (Number) MTU in bytes (576-65535); zero clears the override and uses the system default. Omit to retain the saved value.
- `override_hardware_settings` (Boolean) Override global hardware offload settings for this interface. Omit to retain the saved value.
- `vlan_hardware_filtering` (String) VLAN hardware filtering: enable, disable, or default. Applies when per-interface hardware settings are overridden. Omit to retain the saved value.

### Read-Only

- `id` (String) Interface assignment name, identical to `name`.

## Import

Import an existing assignment by its interface name:

```shell
terraform import opnsense_interface_assignment.pxe opt3
```
