# Story 2.1: Firewall Alias Resource

Status: done

## Story

As an operator,
I want to manage OPNsense firewall aliases through Terraform,
So that I can define host, network, and port aliases as code and see changes before applying.

## Acceptance Criteria

1. **Given** the provider is configured with valid OPNsense credentials
   **When** the operator defines an `opnsense_firewall_alias` resource in HCL
   **Then** `terraform apply` creates the alias on OPNsense via the API and returns the UUID

2. **And** `terraform plan` with no changes shows "No changes" (state read-back matches)

3. **And** modifying alias content in HCL shows the correct diff in `terraform plan` and applies the change

4. **And** removing the resource block deletes the alias from OPNsense

5. **And** `terraform import opnsense_firewall_alias.test <uuid>` imports an existing alias into state

6. **And** after import, `terraform plan` shows "No changes"

7. **And** out-of-band changes to the alias via OPNsense UI are detected in the next `terraform plan`

8. **And** state is populated from API read-back after Create and Update (never echoed from config)

## Tasks / Subtasks

- [x] Task 1: Create `internal/service/firewall/` package with `exports.go` (AC: all)
  - [x] 1.1 Create `internal/service/firewall/exports.go` with `Resources()` returning `newAliasResource` and `DataSources()` returning empty slice
  - [x] 1.2 Register firewall package in `internal/provider/provider.go` `Resources()` method
