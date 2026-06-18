---
baseline_commit: fbaf085e8287ae4f00f786484cd1ab622d77716a
---

# Story 29.2: HAProxy Map File Resource

Status: done

## Story

As an operator,
I want to manage HAProxy map files as an `opnsense_haproxy_mapfile` resource,
So that I can maintain a domain → backend routing table that a `map_use_backend` action consults, enabling host-based routing of many domains through a single HTTPS frontend.

## Context

The canonical OPNsense HAProxy routing pattern for a multi-tenant edge is: one HTTPS frontend → a `map_use_backend` action → a **map file** whose lines are `<host> <backend-name>`. The provider has frontend, backend, acl, server, and healthcheck resources, but **no map file resource** — it is not mentioned anywhere in the planning artifacts. Without it, the `map_use_backend` action from Story 29.1 has nothing to reference, and the only routing available is one-ACL-per-domain (which does not scale and is not how the downstream appliance is built).

## Acceptance Criteria

1. **Given** the `os-haproxy` plugin is installed
   **When** the operator defines an `opnsense_haproxy_mapfile` with a `name`, `type`, and `content`
   **Then** the map file is created via `/api/haproxy/settings/addMapFile` and is referenceable by a `map_use_backend` action (Story 29.1)

2. **And** the schema includes: `name`, `description`, `type` (e.g. `map_str` / domain map), and `content` (the multi-line `<key> <value>` body)

3. **And** `content` updates are a full-replace in-place update (no destroy-and-recreate); a no-op `content` produces "No changes"

4. **And** the resource is tolerant of OPNsense's content normalization (trailing newline / whitespace) so repeated applies do not show perpetual drift — normalize in `fromAPI`/state

5. **And** `terraform import` works by UUID; an acceptance test creates a map file, references it from a `map_use_backend` action, and verifies lifecycle + `CheckDestroy`

