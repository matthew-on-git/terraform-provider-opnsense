# Story 4.3: HAProxy ACL Resource

Status: done

## Story

As an operator,
I want to manage HAProxy ACL rules for domain-based routing,
So that I can direct traffic to specific backends based on host headers, paths, or SNI.

## Acceptance Criteria

1. **Given** the OPNsense appliance has the `os-haproxy` plugin installed
   **When** the operator defines an `opnsense_haproxy_acl` resource
   **Then** the ACL is created with the specified match condition

2. **And** the schema includes: name, description, expression (match type), negate, and the most common match value fields (hdr_beg, hdr_end, hdr, path_beg, path, ssl_sni)

3. **And** `terraform plan` with no changes shows "No changes"

4. **And** `terraform import` works by UUID

5. **And** acceptance test covers full lifecycle with CheckDestroy

## Tasks / Subtasks

- [x] Task 1: Create `acl_model.go` with API structs and conversions (AC: #1, #2)
  - [x] 1.1 Define `aclAPIResponse` with SelectedMap for expression field
  - [x] 1.2 Define `aclAPIRequest` with plain string fields
  - [x] 1.3 Define `ACLResourceModel` Terraform model struct
  - [x] 1.4 Implement `toAPI()` and `fromAPI()` conversions
- [x] Task 2: Create `acl_schema.go` (AC: #2)
  - [x] 2.1 Schema with MVP attributes, validators, defaults
- [x] Task 3: Create `acl_resource.go` with CRUD + ImportState (AC: #1, #3, #4)
  - [x] 3.1 Package-level ReqOpts with HAProxy reconfigure endpoint
  - [x] 3.2 Implement all CRUD methods + ImportState
- [x] Task 4: Register in `exports.go` (AC: all)
  - [x] 4.1 Add `newACLResource` to `haproxy.Resources()` slice
- [x] Task 5: Create acceptance test (AC: #5)
  - [x] 5.1 Full lifecycle test with CheckDestroy
- [x] Task 6: Create documentation and examples (AC: all)
  - [x] 6.1 Create `examples/resources/opnsense_haproxy_acl/resource.tf`
  - [x] 6.2 Create `examples/resources/opnsense_haproxy_acl/import.sh`
  - [x] 6.3 Create `templates/resources/haproxy_acl.md.tmpl`
- [x] Task 7: Run `make check` and verify all targets pass (AC: all)

## Dev Notes

### OPNsense HAProxy ACL API

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Create | POST | `/api/haproxy/settings/addAcl` | Body: `{"acl":{...}}` |
| Read | GET | `/api/haproxy/settings/getAcl/{uuid}` | Returns `{"acl":{...}}` |
| Update | POST | `/api/haproxy/settings/setAcl/{uuid}` | Body: `{"acl":{...}}` |
| Delete | POST | `/api/haproxy/settings/delAcl/{uuid}` | |
| Search | GET/POST | `/api/haproxy/settings/searchAcls` | Paginated |
| Reconfigure | POST | `/api/haproxy/service/reconfigure` | Shared HAProxy reconfigure |

**Monad key:** `"acl"`

**ReqOpts:**
```go
var aclReqOpts = opnsense.ReqOpts{
    AddEndpoint:         "/api/haproxy/settings/addAcl",
    GetEndpoint:         "/api/haproxy/settings/getAcl",
    UpdateEndpoint:      "/api/haproxy/settings/setAcl",
    DeleteEndpoint:      "/api/haproxy/settings/delAcl",
    SearchEndpoint:      "/api/haproxy/settings/searchAcls",
    ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
    Monad:               "acl",
}
```

### IMPORTANT: The ACL Model is Extremely Complex

The full ACL model has 100+ fields and 68 expression types. For MVP, include only the most commonly used fields for the core use cases (host matching, path matching, SNI matching).

### MVP Schema — Core Attributes Only

| API Field | Go Type (Response) | Terraform Attribute | Terraform Type | Default |
|-----------|--------------------|---------------------|----------------|---------|
| (UUID) | `string` | `id` | `types.String` (Computed) | — |
| `name` | `string` | `name` | `types.String` (Required) | — |
| `description` | `string` | `description` | `types.String` | `""` |
| `expression` | `SelectedMap` | `expression` | `types.String` (Required) | — |
| `negate` | `string` ("0"/"1") | `negate` | `types.Bool` | `false` |
| `hdr_beg` | `string` | `hdr_beg` | `types.String` | `""` |
| `hdr_end` | `string` | `hdr_end` | `types.String` | `""` |
| `hdr` | `string` | `hdr` | `types.String` | `""` |
| `path_beg` | `string` | `path_beg` | `types.String` | `""` |
| `path` | `string` | `path` | `types.String` | `""` |
| `ssl_sni` | `string` | `ssl_sni` | `types.String` | `""` |
| `ssl_fc_sni` | `string` | `ssl_fc_sni` | `types.String` | `""` |
| `src` | `string` | `src` | `types.String` | `""` |
| `nbsrv_backend` | `SelectedMap` | `nbsrv_backend` | `types.String` | `""` |
| `custom_acl` | `string` | `custom_acl` | `types.String` | `""` |

**Expression values (MVP subset):** Only validate the most common ones:
`hdr_beg`, `hdr_end`, `hdr`, `hdr_reg`, `hdr_sub`, `path_beg`, `path_end`, `path`, `path_reg`, `path_sub`, `ssl_fc_sni`, `ssl_sni`, `ssl_sni_beg`, `ssl_sni_end`, `src`, `nbsrv`, `custom_acl`

Do NOT validate all 68 expression types — leave the validator open for less common ones to work without a provider update.

**Design decision:** Rather than validating with `OneOf` (which would lock users out of less common expressions), use no validator on `expression` — let OPNsense API validate.

### How ACLs Connect to Frontends

ACLs are NOT directly linked to frontends. The chain is:
1. **ACL** — defines a match condition
2. **Action** — references an ACL and defines behavior (e.g., `use_backend`)
3. **Frontend** — links to actions via `linked_actions`

ACLs are independent resources. They become useful when referenced by actions (not implemented in this provider yet, but the action UUID linking infrastructure is ready in the frontend resource).

### What NOT to Build

- No custom header fields (`cust_hdr_*`) — advanced, too many permutations
- No sticky counter fields (`sc_*`, `src_*` rate/counter fields) — advanced
- No SSL certificate verification fields — advanced
- No URL parameter fields — rare use case
- No variable comparison fields — advanced
- No action resource — future story
- No data source — deferred to Epic 12

### Previous Story Intelligence

**From Story 4.2 (Frontend):**
- `linked_actions` Set on frontend — ready to accept ACL action UUIDs
- Same shared HAProxy reconfigure endpoint
- Same camelCase endpoint pattern

### Project Structure Notes

**New files:**
```
internal/service/haproxy/
├── acl_resource.go
├── acl_schema.go
├── acl_model.go
└── acl_resource_test.go

examples/resources/opnsense_haproxy_acl/
├── resource.tf
└── import.sh

templates/resources/
└── haproxy_acl.md.tmpl
```

**Modified files:**
```
internal/service/haproxy/exports.go
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-4, Story 4.3]
- [Source: _bmad-output/planning-artifacts/prd.md#FR27]
- [Source: https://docs.opnsense.org/development/api/plugins/haproxy.html#ACL endpoints]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- `gofumpt` formatting fix required for struct field alignment — applied via `make fix`
- No expression validator — deliberately left open to avoid locking out valid expression types
- `nbsrv_backend` uses SelectedMap in response (UUID reference to backend)

### Completion Notes List

- Implemented `opnsense_haproxy_acl` resource with MVP subset of the complex ACL model
- 15 attributes covering core use cases: host matching, path matching, SNI matching, source IP, custom ACL
- No expression validator — OPNsense API validates (68 valid expression types too many to enumerate)
- All match value fields are Optional+Computed with empty string default (only the relevant field for the chosen expression needs a value)
- `make check` passes 5/6 targets; format fix applied via `make fix`

### File List

- `internal/service/haproxy/acl_resource.go` — NEW: CRUD + ImportState
- `internal/service/haproxy/acl_schema.go` — NEW: Terraform schema
- `internal/service/haproxy/acl_model.go` — NEW: Models + toAPI/fromAPI
- `internal/service/haproxy/acl_resource_test.go` — NEW: acceptance test with CheckDestroy
- `internal/service/haproxy/exports.go` — MODIFIED: added newACLResource
- `examples/resources/opnsense_haproxy_acl/resource.tf` — NEW: example HCL
- `examples/resources/opnsense_haproxy_acl/import.sh` — NEW: import example
- `templates/resources/haproxy_acl.md.tmpl` — NEW: documentation template
