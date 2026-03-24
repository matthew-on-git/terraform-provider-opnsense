# Story 2.2: HAProxy Server Resource

Status: done

## Story

As an operator,
I want to manage OPNsense HAProxy servers through Terraform,
So that I can define backend server targets as code and validate that the provider works with plugin APIs.

## Acceptance Criteria

1. **Given** the OPNsense appliance has the `os-haproxy` plugin installed
   **When** the operator defines an `opnsense_haproxy_server` resource in HCL
   **Then** the same full CRUD + import + drift detection lifecycle works as with firewall aliases

2. **And** if the HAProxy plugin is not installed, the provider returns a `PluginNotFoundError` with a clear message to install `os-haproxy`

3. **And** the resource schema includes server-specific attributes: name, address, port, weight, ssl, enabled

4. **And** boolean attributes convert correctly between Terraform `types.Bool` and OPNsense `"0"`/`"1"` strings

5. **And** acceptance test covers full lifecycle: Create → Verify → Import → Update → Destroy

6. **And** this validates that the API client pattern works for both core and plugin APIs

## Tasks / Subtasks

- [x] Task 1: Create `internal/service/haproxy/` package with `exports.go` (AC: all)
  - [x] 1.1 Create `internal/service/haproxy/exports.go` with `Resources()` returning `newServerResource` and `DataSources()` returning empty slice
  - [x] 1.2 Register haproxy package in `internal/provider/provider.go` `Resources()` method (append to firewall resources)