6. **And** docs/example show a domain-map with several `host backend` lines (mirroring the appliance's `domain-map`)

## Tasks / Subtasks

- [x] Task 1: `mapfile_model.go` — `mapfileAPIResponse` / `mapfileAPIRequest` / `MapfileResourceModel`; `toAPI`/`fromAPI` (AC: #1, #2)
  - [x] 1.1 Handle `content` as a single multi-line string; preserve line order
  - [x] 1.2 Normalize whitespace/trailing newline in `fromAPI` to prevent drift (AC: #4)
- [x] Task 2: `mapfile_schema.go` — `name`, `description`, `type` (OneOf validator), `content`; no RequiresReplace (AC: #2, #3)
- [x] Task 3: `mapfile_resource.go` — CRUD + ImportState with `mapfileReqOpts` (AC: #1, #5)
- [x] Task 4: Register `newMapfileResource` in `exports.go`
- [x] Task 5: `mapfile_resource_test.go` — mapfile + map_use_backend action chain; assert no-drift second plan (AC: #4, #5)
- [x] Task 6: `mapfile_data_source.go` + schema-parity test
- [x] Task 7: Examples + `templates/resources/haproxy_mapfile.md.tmpl`
- [x] Task 8: `make check`

### Review Findings

- [x] [Review][Patch] Reject whitespace-only mapfile content [`internal/service/haproxy/mapfile_schema.go`]
- [x] [Review][Patch] Acceptance test should assert action references the created mapfile/default backend [`internal/service/haproxy/mapfile_resource_test.go`]
- [x] [Review][Patch] Acceptance test should prove trailing-whitespace no-drift behavior [`internal/service/haproxy/mapfile_resource_test.go`]
- [x] [Review][Patch] Add mapfile data-source read-path test [`internal/service/haproxy/data_source_schema_test.go`]
- [x] [Review][Patch] Reconcile remaining data-source gap count across docs [`docs/index.md`, `templates/index.md.tmpl`, `_bmad-output/planning-artifacts/*`]

## Dev Notes

### OPNsense HAProxy Map File API

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Create | POST | `/api/haproxy/settings/addMapFile` | Body: `{"mapfile":{...}}` |
| Read | GET | `/api/haproxy/settings/getMapFile/{uuid}` | Returns `{"mapfile":{...}}` |
| Update | POST | `/api/haproxy/settings/setMapFile/{uuid}` | Body: `{"mapfile":{...}}` |
| Delete | POST | `/api/haproxy/settings/delMapFile/{uuid}` | |
| Search | GET/POST | `/api/haproxy/settings/searchMapFiles` | Paginated |
| Reconfigure | POST | `/api/haproxy/service/reconfigure` | Shared |

**Monad key:** `"mapfile"`

```go
var mapfileReqOpts = opnsense.ReqOpts{
    AddEndpoint:         "/api/haproxy/settings/addMapFile",
    GetEndpoint:         "/api/haproxy/settings/getMapFile",
    UpdateEndpoint:      "/api/haproxy/settings/setMapFile",
    DeleteEndpoint:      "/api/haproxy/settings/delMapFile",
    SearchEndpoint:      "/api/haproxy/settings/searchMapFiles",
    ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
    Monad:               "mapfile",
}
```

### Schema

| API Field | Response type | TF attr | TF type | Notes |
|-----------|---------------|---------|---------|-------|
| (UUID) | string | `id` | String (Computed) | |
| `name` | string | `name` | String (Required) | |
| `description` | string | `description` | String | |
| `type` | SelectedMap | `type` | String | confirm enum keys via `getMapFile` (domain/string map) |
| `content` | string | `content` | String (Required) | multi-line `<key> <value>` per line |

**Drift caution (AC #4):** OPNsense may return `content` with a normalized trailing newline or re-ordered/whitespace-collapsed lines. In `fromAPI`, trim trailing whitespace and compare line-set semantically where practical; document any normalization in the resource doc so users format `content` consistently (recommend `<<-EOT` heredoc, one `host backend` per line).

### Reference appliance usage

The appliance's `domain-map` is generated from `haproxy_backends[].domain` + aliases, e.g.:
```
grafana.mfsoho.linkridge.net grafana-backend
argocd.mfsoho.linkridge.net argocd-backend
tipsyhive.mfsoho.linkridge.net tipsyhive-backend
thetipsyhive.com tipsyhive-backend
tipsyhive.com tipsyhive-backend
```
A single `map_use_backend` action (Story 29.1) references this map; values are backend **names** (not UUIDs) as HAProxy resolves the map value to a backend at runtime — verify whether OPNsense expects backend name or UUID in the map value on the target version and document it.

### What NOT to build

- No per-line resource — model the whole file as one `content` string (matches the OPNsense object)
- No automatic generation of `content` from backend resources — that belongs to the consumer's HCL (or a future module), not this resource

### Project Structure

New: `internal/service/haproxy/mapfile_{model,schema,resource,resource_test,data_source}.go`, `examples/resources/opnsense_haproxy_mapfile/*`, `templates/resources/haproxy_mapfile.md.tmpl`
Modified: `internal/service/haproxy/exports.go`

### References

- [Source: https://docs.opnsense.org/development/api/plugins/haproxy.html — MapFile endpoints]
- [Pairs with: Story 29.1 (map_use_backend action)]
- [Downstream driver: opnsense-manager ansible/roles/haproxy/tasks/frontend.yml — domain map generation]

## Dev Agent Record

### Debug Log

- Resolved BMad dev-story workflow customization with no prepend/append steps.
- Loaded sprint status and selected first ready story: `29-2-haproxy-mapfile-resource`.
- Captured baseline commit: `fbaf085e8287ae4f00f786484cd1ab622d77716a`.
- Added red-phase model tests first; confirmed `go test ./internal/service/haproxy` failed because mapfile model types did not exist.
- Confirmed upstream OPNsense HAProxy mapfile enum keys from the HAProxy model: `beg`, `dom`, `end`, `int`, `ip`, `reg`, `str`, `sub`.

### Completion Notes

- Added `opnsense_haproxy_mapfile` resource and matching data source under `internal/service/haproxy`.
- Implemented CRUD/import using `/api/haproxy/settings/*MapFile` endpoints and shared HAProxy reconfigure endpoint.
- Implemented content normalization in API conversion and a string plan modifier so heredoc trailing whitespace does not cause perpetual drift.
- Registered the resource and data source in HAProxy exports and updated HAProxy data-source schema count tests.
- Added unit coverage for mapfile conversion and normalization.
- Added acceptance scaffold that creates a mapfile, references it from a `map_use_backend` action, imports it, verifies a no-op plan, and updates content in place.
- Added examples/templates and generated Registry docs for resource and data source.
- Updated support counts from 98 resources / 84 data sources to 99 resources / 85 data sources.
- Resolved code-review findings by rejecting whitespace-only content, strengthening acceptance assertions/no-drift coverage, adding mapfile data-source read coverage, and reconciling remaining data-source gap counts.

### Validation

- `go test ./internal/service/haproxy` failed in red phase before implementation as expected.
- `go test ./pkg/opnsense ./internal/service/haproxy` passed.
- `docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go generate ./tools` passed.
- `make check` passed.
- Post-review `go test ./pkg/opnsense ./internal/service/haproxy` passed.
- Post-review `docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go generate ./tools` passed.
- Post-review `make check` passed.

### File List

- `README.md`
- `_bmad-output/implementation-artifacts/29-2-haproxy-mapfile-resource.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `_bmad-output/planning-artifacts/prd.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `docs/data-sources/haproxy_mapfile.md`
- `docs/index.md`
- `docs/resources/haproxy_mapfile.md`
- `examples/resources/opnsense_haproxy_mapfile/import.sh`
- `examples/resources/opnsense_haproxy_mapfile/resource.tf`
- `internal/service/haproxy/data_source_schema_test.go`
- `internal/service/haproxy/exports.go`
- `internal/service/haproxy/mapfile_data_source.go`
- `internal/service/haproxy/mapfile_model.go`
- `internal/service/haproxy/mapfile_model_test.go`
- `internal/service/haproxy/mapfile_resource.go`
- `internal/service/haproxy/mapfile_resource_test.go`
- `internal/service/haproxy/mapfile_schema.go`
- `internal/service/haproxy/mapfile_validators.go`
- `templates/index.md.tmpl`
- `templates/resources/haproxy_mapfile.md.tmpl`

### Change Log

- 2026-06-10: Implemented HAProxy mapfile resource/data source, docs, examples, tests, and support-count updates; status set to review.
- 2026-06-10: Addressed code review findings; status set to done.
