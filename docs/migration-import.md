# Migration and Import Guide

This provider is designed for brownfield OPNsense environments. Most collection-backed resources import by the UUID that OPNsense already assigned to the object. Singleton or settings-style resources use the stable provider-defined ID documented on the resource page.

Use this guide when moving an existing appliance into Terraform without recreating live objects.

## Migration Principles

Import in small dependency-ordered batches. After each batch, reach a no-change `terraform plan` before importing resources that reference it.

Recommended workflow for every imported object:

1. Find the OPNsense UUID or documented singleton import ID.
2. Write the matching Terraform resource block in HCL.
3. Run `terraform import <address> <id>`.
4. Run `terraform plan`.
5. Adjust HCL until the plan reports no changes.
6. Commit the no-change batch before moving to dependent resources.

This keeps the Terraform state, HCL, and live appliance aligned before references are chained together.

## Finding Import IDs

For UUID-backed resources, use the OPNsense API, the browser developer tools while viewing the object in the OPNsense UI, or any existing automation inventory to identify the UUID.

Example API workflow:

```shell
curl -su "$OPNSENSE_API_KEY:$OPNSENSE_API_SECRET" \
  "$OPNSENSE_URI/api/firewall/alias/searchItem"
```

Then import the matching UUID:

```shell
terraform import opnsense_firewall_alias.internal_networks 11111111-2222-3333-4444-555555555555
```

For singleton or settings-style resources, use the import ID shown in that resource's documentation instead of an OPNsense UUID.

## Dependency Order Checklist

Use this order for full-appliance migrations. Skip domains that are not managed by Terraform in your environment.

1. Independent primitives: firewall aliases, firewall categories, users, groups, certificates, ACME accounts, DNS records, DHCP static reservations, HAProxy health checks, and HAProxy servers.
2. Network foundations: VLANs, VIPs, gateways, static routes, DHCP subnets, and interface-adjacent objects that the provider currently supports.
3. Referenced policies: HAProxy ACLs, FRR/Quagga prefix lists, route maps, BGP neighbors, Unbound ACLs, Monit service definitions, and traffic-shaping primitives.
4. Edge and security policies: firewall filter rules, NAT port forwards, outbound NAT, IPsec child SAs, OpenVPN instances, WireGuard servers, and WireGuard peers.
5. Chained application services: HAProxy backends, HAProxy frontends, ACME certificates, Dynamic DNS accounts, DNS overrides, and dependent VPN or routing objects.
6. Appliance-level validation: run a full `terraform plan`, compare enabled/disabled flags, and verify live OPNsense service behavior before enabling CI-driven applies.

## Independent Resource Example

Start with objects that other resources reference. Firewall aliases are a common first import because rules and NAT entries can depend on them.

```hcl
resource "opnsense_firewall_alias" "internal_networks" {
  name        = "internal_networks"
  type        = "network"
  description = "Internal network ranges"
  content     = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
}
```

```shell
terraform import opnsense_firewall_alias.internal_networks 11111111-2222-3333-4444-555555555555
terraform plan
```

If the plan shows changes, update the HCL to match the live alias before importing firewall rules that reference it.

## HAProxy Chain Example

For HAProxy, import the leaf objects first, then the resources that hold references to those UUIDs.

Recommended order:

1. `opnsense_haproxy_healthcheck`
2. `opnsense_haproxy_server`
3. `opnsense_haproxy_acl`
4. `opnsense_haproxy_backend`
5. `opnsense_haproxy_frontend`

Example HCL for a backend and frontend that reference imported servers and ACLs:

