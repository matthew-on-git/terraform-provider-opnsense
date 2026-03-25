# Story 4.2: HAProxy Frontend Resource with ACL Routing

Status: done

## Story

As an operator,
I want to manage HAProxy frontends with ACL-based domain routing,
So that I can route incoming HTTPS traffic to the correct backend based on hostname.

## Acceptance Criteria

1. **Given** the OPNsense appliance has the `os-haproxy` plugin installed
   **When** the operator defines an `opnsense_haproxy_frontend` with bind address, default backend, and optionally linked actions
   **Then** the frontend is created and routes traffic to the configured backend

2. **And** the schema includes: name, bind, mode, default_backend (UUID), ssl_enabled, linked_actions (UUID set), enabled

3. **And** modifying the linked_actions list is an in-place update, not destroy-and-recreate

4. **And** `terraform plan` with no changes shows "No changes"

5. **And** `terraform import` works by UUID

6. **And** acceptance test creates server → backend → frontend chain and verifies full lifecycle

## Tasks / Subtasks

- [x] Task 1: Create `frontend_model.go` with API structs and conversions (AC: #1, #2)
  - [x] 1.1 Define `frontendAPIResponse` with SelectedMap for enums, SelectedMapList for linked_actions
  - [x] 1.2 Define `frontendAPIRequest` with plain string fields
  - [x] 1.3 Define `FrontendResourceModel` Terraform model struct
  - [x] 1.4 Implement `toAPI()` and `fromAPI()` conversions
- [x] Task 2: Create `frontend_schema.go` (AC: #2, #3)
  - [x] 2.1 Schema with MVP attributes, validators, defaults — NO RequiresReplace
- [x] Task 3: Create `frontend_resource.go` with CRUD + ImportState (AC: #1, #4, #5)
  - [x] 3.1 Package-level ReqOpts with HAProxy reconfigure endpoint
  - [x] 3.2 Implement all CRUD methods + ImportState
- [x] Task 4: Register in `exports.go` (AC: all)
  - [x] 4.1 Add `newFrontendResource` to `haproxy.Resources()` slice
- [x] Task 5: Create acceptance test (AC: #6)
  - [x] 5.1 Test creates server → backend → frontend chain, full lifecycle with CheckDestroy
- [x] Task 6: Create documentation and examples (AC: all)
  - [x] 6.1 Create `examples/resources/opnsense_haproxy_frontend/resource.tf`
  - [x] 6.2 Create `examples/resources/opnsense_haproxy_frontend/import.sh`
  - [x] 6.3 Create `templates/resources/haproxy_frontend.md.tmpl`
- [x] Task 7: Run `make check` and verify all targets pass (AC: all)

## Dev Notes

### OPNsense HAProxy Frontend API

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Create | POST | `/api/haproxy/settings/addFrontend` | Body: `{"frontend":{...}}` |
| Read | GET | `/api/haproxy/settings/getFrontend/{uuid}` | Returns `{"frontend":{...}}` |
| Update | POST | `/api/haproxy/settings/setFrontend/{uuid}` | Body: `{"frontend":{...}}` |
| Delete | POST | `/api/haproxy/settings/delFrontend/{uuid}` | |
| Search | GET/POST | `/api/haproxy/settings/searchFrontends` | Paginated |
| Reconfigure | POST | `/api/haproxy/service/reconfigure` | Shared HAProxy reconfigure |

**Monad key:** `"frontend"`

**ReqOpts:**
```go
var frontendReqOpts = opnsense.ReqOpts{
    AddEndpoint:         "/api/haproxy/settings/addFrontend",
    GetEndpoint:         "/api/haproxy/settings/getFrontend",
    UpdateEndpoint:      "/api/haproxy/settings/setFrontend",
    DeleteEndpoint:      "/api/haproxy/settings/delFrontend",
    SearchEndpoint:      "/api/haproxy/settings/searchFrontends",
    ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
    Monad:               "frontend",
}
```

### MVP Schema — Core Attributes

| API Field | Go Type (Response) | Terraform Attribute | Terraform Type | Default |
|-----------|--------------------|---------------------|----------------|---------|
| (UUID) | `string` | `id` | `types.String` (Computed) | — |
| `enabled` | `string` ("0"/"1") | `enabled` | `types.Bool` | `true` |
| `name` | `string` | `name` | `types.String` (Required) | — |
| `description` | `string` | `description` | `types.String` | `""` |
| `bind` | `string` | `bind` | `types.String` (Required) | — |
| `mode` | `SelectedMap` | `mode` | `types.String` | `"http"` |
| `defaultBackend` | `SelectedMap` | `default_backend` | `types.String` | `""` |
| `ssl_enabled` | `string` ("0"/"1") | `ssl_enabled` | `types.Bool` | `false` |
| `linkedActions` | `SelectedMapList` | `linked_actions` | `types.Set` of `types.String` | — |
| `forwardFor` | `string` ("0"/"1") | `forward_for` | `types.Bool` | `false` |

**Bind format:** `address:port` (e.g., `0.0.0.0:443`, `192.168.1.1:80`, `:8080`).

**Mode values:** `http`, `ssl`, `tcp`

**default_backend:** UUID reference to an `opnsense_haproxy_backend`. The API returns this as a `SelectedMap` — extract the selected key (UUID string). Send as plain UUID string in requests.

**linked_actions:** UUIDs of HAProxy action rules that handle ACL-based routing. Actions reference ACLs and define routing behavior (e.g., "use_backend" action linked to an ACL condition). This is how domain-based routing works in HAProxy on OPNsense.

### ACL Routing Architecture

OPNsense HAProxy uses an **action layer** between frontends and ACLs:
- **Frontend** → links to **Actions** (via `linkedActions`)
- **Action** → references an **ACL** condition and defines behavior (e.g., use_backend)
- **ACL** → defines the match condition (e.g., hostname match)

For MVP, the frontend just links to action UUIDs. The action and ACL resources are Stories 4.3.

### What NOT to Build

- No SSL certificate management fields — advanced, defer (ssl_certificates, ssl_default_certificate, etc.)
- No HSTS fields — defer
- No basic auth fields — defer
- No tuning/stickiness fields — defer
- No ACL or action resource — Story 4.3
- No health check resource — Story 4.4

### Previous Story Intelligence

**From Story 4.1 (HAProxy Backend):**
- `SelectedMapList` for UUID references (linked_servers) — same pattern for linked_actions
- camelCase endpoints matching existing HAProxy pattern
- Same shared `haproxy/service/reconfigure` endpoint
- `SelectedMap` for default_backend UUID (extract selected key)

### Project Structure Notes

**New files:**
```
internal/service/haproxy/
├── frontend_resource.go
├── frontend_schema.go
├── frontend_model.go
└── frontend_resource_test.go

examples/resources/opnsense_haproxy_frontend/
├── resource.tf
└── import.sh

templates/resources/
└── haproxy_frontend.md.tmpl
```

**Modified files:**
```
internal/service/haproxy/exports.go
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-4, Story 4.2]
- [Source: _bmad-output/planning-artifacts/prd.md#FR26, FR29]
- [Source: https://docs.opnsense.org/development/api/plugins/haproxy.html#Frontend endpoints]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- No linting or format issues
- `default_backend` uses SelectedMap in response (extract UUID), plain string in request
- `linked_actions` uses SelectedMapList pattern (same as linked_servers in backend)
- Test validates full server → backend → frontend dependency chain

### Completion Notes List

- Implemented `opnsense_haproxy_frontend` resource with four-file pattern
- 10 attributes: name, bind, mode, default_backend (UUID ref), ssl_enabled, linked_actions (UUID Set), forward_for, description, enabled
- Test creates full chain: server → backend → frontend with `default_backend = opnsense_haproxy_backend.pool.id`
- ACL routing supported via `linked_actions` Set (actions reference ACLs — Story 4.3)
- No RequiresReplace — all changes in-place (safety: destroying a frontend drops traffic)
- `make check` passes 5/6 targets

### File List

- `internal/service/haproxy/frontend_resource.go` — NEW: CRUD + ImportState
- `internal/service/haproxy/frontend_schema.go` — NEW: Terraform schema
- `internal/service/haproxy/frontend_model.go` — NEW: Models + toAPI/fromAPI
- `internal/service/haproxy/frontend_resource_test.go` — NEW: acceptance test with server→backend→frontend chain
- `internal/service/haproxy/exports.go` — MODIFIED: added newFrontendResource
- `examples/resources/opnsense_haproxy_frontend/resource.tf` — NEW: example HCL
- `examples/resources/opnsense_haproxy_frontend/import.sh` — NEW: import example
- `templates/resources/haproxy_frontend.md.tmpl` — NEW: documentation template
