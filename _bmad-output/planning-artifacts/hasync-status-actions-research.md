---
title: HASync Status and Actions Research
date: 2026-06-09
author: BMad Dev
status: current
inputs:
  - OPNsense core API docs fetched 2026-06-09
  - HasyncStatusController.php from opnsense/core master
  - resource-gap-verification.md
  - core-config-gap-analysis.md
  - support-matrix.md
---

# HASync Status and Actions Research

Story 28.2 reviewed the OPNsense `core/hasync_status` API surface to classify each endpoint against Terraform semantics. No provider resource, data source, or action was implemented in this story.

## Source Evidence

Published OPNsense core API docs list `HasyncStatusController.php` under the Core API. Current source extends `ApiControllerBase` and exposes read-only status endpoints plus POST-only operational service actions that execute commands on the HA peer through configd.

## Endpoint Classification

| Endpoint | Method | Source behavior | Mutates state | Terraform fit | Recommendation |
|---|---|---|---|---|---|
| `/api/core/hasync_status/services` | `GET` | Runs `system ha services_cached`, decodes JSON, and returns `searchRecordsetBase` records with synthetic `uid` from service `name` and optional `id`. | No durable config mutation; result depends on HA cache/peer availability. | Read-only data-source candidate. | Keep as Needs research until live response shape and empty/error behavior are captured. |
| `/api/core/hasync_status/version` | `GET` | Runs `system ha exec version` and returns decoded JSON from the HA peer. | No durable config mutation; result depends on peer reachability. | Read-only data-source candidate. | Keep as Needs research until live response shape and peer-offline behavior are captured. |
| `/api/core/hasync_status/remote_service/{action}/{service}/{service_id}` | `GET` in published docs | Source does not define `remoteServiceAction` as a public action; it is a private helper used by POST actions. | Unknown from the documented route; source-backed helper would be operational if reachable. | Needs more live validation. | Do not implement from the documented GET shape until live route behavior and source/docs mismatch are resolved. |
| `/api/core/hasync_status/start/{service?}/{service_id?}` | `POST` | Calls remote service helper with `start` after `exec_sync` and `reload_templates`. | Yes, starts a remote service. | Operational action candidate, not resource/data source. | Do not implement until provider action support and product semantics are explicitly chosen. |
| `/api/core/hasync_status/stop/{service?}/{service_id?}` | `POST` | Calls remote service helper with `stop` after `exec_sync` and `reload_templates`. | Yes, stops a remote service. | Operational action candidate, not resource/data source. | Do not implement until provider action support and product semantics are explicitly chosen. |
| `/api/core/hasync_status/restart/{service?}/{service_id?}` | `POST` | Calls remote service helper with `restart` after `exec_sync` and `reload_templates`. | Yes, restarts a remote service. | Operational action candidate, not resource/data source. | Do not implement until provider action support and product semantics are explicitly chosen. |
| `/api/core/hasync_status/restart_all/{service?}/{service_id?}` | `POST` | Restarts all remote services returned by `system ha exec services`, or falls back to restart helper for a specific service. | Yes, restarts remote services. | Operational action candidate, not resource/data source. | Do not implement until provider action support and product semantics are explicitly chosen. |

## Response Shapes

Known from source:

- `services` returns a paginated/search recordset derived from `system ha services_cached`; each record receives `uid = name` or `name_id` when `id` is present.
- `version` returns decoded JSON from `system ha exec version`; exact keys require live validation.
- service actions return decoded JSON from `system ha exec` or `{"status":"ok","count":N}` for `restart_all`; exact error payloads require live validation.

Privileges were not discoverable from the fetched controller/model evidence. Treat all endpoints as requiring the same privileges as other core HA management APIs until verified on a live appliance.

## Risks

- Status responses depend on HA configuration and remote peer reachability; offline peer behavior must be captured before designing data-source diagnostics.
- Service actions mutate the remote peer immediately and do not represent durable desired state; modeling them as resources would be incorrect.
- Published docs list `remote_service` as `GET`, but current source exposes remote service operations through POST `start`, `stop`, `restart`, and `restart_all` public actions. That mismatch needs live route validation before any action design.

## Recommendation

- Keep HASync status/actions classified as **Needs research**.
- Do not implement durable Terraform resources for any `core/hasync_status` endpoint.
- Treat `services` and `version` as future read-only data-source candidates only after live response and error-shape validation.
- Treat `start`, `stop`, `restart`, and `restart_all` as future Terraform action candidates only after an explicit product decision and provider-framework action support review.
- No immediate implementation follow-up story is added because live HA-peer validation and action-semantics decisions are prerequisites.
