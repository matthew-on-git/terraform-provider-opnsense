# Story 3.2: Firewall Category Resource

Status: done

## Story

As an operator,
I want to manage firewall categories through Terraform,
So that I can organize my firewall rules into logical groups.

## Acceptance Criteria

1. **Given** the provider is configured with valid OPNsense credentials
   **When** the operator defines an `opnsense_firewall_category` resource in HCL
   **Then** `terraform apply` creates the category on OPNsense via the API and returns the UUID

2. **And** `terraform plan` with no changes shows "No changes" (state read-back matches)

3. **And** modifying category attributes in HCL shows the correct diff and applies the change

4. **And** removing the resource block deletes the category from OPNsense

5. **And** `terraform import opnsense_firewall_category.test <uuid>` imports an existing category into state

6. **And** after import, `terraform plan` shows "No changes"

7. **And** acceptance test covers full lifecycle: Create → Verify → Import → Update → Destroy with CheckDestroy

## Tasks / Subtasks

- [x] Task 1: Create `category_model.go` with API structs and conversions (AC: #1, #2)
  - [x] 1.1 Define `categoryAPIResponse` struct with SelectedMap for `auto` field
  - [x] 1.2 Define `categoryAPIRequest` struct with plain string fields
  - [x] 1.3 Define `CategoryResourceModel` Terraform model struct
  - [x] 1.4 Implement `toAPI()` and `fromAPI()` conversions
- [x] Task 2: Create `category_schema.go` with Terraform schema (AC: #1, #3)
  - [x] 2.1 Define schema with name (Required), auto (Optional Bool), color (Optional String with hex validation)
- [x] Task 3: Create `category_resource.go` with CRUD + ImportState (AC: #1-#7)
  - [x] 3.1 Define `categoryReqOpts` with category endpoints (NO ReconfigureEndpoint — categories are metadata)
  - [x] 3.2 Implement Create, Read, Update, Delete, ImportState, Configure, Metadata
- [x] Task 4: Register in `exports.go` (AC: all)
  - [x] 4.1 Add `newCategoryResource` to `firewall.Resources()` slice
- [x] Task 5: Create `category_resource_test.go` acceptance test (AC: #7)
  - [x] 5.1 Full lifecycle test with CheckDestroy
- [x] Task 6: Create documentation and examples (AC: all)
  - [x] 6.1 Create `examples/resources/opnsense_firewall_category/resource.tf`
  - [x] 6.2 Create `examples/resources/opnsense_firewall_category/import.sh`
  - [x] 6.3 Create `templates/resources/firewall_category.md.tmpl`
- [x] Task 7: Run `make check` and verify all targets pass (AC: all)

## Dev Notes

### OPNsense Firewall Category API Endpoints

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Create | POST | `/api/firewall/category/addItem` | Returns `{"result":"saved","uuid":"..."}` |
| Read | GET | `/api/firewall/category/getItem/{uuid}` | Returns `{"category":{...}}` |
| Update | POST | `/api/firewall/category/setItem/{uuid}` | Returns `{"result":"saved"}` |
| Delete | POST | `/api/firewall/category/delItem/{uuid}` | Safe delete — checks if in use |
| Search | GET/POST | `/api/firewall/category/searchItem` | Paginated list |

**CRITICAL: NO reconfigure endpoint.** Categories are metadata labels — they don't affect running firewall configuration and don't need a reconfigure step. Leave `ReconfigureEndpoint` empty in ReqOpts (the `Reconfigure()` function handles this as a no-op).

**Monad key:** `"category"`

**ReqOpts:**
```go
var categoryReqOpts = opnsense.ReqOpts{
    AddEndpoint:    "/api/firewall/category/addItem",
    GetEndpoint:    "/api/firewall/category/getItem",
    UpdateEndpoint: "/api/firewall/category/setItem",
    DeleteEndpoint: "/api/firewall/category/delItem",
    SearchEndpoint: "/api/firewall/category/searchItem",
    Monad:          "category",
    // No ReconfigureEndpoint — categories are metadata, no service reload needed.
}
```

### Category API Model Fields

| API Field | Go Type (Response) | Terraform Attribute | Terraform Type | Conversion |
|-----------|-------------------|---------------------|----------------|------------|
| (UUID) | `string` | `id` | `types.String` (Computed) | Passthrough |
| `name` | `string` | `name` | `types.String` (Required) | Direct |
| `auto` | `SelectedMap` | `auto` | `types.Bool` (Optional, default `true`) | `SelectedMap` → `StringToBool` |
| `color` | `string` | `color` | `types.String` (Optional, default `""`) | Direct |

**Name validation:** No commas allowed (OPNsense enforces regex `/[^,]+/`). Name has a UniqueConstraint.

**Auto field:** Boolean stored as OPNsense SelectedMap `{"0":{"selected":0},"1":{"selected":1}}`. For `toAPI()`, send `"0"` or `"1"` string. For `fromAPI()`, extract selected key from SelectedMap then convert via `StringToBool`.

**Color field:** 6-character hex string (e.g., `"ff0000"` for red). OPNsense validates with regex `/^([0-9a-fA-F]){6,6}$/`. Empty string means no color.

### Dual-Struct Pattern

**Response struct:**
```go
type categoryAPIResponse struct {
    Name  string               `json:"name"`
    Auto  opnsense.SelectedMap `json:"auto"`
    Color string               `json:"color"`
}
```

**Request struct:**
```go
type categoryAPIRequest struct {
    Name  string `json:"name"`
    Auto  string `json:"auto"`
    Color string `json:"color"`
}
```

### Schema Validation for Color

```go
"color": schema.StringAttribute{
    Optional:            true,
    Computed:            true,
    Default:             stringdefault.StaticString(""),
    MarkdownDescription: "Color as 6-digit hex (e.g., `ff0000` for red). Empty for no color.",
    Validators: []validator.String{
        stringvalidator.RegexMatches(
            regexp.MustCompile(`^([0-9a-fA-F]{6})?$`),
            "must be a 6-digit hex color code or empty",
        ),
    },
},
```

**Required import:** `"regexp"` and `"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"`

### Safe Delete Behavior

The `delItem` endpoint has safe delete enabled — OPNsense checks if the category is referenced by any filter rules before allowing deletion. If in use, the API returns a validation error. This error will propagate naturally through `ParseMutationResponse` → `ValidationError` → `resp.Diagnostics.AddError()`.

### What NOT to Build

- No savepoint/rollback flow — categories use NO reconfigure at all (metadata only)
- No filter rule resource — that's Story 3.3
- No data source — follows same pattern as Story 2.3 (deferred to Epic 12)
- No changes to provider.go — category is in the firewall package, already registered

### Previous Story Intelligence

**From Story 3.1 (Savepoint Implementation):**
- `FirewallFilterReconfigure(client)` is available but NOT used here — categories don't need reconfigure
- `Reconfigure()` in `reconfigure.go` handles empty `ReconfigureEndpoint` as a no-op (line 19-21)

**From Story 2.1 (Firewall Alias Resource):**
- Four-file pattern: `_resource.go`, `_schema.go`, `_model.go`, `_resource_test.go`
- Dual-struct: `aliasAPIRequest` (strings for POST) vs `aliasAPIResponse` (SelectedMap for GET)
- CRUD follows: plan → toAPI → Add → Get → fromAPI → state
- CheckDestroy required in acceptance tests
- `_ context.Context` pattern for `fromAPI()` if ctx unused

**From Epic 1 Retrospective:**
- `ctx context.Context` always first parameter
- `make check` must pass all targets

### Project Structure Notes

**New files:**
```
internal/service/firewall/
├── category_resource.go             # NEW: CRUD + ImportState
├── category_schema.go               # NEW: Schema()
├── category_model.go                # NEW: models + toAPI/fromAPI
└── category_resource_test.go        # NEW: acceptance test

examples/resources/opnsense_firewall_category/
├── resource.tf                      # NEW: example HCL
└── import.sh                        # NEW: import example

templates/resources/
└── firewall_category.md.tmpl        # NEW: doc template
```

**Modified files:**
```
internal/service/firewall/exports.go # MODIFIED: add newCategoryResource to Resources()
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-3, Story 3.2]
- [Source: _bmad-output/planning-artifacts/prd.md#FR20 firewall categories]
- [Source: _bmad-output/implementation-artifacts/2-1-firewall-alias-resource.md#Four-file pattern]
- [Source: https://docs.opnsense.org/development/api/core/firewall.html#Category endpoints]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- No linting issues — all `make check` targets pass cleanly (except pre-existing gitleaks scan)
- No ReconfigureEndpoint in ReqOpts — categories are metadata, Reconfigure() handles empty endpoint as no-op
- `auto` field uses SelectedMap in response → cast to string → StringToBool for conversion

### Completion Notes List

- Implemented `opnsense_firewall_category` resource with four-file pattern
- Simplest resource: 3 fields (name, auto, color) — no sets, no integers, no complex types
- No reconfigure endpoint — categories are metadata labels only
- Color validated with hex regex `^([0-9a-fA-F]{6})?$`
- Safe delete behavior handled by OPNsense API (returns validation error if category in use)
- Acceptance test covers Create → Import → Update → Destroy with CheckDestroy
- `make check` passes 5/6 targets

### File List

- `internal/service/firewall/category_resource.go` — NEW: CRUD + ImportState + Configure
- `internal/service/firewall/category_schema.go` — NEW: Terraform schema definition
- `internal/service/firewall/category_model.go` — NEW: CategoryResourceModel, categoryAPIRequest, categoryAPIResponse, toAPI(), fromAPI()
- `internal/service/firewall/category_resource_test.go` — NEW: acceptance test with CheckDestroy
- `internal/service/firewall/exports.go` — MODIFIED: added newCategoryResource to Resources()
- `examples/resources/opnsense_firewall_category/resource.tf` — NEW: example HCL
- `examples/resources/opnsense_firewall_category/import.sh` — NEW: import example
- `templates/resources/firewall_category.md.tmpl` — NEW: documentation template