```hcl
resource "opnsense_haproxy_server" "web1" {
  name    = "web-1"
  address = "10.0.0.10"
  port    = 80
  weight  = 100
}

resource "opnsense_haproxy_server" "web2" {
  name    = "web-2"
  address = "10.0.0.11"
  port    = 80
  weight  = 100
}

resource "opnsense_haproxy_backend" "web_pool" {
  name                 = "web-pool"
  mode                 = "http"
  algorithm            = "roundrobin"
  linked_servers       = [opnsense_haproxy_server.web1.id, opnsense_haproxy_server.web2.id]
  health_check_enabled = true
  forward_for          = true
}

resource "opnsense_haproxy_frontend" "http" {
  name            = "http-frontend"
  bind            = "0.0.0.0:80"
  mode            = "http"
  default_backend = opnsense_haproxy_backend.web_pool.id
  forward_for     = true
}
```

Import commands:

```shell
terraform import opnsense_haproxy_server.web1 22222222-2222-3333-4444-555555555555
terraform import opnsense_haproxy_server.web2 33333333-2222-3333-4444-555555555555
terraform import opnsense_haproxy_backend.web_pool 44444444-2222-3333-4444-555555555555
terraform import opnsense_haproxy_frontend.http 55555555-2222-3333-4444-555555555555
terraform plan
```

Keep UUID references expressed through Terraform resource references after import. Avoid hard-coding UUID strings in dependent resources unless the referenced object remains intentionally unmanaged.

## Firewall Rules and NAT Example

Import aliases and categories first, then firewall rules and NAT entries.

```hcl
resource "opnsense_firewall_alias" "web_ports" {
  name        = "web_ports"
  type        = "port"
  description = "Public web ports"
  content     = ["80", "443"]
}

resource "opnsense_firewall_filter_rule" "allow_web" {
  action           = "pass"
  direction        = "in"
  protocol         = "tcp"
  destination_port = "443"
  description      = "Allow inbound HTTPS"
  log              = true
}

resource "opnsense_firewall_nat_port_forward" "https" {
  interface        = "wan"
  protocol         = "tcp"
  destination_net  = "wanip"
  destination_port = "443"
  target           = "192.0.2.10"
  local_port       = "443"
  description      = "Forward HTTPS to web server"
}
```

```shell
terraform import opnsense_firewall_alias.web_ports 66666666-2222-3333-4444-555555555555
terraform import opnsense_firewall_filter_rule.allow_web 77777777-2222-3333-4444-555555555555
terraform import opnsense_firewall_nat_port_forward.https 88888888-2222-3333-4444-555555555555
terraform plan
```

Firewall ordering can be operationally significant. Import and validate small groups so any ordering drift is visible before more policy is added.

## Routing Example

Import gateways before static routes that depend on them. Import FRR/Quagga primitives before BGP neighbors or route policy that references them.

```hcl
resource "opnsense_system_gateway" "wan_upstream" {
  name        = "WAN_UPSTREAM"
  interface   = "wan"
  gateway     = "198.51.100.1"
  description = "WAN upstream gateway"
}

resource "opnsense_system_route" "k8s_pods" {
  network     = "10.244.0.0/16"
  gateway     = opnsense_system_gateway.wan_upstream.name
  description = "Kubernetes pod network"
}
```

```shell
terraform import opnsense_system_gateway.wan_upstream 99999999-2222-3333-4444-555555555555
terraform import opnsense_system_route.k8s_pods aaaaaaaa-2222-3333-4444-555555555555
terraform plan
```

Gateway groups, interface base assignment, interface IP configuration, PPPoE, and system general settings are upstream-blocked until OPNsense exposes stable APIs. System tunables/sysctl is Coming with a safety/live-validation gate and should remain managed outside Terraform until the provider ships explicit support. Keep those settings managed outside Terraform and document the boundary in your runbook.

## VPN Examples

VPN resources often contain sensitive values or depend on peer identity. Import non-sensitive structure first, then restore secrets through variables.

WireGuard example:

```hcl
resource "opnsense_wireguard_server" "vpn" {
  name           = "wg-vpn"
  private_key    = var.wireguard_server_private_key
  tunnel_address = "10.10.0.1/24"
  port           = "51820"
  description    = "Road-warrior WireGuard server"
}

resource "opnsense_wireguard_peer" "laptop" {
  name           = "laptop"
  public_key     = var.laptop_public_key
  tunnel_address = "10.10.0.2/32"
  keepalive      = 25
}
```

