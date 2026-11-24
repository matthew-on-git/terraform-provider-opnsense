---
page_title: "opnsense_wireguard_peer Resource - opnsense"
subcategory: "WireGuard"
description: |-
  Manages core WireGuard peer on OPNsense.
---
# opnsense_wireguard_peer (Resource)

WireGuard entered OPNsense core in **24.1**; no plugin is required. This provider contract was acceptance-validated on **26.7.7**, not on every earlier core release. Terraform Core was exercised; no OpenTofu execution is claimed by that evidence.

Multi-value addresses and relationships use Terraform sets. Existing scalar state upgrades automatically without replacement; update configuration to set syntax. Server `peers` owns associations; peer `servers` is an API-derived backreference. Peer updates preserve the current API backreference under the association owner lock; this does not create independent association ownership. Do not manage both sides independently in one configuration or through another provider.

PSK and private keys are resource-only Sensitive values. OPNsense may omit them on reads; prior resource state is preserved. Imports cannot recover omitted keys. Sensitive values still reside in resource state: secure the state backend. Data sources never expose secrets.

## Migrating configured server associations

Before upgrading, capture each peer's complete `servers` CSV and every referenced
server's existing peer membership from the API or prior state. Merge all those
memberships; do not replace a server's peer set with only the migrating peer.
Stop the previous configuration/provider from writing these objects first.
Securely preserve and explicitly configure every existing PSK before any apply;
imports cannot recover a PSK omitted by the API. The `psk` default is blank:
omitting it or setting it blank disables an existing PSK on update. Do not treat
an omitted PSK as a request to leave a previously configured appliance PSK alone.

Before (legacy configuration):

```hcl
resource "opnsense_wireguard_peer" "office" {
  # Other peer attributes unchanged.
  tunnel_address = "192.0.2.2/32,2001:db8::2/128"
  servers        = "00000000-0000-0000-0000-000000000001"
}
```

After (new configuration):

```hcl
resource "opnsense_wireguard_peer" "office" {
  # Other peer attributes unchanged; omit the computed-only servers attribute.
  tunnel_address = ["192.0.2.2/32", "2001:db8::2/128"]
}
resource "opnsense_wireguard_server" "office" {
  # Capture the existing server's complete non-secret attributes and supply its
  # private key securely; an import cannot recover a write-only key.
  peers = [opnsense_wireguard_peer.office.id, opnsense_wireguard_peer.existing.id]
}
```

Import every newly managed existing server by UUID, for example
`terraform import opnsense_wireguard_server.office 00000000-0000-0000-0000-000000000001`.
Import existing peers that are not already in this provider's state before the
servers that reference them. Keep already managed UUIDs in place; schema upgrade
is automatic. Repeat this for **all** memberships and require a no-change plan
before any apply. No manual state editing is needed. Never let two configurations
or providers write the same object or association.

## Recovering an ambiguous key update

A confirmed save followed by failed activation retains the newly saved configured
key. An ambiguous save cannot confirm either the previous or planned key when the
API permanently omits secrets. In that case the provider retains the UUID and a
private reconciliation journal, marks the public secret value unconfirmed, and
returns an error. Refresh retries activation, **not the possibly saved mutation**;
blank API secret responses cannot automatically clear that uncertainty.

For explicit operator recovery:

1. Record the exact existing UUID and securely back up Terraform state (which may
   contain secrets). Independently verify/reconcile the desired key on the appliance.
2. Stop other writers. Use `terraform state rm <resource-address>` to release only
   Terraform ownership; this does not delete the appliance object. Then immediately
   `terraform import <resource-address> <the-same-recorded-UUID>` to restore ownership
   without the old private uncertainty marker. Do not run apply between these steps.
3. Configure the independently verified key securely and capture all readable
   fields/associations. Review a fresh plan before apply; an omitted API key cannot
   be recovered by import, so a planned key write is an explicit operator action,
   not automatic replay of the failed mutation. Require convergence afterward.

These supported CLI operations do not require editing a state file or replacing
an appliance object. Keep the backup protected and preserve the same UUID. Do not
remove ownership merely to silence an unresolved key mismatch.

## Example Usage

```terraform
resource "opnsense_wireguard_peer" "this" {
  enabled        = false
  name           = "example_peer"
  public_key     = var.peer_public_key
  psk            = var.peer_psk
  tunnel_address = ["10.10.0.2/32", "fd00::2/128"]
  server_address = "vpn.example.com"
  server_port    = 51820
  keepalive      = 25
}
variable "peer_public_key" {
  type = string
}
variable "peer_psk" {
  type      = string
  sensitive = true
  default   = ""
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Name of the WireGuard peer.
- `public_key` (String) Base64-encoded 256-bit public key.
- `tunnel_address` (Set of String) Allowed tunnel CIDRs.

### Optional

- `enabled` (Boolean) Whether this peer is enabled.
- `keepalive` (Number) Persistent keepalive interval in seconds. Null means unset.
- `psk` (String, Sensitive) Write-only base64-encoded 256-bit pre-shared key. Omission preserves prior state without asserting that an unreadable key is absent. Explicit blank disables PSK.
- `server_address` (String) Remote endpoint address.
- `server_port` (Number) Remote endpoint port. Null means unset.

### Read-Only

- `id` (String) UUID of the WireGuard peer in OPNsense.
- `servers` (Set of String) Authoritative API backreference to server UUIDs. Configure the association through server.peers.

## Import

```shell
terraform import opnsense_wireguard_peer.this 00000000-0000-0000-0000-000000000000
```
