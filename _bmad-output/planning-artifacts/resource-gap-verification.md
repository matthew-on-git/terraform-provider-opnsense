---
title: Remaining Buildable Resource Gap Verification
date: 2026-06-02
author: BMad Dev
status: current
inputs:
  - OPNsense published API docs fetched 2026-06-02
  - core-config-gap-analysis.md
  - support-matrix.md
  - repository implementation inventory
---

# Remaining Buildable Resource Gap Verification

This artifact records Story 27.1 verification for remaining provider-owned resource gaps and Story 27.2 implementation status for the selected Dnsmasq batch.

## Verification Summary

| Candidate | Final classification | Evidence | Handoff |
|---|---|---|---|
| Firewall source NAT | Supported | Existing provider resource `opnsense_firewall_nat_outbound` uses `/api/firewall/source_nat/*_rule`; generated docs describe it as outbound NAT/source NAT. Published OPNsense API docs expose `firewall/source_nat` CRUD/search/toggle endpoints. | Do not create a second source NAT resource. Improve naming/docs only if needed. |
| Interface LAGG | Supported | Published OPNsense API docs expose `interfaces/lagg_settings` item CRUD/search/reconfigure endpoints and `Lagg.xml`. Repository now has generated `opnsense_interface_lagg` resource/data source code. | Live-validated against Vagrant with selectable `em4`/`em5` member interfaces. |
| Kea DHCPv4 option | Needs research | Published OPNsense API docs expose `kea/dhcpv4` option item CRUD/search endpoints, but prior live probing recorded endpoint-not-found on the tested build. | Re-probe target appliance/version before implementation; classify as Coming only after live evidence confirms endpoint availability. |
| Kea DDNS | Needs research | Published OPNsense API docs expose singleton `kea/ddns/get` and `kea/ddns/set` endpoints with model `KeaDdns.xml`, but prior live probing recorded endpoint-not-found on the tested build. | Re-probe target appliance/version before implementation; likely singleton resource if available. |
| Dnsmasq host/domain/range/option/tag/boot | Supported | Story 27.2 added UUID item resources and data sources for `opnsense_dnsmasq_host`, `opnsense_dnsmasq_domain`, `opnsense_dnsmasq_tag`, `opnsense_dnsmasq_range`, `opnsense_dnsmasq_option`, and `opnsense_dnsmasq_boot`. Published OPNsense API docs expose matching item CRUD/search endpoints under `dnsmasq/settings`; all mutations use `dnsmasq/service/reconfigure`. | No further Dnsmasq item-resource handoff remains for these six families. |
| Unbound forward | Supported | Existing provider resource `opnsense_unbound_domain_override` uses `/api/unbound/settings/add_forward`, `get_forward`, `set_forward`, `del_forward`, and `search_forward`. Generated docs describe it as a forwarding rule. | Do not create a second Unbound forward resource. Align planning docs to the existing resource name. |
| OSPF area | Supported | Story 27.3 added UUID item resource and data source `opnsense_quagga_ospf_area` from upstream `OSPF.xml`. Published Quagga API docs expose `quagga/ospfsettings` area item CRUD/search/toggle endpoints. | No remaining OSPF area handoff. |
| HASync configuration | Needs research | Published core API docs expose singleton `core/hasync/get`, `core/hasync/set`, and `core/hasync/reconfigure` using `Hasync.xml`, but Story 27.4 model review found `syncitems` is a dynamic `JsonKeyValueStoreField` populated by `system ha options`. Current generated resource field types cannot safely represent that dynamic key/value-store shape without verified API payload semantics. | Do not implement via the generator until request/response shape for `syncitems` is verified and the provider has a safe Terraform representation for dynamic HA sync item selections; a hand-written resource remains possible after that research. |
| HASync status/actions | Needs research | Story 28.2 confirmed `core/hasync_status/services` and `version` are read-only status endpoints and `start`, `stop`, `restart`, and `restart_all` are POST operational actions. Published `remote_service` GET docs do not match the current source's private helper shape. | Treat `services` and `version` as future data-source candidates only after live response validation. Treat service operations as action candidates only after product/framework decision; never model these endpoints as durable resources. |
| System tunables / sysctl | Supported | Story 28.4 added `opnsense_system_tunable` resource and data source after live validation confirmed `core/tunables` add/get/set/delete/reconfigure behavior with wrapper `sysctl`. | No remaining tunables/sysctl implementation handoff; keep safety warnings in docs because tunables can affect kernel/network behavior. |

## Endpoint Details

