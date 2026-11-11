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
| System general settings | Blocks provider management of legacy general system settings that are not yet exposed through a stable MVC API. | No stable OPNsense API in the current target release (`stable/26.1`). As of 2026-06-14, published core API docs list no durable system general settings endpoint; `core/system` is action/status-only; `Core/Api/SettingsController.php`, `Core/System.xml`, and `Core/Settings.xml` were not found; `core/initial_setup` exists but is wizard-only, memory-mounted, and has broad side effects, so it is unsafe for day-2 Terraform management. | Watch OPNsense release notes and System Settings MVC progress. Adopt only after stable get/set semantics and durable configuration fields are available. |
| Kea ordered parent-option linkage | Blocks ordered UUID links from existing DHCPv4 subnet and reservation owners to generic options. Unordered selected maps would mis-state last-match precedence. | Released core 26.7.8 has the multi `option` relation, but `get_subnet` and `get_reservation` return description-sorted selected maps. The relations do not set `Sorted`, and no action returns an ordered UUID list. | Re-check the next stable core tag. Adopt only after a released parent GET exposes stored order without raw XML or JSON object key order. |

## Stable API Available; Provider Validation Pending

As of 2026-10-25, gateway groups remain **Coming**, not upstream-blocked, on
OPNsense 26.7.7. Isolated direct-API CRUD/tier/trigger/pool/clear/rename,
reference-protected deletion and exact route restoration were exercised. See the
[executed API contract](../_bmad-output/planning-artifacts/gateway-group-api-validation.md).
Story 31.0m owns the missing provider resource/data source and real Terraform
acceptance. These probes do not establish provider support or actual multi-WAN
traffic failover.

Generic Kea options and DDNS settings also have
[disabled-service API proof](../_bmad-output/planning-artifacts/kea-option-ddns-api-validation.md).
Stories 31.0n/o shipped those owners. Ordered parent-option linkage is
upstream-blocked on released core 26.7.8: parent GET returns a
description-sorted selected map and does not expose stored UUID order.
Grid-search CSV and raw XML are not a supported order contract. See the
[ordered-read research](../_bmad-output/planning-artifacts/kea-ordered-option-linkage-api-research.md).

## Not Upstream-Blocked Yet

Some gaps have published or suspected API evidence but still need payload, lifecycle, or target-version verification. These remain **Needs research**, not upstream-blocked:

| Domain | Why it is not in this register |
|---|---|
| PPPoE assignment semantics | Base assignment/IP settings are supported by `opnsense_interface_assignment` on OPNsense 26.7.6+ after disposable-appliance lifecycle/drift/restoration proof. PPPoE remains research-only because its model, secret, and lifecycle semantics were outside that implementation. See the committed [release probe](../_bmad-output/planning-artifacts/opnsense-assignment-api-release-probe.json). |
| System tunables / sysctl | Current OPNsense docs/source expose persistent `core/tunables` item CRUD/search and `reconfigure`; `opnsense_system_tunable` shipped in v0.3.0 and remains listed here only as an example of a former research item that moved out of this register. |
| HASync configuration | `core/hasync` endpoints exist, but `syncitems` uses a dynamic `JsonKeyValueStoreField`; request/response shape needs verification before a safe resource can be designed. |
| HASync status/actions | Endpoints appear operational/status-oriented rather than durable configuration. Data source or action semantics need a separate product decision. |

## Maintenance Workflow

1. After each OPNsense major release, re-check the published API docs and relevant source models/controllers for every blocked domain.
2. If a stable endpoint appears, capture endpoint paths, wrapper keys, model fields, defaults, required constraints, and reconfigure/apply behavior before moving the item to Coming.
3. If no endpoint exists, leave the item blocked and refresh the upstream status or tracked issue/PR reference.
4. If a domain proves intentionally unsupported or unsafe for Terraform management, reclassify it as Not planned with evidence.
5. Keep [`support-matrix.md`](../_bmad-output/planning-artifacts/support-matrix.md), [`core-config-gap-analysis.md`](../_bmad-output/planning-artifacts/core-config-gap-analysis.md), and the provider index ([`docs/index.md`](index.md) / [`templates/index.md.tmpl`](../templates/index.md.tmpl)) aligned whenever an item changes classification.
