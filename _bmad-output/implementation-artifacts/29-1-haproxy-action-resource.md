# Story 29.1: HAProxy Action Resource

Status: done

## Story

As an operator,
I want to manage HAProxy actions as a first-class `opnsense_haproxy_action` resource,
So that I can build ACL-based routing rules (host → backend), domain-map routing, internal-only deny rules, HTTP→HTTPS redirects, and header rewrites — and link them to frontends via the existing `linked_actions` field.

## Context

Story 4.2 shipped `opnsense_haproxy_frontend` with a `linked_actions` (UUID Set) field, and Story 4.3 shipped `opnsense_haproxy_acl`. But **the action resource that sits between them was never built** — Story 4.2 "What NOT to Build" punted it to "Story 4.3," and 4.3 only implemented ACLs. The result: a frontend can link to action UUIDs, and ACLs can define match conditions, but **there is no way to create the actions that tie an ACL to a routing behavior.** Without this resource, the provider can only do single-`default_backend` frontends — no multi-domain host routing, no deny rules, no redirects.

This is the #1 blocker for migrating a real multi-domain OPNsense edge (e.g. the `opnsense-manager` appliance, which routes argocd/grafana/tipsyhive + CNAME aliases through one `https-in` frontend using actions + a domain map) off Ansible.

## Acceptance Criteria

1. **Given** the OPNsense appliance has the `os-haproxy` plugin installed
   **When** the operator defines an `opnsense_haproxy_action` with a `type` and the fields relevant to that type
   **Then** the action is created via `/api/haproxy/settings/addAction` and is linkable from a frontend's `linked_actions`

2. **And** the schema supports at minimum these action `type` values, which cover the appliance hot path:
   - `use_backend` — route to a backend (field: `use_backend` = backend UUID) gated by `linked_acls`
   - `map_use_backend` — route by a domain **map file** (field: `map_use_backend` = mapfile UUID; see Story 29.2)
   - `http-request_deny` — deny when ACL conditions match (the internal-only pattern)
   - `http-request_redirect` — redirect (e.g. HTTP→HTTPS); fields for redirect scheme/location/code
   - `http-request_set-header` — set a header (e.g. `X-Forwarded-Proto: https`); fields `set-header` name + format

3. **And** `linked_acls` is a UUID Set referencing `opnsense_haproxy_acl` resources, and `test_type` (`if` / `unless`) controls match polarity

4. **And** `linked_acls` and other list/ref fields update **in place** (no destroy-and-recreate)

5. **And** `terraform plan` with no changes shows "No changes"; `terraform import` works by UUID

6. **And** an acceptance test builds acl → action → frontend (with `linked_actions = [action.id]`) and verifies the full lifecycle with `CheckDestroy`

7. **And** the existing `opnsense_haproxy_frontend` docs/example are updated to show a real routing chain (acl → action → frontend) replacing the inert ACL shown in the current `haproxy-full-stack` composition

## Tasks / Subtasks

