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

# Configure the OPNsense provider. Credentials and retry settings may also be
# supplied via the OPNSENSE_URI, OPNSENSE_API_KEY, OPNSENSE_API_SECRET,
# OPNSENSE_ALLOW_INSECURE, OPNSENSE_RETRIES, OPNSENSE_MIN_BACKOFF, and
# OPNSENSE_MAX_BACKOFF environment variables (HCL values take priority).
provider "opnsense" {
  uri        = "https://opnsense.example.com"
  api_key    = var.opnsense_api_key
  api_secret = var.opnsense_api_secret

  # Optional retry tuning (whole seconds):
  # retries     = 3
  # min_backoff = 1
  # max_backoff = 30
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

| Setting         | HCL argument  | Environment variable      |
|-----------------|---------------|---------------------------|
| Base URL        | `uri`         | `OPNSENSE_URI`            |
| API key         | `api_key`     | `OPNSENSE_API_KEY`        |
| API secret      | `api_secret`  | `OPNSENSE_API_SECRET`     |
| Skip TLS verify | `insecure`    | `OPNSENSE_ALLOW_INSECURE` |
| Retries         | `retries`     | `OPNSENSE_RETRIES`        |
| Min backoff     | `min_backoff` | `OPNSENSE_MIN_BACKOFF`    |
| Max backoff     | `max_backoff` | `OPNSENSE_MAX_BACKOFF`    |

`retries` is a retry count; `min_backoff` and `max_backoff` are whole seconds.
They default to `3`, `1`, and `30` respectively and must be between `1` and
`2147483647` (`retries = 0` is not supported). Invalid environment values are
ignored so the defaults apply. `api_key` and `api_secret` are marked sensitive.
On configuration the provider validates credentials against
`/api/core/menu/search` and reports a clear diagnostic if authentication fails.

### Retry behavior

The provider retries transient API failures for every resource and data source.
`retries = 3` permits up to four attempts (the initial request plus three
retries). Waits begin at `min_backoff` and double up to `max_backoff`; a
`Retry-After` header on a `429` or `503` response can extend a wait beyond
`max_backoff`. A `min_backoff` greater than `max_backoff` is accepted and the
computed wait is clamped to `max_backoff`.

Retries also apply to writes. If OPNsense processes a write but the response is
lost or fails, a retry can repeat that write. The upper bounds are overflow
guards, not recommended values; large retry/backoff settings can block Terraform
for a very long time.

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

Released implementation inventory through v0.11.0 (reviewed 2026-11-03):

| Capability | Count |
|------------|------:|
| Resources | 112 |
| Data sources | 98 |
| Resource docs | 112 |
| Data source docs | 98 |
| Ephemeral resources / docs | 1 / 1 |
| Remaining data-source gaps | 16 |

| Supported today | Coming / provider-owned follow-up | Needs research | Upstream-blocked / OPNsense dependency |
|-----------------|------------------------------------|----------------|----------------------------------------|
| Firewall, HAProxy, ACME, DNS/Dnsmasq, DHCP/Kea, Dynamic DNS, VPN, FRR/Quagga, interface assignment/IP on OPNsense `26.7.6+` plus LAGG and other interface types, system tunables and gateway groups, Kea generic DHCPv4 options and DDNS general settings, auth, trust, syslog, Monit, cron, and traffic shaping resources. | Data-source parity for 16 singleton or sensitive special-case resources; remaining audited field/policy parity. | Kea ordered parent-option linkage; HASync configuration `syncitems` model shape; HASync status/actions; PPPoE semantics for future interface settings. | System general settings: action/status plus setup-wizard evidence does not establish durable target-release get/set support. |

IPv6 identity-association (`idassoc6`) and legacy track-interface (`track6`) adoption/updates are published in v0.11.0 after [PR #77](https://github.com/matthew-on-git/terraform-provider-opnsense/pull/77). Six optional/computed delegation fields preserve omitted values and unowned settings. Contributor-reported OPNsense 26.7.8_2 / OpenTofu 1.13.2 import and no-change refresh coverage was not independently reproduced and does not establish activation, packet behavior or full write/restore acceptance. Assignment creation, automatic mode migration and upstream DHCPv6-client ownership are not included.

OPNsense core and plugin versions are separate compatibility gates. Released FRR 2.0 compatibility covers the verified installed-plugin/model and migration/lifecycle subgate; advanced BGP field/policy parity remains a separate follow-up, not a blanket compatibility claim for every plugin release. Gateway-group ownership is published in 0.9.0 with release assets and Registry version verified; populated first-tier clearing remains upstream-limited. Kea option/DDNS owners are published in 0.10.0 with separate final disabled-appliance Core lifecycle/restoration evidence; ordered parent linkage, active packet/DNS behavior and OpenTofu lifecycle validation remain separate. Their historical live-test digests remain exact to the recorded checkpoints; the release/tracking-only integration revision is not advertised as a new appliance run.

The full current Supported / Coming / Needs research / Upstream-blocked matrix is maintained in
the repository at
[`_bmad-output/planning-artifacts/support-matrix.md`](https://github.com/matthew-on-git/terraform-provider-opnsense/blob/main/_bmad-output/planning-artifacts/support-matrix.md).
Confirmed upstream blockers and the release-review workflow are documented in
[`docs/upstream-blocked.md`](https://github.com/matthew-on-git/terraform-provider-opnsense/blob/main/docs/upstream-blocked.md).

OpenTofu provider and signing-key registrations are merged; its versions API lists v0.10.1 and download metadata includes the release-signing key. This establishes registry availability, not independently tested OpenTofu installation or resource lifecycle parity.

## Import and Migration Guidance

Most collection-backed resources use OPNsense UUIDs as Terraform IDs; singleton
or settings-style resources use stable provider-defined IDs documented on each
resource. For brownfield migration, import resources in dependency order so
references can be represented cleanly in HCL:

1. Independent primitives: aliases, categories, users, groups, DNS records, and ACME accounts/challenges.
2. Network foundations: interfaces, VLANs, VIPs, gateways, static routes, and trust/CA material.
3. Referenced policies and reusable objects: prefix lists, route maps, HAProxy ACLs, health checks, mapfiles, backends, and ACME certificates.
4. Dependent resources: firewall rules/NAT, HAProxy actions/frontends, VPN tunnels, and certificate-bound listeners.

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
- `max_backoff` (Number) Maximum wait in seconds between retries. Defaults to `30`. Can also be set with the `OPNSENSE_MAX_BACKOFF` environment variable.
- `min_backoff` (Number) Minimum wait in seconds between retries. Defaults to `1`. Can also be set with the `OPNSENSE_MIN_BACKOFF` environment variable.
- `retries` (Number) Maximum number of retries for transient OPNsense API failures. Defaults to `3`. Can also be set with the `OPNSENSE_RETRIES` environment variable.
- `uri` (String) The URI of the OPNsense appliance (e.g., `https://opnsense.example.com`). Can also be set with the `OPNSENSE_URI` environment variable.
