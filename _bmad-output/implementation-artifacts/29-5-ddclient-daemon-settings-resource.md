---
baseline_commit: fbaf085e8287ae4f00f786484cd1ab622d77716a
---

# Story 29.5: ddclient Daemon Settings Resource

Status: done

## Story

As an operator,
I want an `opnsense_ddclient_settings` singleton resource,
So that I can manage the Dynamic DNS daemon's general configuration (enabled, backend, update interval, verbosity, IPv6) — not just individual accounts.

## Context

The provider ships `opnsense_ddclient_account` (per-host DDNS entries) but has **no resource for the os-ddclient general/daemon settings**. Story 9.5 ("Dynamic DNS Provider Configuration") was cancelled as "redundant — 'provider' is the `service` field on the dyndns account." That reasoning conflated two different things: the per-account *provider/service* (correctly on the account) versus the **daemon-level settings** (`enabled`, `daemon_delay`/update interval, `backend` = ddclient vs native, `verbose`, `allowipv6`). The latter is a genuine, untracked gap. The downstream appliance's `dyndns` role sets these via `/api/dyndns/settings/set` + reconfigure; without this resource, an operator must toggle the DynDNS service on and set the interval by hand.

## Acceptance Criteria

1. **Given** the `os-ddclient` plugin is installed
   **When** the operator defines the singleton `opnsense_ddclient_settings` with `enabled`, `backend`, `interval`, etc.
   **Then** the settings are written via `POST /api/dyndns/settings/set` and applied via `/api/dyndns/service/reconfigure`

2. **And** the schema includes at least: `enabled` (Bool), `backend` (String: `ddclient` | `opnsense`/native), `interval` (Number; the `daemon_delay` seconds), `verbose` (Bool), `allow_ipv6` (Bool)

3. **And** it is implemented as a **singleton** (no UUID) using the existing `GetSingleton`/`UpdateSingleton` client primitives (Story 13.1), consistent with `unbound_general`, `quagga_general`, etc.

4. **And** `Read` reflects live state; `terraform plan` with no changes shows "No changes"; `terraform import` brings the singleton under management (fixed/synthetic ID, as other `_general`/settings resources do)

5. **And** `Delete` is a no-op-with-state-removal or a reset-to-defaults (match the convention used by existing singleton settings resources) — document which; do not error on destroy

6. **And** an acceptance test sets settings, asserts idempotent re-plan, and composes cleanly with at least one `opnsense_ddclient_account`

## Tasks / Subtasks

