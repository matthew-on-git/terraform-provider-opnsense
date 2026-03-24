# Research Report: Terraform Plugin Framework — Deep Technical Analysis

**Date:** 2026-03-13
**Author:** Matthew
**Research Type:** Technical (Track 2 of OPNsense Provider Viability Analysis)
**Confidence Level:** High — sourced from official HashiCorp documentation, verified community implementations, and cross-referenced against multiple independent sources.

---

## Executive Summary

The Terraform Plugin Framework is HashiCorp's current and recommended SDK for building Terraform providers. It replaces SDKv2 (which is in maintenance-only mode) and provides stronger type safety, cleaner Go idioms, and features unavailable in SDKv2. Building terraform-provider_opnsense on the Framework is the correct and only defensible choice for a new provider in 2026.

This report covers everything needed to make informed architecture decisions: Framework vs SDKv2 comparison, canonical project structure, resource lifecycle contracts, schema definition, state management, acceptance testing, documentation generation, Registry publishing, reference providers, and common pitfalls.

---

## 1. Framework vs SDKv2

### Current Status (March 2026)

| Aspect | Plugin Framework | SDKv2 |
|---|---|---|
| **Status** | Actively developed, GA | Maintenance-only (security patches) |
| **Recommendation** | HashiCorp's official recommendation for all new providers | Legacy; migrate when possible |
| **Feature development** | Active — new features like ephemeral resources, actions, list resources, write-only arguments | Stopped; only security patches while Terraform 1.x is current |
| **Protocol versions** | 5 and 6 | 5 only |
| **End-of-life** | N/A | No explicit date, but "Terraform 2 will arrive someday and support will end" |

