---
page_title: "Upstream-Blocked Register - opnsense"
subcategory: "Guides"
description: |-
  Tracks OPNsense API gaps that currently block Terraform provider resources.
---

# Upstream-Blocked Register

This register lists provider gaps that cannot be implemented safely until OPNsense exposes a stable API. These are not omitted because of provider neglect; each item needs upstream API availability or an upstream model change before Terraform can manage it predictably.

## Blocked Domains

| Domain | Provider impact | Current upstream status | Maintenance action |
|---|---|---|---|
| Interface assignment / IP config / PPPoE | Blocks a base interface-assignment resource and direct management of physical interface IP/PPP settings. Existing provider resources cover API-backed interface types such as VLAN, bridge, GIF, GRE, loopback, neighbor, and VXLAN. | No stable OPNsense API in the current target release. OPNsense PR #8436 is the tracked upstream item for interface MVC/API work. | Review PR #8436 and related release notes after each OPNsense major release. When an API ships, verify payload shape and lifecycle before creating implementation stories. |
| Gateway group | Blocks a gateway-group resource for multi-WAN failover/load-balancing groups. Existing provider resources cover individual gateways and static routes. | No usable gateway-group MVC/API endpoint is currently tracked. | Recheck published API docs and OPNsense source after each major release; if still absent and user demand remains, consider authoring an upstream MVC API contribution. |
| System general settings | Blocks provider management of legacy general system settings that are not yet exposed through a stable MVC API. | Legacy/non-MVC API status; OPNsense System Settings MVC work remains the upstream dependency. | Watch OPNsense release notes and System Settings MVC progress. Adopt only after stable get/set semantics and durable configuration fields are available. |

## Not Upstream-Blocked Yet

Some gaps have published or suspected API evidence but still need payload, lifecycle, or target-version verification. These remain **Needs research**, not upstream-blocked:

| Domain | Why it is not in this register |
|---|---|
| System tunables / sysctl | Current OPNsense docs/source expose persistent `core/tunables` item CRUD/search and `reconfigure`; this is now Coming with a safety/live-validation gate rather than upstream-blocked. |
| HASync configuration | `core/hasync` endpoints exist, but `syncitems` uses a dynamic `JsonKeyValueStoreField`; request/response shape needs verification before a safe resource can be designed. |
| HASync status/actions | Endpoints appear operational/status-oriented rather than durable configuration. Data source or action semantics need a separate product decision. |
| Kea DHCPv4 option / Kea DDNS | Published docs conflict with earlier live endpoint-not-found evidence. Re-probe target appliances before implementation. |

## Maintenance Workflow

1. After each OPNsense major release, re-check the published API docs and relevant source models/controllers for every blocked domain.
2. If a stable endpoint appears, capture endpoint paths, wrapper keys, model fields, defaults, required constraints, and reconfigure/apply behavior before moving the item to Coming.
3. If no endpoint exists, leave the item blocked and refresh the upstream status or tracked issue/PR reference.
4. If a domain proves intentionally unsupported or unsafe for Terraform management, reclassify it as Not planned with evidence.
5. Keep [`support-matrix.md`](../_bmad-output/planning-artifacts/support-matrix.md), [`core-config-gap-analysis.md`](../_bmad-output/planning-artifacts/core-config-gap-analysis.md), and the provider index ([`docs/index.md`](index.md) / [`templates/index.md.tmpl`](../templates/index.md.tmpl)) aligned whenever an item changes classification.