- [x] Task 1: `action_model.go` — API structs + conversions (AC: #1, #2, #3)
  - [x] 1.1 `actionAPIResponse` with `SelectedMap` for `type`/`test_type`/`use_backend`/`map_use_backend_file`, `SelectedMapList` for `linked_acls`
  - [x] 1.2 `actionAPIRequest` with plain string fields
  - [x] 1.3 `ActionResourceModel` Terraform struct
  - [x] 1.4 `toAPI()` / `fromAPI()` — including type-conditional fields
- [x] Task 2: `action_schema.go` (AC: #2, #3, #4)
  - [x] 2.1 `type` (Required) with `OneOf` validator for the supported action types
  - [x] 2.2 Type-specific optional fields; document which apply to which `type`
  - [x] 2.3 NO RequiresReplace
- [x] Task 3: `action_resource.go` — CRUD + ImportState (AC: #1, #5)
  - [x] 3.1 Package-level `actionReqOpts` (see Dev Notes), shared `haproxy/service/reconfigure`
- [x] Task 4: Register `newActionResource` in `internal/service/haproxy/exports.go` (AC: all)
- [x] Task 5: `action_resource_test.go` — acl → action → frontend chain (AC: #6)
- [x] Task 6: `action_data_source.go` + data-source schema test (parity with other haproxy resources)
- [x] Task 7: Docs + examples — `examples/resources/opnsense_haproxy_action/{resource.tf,import.sh}`, `templates/resources/haproxy_action.md.tmpl`; update `examples/compositions/haproxy-full-stack/main.tf` to wire acl → action → frontend (AC: #7)
- [x] Task 8: `make check` — all targets pass

## Implementation Notes

- Added `opnsense_haproxy_action` resource and matching data source under `internal/service/haproxy`.
- Matched the live OPNsense 25.7 HAProxy action API shape: actions do not expose `enabled`, and HTTP request variants are selected directly through `type` values such as `http-request_redirect`.
- Registered the new resource/data source in HAProxy exports.
- Added unit coverage for API/Terraform action type translation and an acceptance-test scaffold that composes ACL -> action -> frontend.
- Added examples/templates and regenerated `docs/resources/haproxy_action.md` and `docs/data-sources/haproxy_action.md` with containerized `go generate ./tools`.
- Updated `examples/compositions/haproxy-full-stack/main.tf` so the ACL is wired through a real action into `linked_actions`.
- Fixed `opnsense_haproxy_frontend` live read compatibility for OPNsense 25.7 by decoding `bind` from the selected-map response format.
- Code-review patches applied: plain-string fallback for `SelectedMapList`, action-type config validation for required fields, `deny_status` constrained to HTTP status range, frontend standalone example/docs wired through ACL -> action -> frontend, and acceptance destroy checks expanded to all resources created by the action test.

## Validation

- `go test ./internal/service/haproxy` passed.
- `go test ./pkg/opnsense ./internal/service/haproxy` passed after review patches.
- `docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go generate ./tools` passed.
- `make check` passed.
- Vagrant live acceptance passed against OPNsense 25.7 with `os-haproxy` installed: `TF_ACC=1 ... go test -run TestAccHAProxyAction_basic -count=1 ./internal/service/haproxy` in the dev-toolchain container via an SSH tunnel to the VM API.
- Additional focused Vagrant regression checks passed with the same VM/tunnel/container method: `TestAccHAProxyFrontend_basic` and `TestAccHAProxyBackend_basic`.
- Plain `vagrant` is still affected by host Ruby/mise leakage; use the documented clean wrapper/alias (`vg`) from `test/README.md`.

## Dev Notes

### OPNsense HAProxy Action API

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Create | POST | `/api/haproxy/settings/addAction` | Body: `{"action":{...}}` |
| Read | GET | `/api/haproxy/settings/getAction/{uuid}` | Returns `{"action":{...}}` |
| Update | POST | `/api/haproxy/settings/setAction/{uuid}` | Body: `{"action":{...}}` |
| Delete | POST | `/api/haproxy/settings/delAction/{uuid}` | |
| Search | GET/POST | `/api/haproxy/settings/searchActions` | Paginated |
| Reconfigure | POST | `/api/haproxy/service/reconfigure` | Shared HAProxy reconfigure |

**Monad key:** `"action"`

```go
var actionReqOpts = opnsense.ReqOpts{
    AddEndpoint:         "/api/haproxy/settings/addAction",
    GetEndpoint:         "/api/haproxy/settings/getAction",
    UpdateEndpoint:      "/api/haproxy/settings/setAction",
    DeleteEndpoint:      "/api/haproxy/settings/delAction",
    SearchEndpoint:      "/api/haproxy/settings/searchActions",
    ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
    Monad:               "action",
}
```

### Schema — core attributes

| API Field | Response type | Terraform attr | TF type | Notes |
|-----------|---------------|----------------|---------|-------|
| (UUID) | string | `id` | String (Computed) | |
| `name` | string | `name` | String (Required) | |
| `description` | string | `description` | String | |
| `testType` | SelectedMap | `test_type` | String | `if` (default) / `unless` |
| `linkedAcls` | SelectedMapList | `linked_acls` | Set(String) | ACL UUIDs |
| `type` | SelectedMap | `type` | String (Required) | the action type (see AC #2) |
| `use_backend` | SelectedMap | `use_backend` | String | backend UUID (type=`use_backend`) |
| `map_use_backend_file` | SelectedMap | `mapfile` | String | mapfile UUID (type=`map_use_backend`, Story 29.2) |
| `http_request_redirect` | string | `redirect` | String | redirect rule text (type=`http-request_redirect`) |
| `http_request_set_header_name` / `http_request_set_header_content` | string | `set_header_name` / `set_header_content` | String | type=`http-request_set-header` |

**Field naming note:** Live OPNsense 25.7 exposes the action verb as the `type` API field. Terraform also surfaces it as `type`. The map backend field is `map_use_backend_file` in the API and `mapfile` in Terraform to match the OPNsense UI label.

**Pattern reuse:** `linked_acls` mirrors `linked_servers` (backend) and `linked_actions` (frontend) — both are the `SelectedMapList` pattern. `use_backend`/`mapfile`/`test_type` are `SelectedMap` single-key extraction (same as `default_backend` in the frontend).

### Reference appliance usage (what this must express)

From the downstream `opnsense-manager` Ansible HAProxy role, the actions in production are:
- `route-by-domain-map` → `map_use_backend` referencing the `domain-map` map file (Story 29.2)
- `use-wildcard-<svc>` → `use_backend` gated by a `host-ends-with-<svc>` ACL
- `deny-external-<svc>` → `http-request_deny` gated by host ACL + `unless` internal-network ACL
- `redirect-to-https` → `http-request_redirect`
- `set-x-forwarded-proto-https` → `http-request_set-header`

### What NOT to build

- No new ACL fields — `opnsense_haproxy_acl` (Story 4.3) already covers match conditions
- No inline ACL creation inside the action — keep ACLs as separate resources referenced by UUID
- Do not attempt to model every exotic action verb — ship the five in AC #2; leave the rest as a documented follow-up

### Previous Story Intelligence

- Story 4.2 (frontend): `SelectedMap`/`SelectedMapList` conversions, shared reconfigure endpoint, no RequiresReplace (mutating live traffic resources should be in-place)
- Story 4.3 (acl): the four-file pattern + data source + schema parity test

### Project Structure

New: `internal/service/haproxy/action_{model,schema,resource,resource_test,data_source}.go`, `examples/resources/opnsense_haproxy_action/*`, `templates/resources/haproxy_action.md.tmpl`
Modified: `internal/service/haproxy/exports.go`, `examples/compositions/haproxy-full-stack/main.tf`

### References

- [Source: _bmad-output/implementation-artifacts/4-2-haproxy-frontend-resource-with-acl-routing.md#What-NOT-to-Build]
- [Source: _bmad-output/implementation-artifacts/4-3-haproxy-acl-resource.md]
- [Source: https://docs.opnsense.org/development/api/plugins/haproxy.html — Action endpoints]
- [Downstream driver: opnsense-manager ansible/roles/haproxy/tasks/frontend.yml — actions wired via linkedActions]