| Candidate | Endpoints | CRUD/search behavior | Wrapper key | Lifecycle |
|---|---|---|---|---|
| `opnsense_firewall_nat_outbound` / source NAT | `POST /api/firewall/source_nat/add_rule`, `GET /api/firewall/source_nat/get_rule/{uuid?}`, `POST /api/firewall/source_nat/set_rule/{uuid}`, `POST /api/firewall/source_nat/del_rule/{uuid}`, `GET,POST /api/firewall/source_nat/search_rule`, `POST /api/firewall/source_nat/apply` | Confirmed in repository and OPNsense docs. | `rule` | UUID item resource; currently supported. |
| Interface LAGG | `POST /api/interfaces/lagg_settings/add_item`, `GET /api/interfaces/lagg_settings/get_item/{uuid?}`, `POST /api/interfaces/lagg_settings/set_item/{uuid}`, `POST /api/interfaces/lagg_settings/del_item/{uuid}`, `GET,POST /api/interfaces/lagg_settings/search_item`, `POST /api/interfaces/lagg_settings/reconfigure` | Published API docs confirm standard mutable item pattern; live member-interface behavior was validated with Vagrant `em4`/`em5`. | `lagg` request/response wrapper confirmed by live API. | Supported UUID item resource. |
| Kea DHCPv4 option | `POST /api/kea/dhcpv4/add_option`, `GET /api/kea/dhcpv4/get_option/{uuid?}`, `POST /api/kea/dhcpv4/set_option/{uuid}`, `POST /api/kea/dhcpv4/del_option/{uuid}`, `GET,POST /api/kea/dhcpv4/search_option`, `POST /api/kea/service/reconfigure` | Published API docs list item endpoints; prior live appliance evidence conflicts, so buildability is unconfirmed. | `option` from published endpoint naming; confirm model wrapper only after live endpoint availability is resolved. | Needs research; UUID candidate only after live recheck passes. |
| Kea DDNS | `GET /api/kea/ddns/get`, `POST /api/kea/ddns/set`, `POST /api/kea/service/reconfigure` | Published API docs list singleton get/set; prior live appliance evidence conflicts, so buildability is unconfirmed. | `ddns` or model-root wrapper unresolved until live endpoint availability is resolved. | Needs research; singleton candidate only after live recheck passes. |
| Dnsmasq host/domain/range/option/tag/boot | `POST /api/dnsmasq/settings/add_{host,domain,range,option,tag,boot}`, `GET /api/dnsmasq/settings/get_{host,domain,range,option,tag,boot}/{uuid?}`, `POST /api/dnsmasq/settings/set_{host,domain,range,option,tag,boot}/{uuid}`, `POST /api/dnsmasq/settings/del_{host,domain,range,option,tag,boot}/{uuid}`, `GET,POST /api/dnsmasq/settings/search_{host,domain,range,option,tag,boot}`, `POST /api/dnsmasq/service/reconfigure` | Published API docs confirm item endpoints for all six implemented families. | `host`, `domainoverride` for domain, `range`, `option`, `tag`, `boot` | UUID item resources; currently supported. |
| Unbound forward | `POST /api/unbound/settings/add_forward`, `GET /api/unbound/settings/get_forward/{uuid?}`, `POST /api/unbound/settings/set_forward/{uuid}`, `POST /api/unbound/settings/del_forward/{uuid}`, `GET,POST /api/unbound/settings/search_forward`, `POST /api/unbound/service/reconfigure` | Confirmed in repository as `opnsense_unbound_domain_override`; published API docs confirm forward endpoints. | `forward` | UUID item resource; currently supported under domain override naming. |
| OSPF area | `POST /api/quagga/ospfsettings/add_area`, `GET /api/quagga/ospfsettings/get_area/{uuid?}`, `POST /api/quagga/ospfsettings/set_area/{uuid}`, `POST /api/quagga/ospfsettings/del_area/{uuid}`, `GET,POST /api/quagga/ospfsettings/search_area`, `POST /api/quagga/service/reconfigure` | Published Quagga API docs and `OspfsettingsController.php` confirm item endpoints. | `area` | UUID item resource; currently supported. |
| HASync configuration | `GET /api/core/hasync/get`, `POST /api/core/hasync/set`, `POST /api/core/hasync/reconfigure` | Published core API docs and `HasyncController.php` support singleton get/set/reconfigure, but live request/response payload shape was not captured. | Likely `hasync`; confirm from live/source payload before implementation. | Needs research; singleton candidate only after `syncitems` dynamic key/value-store representation is safe. |
| HASync status/actions | `GET /api/core/hasync_status/services`, `GET /api/core/hasync_status/version`, documented `GET /api/core/hasync_status/remote_service/{action}/{service}/{service_id}`, plus source-backed `POST /api/core/hasync_status/{start,stop,restart,restart_all}/{service?}/{service_id?}` | Published docs and `HasyncStatusController.php` confirm status/action surface, but documented `remote_service` shape differs from current source public actions. | None for read-only status; service/action parameters for POST operations. | Needs research; data-source candidates for `services`/`version` after live validation, action candidates only after product/framework decision. |
| System tunables / sysctl | `POST /api/core/tunables/add_item`, `GET /api/core/tunables/get_item/{uuid?}`, `POST /api/core/tunables/set_item/{uuid}`, `POST /api/core/tunables/del_item/{uuid}`, `GET,POST /api/core/tunables/search_item`, `POST /api/core/tunables/reconfigure`, `POST /api/core/tunables/reset` | Published core API docs and `TunablesController.php` confirm item CRUD/search; Story 28.4 live validation confirmed add/read/update/delete/reconfigure on the target appliance. `reset` restores factory tunables and is not used for normal resource lifecycle. | `sysctl` for add/get/set item; model path `item`. | Supported UUID item resource and data source. |

## Story 27.2 Recommendation

Remaining implementation targets after the Story 27.3 OSPF area resource:

1. `opnsense_interface_lagg` if a live appliance with assignable member interfaces is available.
2. HASync singleton only after `syncitems` request/response shape and Terraform representation are verified.

Avoid duplicate resources for source NAT and Unbound forward; both are already supported under existing provider names.