**Source:** [Plugin Framework Benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits), [SDKv2 Home](https://developer.hashicorp.com/terraform/plugin/sdkv2)

### Key Technical Advantages of the Framework

**1. Type Safety and Concise Abstractions**

- SDKv2 uses `map[string]*schema.Schema` and `helper/schema.ResourceData` — no compile-time safety, attribute names are strings that can be misspelled and only fail at runtime.
- The Framework uses distinct packages per concept (`datasource`, `provider`, `resource`) and interface types (`resource.Resource`, `resource.ResourceWithImportState`) that produce compiler errors for missing required functionality.
- SDKv2's `ResourceData` merges configuration, plan, and state into one ambiguous type. The Framework separates them: `req.Config`, `req.Plan`, `req.State` are distinct, and null vs unknown values are distinguishable (SDKv2 returns both as zero-values like empty strings).

**2. Request-Response Pattern**

- The Framework uses explicit request/response types for every operation. Each CRUD method receives a typed request and writes to a typed response. This enables the framework to add fields over time without breaking method signatures.
- SDKv2 duplicated fields (`Create`, `CreateContext`, `CreateWithoutTimeout`) to maintain backward compatibility, complicating discovery and static analysis.

**3. Context Threading**

- The Framework passes `context.Context` consistently throughout all provider logic, enabling rich automatic logging.
- SDKv2 uses context inconsistently and cannot be updated without breaking changes.

**4. Extensibility**

- Framework provides extensible validation (not limited to fixed options like `ConflictsWith`).
- Custom attribute types, nested attributes without map restrictions, and encapsulation within service packages.
- Plan modification is a first-class concept with composable modifiers.

**5. Features Only Available in the Framework**

- Dynamic attributes (runtime-determined types)
- Provider-defined functions (custom functions in HCL)
- Ephemeral resources (not persisted to plan/state)
- Actions (non-CRUD operations)
- List resources (discovery of unmanaged resources)
- Write-only arguments (Terraform 1.11+)
- Unrestricted type system with custom attribute types and nested attributes

**Source:** [Plugin Framework Benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits)

### Migration Path

For existing SDKv2 providers, `terraform-plugin-mux` enables incremental migration — individual resources can be moved to the Framework while others remain on SDKv2. This is not relevant for terraform-provider_opnsense (greenfield), but is worth noting for architecture awareness.

**Source:** [Migrate from SDKv2](https://developer.hashicorp.com/terraform/plugin/framework/migrating)

### Verdict for terraform-provider_opnsense

The Framework is the only defensible choice. SDKv2 is in maintenance mode, lacks critical features, and will eventually lose support. A new provider started in 2026 on SDKv2 would need to be rewritten.

---

## 2. Go Project Structure

### Canonical Layout

Based on the official [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework) template and verified against well-built community providers:

```
terraform-provider-opnsense/
├── .github/
│   └── workflows/
│       └── release.yml              # GitHub Actions: goreleaser on tag push
├── .goreleaser.yml                   # Cross-platform build + signing config
├── .golangci.yml                     # Go linter configuration
├── docs/                             # Generated documentation (tfplugindocs output)
│   ├── index.md                      # Provider overview
│   ├── resources/                    # Resource docs
│   └── data-sources/                 # Data source docs
├── examples/
│   ├── provider/
│   │   └── provider.tf               # Provider configuration example
│   ├── resources/
│   │   └── opnsense_firewall_alias/
│   │       ├── resource.tf           # Resource example
│   │       └── import.sh             # Import example
│   └── data-sources/
│       └── opnsense_firewall_alias/
│           └── data-source.tf        # Data source example
├── internal/
│   └── provider/                     # Core provider implementation
│       ├── provider.go               # Provider definition (Schema, Configure, Resources, DataSources)
│       ├── provider_test.go          # Provider test setup (shared factories, config)
│       ├── <resource_name>_resource.go      # Resource implementation
│       ├── <resource_name>_resource_test.go # Acceptance tests
│       ├── <datasource_name>_data_source.go
│       └── <datasource_name>_data_source_test.go
├── templates/                        # tfplugindocs templates (*.md.tmpl)
├── tools/
│   └── tools.go                      # Tool dependencies (tfplugindocs)
├── main.go                           # Entry point
├── go.mod                            # Go module definition
├── go.sum
├── GNUmakefile                       # Build targets (generate, testacc, install)
├── terraform-registry-manifest.json  # Registry metadata
├── LICENSE                           # MPL-2.0 (standard for Terraform providers)
└── CHANGELOG.md
```

**Source:** [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework)

### Variation for Large Providers: Separate API Client Package

Well-built community providers (e.g., marshallford/terraform-provider-pfsense) separate API client logic from provider logic:

```
├── internal/
│   └── provider/          # Terraform resource/datasource implementations
├── pkg/
│   └── opnsense/          # Independent API client library
│       ├── client.go      # HTTP client, auth, error handling
│       ├── firewall.go    # Firewall API operations
│       ├── haproxy.go     # HAProxy API operations
│       └── ...
```

This is explicitly recommended by HashiCorp as a best practice: "Terraform should always consume an independent client library which implements the core logic for communicating with the upstream. Do not try to implement this type of logic in the provider itself."

**Source:** [Best Practices](https://developer.hashicorp.com/terraform/plugin/best-practices), [marshallford/terraform-provider-pfsense](https://github.com/marshallford/terraform-provider-pfsense)

### main.go

The entry point is minimal — it starts the provider server:

```go
package main

import (
    "context"
    "flag"
    "log"

    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/matthew-on-git/terraform-provider-opnsense/internal/provider"
)

var version string = "dev"

func main() {
    var debug bool
    flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
    flag.Parse()

    opts := providerserver.ServeOpts{
        Address: "registry.terraform.io/matthew-on-git/opnsense",
        Debug:   debug,
    }

    err := providerserver.Serve(context.Background(), provider.New(version), opts)
    if err != nil {
        log.Fatal(err.Error())
    }
}
```

**Source:** [Implement a Provider](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider)

### Provider Definition

```go
package provider

import (
    "context"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/provider/schema"
    "github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &opnsenseProvider{}

type opnsenseProvider struct {
    version string
}

func New(version string) func() provider.Provider {
    return func() provider.Provider {
        return &opnsenseProvider{version: version}
    }
}

func (p *opnsenseProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "opnsense"
    resp.Version = p.version
}

func (p *opnsenseProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "api_key": schema.StringAttribute{
                Optional:    true,
                Sensitive:   true,
                Description: "OPNsense API key. Can also be set with OPNSENSE_API_KEY env var.",
            },
            "api_secret": schema.StringAttribute{
                Optional:    true,
                Sensitive:   true,
                Description: "OPNsense API secret. Can also be set with OPNSENSE_API_SECRET env var.",
            },
            "uri": schema.StringAttribute{
                Optional:    true,
                Description: "OPNsense base URI (e.g., https://opnsense.example.com). Can also be set with OPNSENSE_URI env var.",
            },
            "insecure": schema.BoolAttribute{
                Optional:    true,
                Description: "Disable TLS certificate verification. Can also be set with OPNSENSE_INSECURE env var.",
            },
        },
    }
}

func (p *opnsenseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // 1. Retrieve config values
    // 2. Fall back to environment variables
    // 3. Validate
    // 4. Create API client
    // 5. Set resp.DataSourceData and resp.ResourceData to the client
}

func (p *opnsenseProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewFirewallAliasResource,
        // ... all resources
    }
}

func (p *opnsenseProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        NewFirewallAliasDataSource,
        // ... all data sources
    }
}
```

**Source:** [Implement a Provider Tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider)

### Required Go Modules

The core modules for a Framework-based provider:

| Module | Purpose |
|---|---|
| `github.com/hashicorp/terraform-plugin-framework` | Core SDK — provider, resource, datasource, schema, types |
| `github.com/hashicorp/terraform-plugin-go` | Low-level protocol bindings (transitive dependency) |
| `github.com/hashicorp/terraform-plugin-log` | Structured logging |
| `github.com/hashicorp/terraform-plugin-testing` | Acceptance test framework |
| `github.com/hashicorp/terraform-plugin-docs` | Documentation generation (tool dependency) |
| `github.com/hashicorp/terraform-plugin-framework-validators` | Pre-built validators (string length, regex, etc.) |

Go version requirement: Go 1.24+ (the scaffolding repo requires >= 1.24 as of March 2026; the framework module itself supports the two latest Go major releases).

**Source:** [terraform-plugin-framework on pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework), [scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework)

---

## 3. Resource Lifecycle

### The CRUD + Import Contract

Every managed resource implements the `resource.Resource` interface with four required CRUD methods plus optional ImportState. The Framework calls these methods via the Terraform Plugin Protocol (gRPC) in response to Terraform CLI operations.

| Method | When Called | Input | Output | Contract |
|---|---|---|---|---|
| **Create** | `terraform apply` when resource is new | Plan data + Config | State | Call API to create. Write full resource state. All planned values must match. |
| **Read** | `terraform plan`, `terraform apply` (refresh), `terraform import` | Prior state | Updated state | Call API to read current state. Write full state. If resource gone, call `resp.State.RemoveResource()`. |
| **Update** | `terraform apply` when resource exists and plan differs from state | Prior state + Plan + Config | Updated state | Call API to update in-place. Write full state. All planned values must match. |
| **Delete** | `terraform apply` when resource is removed from config, or `terraform destroy` | Prior state | (empty) | Call API to delete. Framework automatically removes state if no error. |
| **ImportState** | `terraform import` | Import ID string | Partial state (enough for Read) | Parse import ID, set minimum state for Read to succeed. |

### Method Signatures

```go
// Create
func (r *MyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse)

// Read
func (r *MyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse)

// Update
func (r *MyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse)

// Delete
func (r *MyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse)

// ImportState (optional — implement resource.ResourceWithImportState)
func (r *MyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse)
```

### Critical Contract Rules

1. **State must match plan:** After Create or Update, all attribute values in `resp.State` must match their planned values. If they don't, Terraform raises `"Provider produced inconsistent result"`. This means you must either use plan modifiers to declare what will change, or ensure your API returns values matching what was planned.

2. **Read must handle missing resources:** If the API returns 404 (resource deleted out-of-band), call `resp.State.RemoveResource(ctx)`. This tells Terraform the resource is gone and it should be recreated on next apply.

3. **Delete is silent on success:** If Delete returns no errors, the Framework automatically removes the resource from state. Do not explicitly clear state in Delete.

4. **All state must be populated:** The Framework does not automatically carry forward state values. Every Read/Create/Update must write the complete state.

### Data Flow Pattern

```
Create:
  req.Plan.Get(ctx, &model)   →   API call   →   resp.State.Set(ctx, &model)

Read:
  req.State.Get(ctx, &model)  →   API call   →   resp.State.Set(ctx, &model)

Update:
  req.Plan.Get(ctx, &model)   →   API call   →   resp.State.Set(ctx, &model)

Delete:
  req.State.Get(ctx, &model)  →   API call   →   (no state write needed)

ImportState:
  req.ID                       →   parse      →   resp.State.SetAttribute(ctx, path.Root("id"), value)
```

### Configure Method (Receiving the API Client)

Resources receive the provider's API client via an optional Configure method:

```go
var _ resource.ResourceWithConfigure = &MyResource{}

func (r *MyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    client, ok := req.ProviderData.(*opnsense.Client)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *opnsense.Client, got: %T", req.ProviderData),
        )
        return
    }
    r.client = client
}
```

### Import State

Simple case — passthrough ID:

```go
func (r *MyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

Complex case — composite key (common for OPNsense where UUIDs identify resources):

```go
func (r *MyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    idParts := strings.Split(req.ID, ",")
    if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
        resp.Diagnostics.AddError(
            "Unexpected Import Identifier",
            fmt.Sprintf("Expected format: parent_id,child_id. Got: %q", req.ID),
        )
        return
    }
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("parent_id"), idParts[0])...)
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), idParts[1])...)
}
```

**Source:** [Resources](https://developer.hashicorp.com/terraform/plugin/framework/resources), [Resource Import](https://developer.hashicorp.com/terraform/plugin/framework/resources/import), [Resource Create Tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-resource-create)

---

## 4. Schema Definition

### Attribute Types

| Framework Type | Go Value Type | Terraform Type | Common Use |
|---|---|---|---|
| `schema.StringAttribute` | `types.String` | `string` | Names, IDs, descriptions, enum-like values |
| `schema.Int64Attribute` | `types.Int64` | `number` | Ports, counts, weights |
| `schema.Int32Attribute` | `types.Int32` | `number` | Smaller integers |
| `schema.Float64Attribute` | `types.Float64` | `number` | Decimal values |
| `schema.BoolAttribute` | `types.Bool` | `bool` | Enabled/disabled flags |
| `schema.ListAttribute` | `types.List` | `list(type)` | Ordered collections of single type |
| `schema.SetAttribute` | `types.Set` | `set(type)` | Unordered unique collections |
| `schema.MapAttribute` | `types.Map` | `map(type)` | String-keyed maps of single type |
| `schema.SingleNestedAttribute` | Custom struct | `object({...})` | Structured sub-objects |
| `schema.ListNestedAttribute` | `[]CustomStruct` | `list(object({...}))` | Lists of structured objects |
| `schema.SetNestedAttribute` | `[]CustomStruct` | `set(object({...}))` | Sets of structured objects |
| `schema.MapNestedAttribute` | `map[string]CustomStruct` | `map(object({...}))` | Maps of structured objects |

**Source:** [Attributes](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes), [String Attributes](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string)

### Required / Optional / Computed Combinations

| Combination | Meaning | Use Case |
|---|---|---|
| `Required: true` | Practitioner must provide a value | User-specified fields (name, type, address) |
| `Optional: true` | Practitioner may provide a value or leave null | Optional configuration (description, tags) |
| `Computed: true` | Provider sets the value; practitioner cannot configure | Server-generated fields (id, created_at, uuid) |
| `Optional: true, Computed: true` | Practitioner may provide; provider fills in if null | Fields with server defaults (port defaults to 443) |

**At least one of Required, Optional, or Computed must be true.** Setting Computed alone means any user-provided value is automatically rejected with an error.

### Sensitive Fields

```go
"api_secret": schema.StringAttribute{
    Required:  true,
    Sensitive: true,
    Description: "API secret for authentication.",
}
```

Setting `Sensitive: true` masks the value in all Terraform plan/apply output. Use for credentials, keys, and secrets.

### Validators

Pre-built validators from `terraform-plugin-framework-validators`:

```go
import (
    "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

"name": schema.StringAttribute{
    Required: true,
    Validators: []validator.String{
        stringvalidator.LengthBetween(1, 255),
        stringvalidator.RegexMatches(
            regexp.MustCompile(`^[a-zA-Z0-9_]+$`),
            "must contain only alphanumeric characters and underscores",
        ),
    },
}

"port": schema.Int64Attribute{
    Optional: true,
    Computed: true,
    Validators: []validator.Int64{
        int64validator.Between(1, 65535),
    },
}
```

Custom validators implement the relevant `validator.<Type>` interface (e.g., `validator.String` requires `ValidateString(ctx, req, resp)`). Validators must handle null and unknown values (typically by returning early).

**Source:** [Validation](https://developer.hashicorp.com/terraform/plugin/framework/validation)

### Specialized Type Modules

| Module | Types |
|---|---|
| `terraform-plugin-framework-jsontypes` | JSON-encoded string validation |
| `terraform-plugin-framework-nettypes` | IPv4/IPv6 addresses, CIDR blocks |
| `terraform-plugin-framework-timetypes` | RFC3339 timestamps |

### Plan Modifiers

```go
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(), // ID doesn't change after creation
    },
},
"name": schema.StringAttribute{
    Required: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(), // Changing name forces recreation
    },
},
```

### Default Values

```go
"enabled": schema.BoolAttribute{
    Optional: true,
    Computed: true,
    Default:  booldefault.StaticValue(true),
    Description: "Whether the resource is enabled. Defaults to true.",
},
"port": schema.Int64Attribute{
    Optional: true,
    Computed: true,
    Default:  int64default.StaticValue(443),
},
```

### Nested Attributes Example

```go
"items": schema.ListNestedAttribute{
    Required: true,
    NestedObject: schema.NestedAttributeObject{
        Attributes: map[string]schema.Attribute{
            "name": schema.StringAttribute{Required: true},
            "weight": schema.Int64Attribute{
                Optional: true,
                Computed: true,
                Default:  int64default.StaticValue(1),
            },
        },
    },
},
```

### Data Model Structs

Schema maps to Go structs via `tfsdk` struct tags:

```go
type firewallAliasModel struct {
    ID          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    Type        types.String `tfsdk:"type"`
    Description types.String `tfsdk:"description"`
    Enabled     types.Bool   `tfsdk:"enabled"`
    Content     types.List   `tfsdk:"content"`  // types.List of types.String
}
```

**Source:** [String Attributes](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string), [Schemas](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas)

---

## 5. State Management

### Plan Modification

Plan modifiers run after validation but before apply. Execution order:

1. Null config attributes set to default values
2. Computed attributes (null in config) marked as unknown if different from state
3. Attribute-level plan modifiers execute
4. Resource-level plan modifiers execute

**Built-in plan modifiers** (per type, e.g., `stringplanmodifier`, `int64planmodifier`, `boolplanmodifier`):

| Modifier | Purpose |
|---|---|
| `UseStateForUnknown()` | Copies prior state value to plan, reducing "(known after apply)" noise. Use for computed values that don't change (IDs, creation timestamps). |
| `RequiresReplace()` | Forces resource destruction and recreation when value changes. Use for immutable fields. |
| `RequiresReplaceIf(func)` | Conditional replacement based on provider logic. |
| `RequiresReplaceIfConfigured()` | Only triggers replacement if practitioner configured a non-null value. |

**Custom plan modifiers** implement `planmodifier.<Type>` interfaces. Common use case: marking a resource for replacement when an upstream API doesn't support in-place updates.

**Resource-level plan modifiers** implement `resource.ResourceWithModifyPlan` for cross-attribute logic.

**Source:** [Plan Modification](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)

### Default Values

Defaults are set via the `Default` field on attributes. They execute before plan modifiers. Available default implementations:

- `stringdefault.StaticString("value")`
- `int64default.StaticValue(42)`
- `booldefault.StaticValue(true)`
- `float64default.StaticValue(3.14)`

Custom defaults implement the `<type>default.Describable<Type>` interface.

### State Upgrades

When schema changes are incompatible (e.g., changing an attribute type), state upgrades migrate stored state:

1. Set `Version` field in `schema.Schema` (increment by 1 per upgrade).
2. Implement `resource.ResourceWithUpgradeState` interface.
3. Return a map from prior version numbers to `resource.StateUpgrader` structs.

```go
func (r *MyResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
    return map[int64]resource.StateUpgrader{
        0: {
            PriorSchema: &schema.Schema{
                Attributes: map[string]schema.Attribute{
                    "id": schema.StringAttribute{Computed: true},
                    // ... prior schema v0 definition
                },
            },
            StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
                var priorState myResourceModelV0
                resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
                // Transform to current schema
                upgradedState := myResourceModel{
                    ID: priorState.ID,
                    // ... field transformations
                }
                resp.Diagnostics.Append(resp.State.Set(ctx, &upgradedState)...)
            },
        },
    }
}
```

**Critical rules:**
- Each upgrader must wholly upgrade from the prior version to the current version (no intermediate chaining).
- The framework does NOT copy any prior state data automatically — `resp.State` must be fully populated.
- Unknown values in the response state cause an error.
- Always increment `SchemaVersion` when making incompatible changes. Forgetting this causes Terraform to read old state with new schema, producing corrupt state.

**Source:** [State Upgrade](https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade)

---

## 6. Acceptance Testing

### Overview

Acceptance tests execute real Terraform commands (`terraform apply`, `terraform destroy`) against actual API endpoints. They validate the full lifecycle of provider resources.

### Test Setup

Provider test configuration (shared across all tests):

```go
// internal/provider/provider_test.go
package provider