- [x] Task 2: Create `server_model.go` with API structs and conversions (AC: #1, #3, #4)
  - [x] 2.1 Define `serverAPIResponse` struct with SelectedMap for mode field and string for booleans
  - [x] 2.2 Define `serverAPIRequest` struct with plain string fields for POST requests
  - [x] 2.3 Define `ServerResourceModel` Terraform model struct with `tfsdk` tags
  - [x] 2.4 Implement `toAPI()` converting Terraform model → API request (Bool→"0"/"1", Int64→string)
  - [x] 2.5 Implement `fromAPI()` converting API response → Terraform model ("0"/"1"→Bool, string→Int64)
- [x] Task 3: Create `server_schema.go` with Terraform schema (AC: #1, #3)
  - [x] 3.1 Define schema with all server attributes, validators, plan modifiers, and defaults
- [x] Task 4: Create `server_resource.go` with CRUD + ImportState (AC: #1, #2, #5)
  - [x] 4.1 Implement `Create` — plan → toAPI → Add → Get → fromAPI → state
  - [x] 4.2 Implement `Read` — Get → fromAPI → state (RemoveResource on NotFoundError)
  - [x] 4.3 Implement `Update` — plan → toAPI → Update → Get → fromAPI → state
  - [x] 4.4 Implement `Delete` — Delete by UUID (handle NotFoundError gracefully)
  - [x] 4.5 Implement `ImportState` — passthrough ID
  - [x] 4.6 Implement `Configure` — extract `*opnsense.Client` from provider data
- [x] Task 5: Create `server_resource_test.go` acceptance test (AC: #1, #4, #5)
  - [x] 5.1 Full lifecycle test: Create → Verify → Import → Update → Destroy with CheckDestroy
- [x] Task 6: Create documentation and examples (AC: all)
  - [x] 6.1 Create `examples/resources/opnsense_haproxy_server/resource.tf`
  - [x] 6.2 Create `examples/resources/opnsense_haproxy_server/import.sh`
  - [x] 6.3 Create `templates/resources/haproxy_server.md.tmpl`
- [x] Task 7: Run `make check` and verify all targets pass (AC: all)

## Dev Notes

### OPNsense HAProxy Server API Endpoints

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Create | POST | `/api/haproxy/settings/addServer` | Returns `{"result":"saved","uuid":"..."}` |
| Read | GET | `/api/haproxy/settings/getServer/{uuid}` | Returns `{"server":{...}}` |
| Update | POST | `/api/haproxy/settings/setServer/{uuid}` | Returns `{"result":"saved"}` |
| Delete | POST | `/api/haproxy/settings/delServer/{uuid}` | Returns `{"result":"deleted"}` |
| Search | GET/POST | `/api/haproxy/settings/searchServers` | Paginated list |
| Reconfigure | POST | `/api/haproxy/service/reconfigure` | Activates changes |

**Plugin:** `os-haproxy` — must be installed on OPNsense. HTTP 404 on any endpoint triggers `PluginNotFoundError` automatically via `CheckHTTPError` in `pkg/opnsense/errors.go`.

**Monad key:** `"server"`

**ReqOpts for this resource:**
```go
var serverReqOpts = opnsense.ReqOpts{
    AddEndpoint:         "/api/haproxy/settings/addServer",
    GetEndpoint:         "/api/haproxy/settings/getServer",
    UpdateEndpoint:      "/api/haproxy/settings/setServer",
    DeleteEndpoint:      "/api/haproxy/settings/delServer",
    SearchEndpoint:      "/api/haproxy/settings/searchServers",
    ReconfigureEndpoint: "/api/haproxy/service/reconfigure",
    Monad:               "server",
}
```

### Server API Model Fields

| API Field | Go Type (Response) | Terraform Attribute | Terraform Type | Conversion |
|-----------|-------------------|---------------------|----------------|------------|
| (UUID) | `string` | `id` | `types.String` (Computed) | Passthrough |
| `name` | `string` | `name` | `types.String` (Required) | Direct |
| `description` | `string` | `description` | `types.String` (Optional) | Direct |
| `address` | `string` | `address` | `types.String` (Required) | Direct |
| `port` | `string` | `port` | `types.Int64` (Required) | `StringToInt64`/`Int64ToString` |
| `weight` | `string` | `weight` | `types.Int64` (Optional, Computed) | `StringToInt64`/`Int64ToString` |
| `mode` | `SelectedMap` | `mode` | `types.String` (Optional, Computed) | `SelectedMap` → `string(sm)` |
| `ssl` | `string` ("0"/"1") | `ssl` | `types.Bool` (Optional, default `false`) | `BoolToString`/`StringToBool` |
| `sslVerify` | `string` ("0"/"1") | `ssl_verify` | `types.Bool` (Optional, default `true`) | `BoolToString`/`StringToBool` |
| `enabled` | `string` ("0"/"1") | `enabled` | `types.Bool` (Optional, default `true`) | `BoolToString`/`StringToBool` |

**Mode values:** `active`, `backup`, `disabled` (default: `active`)

**Integer fields:** OPNsense stores `port` and `weight` as strings. Use `opnsense.StringToInt64()` and `opnsense.Int64ToString()` for conversion. Handle empty strings for optional `weight` (empty → null in Terraform).

**Boolean fields:** Three boolean fields (`ssl`, `ssl_verify`, `enabled`) all use `"0"`/`"1"` strings in the API. Use `opnsense.BoolToString()` / `opnsense.StringToBool()`.

### Plugin Detection — ALREADY HANDLED

`PluginNotFoundError` is handled automatically by the API client. When OPNsense returns HTTP 404 for any endpoint, `CheckHTTPError()` in `pkg/opnsense/errors.go:113` creates a `PluginNotFoundError` with the module name extracted from the endpoint path (e.g., `"haproxy"`). The error message reads: `"plugin 'haproxy' is not installed on OPNsense"`.

**No special plugin detection code is needed in the resource.** The generic CRUD functions already call `CheckHTTPError` which returns `PluginNotFoundError` on 404. The resource just needs standard error handling — the error surfaces to the user via `resp.Diagnostics.AddError()`.

### Dual-Struct Pattern (MUST follow from Story 2.1)

**Response struct** (for unmarshaling GET responses — uses SelectedMap for enum fields):
```go
type serverAPIResponse struct {
    Name        string               `json:"name"`
    Description string               `json:"description"`
    Address     string               `json:"address"`
    Port        string               `json:"port"`
    Weight      string               `json:"weight"`
    Mode        opnsense.SelectedMap `json:"mode"`
    SSL         string               `json:"ssl"`
    SSLVerify   string               `json:"sslVerify"`
    Enabled     string               `json:"enabled"`
}
```

**Request struct** (for marshaling POST requests — plain strings only):
```go
type serverAPIRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Address     string `json:"address"`
    Port        string `json:"port"`
    Weight      string `json:"weight"`
    Mode        string `json:"mode"`
    SSL         string `json:"ssl"`
    SSLVerify   string `json:"sslVerify"`
    Enabled     string `json:"enabled"`
}
```

### Type Conversion Patterns

**toAPI() — key conversions:**
```go
Port:    opnsense.Int64ToString(m.Port.ValueInt64()),
Weight:  opnsense.Int64ToString(m.Weight.ValueInt64()),  // handle null/unknown
SSL:     opnsense.BoolToString(m.SSL.ValueBool()),
Enabled: opnsense.BoolToString(m.Enabled.ValueBool()),
Mode:    m.Mode.ValueString(),
```

**fromAPI() — key conversions:**
```go
// Port (required, always has value)
portVal, _ := opnsense.StringToInt64(a.Port)
m.Port = types.Int64Value(portVal)

// Weight (optional — may be empty string from API)
if a.Weight != "" {
    weightVal, _ := opnsense.StringToInt64(a.Weight)
    m.Weight = types.Int64Value(weightVal)
} else {
    m.Weight = types.Int64Null()
}

// Booleans
m.SSL = types.BoolValue(opnsense.StringToBool(a.SSL))
m.SSLVerify = types.BoolValue(opnsense.StringToBool(a.SSLVerify))
m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))

// Mode (SelectedMap → string)
m.Mode = types.StringValue(string(a.Mode))
```

### Schema Pattern for Integer Attributes

```go
"port": schema.Int64Attribute{
    Required:            true,
    MarkdownDescription: "Port number of the backend server (1-65535).",
    Validators: []validator.Int64{
        int64validator.Between(1, 65535),
    },
},
"weight": schema.Int64Attribute{
    Optional:            true,
    Computed:            true,
    MarkdownDescription: "Load balancing weight (0-256). Higher values receive more traffic.",
},
```

**Required imports for int64 validator:**
```go
"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
"github.com/hashicorp/terraform-plugin-framework/schema/validator"
```

### Provider Registration

Update `internal/provider/provider.go` to append haproxy resources:
```go
import (
    "github.com/matthew-on-git/terraform-provider-opnsense/internal/service/firewall"
    "github.com/matthew-on-git/terraform-provider-opnsense/internal/service/haproxy"
)

func (p *OpnsenseProvider) Resources(_ context.Context) []func() resource.Resource {
    var resources []func() resource.Resource
    resources = append(resources, firewall.Resources()...)
    resources = append(resources, haproxy.Resources()...)
    return resources
}
```

### Acceptance Test Pattern

**Test name:** `TestAccHAProxyServer_basic`

**Test config function:**
```go
func testAccHAProxyServerConfig(name, address string, port int) string {
    return fmt.Sprintf(`
resource "opnsense_haproxy_server" "test" {
  name    = %[1]q
  address = %[2]q
  port    = %[3]d
}
`, name, address, port)
}
```

**Test steps must include:** Create → Verify (id, name, address, port, enabled=true, ssl=false) → Import → Update (change port) → Verify → Destroy with CheckDestroy.

**CheckDestroy must check for resource type `"opnsense_haproxy_server"`.**

### What NOT to Build

- No HAProxy controller in `pkg/opnsense/haproxy/` — use generic CRUD functions directly with `ReqOpts`
- No backend, frontend, ACL, or health check resources — those are Epic 4
- No data source — that pattern is proven in Story 2.3
- No custom validators beyond framework-provided `int64validator.Between` and `stringvalidator.OneOf`

### Previous Story Intelligence

**From Story 2.1 (Firewall Alias Resource):**
- Used dual-struct pattern: `aliasAPIRequest` (plain strings for POST) and `aliasAPIResponse` (SelectedMap for GET) — follow this exactly
- `fromAPI()` parameter `ctx` was flagged unused by `revive` — use `_ context.Context` if ctx not needed
- `CheckDestroy` is required in acceptance tests (added during code review)
- `gitleaks` scan fails on main with 9 pre-existing false positives — not a blocker
- `make check` passes 5/6 targets (lint, format, test, security, docs); scan pre-existing failure
- `types.Set` used for unordered collections; `types.Int64` for integer fields
- `stringdefault.StaticString("")` and `booldefault.StaticBool(true)` for defaults
- Schema uses `MarkdownDescription` (not `Description`) for all attributes
- Test resource names use `"test"`: `resource "opnsense_haproxy_server" "test"`
- Test names prefixed with `tf_test_` to avoid conflicts with real data

**From Epic 1 Retrospective:**
- `ctx context.Context` is ALWAYS the first function parameter (revive enforces)
- `make check` must pass all targets before marking done
- DevRail container: `ghcr.io/devrail-dev/dev-toolchain:1.8.1`

### Project Structure Notes

**New files this story creates:**
```
internal/
└── service/
    └── haproxy/
        ├── exports.go                       # NEW: Resources() / DataSources()
        ├── server_resource.go               # NEW: CRUD + ImportState
        ├── server_schema.go                 # NEW: Schema()
        ├── server_model.go                  # NEW: models + toAPI/fromAPI
        └── server_resource_test.go          # NEW: acceptance tests

examples/
└── resources/
    └── opnsense_haproxy_server/
        ├── resource.tf                      # NEW: example HCL
        └── import.sh                        # NEW: import example

templates/
└── resources/
    └── haproxy_server.md.tmpl               # NEW: doc template
```

**Modified files:**
```
internal/provider/provider.go                # MODIFIED: register haproxy.Resources()
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-2, Story 2.2]
- [Source: _bmad-output/planning-artifacts/architecture.md#AR11 four-file pattern]
- [Source: _bmad-output/planning-artifacts/architecture.md#AR8 PluginNotFoundError]
- [Source: _bmad-output/planning-artifacts/prd.md#FR24 HAProxy servers]
- [Source: _bmad-output/implementation-artifacts/2-1-firewall-alias-resource.md#Patterns and learnings]
- [Source: https://docs.opnsense.org/development/api/plugins/haproxy.html#Server endpoints]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- No linting issues — all `make check` targets pass cleanly (except pre-existing gitleaks scan)
- Provider Resources() method changed from returning single slice to appending firewall + haproxy slices
- Integer fields (port, weight) require `Int64ToString`/`StringToInt64` from `pkg/opnsense/types.go`
- Weight is optional — `fromAPI()` sets `types.Int64Null()` when API returns empty string

### Completion Notes List

- Implemented complete `opnsense_haproxy_server` resource with four-file pattern in `internal/service/haproxy/`
- CRUD operations follow identical pattern to Story 2.1 — validates pattern generalizes to plugin APIs
- Three boolean fields (`ssl`, `ssl_verify`, `enabled`) all convert correctly via `BoolToString`/`StringToBool`
- Integer fields (`port`, `weight`) convert via `Int64ToString`/`StringToInt64` with null handling for optional weight
- `SelectedMap` used for `mode` field (active/backup/disabled)
- Plugin detection via `PluginNotFoundError` handled automatically by existing `CheckHTTPError` — no custom code needed
- Schema includes `int64validator.Between` for port (1-65535) and weight (0-256), `stringvalidator.OneOf` for mode
- Provider registration updated to append haproxy.Resources() to firewall.Resources()
- Acceptance test covers full lifecycle with CheckDestroy
- `make check` passes 5/6 targets (lint, format, test, security, docs); scan fails due to pre-existing gitleaks findings

### File List

- `internal/service/haproxy/exports.go` — NEW: Resources() and DataSources() registration
- `internal/service/haproxy/server_resource.go` — NEW: CRUD + ImportState + Configure
- `internal/service/haproxy/server_schema.go` — NEW: Terraform schema definition
- `internal/service/haproxy/server_model.go` — NEW: ServerResourceModel, serverAPIRequest, serverAPIResponse, toAPI(), fromAPI()
- `internal/service/haproxy/server_resource_test.go` — NEW: acceptance test with CheckDestroy
- `internal/provider/provider.go` — MODIFIED: register haproxy.Resources(), append to firewall resources
- `examples/resources/opnsense_haproxy_server/resource.tf` — NEW: example HCL
- `examples/resources/opnsense_haproxy_server/import.sh` — NEW: import example
- `templates/resources/haproxy_server.md.tmpl` — NEW: documentation template