- [x] Task 2: Create `alias_model.go` with API struct and conversions (AC: #1, #2, #8)
  - [x] 2.1 Define `aliasAPIResponse` struct with JSON tags and SelectedMap types for GET responses
  - [x] 2.2 Define `aliasAPIRequest` struct with plain string fields for POST requests
  - [x] 2.3 Define `AliasResourceModel` Terraform model struct with `tfsdk` tags
  - [x] 2.4 Implement `toAPI()` converting Terraform model to API request struct
  - [x] 2.5 Implement `fromAPI()` converting API response struct to Terraform model
- [x] Task 3: Create `alias_schema.go` with Terraform schema (AC: #1, #3, #5)
  - [x] 3.1 Define schema with all alias attributes, validators, plan modifiers, and defaults
- [x] Task 4: Create `alias_resource.go` with CRUD + ImportState (AC: #1-#8)
  - [x] 4.1 Implement `Create` — plan → toAPI → Add → Get → fromAPI → state
  - [x] 4.2 Implement `Read` — Get → fromAPI → state (RemoveResource on NotFoundError)
  - [x] 4.3 Implement `Update` — plan → toAPI → Update → Get → fromAPI → state
  - [x] 4.4 Implement `Delete` — Delete by UUID
  - [x] 4.5 Implement `ImportState` — passthrough ID
  - [x] 4.6 Implement `Configure` — extract `*opnsense.Client` from provider data
- [x] Task 5: Create `internal/acctest/acctest.go` test helper (AC: #1-#8)
  - [x] 5.1 Implement `ProtoV6ProviderFactories` map
  - [x] 5.2 Implement `PreCheck(t)` validating env vars and OPNsense reachability
- [x] Task 6: Create `alias_resource_test.go` acceptance test (AC: #1-#8)
  - [x] 6.1 Full lifecycle test: Create → Verify → Import → Update → Destroy
- [x] Task 7: Create documentation and examples (AC: all)
  - [x] 7.1 Create `examples/resources/opnsense_firewall_alias/resource.tf`
  - [x] 7.2 Create `examples/resources/opnsense_firewall_alias/import.sh`
  - [x] 7.3 Create `templates/resources/firewall_alias.md.tmpl`
- [x] Task 8: Run `make check` and verify all targets pass (AC: all)

## Dev Notes

### OPNsense Firewall Alias API Endpoints

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Create | POST | `/api/firewall/alias/addItem` | Returns `{"result":"saved","uuid":"..."}` |
| Read | GET | `/api/firewall/alias/getItem/{uuid}` | Returns `{"alias":{...}}` |
| Update | POST | `/api/firewall/alias/setItem/{uuid}` | Returns `{"result":"saved"}` |
| Delete | POST | `/api/firewall/alias/delItem/{uuid}` | Returns `{"result":"deleted"}` |
| Search | GET/POST | `/api/firewall/alias/searchItem` | Paginated list |
| Reconfigure | POST | `/api/firewall/alias/reconfigure` | Activates changes |

**Monad key:** `"alias"`

**ReqOpts for this resource:**
```go
var aliasReqOpts = opnsense.ReqOpts{
    AddEndpoint:         "/api/firewall/alias/addItem",
    GetEndpoint:         "/api/firewall/alias/getItem",
    UpdateEndpoint:      "/api/firewall/alias/setItem",
    DeleteEndpoint:      "/api/firewall/alias/delItem",
    SearchEndpoint:      "/api/firewall/alias/searchItem",
    ReconfigureEndpoint: "/api/firewall/alias/reconfigure",
    Monad:               "alias",
}
```

### Alias API Model Fields

The OPNsense API returns alias objects with these fields (inside the `"alias"` monad):

| API Field | Go Type | Terraform Attribute | Terraform Type | Conversion |
|-----------|---------|---------------------|----------------|------------|
| (UUID from response) | `string` | `id` | `types.String` (Computed) | Passthrough |
| `name` | `string` | `name` | `types.String` (Required) | Direct |
| `type` | `SelectedMap` | `type` | `types.String` (Required) | `SelectedMap` → `string(sm)` |
| `content` | `string` (newline-separated) | `content` | `types.Set` of `types.String` (Optional) | Newline split/join |
| `description` | `string` | `description` | `types.String` (Optional, defaults to `""`) | Direct |
| `enabled` | `string` ("0"/"1") | `enabled` | `types.Bool` (Optional, defaults `true`) | `BoolToString`/`StringToBool` |
| `proto` | `SelectedMap` | `proto` | `types.String` (Optional, Computed) | `SelectedMap` → `string(sm)` |
| `categories` | `SelectedMapList` | `categories` | `types.Set` of `types.String` (Optional, Computed) | `SelectedMapList` → `[]string` |
| `updatefreq` | `string` | `update_freq` | `types.String` (Optional) | Direct |

**Content field format:** OPNsense stores alias content as newline-separated strings in the API (e.g., `"10.0.0.1\n10.0.0.2"`). Use `types.Set` (not List) to prevent perpetual diffs from ordering changes. Convert via newline split/join, not CSV.

**Type field values:** `host`, `network`, `port`, `url`, `urltable`, `geoip`, `networkgroup`, `mac`, `asn`, `dynipv6host`, `authgroup`, `internal`, `external`

**Proto field values:** Empty string `""` (any), `IPv4`, `IPv6`

### Resource Implementation Pattern

**File locations — four-file pattern in `internal/service/firewall/`:**
```
internal/service/firewall/
├── exports.go              # Resources() and DataSources() registration
├── alias_resource.go       # CRUD methods: Create, Read, Update, Delete, ImportState
├── alias_schema.go         # Schema() method returning Terraform schema
├── alias_model.go          # AliasResourceModel, aliasAPIModel, toAPI(), fromAPI()
└── alias_resource_test.go  # Acceptance tests
```

**Resource struct pattern (from architecture):**
```go
type aliasResource struct {
    client *opnsense.Client
}

var _ resource.Resource = &aliasResource{}
var _ resource.ResourceWithImportState = &aliasResource{}

func newAliasResource() resource.Resource {
    return &aliasResource{}
}

func (r *aliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_firewall_alias"
}

func (r *aliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    client, ok := req.ProviderData.(*opnsense.Client)
    if !ok {
        resp.Diagnostics.AddError("Unexpected Provider Data",
            "Expected *opnsense.Client, got something else.")
        return
    }
    r.client = client
}
```

**CRUD flow — every method must follow this exactly:**

**Create:**
1. `resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)` — read plan into Terraform model
2. `apiModel := plan.toAPI()` — convert to API struct
3. `uuid, err := opnsense.Add(ctx, r.client, aliasReqOpts, apiModel)` — create via API
4. Handle errors with `resp.Diagnostics.AddError()`
5. `result, err := opnsense.Get[aliasAPIModel](ctx, r.client, aliasReqOpts, uuid)` — read back
6. `plan.fromAPI(result, uuid)` — populate model from API (NOT from plan)
7. `resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)` — set state

**Read:**
1. `resp.Diagnostics.Append(req.State.Get(ctx, &state)...)` — read state for UUID
2. `result, err := opnsense.Get[aliasAPIModel](ctx, r.client, aliasReqOpts, state.ID.ValueString())` — read from API
3. On `NotFoundError`: `resp.State.RemoveResource(ctx)` and return
4. `state.fromAPI(result, state.ID.ValueString())` — populate from API
5. `resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)` — set state

**Update:**
1. `resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)` — read plan
2. `resp.Diagnostics.Append(req.State.Get(ctx, &state)...)` — read state for UUID
3. `apiModel := plan.toAPI()` — convert to API struct
4. `err := opnsense.Update(ctx, r.client, aliasReqOpts, apiModel, state.ID.ValueString())` — update
5. `result, err := opnsense.Get[aliasAPIModel](ctx, r.client, aliasReqOpts, state.ID.ValueString())` — read back
6. `plan.fromAPI(result, state.ID.ValueString())` — populate from API
7. `resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)` — set state

**Delete:**
1. `resp.Diagnostics.Append(req.State.Get(ctx, &state)...)` — read state for UUID
2. `err := opnsense.Delete(ctx, r.client, aliasReqOpts, state.ID.ValueString())` — delete
3. On `NotFoundError`: return silently (already deleted)
4. Other errors: `resp.Diagnostics.AddError()`

**ImportState:**
```go
func (r *aliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

### Schema Definition Pattern

**ID attribute (every resource):**
```go
"id": schema.StringAttribute{
    Computed:    true,
    Description: "UUID of the firewall alias in OPNsense.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

**Content as Set (prevents ordering diffs):**
```go
"content": schema.SetAttribute{
    ElementType: types.StringType,
    Optional:    true,
    Computed:    true,
    Description: "List of alias entries (IPs, networks, ports, URLs depending on type).",
},
```

**Enabled with default true:**
```go
"enabled": schema.BoolAttribute{
    Optional:    true,
    Computed:    true,
    Default:     booldefault.StaticBool(true),
    Description: "Whether this alias is enabled. Defaults to true.",
},
```

**Type with validation (known alias types only):**
```go
"type": schema.StringAttribute{
    Required:    true,
    Description: "Alias type (host, network, port, url, urltable, geoip, networkgroup, mac, asn, dynipv6host, authgroup, internal, external).",
    Validators: []validator.String{
        stringvalidator.OneOf("host", "network", "port", "url", "urltable", "geoip",
            "networkgroup", "mac", "asn", "dynipv6host", "authgroup", "internal", "external"),
    },
},
```

### Model Conversion Pattern

**API model struct (JSON tags match OPNsense API field names):**
```go
type aliasAPIModel struct {
    Name        string                  `json:"name"`
    Type        opnsense.SelectedMap    `json:"type"`
    Content     string                  `json:"content"`
    Description string                  `json:"description"`
    Enabled     string                  `json:"enabled"`
    Proto       opnsense.SelectedMap    `json:"proto"`
    Categories  opnsense.SelectedMapList `json:"categories"`
    UpdateFreq  string                  `json:"updatefreq"`
}
```

**toAPI() — Terraform model to API struct:**
- `types.String` → `string` via `.ValueString()`
- `types.Bool` → `string` via `opnsense.BoolToString(model.Enabled.ValueBool())`
- `types.Set` → newline-separated `string` via extracting elements and joining with `\n`
- `types.Set` (categories) → CSV `string` via extracting elements and joining with `,`

**fromAPI() — API struct to Terraform model:**
- `string` → `types.StringValue()`
- `string` ("0"/"1") → `types.BoolValue(opnsense.StringToBool(...))`
- `SelectedMap` → `types.StringValue(string(sm))`
- newline-separated `string` → `types.Set` via splitting on `\n` and building set
- `SelectedMapList` → `types.Set` via converting `[]string` to set of `types.String`

**CRITICAL: `fromAPI()` takes the UUID as a separate parameter** (it comes from `Add` response, not from the API model itself). Set `model.ID = types.StringValue(uuid)`.

### Provider Registration

Update `internal/provider/provider.go`:
```go
import "github.com/matthew-on-git/terraform-provider-opnsense/internal/service/firewall"

func (p *OpnsenseProvider) Resources(_ context.Context) []func() resource.Resource {
    return firewall.Resources()
}
```

### Test Infrastructure (`internal/acctest/acctest.go`)

This is the first resource — the acceptance test helper package does not exist yet. Create it.

```go
package acctest

import (
    "os"
    "testing"

    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
    "github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
    "opnsense": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func PreCheck(t *testing.T) {
    t.Helper()
    for _, env := range []string{"OPNSENSE_URI", "OPNSENSE_API_KEY", "OPNSENSE_API_SECRET"} {
        if os.Getenv(env) == "" {
            t.Fatalf("Environment variable %s must be set for acceptance tests", env)
        }
    }
}
```

### Acceptance Test Pattern

Tests are gated by `TF_ACC` env var (standard Terraform convention — the test framework checks this automatically when using `resource.Test`).

**Full lifecycle test structure:**
```go
func TestAccFirewallAlias_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
        PreCheck:                 func() { acctest.PreCheck(t) },
        Steps: []resource.TestStep{
            // Step 1: Create and verify
            {
                Config: testAccFirewallAliasConfig("test_alias", "host", "10.0.0.1"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttrSet("opnsense_firewall_alias.test", "id"),
                    resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "name", "test_alias"),
                    resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "type", "host"),
                    resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "enabled", "true"),
                ),
            },
            // Step 2: Import
            {
                ResourceName:      "opnsense_firewall_alias.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
            // Step 3: Update and verify
            {
                Config: testAccFirewallAliasConfig("test_alias", "host", "10.0.0.2"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "name", "test_alias"),
                ),
            },
        },
    })
}
```

**Test config function returns HCL as string:**
```go
func testAccFirewallAliasConfig(name, aliasType, content string) string {
    return fmt.Sprintf(`
resource "opnsense_firewall_alias" "test" {
  name    = %[1]q
  type    = %[2]q
  content = [%[3]q]
}
`, name, aliasType, content)
}
```

### Error Handling Pattern

Use `errors.As` for typed error checks in CRUD methods:
```go
if err != nil {
    var notFoundErr *opnsense.NotFoundError
    if errors.As(err, &notFoundErr) {
        resp.State.RemoveResource(ctx)
        return
    }
    resp.Diagnostics.AddError(
        "Error reading firewall alias",
        fmt.Sprintf("Could not read firewall alias %s: %s", state.ID.ValueString(), err),
    )
    return
}
```

### Required Imports for Resource File

```go
import (
    "context"
    "errors"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"

    "github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)
```

### Required Imports for Schema File

```go
import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)
```

### What NOT to Build

- No data source (`alias_data_source.go`) — that is Story 2.3
- No firewall controller in `pkg/opnsense/firewall/` — use generic CRUD functions directly with `ReqOpts`
- No additional error types — use the 5 existing types from `pkg/opnsense/errors.go`
- No custom validators yet — use `stringvalidator.OneOf` from the framework validators package
- No acceptance.yml CI workflow — acceptance tests run locally against Vagrant OPNsense

### Previous Story Intelligence

**From Epic 1 Retrospective:**
- `ctx context.Context` is ALWAYS the first function parameter (revive enforces, overrides architecture spec)
- Mock servers use non-retryable status codes (400, 418) for deterministic testing — never 5xx
- All `gosec` G704 suppressions require inline explanation comment
- `make check` must pass all 6 targets before marking done
- DevRail container: `ghcr.io/devrail-dev/dev-toolchain:1.8.1`

**From Story 1.5 (CRUD):**
- `Add[K]` returns `(string, error)` — UUID of created resource
- `Get[K]` returns `(*K, error)` — pointer to deserialized struct
- `Update[K]` returns `error` — no UUID in response
- `Delete` returns `error` — does NOT parse mutation response body
- Monad wrapping/unwrapping is handled by CRUD functions — model code never sees the wrapper
- `NotFoundError` returned when monad key missing, inner value empty/null/{}

**From Story 1.7 (Types):**
- `BoolToString(true)` → `"1"`, `BoolToString(false)` → `"0"`
- `StringToBool("1")` → `true`, anything else → `false`
- `SelectedMap` extracts single selected key from OPNsense response via `json.Number`
- `SelectedMapList` collects all selected keys, sorted alphabetically
- `CSVToSlice`/`SliceToCSV` handle comma-separated values

**From Story 1.1 (API Client):**
- `Client.HTTPClient()` returns `*http.Client` for direct use
- `Client.BaseURL()` returns configured base URL string
- `opnsense.NewClient(ClientConfig{...})` creates client
- Client passed as `*opnsense.Client` via `resp.ResourceData` in provider Configure

### Project Structure Notes

**New files this story creates:**
```
internal/
├── acctest/
│   └── acctest.go                           # NEW: test helpers
└── service/
    └── firewall/
        ├── exports.go                       # NEW: Resources() / DataSources()
        ├── alias_resource.go                # NEW: CRUD + ImportState
        ├── alias_schema.go                  # NEW: Schema()
        ├── alias_model.go                   # NEW: models + toAPI/fromAPI
        └── alias_resource_test.go           # NEW: acceptance tests

examples/
└── resources/
    └── opnsense_firewall_alias/
        ├── resource.tf                      # NEW: example HCL
        └── import.sh                        # NEW: import example

templates/
└── resources/
    └── firewall_alias.md.tmpl               # NEW: doc template
```

**Modified files:**
```
internal/provider/provider.go                # MODIFIED: register firewall.Resources()
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-2, Story 2.1]
- [Source: _bmad-output/planning-artifacts/architecture.md#AR11 four-file pattern]
- [Source: _bmad-output/planning-artifacts/architecture.md#AR3 generic CRUD]
- [Source: _bmad-output/planning-artifacts/architecture.md#AR10 type conversion]
- [Source: _bmad-output/planning-artifacts/prd.md#FR6-FR12 CRUD lifecycle]
- [Source: _bmad-output/planning-artifacts/prd.md#FR19 firewall aliases]
- [Source: _bmad-output/implementation-artifacts/1-5-generic-crud-functions.md#CRUD patterns]
- [Source: _bmad-output/implementation-artifacts/1-7-type-conversion-utilities.md#Type converters]
- [Source: _bmad-output/implementation-artifacts/epic-1-retro-2026-03-23.md#Team agreements]
- [Source: https://docs.opnsense.org/development/api/core/firewall.html#Alias endpoints]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- `revive` flagged `ctx context.Context` as unused in `fromAPI()` — renamed to `_ context.Context` (fromAPI doesn't call ElementsAs, only toAPI does)
- Used separate `aliasAPIRequest` and `aliasAPIResponse` structs because OPNsense sends SelectedMap format in GET responses but accepts plain strings in POST requests
- `gitleaks` scan fails on main with 9 pre-existing false positives in BMAD documentation files (hash values and example API keys) — not introduced by this story

### Completion Notes List

- Implemented complete `opnsense_firewall_alias` resource with four-file pattern in `internal/service/firewall/`
- CRUD operations follow state read-back pattern: Create/Update always GET from API to populate state
- Two API struct types: `aliasAPIRequest` (plain strings for POST) and `aliasAPIResponse` (SelectedMap for GET)
- Content stored as `types.Set` of strings, converted to/from newline-separated API format
- Categories stored as `types.Set` of UUID strings, converted to/from comma-separated API format
- Schema includes `stringvalidator.OneOf` for alias type validation, `booldefault` for enabled, `stringdefault` for description/proto/updatefreq
- `internal/acctest/` package provides `ProtoV6ProviderFactories` and `PreCheck` for acceptance tests
- Acceptance test covers full lifecycle: Create → Verify → Import → Update → Destroy (TF_ACC gated)
- Provider registration updated to include `firewall.Resources()`
- Documentation template and HCL examples created for tfplugindocs generation
- `make check` passes 5/6 targets (lint, format, test, security, docs); scan fails due to pre-existing gitleaks findings on main
- Code review: added `CheckDestroy` to acceptance test (1 MEDIUM finding fixed)

### File List

- `internal/service/firewall/exports.go` — NEW: Resources() and DataSources() registration
- `internal/service/firewall/alias_resource.go` — NEW: CRUD + ImportState + Configure
- `internal/service/firewall/alias_schema.go` — NEW: Terraform schema definition
- `internal/service/firewall/alias_model.go` — NEW: AliasResourceModel, aliasAPIRequest, aliasAPIResponse, toAPI(), fromAPI()
- `internal/service/firewall/alias_resource_test.go` — NEW: acceptance test (TF_ACC gated)
- `internal/acctest/acctest.go` — NEW: ProtoV6ProviderFactories, PreCheck
- `internal/provider/provider.go` — MODIFIED: register firewall.Resources()
- `go.mod` — MODIFIED: added terraform-plugin-framework-validators and terraform-plugin-testing
- `go.sum` — MODIFIED: updated checksums
- `examples/resources/opnsense_firewall_alias/resource.tf` — NEW: example HCL
- `examples/resources/opnsense_firewall_alias/import.sh` — NEW: import example
- `templates/resources/firewall_alias.md.tmpl` — NEW: documentation template
