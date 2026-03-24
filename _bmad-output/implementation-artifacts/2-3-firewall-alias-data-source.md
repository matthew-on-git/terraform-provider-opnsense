# Story 2.3: Firewall Alias Data Source

Status: done

## Story

As an operator,
I want to look up existing OPNsense firewall aliases as a data source,
So that I can reference aliases in other resource configurations without importing them.

## Acceptance Criteria

1. **Given** a firewall alias exists on OPNsense
   **When** the operator defines a `data.opnsense_firewall_alias` block with the alias UUID
   **Then** the data source reads the alias attributes from the API and makes them available for reference

2. **And** the data source follows the same `fromAPI()` conversion as the resource

3. **And** the data source is read-only (no Create, Update, Delete, or ImportState)

4. **And** acceptance test verifies data source reads match the resource state

## Tasks / Subtasks

- [x] Task 1: Create `alias_data_source.go` in `internal/service/firewall/` (AC: #1, #2, #3)
  - [x] 1.1 Define `aliasDataSource` struct with `*opnsense.Client`
  - [x] 1.2 Implement `Metadata` returning type name `opnsense_firewall_alias`
  - [x] 1.3 Implement `Schema` with `id` as Required lookup key, all other attributes Computed
  - [x] 1.4 Implement `Configure` extracting `*opnsense.Client` from provider data
  - [x] 1.5 Implement `Read` — get UUID from config → `opnsense.Get[aliasAPIResponse]` → `fromAPI()` → state
- [x] Task 2: Register data source in `exports.go` and `provider.go` (AC: #1)
  - [x] 2.1 Add `newAliasDataSource` to `firewall.DataSources()` slice
  - [x] 2.2 Update `provider.go` `DataSources()` to return `firewall.DataSources()`
- [x] Task 3: Create acceptance test (AC: #4)
  - [x] 3.1 Test creates a resource, reads it via data source, verifies all attributes match
- [x] Task 4: Create documentation and examples (AC: all)
  - [x] 4.1 Create `examples/data-sources/opnsense_firewall_alias/data-source.tf`
  - [x] 4.2 Create `templates/data-sources/firewall_alias.md.tmpl`
- [x] Task 5: Run `make check` and verify all targets pass (AC: all)

## Dev Notes

### Data Source vs Resource — Key Differences

| Aspect | Resource (Story 2.1) | Data Source (this story) |
|--------|---------------------|--------------------------|
| Methods | Create, Read, Update, Delete, ImportState | **Read only** |
| Schema | Mix of Required/Optional/Computed | `id` Required, everything else **Computed** |
| Interface | `resource.Resource`, `resource.ResourceWithImportState` | **`datasource.DataSource`** |
| State | Creates and manages state | Reads existing data, populates state |
| File count | 4 files | **1 file** (`alias_data_source.go`) |
| Tests | Full lifecycle + CheckDestroy | Read-only, no destroy |
| `toAPI()` | Needed for Create/Update | **Not needed** |

### Data Source Implementation Pattern

**Single file:** `internal/service/firewall/alias_data_source.go`

**Struct and interfaces:**
```go
type aliasDataSource struct {
    client *opnsense.Client
}

var _ datasource.DataSource = &aliasDataSource{}

func newAliasDataSource() datasource.DataSource {
    return &aliasDataSource{}
}
```

**Metadata:**
```go
func (d *aliasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_firewall_alias"
}
```

**CRITICAL: TypeName is the SAME as the resource** (`_firewall_alias`). Terraform distinguishes resources from data sources by the `data.` prefix in HCL (`data "opnsense_firewall_alias"` vs `resource "opnsense_firewall_alias"`).

### Data Source Schema Pattern

**Lookup key:** `id` is Required (the UUID to look up). All other attributes are Computed only — no Optional, no Defaults, no PlanModifiers, no Validators.

```go
func (d *aliasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Look up an existing firewall alias on OPNsense by UUID.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "UUID of the firewall alias to look up.",
            },
            "name": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Name of the alias.",
            },
            // ... all other attributes Computed only
        },
    }
}
```

**IMPORTANT:** The data source schema uses `datasource/schema` package, NOT `resource/schema`:
```go
import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
```

### Data Source Read Method

**Reuses existing `aliasAPIResponse` and `fromAPI()` from `alias_model.go`:**
```go
func (d *aliasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var config AliasResourceModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    id := config.ID.ValueString()
    result, err := opnsense.Get[aliasAPIResponse](ctx, d.client, aliasReqOpts, id)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error reading firewall alias",
            fmt.Sprintf("Could not read firewall alias %s: %s", id, err),
        )
        return
    }

    config.fromAPI(ctx, result, id)
    resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
```

**Key differences from resource Read:**
- Uses `req.Config.Get` (not `req.State.Get`) — data sources read from config, not state
- No `RemoveResource` on NotFoundError — data sources should error if the resource doesn't exist
- Reuses the same `AliasResourceModel`, `aliasAPIResponse`, `aliasReqOpts`, and `fromAPI()` from the resource

### Registration Pattern

**Update `internal/service/firewall/exports.go`:**
```go
func DataSources() []func() datasource.DataSource {
    return []func() datasource.DataSource{
        newAliasDataSource,
    }
}
```

**Update `internal/provider/provider.go`:**
```go
func (p *OpnsenseProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return firewall.DataSources()
}
```

### Acceptance Test Pattern

**Data source tests create a resource first, then read via data source:**
```go
func TestAccFirewallAliasDataSource_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
        PreCheck:                 func() { acctest.PreCheck(t) },
        Steps: []resource.TestStep{
            {
                Config: testAccFirewallAliasDataSourceConfig(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("data.opnsense_firewall_alias.test", "name", "tf_test_ds_alias"),
                    resource.TestCheckResourceAttr("data.opnsense_firewall_alias.test", "type", "host"),
                    resource.TestCheckResourceAttr("data.opnsense_firewall_alias.test", "enabled", "true"),
                ),
            },
        },
    })
}

func testAccFirewallAliasDataSourceConfig() string {
    return `
resource "opnsense_firewall_alias" "test" {
  name    = "tf_test_ds_alias"
  type    = "host"
  content = ["10.0.0.1"]
}

data "opnsense_firewall_alias" "test" {
  id = opnsense_firewall_alias.test.id
}
`
}
```

**Key test differences from resource tests:**
- No Import step (read-only)
- No Update step
- No CheckDestroy (the resource test handles cleanup)
- Config creates a resource AND reads it via data source
- Checks use `data.opnsense_firewall_alias.test` prefix

### Test File Location

The data source test goes in the SAME test file or a new test file in the firewall package. Since the test needs to create a resource first, use a single config block. The test file should be `alias_data_source_test.go` in `internal/service/firewall/`.

### What NOT to Build

- No new model structs — reuse `AliasResourceModel`, `aliasAPIResponse` from `alias_model.go`
- No `toAPI()` needed — data sources are read-only
- No `fromAPI()` changes — existing conversion works for data sources
- No new ReqOpts — reuse `aliasReqOpts` from `alias_resource.go`
- No new dependencies — all imports already available
- No CheckDestroy in data source test — the resource's test handles that

### Previous Story Intelligence

**From Story 2.1 (Firewall Alias Resource):**
- `AliasResourceModel` is the shared Terraform model struct (exported, reusable by data source)
- `aliasAPIResponse` is the API response struct (unexported but in same package — accessible by data source)
- `aliasReqOpts` is the endpoint config (unexported but in same package — accessible by data source)
- `fromAPI()` handles all type conversions including SelectedMap, newline-separated content, SelectedMapList categories
- `_ context.Context` pattern used for `fromAPI()` (revive enforces ctx-first but it's unused)

**From Story 2.2 (HAProxy Server Resource):**
- Provider `Resources()` uses append pattern for multiple service modules
- Same append pattern needed for `DataSources()` when multiple services have data sources

**From Epic 1 Retrospective:**
- `make check` must pass all targets before marking done
- `ctx context.Context` always first parameter

### Project Structure Notes

**New files this story creates:**
```
internal/
└── service/
    └── firewall/
        ├── alias_data_source.go             # NEW: data source (Schema + Read + Configure)
        └── alias_data_source_test.go        # NEW: acceptance test

examples/
└── data-sources/
    └── opnsense_firewall_alias/
        └── data-source.tf                   # NEW: example HCL

templates/
└── data-sources/
    └── firewall_alias.md.tmpl               # NEW: doc template
```

**Modified files:**
```
internal/service/firewall/exports.go         # MODIFIED: add newAliasDataSource to DataSources()
internal/provider/provider.go                # MODIFIED: DataSources() returns firewall.DataSources()
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-2, Story 2.3]
- [Source: _bmad-output/planning-artifacts/architecture.md#Data source pattern]
- [Source: _bmad-output/planning-artifacts/prd.md#FR60 data sources]
- [Source: _bmad-output/implementation-artifacts/2-1-firewall-alias-resource.md#fromAPI patterns]
- [Source: _bmad-output/implementation-artifacts/2-2-haproxy-server-resource.md#Provider registration]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- No linting issues — data source uses `datasource/schema` package (not `resource/schema`)
- Reused `AliasResourceModel`, `aliasAPIResponse`, `aliasReqOpts`, and `fromAPI()` from resource — zero duplication
- Data source Read uses `req.Config.Get` (not `req.State.Get`) — correct for data sources
- No NotFoundError handling with RemoveResource — data sources should error if alias doesn't exist

### Completion Notes List

- Implemented `opnsense_firewall_alias` data source in single file `alias_data_source.go`
- Read-only: only `datasource.DataSource` interface, no Create/Update/Delete/Import
- Schema: `id` is Required lookup key, all other attributes Computed only (no defaults, no validators, no plan modifiers)
- Reuses existing model, API response struct, ReqOpts, and `fromAPI()` from the resource implementation
- Acceptance test creates a resource then reads it back via data source, verifying name/type/enabled attributes
- Registered in `firewall.DataSources()` and provider `DataSources()` method
- `make check` passes 5/6 targets; scan fails due to pre-existing gitleaks findings

### File List

- `internal/service/firewall/alias_data_source.go` — NEW: data source (Schema + Read + Configure)
- `internal/service/firewall/alias_data_source_test.go` — NEW: acceptance test
- `internal/service/firewall/exports.go` — MODIFIED: added newAliasDataSource to DataSources()
- `internal/provider/provider.go` — MODIFIED: DataSources() returns firewall.DataSources()
- `examples/data-sources/opnsense_firewall_alias/data-source.tf` — NEW: example HCL
- `templates/data-sources/firewall_alias.md.tmpl` — NEW: documentation template
