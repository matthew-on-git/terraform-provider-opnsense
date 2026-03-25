# Story 3.4: Firewall NAT Port Forward Resource

Status: done

## Story

As an operator,
I want to manage NAT port-forward rules through Terraform,
So that I can expose internal services through the firewall as code.

## Acceptance Criteria

1. **Given** the provider is configured with valid OPNsense credentials
   **When** the operator defines an `opnsense_firewall_nat_port_forward` resource in HCL
   **Then** `terraform apply` creates the NAT rule on OPNsense via the API and returns the UUID

2. **And** `terraform plan` with no changes shows "No changes"

3. **And** modifying attributes shows the correct diff and applies the change

4. **And** removing the resource block deletes the NAT rule

5. **And** `terraform import` works by UUID, and after import `terraform plan` shows "No changes"

6. **And** acceptance test covers full lifecycle: Create → Verify → Import → Update → Destroy with CheckDestroy

## Tasks / Subtasks

- [x] Task 1: Create `nat_port_forward_model.go` with API structs and conversions (AC: #1, #2)
  - [x] 1.1 Define `natPortForwardAPIResponse` with SelectedMap for enum fields, nested dot-key JSON tags
  - [x] 1.2 Define `natPortForwardAPIRequest` with plain string fields, dot-key JSON tags
  - [x] 1.3 Define `NatPortForwardResourceModel` Terraform model struct
  - [x] 1.4 Implement `toAPI()` and `fromAPI()` conversions
- [x] Task 2: Create `nat_port_forward_schema.go` (AC: #1, #3)
  - [x] 2.1 Schema with target, local_port, interface, protocol, source/destination fields, enabled, description
- [x] Task 3: Create `nat_port_forward_resource.go` with CRUD + ImportState (AC: #1-#6)
  - [x] 3.1 Package-level ReqOpts with ReconfigureEndpoint `/api/firewall/d_nat/apply`
  - [x] 3.2 Implement all CRUD methods + ImportState
- [x] Task 4: Register in `exports.go` (AC: all)
  - [x] 4.1 Add `newNatPortForwardResource` to `firewall.Resources()` slice
- [x] Task 5: Create acceptance test (AC: #6)
  - [x] 5.1 Full lifecycle test with CheckDestroy
- [x] Task 6: Create documentation and examples (AC: all)
  - [x] 6.1 Create `examples/resources/opnsense_firewall_nat_port_forward/resource.tf`
  - [x] 6.2 Create `examples/resources/opnsense_firewall_nat_port_forward/import.sh`
  - [x] 6.3 Create `templates/resources/firewall_nat_port_forward.md.tmpl`
- [x] Task 7: Run `make check` and verify all targets pass (AC: all)

## Dev Notes

### OPNsense DNat (Port Forward) API

**Controller:** `DNatController.php` extends `FilterBaseController`
**Base URL:** `/api/firewall/d_nat/`
**IMPORTANT:** Endpoints use snake_case (`add_rule`, `get_rule`), NOT camelCase.

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Create | POST | `/api/firewall/d_nat/add_rule` | Body: `{"rule":{...}}`, returns UUID |
| Read | GET | `/api/firewall/d_nat/get_rule/{uuid}` | Returns `{"rule":{...}}` |
| Update | POST | `/api/firewall/d_nat/set_rule/{uuid}` | Body: `{"rule":{...}}` |
| Delete | POST | `/api/firewall/d_nat/del_rule/{uuid}` | |
| Search | GET/POST | `/api/firewall/d_nat/search_rule` | Paginated |
| Apply | POST | `/api/firewall/d_nat/apply` | Activates changes |

**Monad key:** `"rule"`

**ReqOpts:**
```go
var natPortForwardReqOpts = opnsense.ReqOpts{
    AddEndpoint:         "/api/firewall/d_nat/add_rule",
    GetEndpoint:         "/api/firewall/d_nat/get_rule",
    UpdateEndpoint:      "/api/firewall/d_nat/set_rule",
    DeleteEndpoint:      "/api/firewall/d_nat/del_rule",
    SearchEndpoint:      "/api/firewall/d_nat/search_rule",
    ReconfigureEndpoint: "/api/firewall/d_nat/apply",
    Monad:               "rule",
}
```

### CRITICAL: Nested Dot-Key JSON Fields

The DNat model uses **dot-separated nested field names** in the API JSON:
```json
{
  "rule": {
    "disabled": "0",
    "interface": "wan",
    "protocol": "tcp",
    "source.network": "any",
    "source.port": "",
    "source.not": "0",
    "destination.network": "wanip",
    "destination.port": "443",
    "destination.not": "0",
    "target": "10.0.0.1",
    "local-port": "443",
    "log": "0",
    "descr": ""
  }
}
```

Go JSON tags handle dots fine: `json:"source.network"`. The field name in the API uses `descr` (not `description`).

### MVP Schema — Core Attributes

| API Field | Go Type (Response) | Terraform Attribute | Terraform Type | Default |
|-----------|--------------------|---------------------|----------------|---------|
| (UUID) | `string` | `id` | `types.String` (Computed) | — |
| `disabled` | `string` ("0"/"1") | `enabled` | `types.Bool` | `true` |
| `interface` | `SelectedMap` | `interface` | `types.String` (Required) | — |
| `ipprotocol` | `SelectedMap` | `ip_protocol` | `types.String` | `"inet"` |
| `protocol` | `SelectedMap` | `protocol` | `types.String` | `"any"` |
| `source.network` | `string` | `source_net` | `types.String` | `"any"` |
| `source.port` | `string` | `source_port` | `types.String` | `""` |
| `source.not` | `string` ("0"/"1") | `source_not` | `types.Bool` | `false` |
| `destination.network` | `string` | `destination_net` | `types.String` | — |
| `destination.port` | `string` | `destination_port` | `types.String` (Required) | — |
| `destination.not` | `string` ("0"/"1") | `destination_not` | `types.Bool` | `false` |
| `target` | `string` | `target` | `types.String` (Required) | — |
| `local-port` | `string` | `local_port` | `types.String` (Required) | — |
| `log` | `string` ("0"/"1") | `log` | `types.Bool` | `false` |
| `descr` | `string` | `description` | `types.String` | `""` |
| `categories` | `SelectedMapList` | `categories` | `types.Set` of `types.String` | — |

**NOTE on `disabled` vs `enabled`:** The DNat model uses `disabled` (inverted logic). In `toAPI()`, send the INVERTED value: `Disabled: opnsense.BoolToString(!m.Enabled.ValueBool())`. In `fromAPI()`, invert: `m.Enabled = types.BoolValue(!opnsense.StringToBool(a.Disabled))`.

**NOTE on `descr`:** The API field is `descr` (not `description`). Map it to Terraform attribute `description` for consistency.

**NOTE on `local-port`:** The API field has a hyphen. Go JSON tag: `json:"local-port"`.

### API Model Structs

**Response:**
```go
type natPortForwardAPIResponse struct {
    Disabled        string                   `json:"disabled"`
    Interface       opnsense.SelectedMap     `json:"interface"`
    IPProtocol      opnsense.SelectedMap     `json:"ipprotocol"`
    Protocol        opnsense.SelectedMap     `json:"protocol"`
    SourceNetwork   string                   `json:"source.network"`
    SourcePort      string                   `json:"source.port"`
    SourceNot       string                   `json:"source.not"`
    DestNetwork     string                   `json:"destination.network"`
    DestPort        string                   `json:"destination.port"`
    DestNot         string                   `json:"destination.not"`
    Target          string                   `json:"target"`
    LocalPort       string                   `json:"local-port"`
    Log             string                   `json:"log"`
    Description     string                   `json:"descr"`
    Categories      opnsense.SelectedMapList `json:"categories"`
}
```

**Request:**
```go
type natPortForwardAPIRequest struct {
    Disabled        string `json:"disabled"`
    Interface       string `json:"interface"`
    IPProtocol      string `json:"ipprotocol"`
    Protocol        string `json:"protocol"`
    SourceNetwork   string `json:"source.network"`
    SourcePort      string `json:"source.port"`
    SourceNot       string `json:"source.not"`
    DestNetwork     string `json:"destination.network"`
    DestPort        string `json:"destination.port"`
    DestNot         string `json:"destination.not"`
    Target          string `json:"target"`
    LocalPort       string `json:"local-port"`
    Log             string `json:"log"`
    Description     string `json:"descr"`
    Categories      string `json:"categories"`
}
```

### Standard ReconfigureEndpoint (NOT Savepoint)

Uses standard `ReconfigureEndpoint` (package-level ReqOpts). The `apply` endpoint without a revision parameter just applies changes immediately. No instance-level ReqOpts needed.

### What NOT to Build

- No savepoint flow — use standard apply endpoint
- No `nordr`, `poolopts`, `natreflection`, `pass`, `tag`, `tagged`, `nosync` fields — advanced, defer
- No outbound NAT — that's Story 3.5
- No data source — deferred to Epic 12

### Previous Story Intelligence

**From Story 3.3 (Filter Rule):**
- Complex model with 18 attributes — similar complexity here
- Instance-level ReqOpts pattern NOT needed here (standard ReconfigureEndpoint)
- SelectedMap for enum fields, SelectedMapList for categories

**From Story 3.2 (Category):**
- Package-level ReqOpts pattern (simpler, use for NAT)
- Clean CRUD pattern to follow

### Project Structure Notes

**New files:**
```
internal/service/firewall/
├── nat_port_forward_resource.go
├── nat_port_forward_schema.go
├── nat_port_forward_model.go
└── nat_port_forward_resource_test.go

examples/resources/opnsense_firewall_nat_port_forward/
├── resource.tf
└── import.sh

templates/resources/
└── firewall_nat_port_forward.md.tmpl
```

**Modified files:**
```
internal/service/firewall/exports.go
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-3, Story 3.4]
- [Source: _bmad-output/planning-artifacts/prd.md#FR22]
- [Source: https://docs.opnsense.org/development/api/core/firewall.html#DNat endpoints]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- `gofumpt` formatting fix required — struct field alignment adjusted by `make fix`
- Dot-separated JSON tags (`json:"source.network"`, `json:"local-port"`) work correctly in Go
- Inverted `disabled` ↔ `enabled` logic verified: `!m.Enabled.ValueBool()` in toAPI, `!opnsense.StringToBool(a.Disabled)` in fromAPI
- DNat endpoints use snake_case (`add_rule`, `get_rule`) unlike other firewall controllers (camelCase)

### Completion Notes List

- Implemented `opnsense_firewall_nat_port_forward` resource with four-file pattern
- DNat controller at `/api/firewall/d_nat/` with snake_case endpoints
- Dot-separated JSON tags for nested fields: `source.network`, `source.port`, `destination.network`, etc.
- Inverted boolean: API `disabled` field mapped to Terraform `enabled` attribute (inverted in toAPI/fromAPI)
- API field `descr` mapped to Terraform attribute `description`
- API field `local-port` (hyphenated) mapped to Terraform attribute `local_port` (underscored)
- Standard ReconfigureEndpoint (`/api/firewall/d_nat/apply`) — package-level ReqOpts
- 16 attributes including 4 Required fields: interface, destination_net, destination_port, target, local_port
- `make check` passes 5/6 targets; format fix applied via `make fix`

### File List

- `internal/service/firewall/nat_port_forward_resource.go` — NEW: CRUD + ImportState
- `internal/service/firewall/nat_port_forward_schema.go` — NEW: Terraform schema
- `internal/service/firewall/nat_port_forward_model.go` — NEW: Models + toAPI/fromAPI with dot-key JSON tags
- `internal/service/firewall/nat_port_forward_resource_test.go` — NEW: acceptance test with CheckDestroy
- `internal/service/firewall/exports.go` — MODIFIED: added newNatPortForwardResource
- `examples/resources/opnsense_firewall_nat_port_forward/resource.tf` — NEW: example HCL
- `examples/resources/opnsense_firewall_nat_port_forward/import.sh` — NEW: import example
- `templates/resources/firewall_nat_port_forward.md.tmpl` — NEW: documentation template