```shell
terraform import opnsense_wireguard_server.vpn bbbbbbbb-2222-3333-4444-555555555555
terraform import opnsense_wireguard_peer.laptop cccccccc-2222-3333-4444-555555555555
terraform plan
```

IPsec example:

```hcl
resource "opnsense_ipsec_connection" "site_a" {
  description  = "Site A tunnel"
  remote_addrs = "198.51.100.10"
  version      = "2"
}
```

```shell
terraform import opnsense_ipsec_connection.site_a dddddddd-2222-3333-4444-555555555555
terraform plan
```

For pre-shared keys, static keys, and private keys, keep values in sensitive Terraform variables or an external secret store. If OPNsense does not return a write-only field through the API, Terraform cannot reconstruct it from import alone.

## Reaching a No-Change Plan

After every import, run:

```shell
terraform plan
```

Common fixes when the plan is not empty:

1. Add attributes that were omitted from HCL but are set on the live object.
2. Match enabled/disabled flags exactly.
3. Normalize lists to the order returned by OPNsense when order is semantically meaningful.
4. Replace literal UUID strings with references to imported Terraform resources where possible.
5. Supply retained sensitive write-only values through variables so Terraform can compare configured intent; if the original value is unknown, rotate it intentionally instead.
6. Decide explicitly whether live drift should be accepted into HCL or corrected by Terraform.

Do not import the next dependency layer until the current layer reaches a no-change plan.

## Sensitive and Write-Only Fields

Some OPNsense APIs do not return secret material after it is written. Examples include private keys, pre-shared keys, static keys, passwords, tokens, and similar credential fields.

For these fields:

1. Store the value in a sensitive Terraform variable or secret manager.
2. Populate the resource block with that variable before import or before the first post-import plan.
3. Avoid pasting secret values into committed HCL.
4. Expect Terraform to show a change if the configured secret differs from the live secret or cannot be read back.

If the live secret is unknown, rotate it through Terraform during a planned maintenance window instead of trying to recover it from state.

## Upstream-Blocked Domains

Some appliance settings are not currently manageable because OPNsense does not expose stable APIs for them. Current upstream-blocked areas include interface base assignment, interface IP configuration, PPPoE, gateway groups, and system general settings. System tunables/sysctl is not upstream-blocked; current `core/tunables` evidence makes it Coming with a safety/live-validation gate.

During migration:

1. Leave upstream-blocked settings managed in OPNsense or by an existing external process.
2. Document those boundaries next to your Terraform configuration.
3. Avoid modeling unsupported settings with local-only data or placeholder resources.
4. Revisit the support matrix when OPNsense exposes stable API coverage.

## Troubleshooting

### Import Succeeds but Plan Wants to Recreate

Confirm that the import ID belongs to the same resource type and that the HCL resource address matches the imported object. Re-importing a UUID into the wrong address can produce misleading diffs.

### Plan Shows Attribute Drift Immediately After Import

Update HCL to match live OPNsense values, including defaults that the UI set implicitly. If the live value is wrong, make that a deliberate Terraform change in a separate apply after the import baseline is stable.

### Dependent Resource Cannot Find a Referenced Object

Import the referenced object first and use its `.id` or documented attribute in the dependent resource. If the referenced object is intentionally unmanaged, keep the UUID literal and add a comment explaining the boundary.

### No-Change Plan Is Blocked by a Secret

Set the secret through a sensitive variable. If the original value is unavailable, schedule a controlled rotation and let Terraform become the source of truth for the replacement.

### API Endpoint Is Missing or Unstable

Check the support matrix before adding workarounds. If the domain is upstream-blocked, keep it outside Terraform until stable OPNsense API support exists.