import (
    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
    "opnsense": providerserver.NewProtocol6WithError(New("test")()),
}

const providerConfig = `
provider "opnsense" {
  uri        = "https://opnsense.test.local"
  api_key    = "test_key"
  api_secret = "test_secret"
  insecure   = true
}
`
```

### Test Structure

```go
// internal/provider/firewall_alias_resource_test.go
package provider

import (
    "testing"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallAliasResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create and Read
            {
                Config: providerConfig + `
resource "opnsense_firewall_alias" "test" {
  name    = "test_alias"
  type    = "host"
  content = ["192.168.1.1", "192.168.1.2"]
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttrSet("opnsense_firewall_alias.test", "id"),
                    resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "name", "test_alias"),
                    resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "type", "host"),
                    resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "content.#", "2"),
                ),
            },
            // Step 2: ImportState
            {
                ResourceName:            "opnsense_firewall_alias.test",
                ImportState:             true,
                ImportStateVerify:       true,
                ImportStateVerifyIgnore: []string{"last_updated"},
            },
            // Step 3: Update and Read
            {
                Config: providerConfig + `
resource "opnsense_firewall_alias" "test" {
  name    = "test_alias"
  type    = "host"
  content = ["192.168.1.1", "192.168.1.2", "192.168.1.3"]
}
`,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("opnsense_firewall_alias.test", "content.#", "3"),
                ),
            },
            // Step 4: Delete is automatic — framework destroys after last step
        },
    })
}
```

### Common Check Functions

| Function | Purpose |
|---|---|
| `resource.TestCheckResourceAttr(addr, key, value)` | Verify exact attribute value |
| `resource.TestCheckResourceAttrSet(addr, key)` | Verify attribute exists with any value |
| `resource.TestCheckResourceAttrPair(addr1, key1, addr2, key2)` | Verify two attributes match |
| `resource.TestCheckNoResourceAttr(addr, key)` | Verify attribute is not set |
| `resource.ComposeAggregateTestCheckFunc(checks...)` | Combine multiple checks (runs all, reports all failures) |
| `resource.TestCheckOutput(name, value)` | Verify Terraform output value |

### Running Tests

```bash
# Run all acceptance tests
TF_ACC=1 go test ./internal/provider/ -v -timeout 120m

# Run specific test
TF_ACC=1 go test ./internal/provider/ -run='TestAccFirewallAliasResource' -v -timeout 120m

# With count=1 to bypass test caching
TF_ACC=1 go test -count=1 ./internal/provider/ -v
```

**The `TF_ACC=1` environment variable is mandatory.** Without it, tests are silently skipped.

### Important: The `id` Attribute

**The Framework does NOT implicitly add an `id` attribute** (SDKv2 did this automatically). You must explicitly define it in your schema:

```go
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

Omitting this causes test failures with: `"no id found in attributes"`.

### Schema Unit Tests

Validate schema definitions without hitting APIs:

```go
func TestFirewallAliasResourceSchema(t *testing.T) {
    t.Parallel()
    ctx := context.Background()
    schemaResponse := &fwresource.SchemaResponse{}
    NewFirewallAliasResource().Schema(ctx, fwresource.SchemaRequest{}, schemaResponse)
    if schemaResponse.Diagnostics.HasError() {
        t.Fatalf("Schema method diagnostics: %+v", schemaResponse.Diagnostics)
    }
    diagnostics := schemaResponse.Schema.ValidateImplementation(ctx)
    if diagnostics.HasError() {
        t.Fatalf("Schema validation diagnostics: %+v", diagnostics)
    }
}
```

**Source:** [Acceptance Testing Tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-acceptance-testing), [Acceptance Testing Reference](https://developer.hashicorp.com/terraform/plugin/framework/acctests)

---

## 7. Documentation Generation

### tfplugindocs

[terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) (CLI: `tfplugindocs`) generates Terraform Registry-compatible documentation from provider schemas and templates.

### How It Works

1. Builds the provider binary via `go build`
2. Extracts schema using `terraform providers schema -json`
3. Processes templates from `templates/` directory
4. Copies examples from `examples/` directory
5. Outputs rendered markdown to `docs/`

### Directory Structure

**Input — Templates** (`templates/`):
```
templates/
├── index.md.tmpl                          # Provider overview
├── resources/
│   └── firewall_alias.md.tmpl             # Resource doc template
└── data-sources/
    └── firewall_alias.md.tmpl             # Data source doc template
```

**Input — Examples** (`examples/`):
```
examples/
├── provider/
│   └── provider.tf                         # Provider config example
├── resources/
│   └── opnsense_firewall_alias/
│       ├── resource.tf                     # Resource example (HCL)
│       └── import.sh                       # Import command example
└── data-sources/
    └── opnsense_firewall_alias/
        └── data-source.tf                  # Data source example
```

**Output** (`docs/`):
```
docs/
├── index.md                               # Provider overview
├── resources/
│   └── firewall_alias.md                  # Resource documentation
└── data-sources/
    └── firewall_alias.md                  # Data source documentation
```

### Template Variables

Templates use Go `text/template` syntax with these variables:

- `{{ .SchemaMarkdown }}` — Auto-generated schema table (arguments, attributes, types)
- `{{ .Description }}` — From schema Description field
- `{{ .ProviderShortName }}` — e.g., "opnsense"
- `{{ .Name }}` — Resource/datasource name
- `{{ .HasExamples }}` / `{{ .ExampleFiles }}` — Example file references
- `{{ .HasImport }}` / `{{ .ImportFile }}` — Import command reference

### Running

```bash
# Via Makefile
make generate

# Directly
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate
```

### Validation

```bash
tfplugindocs validate
```

Checks: proper directory structure, file extensions, frontmatter, file sizes, and that docs match schema definitions.

### Installation via tools/tools.go

```go
//go:build tools

package tools

import (
    _ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
```

**Source:** [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs)

---

## 8. Registry Publishing

### Requirements

1. **Repository naming:** Must be `terraform-provider-{NAME}` (lowercase only, public repository on GitHub)
2. **Semantic versioning:** Tags must be `v{MAJOR}.{MINOR}.{PATCH}` (e.g., `v0.1.0`)
3. **No branch matching tag name:** A branch named `v0.1.0` would conflict with the tag
4. **Documentation:** `docs/` directory with index.md and per-resource/datasource docs
5. **Registry manifest file:** `terraform-registry-manifest.json` in the repo root

### terraform-registry-manifest.json

```json
{
  "version": 1,
  "metadata": {
    "protocol_versions": ["6.0"]
  }
}
```

The `version` field is the manifest format version (always `1`), not the provider version. `protocol_versions` should be `["6.0"]` for Framework providers.

### GPG Key Setup

**Generate an RSA key** (the Registry does NOT support ECC, the default):

```bash
gpg --full-generate-key
# Select: (1) RSA and RSA
# Key size: 4096
# Expiration: 0 (does not expire)
```

**Export public key** (add to Registry at registry.terraform.io/settings/gpg-keys):

```bash
gpg --armor --export "YOUR_KEY_ID"
```

**Export private key** (add to GitHub Secrets as `GPG_PRIVATE_KEY`):

```bash
gpg --armor --export-secret-keys "YOUR_KEY_ID"
```

### GitHub Repository Secrets

| Secret | Value |
|---|---|
| `GPG_PRIVATE_KEY` | ASCII-armored private GPG key (including `-----BEGIN...` and `-----END...` lines) |
| `PASSPHRASE` | GPG key passphrase |

### .goreleaser.yml

Copy from the [scaffolding-framework repository](https://github.com/hashicorp/terraform-provider-scaffolding-framework). This configuration:
- Builds for multiple platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
- Creates SHA256 checksums
- Signs the checksums file with GPG
- Creates a GitHub Release with all artifacts

### GitHub Actions Workflow

**Option A: HashiCorp's reusable workflow** (recommended):

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags:
      - 'v*'

jobs:
  terraform-provider-release:
    name: 'Terraform Provider Release'
    uses: hashicorp/ghaction-terraform-provider-release/.github/workflows/community.yml@v4
    secrets:
      gpg-private-key: '${{ secrets.GPG_PRIVATE_KEY }}'
    with:
      setup-go-version-file: 'go.mod'
```

**Option B: Custom workflow with goreleaser** (more control):

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'
      - name: Import GPG key
        uses: crazy-max/ghaction-import-gpg@v6
        id: import_gpg
        with:
          gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.PASSPHRASE }}
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPG_FINGERPRINT: ${{ steps.import_gpg.outputs.fingerprint }}
```

### Release Process

```bash
git tag v0.1.0
git push origin v0.1.0
```

The GitHub Action triggers automatically, builds all platform binaries, signs them, and creates a GitHub Release.

### Publishing to Registry

1. Sign in to registry.terraform.io with your GitHub account
2. Navigate to **Publish > Provider**
3. Select the repository
4. Add GPG public key at **Settings > Signing Keys**
5. Publish — creates a webhook on the repo for future `release` events

**Warning:** Publishing is permanent. You cannot un-publish a provider.

**Source:** [Publishing Providers](https://developer.hashicorp.com/terraform/registry/providers/publishing), [Release and Publish Tutorial](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-release-publish), [ghaction-terraform-provider-release](https://github.com/hashicorp/ghaction-terraform-provider-release)

---

## 9. Reference Providers

### Providers to Study

#### 1. browningluke/terraform-provider-opnsense
- **URL:** [GitHub](https://github.com/browningluke/terraform-provider-opnsense) | [Registry](https://registry.terraform.io/providers/browningluke/opnsense/latest/docs)
- **Framework:** Terraform Plugin Framework
- **Relevance:** Direct competitor / predecessor for the same target API (OPNsense)
- **Architecture:** Resources organized by OPNsense module hierarchy (Firewall/Alias, Interfaces/Vlan). Separate API client. GoReleaser for releases.
- **Coverage:** Core APIs only — firewall aliases/rules, NAT, interfaces, IPsec. No plugin APIs (no HAProxy, FRR, ACME, etc.).
- **Status:** Pre-v1.0, schemas subject to change. Comprehensive acceptance tests are the gate for v1.0.
- **Key lesson:** Clean project structure following the scaffolding template. Good example of two-tier API coverage strategy. Shows how to handle OPNsense API auth and the reconfigure model.

#### 2. RyanNgWH/terraform-provider-opnsense
- **URL:** [GitHub](https://github.com/RyanNgWH/terraform-provider-opnsense)
- **Framework:** Terraform Plugin Framework
- **Relevance:** Another OPNsense provider, newer
- **Coverage:** Networking (interface groups, gateways), firewall (aliases, rules, categories, NAT), services (captive portal, traffic shaping)
- **Go version:** 1.25+ required
- **Targets:** OPNsense CE 25.7.8+, Terraform 1.8+, OpenTofu 1.10+
- **Key lesson:** Shows granular ACL-based API privilege requirements in documentation. Active development.

#### 3. marshallford/terraform-provider-pfsense
- **URL:** [GitHub](https://github.com/marshallford/terraform-provider-pfsense)
- **Framework:** Terraform Plugin Framework
- **Relevance:** Same problem domain (network appliance firewall management), different target (pfSense)
- **Architecture:** **Exemplary client library separation** — `pkg/pfsense/` contains the API client, `internal/provider/` contains Terraform resource implementations. This is the HashiCorp-recommended pattern.
- **Build system:** GNUmakefile, GoReleaser, golangci-lint, ShellCheck, yamllint
- **Status:** Pre-v1.0, 24 releases, active development
- **Key lesson:** The `pkg/pfsense/` client library pattern is the best reference for how to separate API client concerns from provider concerns. Directly applicable to terraform-provider_opnsense.

#### 4. Mastercard/terraform-provider-restapi
- **URL:** [GitHub](https://github.com/Mastercard/terraform-provider-restapi)
- **Relevance:** Generic REST API provider — shows patterns for wrapping arbitrary REST APIs
- **Key lesson:** Demonstrates generic CRUD mapping to REST endpoints. Useful for understanding the mapping between HTTP methods and Terraform lifecycle operations.

#### 5. hashicorp/terraform-provider-hashicups (Tutorial Provider)
- **URL:** Referenced throughout HashiCorp tutorials
- **Relevance:** The official learning provider with complete documentation at every step
- **Key lesson:** The gold standard for understanding the Framework patterns. Follow the full tutorial series before starting implementation.

**Source:** [browningluke/opnsense](https://github.com/browningluke/terraform-provider-opnsense), [RyanNgWH/opnsense](https://github.com/RyanNgWH/terraform-provider-opnsense), [marshallford/pfsense](https://github.com/marshallford/terraform-provider-pfsense), [Mastercard/restapi](https://github.com/Mastercard/terraform-provider-restapi)

---

## 10. Common Pitfalls

### Pitfall 1: Conflating API Client with Provider Code

**The mistake:** Writing HTTP/REST logic directly inside resource Create/Read/Update/Delete methods.

**The fix:** Build an independent API client library (`pkg/opnsense/` or `internal/opnsense/`) that knows nothing about Terraform. The provider code consumes this client. This enables testing the API client independently, reuse outside Terraform, and cleaner resource implementations.

**Source:** [Best Practices](https://developer.hashicorp.com/terraform/plugin/best-practices)

### Pitfall 2: Forgetting the Explicit `id` Attribute

**The mistake:** Assuming the Framework auto-creates an `id` attribute like SDKv2 did.

**The fix:** Always explicitly define `id` in your schema with `Computed: true` and `UseStateForUnknown()` plan modifier. Omitting it causes test failures and import failures.

**Source:** [Acceptance Testing](https://developer.hashicorp.com/terraform/plugin/framework/acctests)

### Pitfall 3: Not Handling Null and Unknown Values

**The mistake:** Treating null (not set) and unknown (computed later) identically, or assuming values are always present.

**The fix:** The Framework distinguishes three states: null (`IsNull()`), unknown (`IsUnknown()`), and known (has a value). Validators and CRUD methods must check for null/unknown before accessing values. Ignoring this causes panics or incorrect behavior.

**Source:** [Plugin Framework Benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits)

### Pitfall 4: State/Plan Inconsistency

**The mistake:** Setting state values that don't match planned values after Create or Update.

**The fix:** After an API call, always populate state from the API response (not from the plan input). If the API transforms or normalizes values (e.g., lowercasing a name), use plan modifiers to declare this. The error message `"Provider produced inconsistent result"` means your state doesn't match the plan.

**Source:** [Plan Modification](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)

### Pitfall 5: Not Handling Resource Deletion Outside Terraform

**The mistake:** Read method fails hard when the API returns 404 for a deleted resource.

**The fix:** In Read, check for 404 and call `resp.State.RemoveResource(ctx)` to tell Terraform the resource was deleted externally. Terraform will then show it needs to be recreated on next plan.

**Source:** [Custom Terraform Providers Blog](https://superorbital.io/blog/custom-terraform-providers/)

### Pitfall 6: Forgetting SchemaVersion on Breaking Changes

**The mistake:** Changing attribute types or structures without incrementing `schema.Schema.Version`.

**The fix:** Always increment SchemaVersion and provide a StateUpgrader. Without this, Terraform tries to read old state with the new schema, producing silent corruption.

**Source:** [State Upgrade](https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade)

### Pitfall 7: Echoing Config Instead of API Response

**The mistake:** In Create/Update, saving the user's input to state instead of the API's response.

**The fix:** Always call the API, then populate state from the response. This catches server-side defaults, normalizations, and computed fields. The API response is the source of truth, not the user's configuration.

**Source:** [Custom Terraform Providers Blog](https://superorbital.io/blog/custom-terraform-providers/)

### Pitfall 8: Missing nil Check in Configure

**The mistake:** Type-asserting `req.ProviderData` without checking for nil first, causing panics during early provider initialization.

**The fix:** Always guard with `if req.ProviderData == nil { return }` before the type assertion.

### Pitfall 9: Not Using UseStateForUnknown for Immutable Computed Fields

**The mistake:** Computed fields like `id` or `created_at` show as "(known after apply)" on every plan, even when nothing changed.

**The fix:** Add `stringplanmodifier.UseStateForUnknown()` (or equivalent) to computed fields that don't change after creation. This copies the prior state value to the plan, eliminating noise.

### Pitfall 10: List/Set Element Ordering

**The mistake:** Assuming list elements maintain their order between API calls, or using lists when order doesn't matter.

**The fix:** Use `SetAttribute`/`SetNestedAttribute` when element order is not significant. For lists, plan modifiers that reference prior state must account for element reordering. This is especially relevant for OPNsense resources where the API may return items in different orders than they were created.

---

## Appendix A: OPNsense-Specific Architecture Implications

Based on the reference providers and the OPNsense API characteristics identified in Track 1 research, key architecture decisions for terraform-provider_opnsense:

1. **Client library separation:** Follow marshallford/pfsense pattern with `pkg/opnsense/` (or `internal/opnsense/`) containing the HTTP client, auth handling, and per-module API methods. The provider code in `internal/provider/` should only deal with Terraform types and schema.

2. **Reconfigure pattern:** OPNsense requires a service `reconfigure` API call after most mutations. The API client should handle this transparently (e.g., a `CreateAndReconfigure` method or a middleware pattern).

3. **UUID-based IDs:** OPNsense resources are identified by UUIDs. Import will typically use `resource.ImportStatePassthroughID` with the UUID.

4. **Plugin API consistency:** The OPNsense plugin APIs (HAProxy, FRR, ACME, etc.) follow a consistent pattern — this means a shared client architecture can handle the common CRUD+reconfigure cycle while individual resource types handle schema differences.

5. **Naming convention:** Resources should follow `opnsense_{service}_{resource}` pattern (e.g., `opnsense_haproxy_server`, `opnsense_frr_bgp_neighbor`, `opnsense_firewall_alias`), matching OPNsense's module hierarchy.

---

## Appendix B: Key Go Module Versions (March 2026)

| Module | Minimum Version | Notes |
|---|---|---|
| Go | 1.24 | Scaffolding template requirement |
| `terraform-plugin-framework` | Latest GA | Core SDK |
| `terraform-plugin-testing` | Latest GA | Acceptance test framework |
| `terraform-plugin-go` | Latest GA | Protocol bindings (transitive) |
| `terraform-plugin-log` | Latest GA | Structured logging |
| `terraform-plugin-docs` | Latest | Documentation generator (tool dep) |
| `terraform-plugin-framework-validators` | Latest | Pre-built validators |
| Terraform CLI (for testing) | >= 1.8 | Required for acceptance tests |

---

## Sources

### Official HashiCorp Documentation
- [Terraform Plugin Framework Overview](https://developer.hashicorp.com/terraform/plugin/framework)
- [Plugin Framework Benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits)
- [SDKv2 Home](https://developer.hashicorp.com/terraform/plugin/sdkv2)
- [Best Practices](https://developer.hashicorp.com/terraform/plugin/best-practices)
- [Resources](https://developer.hashicorp.com/terraform/plugin/framework/resources)
- [Resource Import](https://developer.hashicorp.com/terraform/plugin/framework/resources/import)
- [Plan Modification](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)
- [State Upgrade](https://developer.hashicorp.com/terraform/plugin/framework/resources/state-upgrade)
- [Schemas](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas)
- [Attributes](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes)
- [String Attributes](https://developer.hashicorp.com/terraform/plugin/framework/handling-data/attributes/string)
- [Validation](https://developer.hashicorp.com/terraform/plugin/framework/validation)
- [Acceptance Testing](https://developer.hashicorp.com/terraform/plugin/framework/acctests)
- [Publishing Providers](https://developer.hashicorp.com/terraform/registry/providers/publishing)
- [Migrate from SDKv2](https://developer.hashicorp.com/terraform/plugin/framework/migrating)

### Official HashiCorp Tutorials
- [Implement a Provider](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider)
- [Resource Create](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-resource-create)
- [Acceptance Testing](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-acceptance-testing)
- [Release and Publish](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-release-publish)

### Official HashiCorp Repositories
- [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework)
- [terraform-plugin-framework](https://github.com/hashicorp/terraform-plugin-framework)
- [terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs)
- [ghaction-terraform-provider-release](https://github.com/hashicorp/ghaction-terraform-provider-release)

### Reference Community Providers
- [browningluke/terraform-provider-opnsense](https://github.com/browningluke/terraform-provider-opnsense)
- [RyanNgWH/terraform-provider-opnsense](https://github.com/RyanNgWH/terraform-provider-opnsense)
- [marshallford/terraform-provider-pfsense](https://github.com/marshallford/terraform-provider-pfsense)
- [Mastercard/terraform-provider-restapi](https://github.com/Mastercard/terraform-provider-restapi)

### Community Articles
- [Building Custom Providers with the Plugin Framework — SuperOrbital](https://superorbital.io/blog/custom-terraform-providers/)
- [How to Use the Terraform Plugin Framework — OneUptime](https://oneuptime.com/blog/post/2026-02-23-terraform-plugin-framework/view)
- [terraform-plugin-framework on pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework)