- [x] Task 1: `settings_model.go` — `settingsAPIResponse`/`Request` + `SettingsResourceModel`; bool "0"/"1" and number conversions (AC: #1, #2)
- [x] Task 2: `settings_schema.go` — fields per AC #2; sensible defaults; `backend` OneOf validator (AC: #2)
- [x] Task 3: `settings_resource.go` — singleton CRUD via `GetSingleton`/`UpdateSingleton`; `ddclientSettingsReqOpts` (AC: #1, #3)
  - [x] 3.1 Define the no-UUID get/set semantics (Story 13.1 primitives)
  - [x] 3.2 ImportState with the singleton convention used elsewhere (AC: #4)
- [x] Task 4: Register `newSettingsResource` in `internal/service/ddclient/exports.go`
- [x] Task 5: `settings_resource_test.go` — set + idempotency + compose with account (AC: #6)
- [x] Task 6: `settings_data_source.go` (optional, parity) + schema test
- [x] Task 7: Examples + `templates/resources/ddclient_settings.md.tmpl`
- [x] Task 8: `make check`

## Dev Notes

### OPNsense DynDNS settings API

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Read | GET | `/api/dyndns/settings/get` | singleton — no UUID; unwrap `ddclient.general` monad |
| Update | POST | `/api/dyndns/settings/set` | singleton — no UUID |
| Reconfigure | POST | `/api/dyndns/service/reconfigure` | apply (restarts daemon) |
| Status | GET | `/api/dyndns/service/status` | optional health read |

```go
var ddclientSettingsReqOpts = opnsense.ReqOpts{
    GetEndpoint:         "/api/dyndns/settings/get",
    UpdateEndpoint:      "/api/dyndns/settings/set",
    ReconfigureEndpoint: "/api/dyndns/service/reconfigure",
    Monad:               "ddclient.general",
}
```

### Schema

| API Field | TF attr | TF type | Notes |
|-----------|---------|---------|-------|
| `enabled` | `enabled` | Bool | "0"/"1" |
| `backend` | `backend` | String | `ddclient` or native `opnsense`; upstream default `opnsense` |
| `daemon_delay` | `interval` | Number (Int64) | seconds between updates; upstream range 1..86400; default 300 |
| `verbose` | `verbose` | Bool | |
| `allowipv6` | `allow_ipv6` | Bool | |

Confirmed from upstream `SettingsController` and `DynDNS.xml`: the settings API wraps fields under `ddclient.general`, uses `backend` values `ddclient`/`opnsense`, defaults `backend` to `opnsense`, and constrains `daemon_delay` to 1..86400 seconds.

### Singleton pattern (reuse, do not reinvent)

Story 13.1 added `GetSingleton[K]` / `UpdateSingleton[K]` precisely for get/set endpoints with no `/{id}`. Use them. Mirror an existing singleton resource (`internal/service/unbound/general_*.go` or `quagga/general_*.go`) for the ImportState/Delete conventions so behavior is consistent across the provider.

### Why 9.5's cancellation was wrong (for the retro/record)

9.5 was scoped as "provider configuration" and dismissed because per-account `service` already selects the provider. But daemon settings (`enabled`/`daemon_delay`/`backend`/`verbose`/`allowipv6`) are a separate OPNsense object (`/dyndns/settings`) with no account-level equivalent. This story covers that object, under a clearer name (`_settings`, not `_provider`).

### What NOT to build

- No account fields — `opnsense_ddclient_account` owns those
- No per-provider credential modeling — that is account-level

### References

- [Source: 9-5 cancellation note in sprint-status.yaml — "CANCELLED — redundant"]
- [Source: 13-1-singleton-get-set-client-support.md — GetSingleton/UpdateSingleton]
- [Pattern: internal/service/unbound/general_*.go, internal/service/quagga/general_*.go]
- [Downstream driver: opnsense-manager ansible/roles/dyndns/tasks/main.yml — /dyndns/settings/set]

## Dev Agent Record

### Implementation Plan

- Mirror existing singleton resource behavior: constant ID, `UpdateSingleton` on create/update, `GetSingleton` readback, no-op delete, and fixed-ID import.
- Map OPNsense `general` settings fields into Terraform attributes: `enabled`, `backend`, `interval`, `verbose`, and `allow_ipv6`.
- Add optional data-source parity for the singleton settings object.
- Document destroy semantics and daemon/account composition in examples.

### Debug Log

- Red phase: `go test ./internal/service/ddclient` failed because settings model/resource symbols did not exist.
- Implemented `opnsense_ddclient_settings` resource and data source with singleton `/api/dyndns/settings/get` and `/api/dyndns/settings/set` endpoints.
- Added model conversion, registration/schema, data-source read, and acceptance scaffold coverage.
- Generated Registry docs with containerized `go generate ./tools`.
- Updated current-facing support counts to 100 resources / 86 data sources.
- Verification passed: `go test ./pkg/opnsense ./internal/service/ddclient`.
- Required project gate passed: `make check`.
- Review patches applied: switched singleton monad to `ddclient.general`, validated the data-source singleton ID, matched upstream default `backend = opnsense`, and constrained `interval` to 1..86400 seconds.
- Review verification passed: `go test ./pkg/opnsense ./internal/service/ddclient` and `make check`.

### Completion Notes

- Added `opnsense_ddclient_settings` singleton resource with no-op destroy/state removal semantics.
- Added `opnsense_ddclient_settings` data source for parity.
- Added docs/examples for settings and account composition.
- Dynamic DNS support now includes both per-account entries and daemon-level settings.

## File List

- `internal/service/ddclient/settings_model.go`
- `internal/service/ddclient/settings_schema.go`
- `internal/service/ddclient/settings_resource.go`
- `internal/service/ddclient/settings_data_source.go`
- `internal/service/ddclient/settings_model_test.go`
- `internal/service/ddclient/data_source_schema_test.go`
- `internal/service/ddclient/account_resource_test.go`
- `internal/service/ddclient/exports.go`
- `examples/resources/opnsense_ddclient_settings/resource.tf`
- `examples/resources/opnsense_ddclient_settings/import.sh`
- `examples/data-sources/opnsense_ddclient_settings/data-source.tf`
- `templates/resources/ddclient_settings.md.tmpl`
- `templates/data-sources/ddclient_settings.md.tmpl`
- `docs/resources/ddclient_settings.md`
- `docs/data-sources/ddclient_settings.md`
- `docs/index.md`
- `templates/index.md.tmpl`
- `README.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/implementation-artifacts/29-5-ddclient-daemon-settings-resource.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/post-release-epics.md`

## Change Log

- 2026-06-10: Implemented ddclient daemon settings singleton resource/data source, docs/examples, support-count updates, and validation.

### Review Findings

- [x] [Review][Patch] Use the ddclient.general singleton wrapper confirmed by upstream `SettingsController` and `DynDNS.xml` [internal/service/ddclient/settings_resource.go:25]
- [x] [Review][Patch] Validate the singleton data-source ID instead of accepting arbitrary values [internal/service/ddclient/settings_data_source.go:28]
- [x] [Review][Patch] Match upstream ddclient settings model defaults and interval bounds (`backend = opnsense`, `daemon_delay` 1..86400) [internal/service/ddclient/settings_schema.go:33]
