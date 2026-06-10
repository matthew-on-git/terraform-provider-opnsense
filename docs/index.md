---
page_title: "Provider: OPNsense"
description: |-
  The OPNsense provider manages firewall, networking, VPN, DNS, DHCP, load
  balancing, certificate, and routing configuration on an OPNsense appliance
  through its MVC API.
---

# OPNsense Provider

The OPNsense provider lets you manage an [OPNsense](https://opnsense.org/)
appliance declaratively through its REST/MVC API — firewall rules and aliases,
HAProxy load balancing, ACME certificates, IPsec/WireGuard/OpenVPN tunnels,
Unbound/dnsmasq DNS, Kea DHCP, FRR dynamic routing, interfaces, users, and more.

## Example Usage

```terraform
terraform {
  required_providers {
    opnsense = {
      source = "matthew-on-git/opnsense"
    }
  }
}

# Configure the OPNsense provider. Credentials may also be supplied via the
# OPNSENSE_URI, OPNSENSE_API_KEY, OPNSENSE_API_SECRET, and OPNSENSE_ALLOW_INSECURE
# environment variables (HCL values take priority).
provider "opnsense" {
  uri        = "https://opnsense.example.com"
  api_key    = var.opnsense_api_key
  api_secret = var.opnsense_api_secret
}

# Quickstart: a firewall alias and a rule that references it.
resource "opnsense_firewall_alias" "web_servers" {
  name        = "web_servers"
  type        = "host"
  description = "Public web servers"
  content     = ["10.0.0.10", "10.0.0.11"]
}

resource "opnsense_firewall_filter_rule" "allow_https" {
  action           = "pass"
  direction        = "in"
  protocol         = "tcp"
  destination_port = "443"
  description      = "Allow inbound HTTPS"
}
```

## Authentication

The provider authenticates with an OPNsense **API key + secret** (created under
*System → Access → Users → API keys*). Each request uses HTTP Basic auth over
TLS.

Credentials and connection settings can be supplied two ways, with **HCL taking
priority over environment variables**:

| Setting        | HCL argument  | Environment variable      |
|----------------|---------------|---------------------------|
| Base URL       | `uri`         | `OPNSENSE_URI`            |
| API key        | `api_key`     | `OPNSENSE_API_KEY`        |
| API secret     | `api_secret`  | `OPNSENSE_API_SECRET`     |
| Skip TLS verify| `insecure`    | `OPNSENSE_ALLOW_INSECURE` |

`api_key` and `api_secret` are marked sensitive. On configuration the provider
validates credentials against `/api/core/firmware/status` and reports a clear
diagnostic if authentication fails.

```hcl
# Environment-variable configuration (no secrets in HCL):
provider "opnsense" {}
```

## Minimum OPNsense version

The provider targets **OPNsense 26.1.x** and is verified against the current
release series; **24.1+** is the practical minimum (the MVC API endpoints this
provider relies on must be present). Resources for optional features require the
corresponding plugin to be installed (for example `os-haproxy`, `os-acme-client`,
`os-frr`, `os-wireguard`).

## Required API user permissions

The API user must hold the privileges for the modules you manage. Grant the
matching *effective privileges* to the user (or a group) under
*System → Access*:

| Module / resources                              | Privilege area                |
|-------------------------------------------------|-------------------------------|
| `opnsense_firewall_*`                           | Firewall: Aliases / Rules / NAT |
| `opnsense_haproxy_*`                            | Services: HAProxy             |
| `opnsense_acme_*`                               | Services: ACME Client         |
| `opnsense_unbound_*` / `opnsense_dnsmasq_*`     | Services: Unbound / Dnsmasq DNS |
| `opnsense_kea_*` / `opnsense_dhcpv4_*`          | Services: Kea / DHCP          |
| `opnsense_quagga_*`                             | Services: FRR                 |
| `opnsense_ipsec_*` / `opnsense_wireguard_*` / `opnsense_openvpn_*` | VPN: IPsec / WireGuard / OpenVPN |
| `opnsense_interface_*` / `opnsense_system_*`    | Interfaces / System           |
| `opnsense_trust_*`                              | System: Trust (certificates)  |
| `opnsense_auth_*`                               | System: Access (users/groups) |

For automation, an administrator-equivalent API user is simplest; for least
privilege, grant only the areas above that your configuration touches.

## Support Matrix

The provider uses four public support states:

| Status | Meaning |
|--------|---------|
| Supported | Implemented in this provider release and documented in the Registry. |
| Coming | Provider-owned implementation remains after endpoint availability and durable Terraform semantics are sufficiently verified. No delivery date is implied. |
| Needs research | Published or suspected endpoint evidence exists, but target-version availability or durable Terraform semantics are not confirmed. |
| Upstream-blocked | OPNsense does not currently expose a stable usable API, or an upstream dependency must land first. |

Current provider baseline:

| Capability | Count |
|------------|------:|
| Resources | 97 |
| Data sources | 83 |
| Resource docs | 97 |
| Data source docs | 83 |
| Remaining data-source gaps | 15 |

| Supported today | Coming / provider-owned follow-up | Needs research | Upstream-blocked / OPNsense dependency |
|-----------------|------------------------------------|----------------|----------------------------------------|
| Firewall, HAProxy, ACME, DNS/Dnsmasq, DHCP/Kea, Dynamic DNS, VPN, FRR/Quagga, interfaces, auth, trust, syslog, Monit, cron, and traffic shaping resources. | Data-source parity for 15 singleton or sensitive special-case resources; interface LAGG with live member validation; system tunables/sysctl with safety/live-validation gate. | Kea DHCPv4 option/DDNS live endpoint conflict; HASync configuration `syncitems` model shape; HASync status/actions. | Interface base assignment/IP config/PPPoE, gateway groups, and system general settings until OPNsense exposes stable APIs. |

The full current Supported / Coming / Needs research / Upstream-blocked matrix is maintained in
the repository at
[`_bmad-output/planning-artifacts/support-matrix.md`](https://github.com/matthew-on-git/terraform-provider-opnsense/blob/main/_bmad-output/planning-artifacts/support-matrix.md).
Confirmed upstream blockers and the release-review workflow are documented in
[`docs/upstream-blocked.md`](https://github.com/matthew-on-git/terraform-provider-opnsense/blob/main/docs/upstream-blocked.md).

## Import and Migration Guidance

Most collection-backed resources use OPNsense UUIDs as Terraform IDs; singleton
or settings-style resources use stable provider-defined IDs documented on each
resource. For brownfield migration, import resources in dependency order so
references can be represented cleanly in HCL:

1. Independent primitives: aliases, categories, users, groups, certificates, DNS records.
2. Network foundations: interfaces, VLANs, VIPs, gateways, static routes.
3. Referenced policies: prefix lists, route maps, ACLs, health checks.
4. Dependent resources: firewall rules/NAT, HAProxy backends/frontends, VPN tunnels, ACME certificates.

After each import, run `terraform plan` and adjust HCL until Terraform reports no
changes. This confirms the imported state matches the live appliance before more
dependent resources are added.

For a full brownfield workflow with dependency-ordered examples, see
[`docs/migration-import.md`](https://github.com/matthew-on-git/terraform-provider-opnsense/blob/main/docs/migration-import.md).

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_key` (String, Sensitive) The API key for OPNsense authentication. Can also be set with the `OPNSENSE_API_KEY` environment variable.
- `api_secret` (String, Sensitive) The API secret for OPNsense authentication. Can also be set with the `OPNSENSE_API_SECRET` environment variable.
- `insecure` (Boolean) Whether to disable TLS certificate verification. Required for self-signed certificates. Defaults to `false`. Can also be set with the `OPNSENSE_ALLOW_INSECURE` environment variable.
- `uri` (String) The URI of the OPNsense appliance (e.g., `https://opnsense.example.com`). Can also be set with the `OPNSENSE_URI` environment variable.
