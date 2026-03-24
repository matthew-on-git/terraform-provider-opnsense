---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
status: complete
completedAt: '2026-03-18'
inputDocuments:
  - "prd.md"
  - "product-brief-terraform-provider_opnsense-2026-03-13.md"
  - "research/technical-opnsense-api-terraform-provider-framework-research-2026-03-13.md"
  - "research/technical-terraform-plugin-framework-research-2026-03-13.md"
workflowType: 'architecture'
project_name: 'terraform-provider_opnsense'
user_name: 'Matthew'
date: '2026-03-18'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**
68 FRs across 12 capability areas. The architecture must support:
- Provider configuration with credential validation and OPNsense version detection (FR1-5)
- 13 cross-cutting lifecycle behaviors that apply to ALL resources: CRUD, import, drift detection, reconfigure routing, firewall rollback, mutex serialization, state read-back, plan modifiers, partial failure consistency, reconfigure failure handling (FR6-18)
- ~38 resource types across 8 plugin/service areas (HAProxy, FRR/BGP, ACME, DHCPv4, Unbound DNS, Dynamic DNS, WireGuard, IPsec) plus 4 core areas (firewall, interfaces, routing, system) (FR19-59)
- ~38 data sources mirroring all resource types plus system info (FR60-61)
- Structured error handling with field-level validation diagnostics, permission-specific messages, and connection error differentiation (FR62-65)
- Auto-generated documentation with composition examples for every resource (FR66-68)

**Non-Functional Requirements:**
34 NFRs across 6 quality categories driving architectural decisions:
- **Performance:** 60s full plan refresh, 10s per CRUD operation, configurable read concurrency limit to protect OPNsense PHP-FPM
- **Security:** Write-only attributes for credentials, Sensitive marking, TLS with configurable verification
- **Reliability:** Idempotent operations, state consistency on partial failure, automatic retry with backoff, schema version migration
- **Compatibility:** Terraform CLI 1.0+ (Protocol v6), OpenTofu, GitLab HTTP state backend, Terraform Cloud, statically linked binary
- **Code Quality:** >80% acceptance test coverage, golangci-lint, semver, DevRail compliance
- **Error Quality:** Resource-type + operation + API response + suggested action in every error message

**Scale & Complexity:**

- Primary domain: Go infrastructure provider (Terraform Plugin Framework v6)
- Complexity level: High
- Estimated architectural layers: 7 (provider, service modules, API client core, generated API controllers, code generation pipeline, test framework, documentation generation)
- Resource count: ~38 resources + ~38 data sources
- Implementation tiers: 9 (dependency-ordered)

### Technical Constraints & Dependencies

| Constraint | Source | Impact |
|---|---|---|
| Go language required | Terraform Plugin Framework | No language choice — entire codebase is Go |
| Plugin Framework v6 | PRD FR, technical research | Determines resource lifecycle contract, schema patterns, plan modifiers |
| OPNsense REST API (HTTP Basic Auth) | Domain | API client must handle auth, JSON, non-standard error responses |
| **config.xml integrity (safety-critical)** | Domain (OPNsense architecture) | OPNsense stores all configuration in a single monolithic XML file. Concurrent writes risk XML corruption that can brick the appliance. The global mutex is not just a performance concern — it prevents data corruption. |
| Global mutex for mutations | Domain + safety-critical | All CRUD operations serialized to prevent config.xml corruption; reads parallel |
| Two-phase reconfigure lifecycle | Domain (OPNsense API pattern) | Every mutation = CRUD call + service reconfigure call |
| Firewall savepoint/rollback | Domain (safety-critical) | Firewall filter resources use 3-step apply, not standard reconfigure |
| HTTP 200 on validation errors | Domain (OPNsense API quirk) | API client must parse response body, not HTTP status |
| Blank defaults for missing UUIDs | Domain (OPNsense API quirk) | Must detect "empty" records vs real data |
| Write-only fields | Domain (passwords, PSKs) | Sensitive + UseStateForUnknown; no drift detection on these fields |
| Request body wrapper key | Domain (OPNsense API pattern) | All POST bodies wrapped in resource-type key |
| **API pagination on search endpoints** | Domain (OPNsense API) | Search/list endpoints return paginated results with `rowCount`/`current` params. API client must iterate all pages for complete results. Response format differs between `search` (list) and `get` (single item). |
| **Plugin installation prerequisite** | Domain (OPNsense plugin system) | Plugin resources (HAProxy, FRR, ACME, etc.) require the corresponding OPNsense plugin to be pre-installed. Provider must detect missing plugins and report clear errors (FR64), not fail with cryptic 404s. |
| OPNsense version 26.1.x target | Domain | Version detection, endpoint routing may vary across majors |
| Terraform Registry publishing | Distribution requirement | Requires GitHub mirror, GoReleaser, GPG RSA signing |
| GitLab CI primary | Matthew's infrastructure | Development on self-hosted GitLab, releases mirror to GitHub |

### Cross-Cutting Concerns Identified

| Concern | Affects | Architectural Implication |
|---|---|---|
| **Mutex-protected CRUD** | All 38 resources | Global mutex in API client core; protects config.xml integrity; transparent to resource authors |
| **Auto-reconfigure** | All 38 resources | Standard: ReconfigureEndpoint string. Firewall filter: ReconfigureFunc for savepoint flow |
| **Response parsing** | All API calls | Every response checked for `result != "saved"`, validation errors extracted |
| **API pagination** | Search/list operations, data sources | API client must transparently iterate all pages; handle both search (list) and get (single) response formats |
| **Type conversion** | All resources | String bools ("0"/"1"), SelectedMap enums, SelectedMapList multi-selects, CSVList fields |
| **State read-back** | All resources | After Create/Update, always GET from API to populate state (never echo config) |
| **UUID references** | Linked resources (HAProxy, BGP, firewall) | types.String with UUID validation; Terraform DAG handles dependency ordering |
| **Import** | All 38 resources | ImportStatePassthroughID with UUID; Read populates remaining state |
| **Plugin detection** | All plugin resources (HAProxy, FRR, ACME, DDNS, WireGuard, IPsec) | Detect plugin availability on first operation; report clear error with plugin name if missing |
| **Code generation** | API client controller layer | YAML schemas → Go structs + CRUD methods; reduces per-resource boilerplate |
| **Acceptance testing** | All resources | Full lifecycle test (create→read→import→update→delete) per resource against real OPNsense |
| **Documentation** | All resources | tfplugindocs from schema + templates + example HCL; ships with every resource |

### Architectural Layers (7)

| Layer | Location | Ownership | Purpose |
|---|---|---|---|
| **1. Provider** | `internal/provider/` | Hand-written | Entry point, schema, Configure, resource/datasource registry |
| **2. Service modules** | `internal/service/{module}/` | Hand-written | Per-resource CRUD, schema, model, data source, tests |
| **3. API client core** | `pkg/opnsense/` | Hand-written | HTTP client, auth, mutex, error types, generic CRUD, pagination, type converters |
| **4. API controllers** | `pkg/opnsense/{module}/` | Generated from YAML | Per-module Go structs, endpoint config (ReqOpts), typed CRUD methods |
| **5. Code generation** | `internal/generate/` + `schema/` | Hand-written (pipeline) + YAML (input) | YAML schemas → Go code via go generate |
| **6. Test framework** | `internal/acctest/` | Hand-written | Provider factories, pre-check helpers, test OPNsense configuration |
| **7. Documentation** | `templates/` + `examples/` + `tools/` | Hand-written (templates) + generated (docs/) | tfplugindocs pipeline, composition examples |

## Starter Template Evaluation

### Primary Technology Domain

**Go infrastructure provider (Terraform Plugin Framework v6)** — determined by project requirements. No technology choice to make — the Terraform Plugin Framework mandates Go, and the framework version determines the minimum Go version.

### Starter Options Considered

| Option | Description | Status |
|---|---|---|
| **hashicorp/terraform-provider-scaffolding-framework** | Official HashiCorp scaffold for Plugin Framework providers | **Selected** — canonical starting point |
| Custom scaffold from scratch | Build project structure manually | Rejected — unnecessary when official scaffold exists |
| Fork browningluke/terraform-provider-opnsense | Start from existing OPNsense provider | Rejected — clean-slate architecture is a project goal |

### Selected Starter: hashicorp/terraform-provider-scaffolding-framework

**Rationale:** Official HashiCorp template repository for new Plugin Framework providers. Includes all required configuration files (GoReleaser, golangci-lint, Registry manifest, GNUmakefile, GitHub Actions), a sample resource and data source, and follows current best practices.

**Initialization:**

```bash
git clone https://github.com/hashicorp/terraform-provider-scaffolding-framework terraform-provider-opnsense
cd terraform-provider-opnsense
go mod edit -module github.com/matthew-on-git/terraform-provider-opnsense
grep -rl "scaffolding" --include="*.go" | xargs sed -i 's|hashicorp/scaffolding|matthew-on-git/opnsense|g'
go mod tidy
mkdir -p pkg/opnsense internal/service internal/generate schema
```

**Verified Current Versions (as of 2026-03-18):**

| Component | Version | Source |
|---|---|---|
| Go | 1.25.0 | terraform-plugin-framework go.mod |
| terraform-plugin-framework | v1.19.0 | CHANGELOG.md (released 2026-03-10) |
| terraform-plugin-go | v0.31.0 | framework go.mod |
| terraform-plugin-log | v0.10.0 | framework go.mod |
| Terraform CLI minimum | 1.0+ | Scaffold README |
| Protocol version | 6.0 | terraform-registry-manifest.json |

**Architectural Decisions Provided by Starter:**

- **Language & Runtime:** Go 1.25.0, statically linked (`CGO_ENABLED=0`), cross-compiled for linux/darwin/windows × amd64/arm64/386/arm
- **Build Tooling:** GNUmakefile with targets: `build`, `install`, `lint`, `fmt`, `generate`, `test`, `testacc`. GoReleaser v2 config for cross-compilation + GPG signing.
- **Code Quality:** golangci-lint v2 config with 15+ linters including `depguard` (prevents importing deprecated SDKv2).
- **Testing Framework:** `terraform-plugin-testing` with `ProtoV6ProviderFactories`. TF_ACC gate for acceptance tests.
- **Documentation:** tfplugindocs via `tools/tools.go` with `go generate`. Templates in `templates/`, examples in `examples/`, output to `docs/`.
- **Registry Publishing:** `terraform-registry-manifest.json` declaring Protocol v6.0. GitHub Actions release workflow triggered by `v*` tags.

**What the scaffold does NOT provide (we add):**
- `pkg/opnsense/` — API client (core + generated controllers)
- `internal/service/` — Per-module resource packages
- `internal/generate/` + `schema/` — Code generation pipeline
- `internal/acctest/` — Test helpers
- `internal/validators/` — Shared validators
- DevRail integration (Makefile wrapper, container-based tooling)

**Note:** Project initialization using the scaffold should be the first implementation story.

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
1. API client interface design — defines how every resource interacts with OPNsense
2. Type conversion strategy — every resource needs this
3. Error handling hierarchy — affects all CRUD operations

**Important Decisions (Shape Architecture):**
4. Code generation approach — reduces long-term effort
5. Testing infrastructure — Vagrant locally, QEMU in CI
6. CI/CD pipeline — GitHub Actions primary

**Deferred Decisions (Post-MVP):**
- Multi-version OPNsense support strategy (version-aware endpoint routing)
- Resource skeleton code generation (extend code gen to Terraform layer)
- Plugin detection caching strategy

### API Client Design

**Decision:** Separate `pkg/opnsense/` package with generic CRUD functions, custom HTTP transport, and per-module controllers.

**Core Client Interface:**

| Component | Decision | Rationale |
|---|---|---|
| HTTP transport | `go-retryablehttp` with custom `RoundTripper` for Basic Auth | HashiCorp standard; automatic retry on transient failures |
| Authentication | `apiKeyTransport` injecting Basic Auth on every request | OPNsense uses HTTP Basic Auth with API key as username, secret as password |
| TLS | Configurable `InsecureSkipVerify` via provider config | Self-signed certs common in homelab/enterprise |
| Connection pooling | HTTP keep-alive enabled by default | Reduces TCP handshake overhead across sequential API calls |
| Read concurrency | `semaphore.Weighted` limiting concurrent read operations | Prevents overwhelming OPNsense PHP-FPM worker pool (configurable, default 10) |
| Write serialization | Global mutex (single key) on all Create/Update/Delete | Protects config.xml integrity — safety-critical |
| Pagination | Transparent page iteration in Search/List operations | API client handles `rowCount`/`current` params internally |

**Generic CRUD Pattern (Go generics):**

```go
type ReqOpts struct {
    AddEndpoint         string
    GetEndpoint         string
    UpdateEndpoint      string
    DeleteEndpoint      string
    SearchEndpoint      string
    ReconfigureEndpoint string                    // Standard reconfigure
    ReconfigureFunc     func(ctx) error           // Override for firewall savepoint flow
    Monad               string                    // Request body wrapper key
}

func Add[K any](c *Client, ctx context.Context, opts ReqOpts, resource *K) (string, error)
func Get[K any](c *Client, ctx context.Context, opts ReqOpts, id string) (*K, error)
func Update[K any](c *Client, ctx context.Context, opts ReqOpts, resource *K, id string) error
func Delete(c *Client, ctx context.Context, opts ReqOpts, id string) error
func Search[K any](c *Client, ctx context.Context, opts ReqOpts, params SearchParams) ([]K, error)
```

**Affects:** Every resource and data source implementation.

### Type Conversion Strategy

**Decision:** Three-layer type conversion: OPNsense API types ↔ Go model types ↔ Terraform Framework types.

| OPNsense API Type | Go Model Type | Terraform Type | Conversion |
|---|---|---|---|
| `"0"` / `"1"` (string bool) | `string` | `types.Bool` | `BoolToString()` / `StringToBool()` |
| SelectedMap (`{"key": {"selected": 1}}`) | `SelectedMap` (custom) | `types.String` | Custom `UnmarshalJSON` extracts selected key |
| SelectedMapList (multi-select) | `SelectedMapList` (custom) | `types.Set` of `types.String` | Custom unmarshal returns `[]string` of selected keys |
| CSV list (`"a,b,c"`) | `string` | `types.List` of `types.String` | Split/join on `,` |
| UUID reference | `string` | `types.String` with UUID validator | Passthrough with validation |
| Integer as string (`"443"`) | `string` | `types.Int64` | `strconv.Atoi()` / `strconv.Itoa()` |
| Write-only field | `string` (never read back) | `types.String` (Sensitive, UseStateForUnknown) | Set on write, preserve state on read |

**Conversion location:** `{resource}_model.go` contains `toAPI()` and `fromAPI()` methods that handle all conversions for a given resource. Shared converters live in `pkg/opnsense/types.go`.

**Affects:** Every resource's model layer.

### Error Handling Hierarchy

**Decision:** Three custom error types in `pkg/opnsense/errors.go`, mapped to Terraform diagnostics in the resource layer.

| Error Type | Trigger | API Client Behavior | Terraform Resource Behavior |
|---|---|---|---|
| `NotFoundError` | GET returns blank/default record or JSON unmarshal type error | Return `NotFoundError` | Call `resp.State.RemoveResource()` — resource was deleted out-of-band |
| `ValidationError` | Mutation response has `result != "saved"` | Parse `validations` map, return `ValidationError` with field names | `resp.Diagnostics.AddAttributeError()` per field |
| `AuthError` | HTTP 401 or 403 | Return `AuthError` with status code | `resp.Diagnostics.AddError()` with permission guidance |
| `ServerError` | HTTP 500+, connection refused, timeout | Handled by `go-retryablehttp` (auto-retry). If all retries exhausted, return `ServerError` | `resp.Diagnostics.AddError()` with connection troubleshooting |
| `PluginNotFoundError` | HTTP 404 on a plugin API endpoint | Return `PluginNotFoundError` with plugin name | `resp.Diagnostics.AddError()` telling user to install the plugin |

**Affects:** API client core and every resource's CRUD methods.

### Code Generation Approach

**Decision:** `text/template` (Go standard library) generating API client structs and controller methods from YAML schemas. NOT jennifer.

**Rationale:** text/template uses the same concept as Ansible/Jinja2 templates — `.tmpl` files with `{{.FieldName}}` placeholders. Proven by browningluke/opnsense-go for exactly this use case. Jennifer (programmatic Go AST generation) adds unnecessary learning curve for someone new to Go.

**Critical implementation rule:** Hand-write the first 2-3 API client modules (firewall, haproxy) in Tier 0. Prove the pattern works. THEN extract into YAML → text/template generation. Do not build the code gen pipeline before validating the pattern it generates.

**What gets generated:**
- Go structs with JSON tags matching OPNsense API field names
- `ReqOpts` configuration per resource (endpoints, monad, reconfigure)
- Typed CRUD method wrappers calling generic functions
- Per-module `controller.go` aggregating all resources

**What stays hand-written:**
- API client core (`pkg/opnsense/client.go`, `crud.go`, `mutex.go`, `errors.go`, `types.go`)
- All Terraform resource implementations (`internal/service/*/`)
- Code generation pipeline itself (`internal/generate/`)
- Tests, validators, provider configuration

**YAML schema format (modeled on browningluke):**
```yaml
name: haproxy
reconfigureEndpoint: "/haproxy/service/reconfigure"
resources:
  - name: Server
    monad: server
    endpoints:
      add: "/haproxy/settings/addServer"
      get: "/haproxy/settings/getServer"
      update: "/haproxy/settings/setServer"
      delete: "/haproxy/settings/delServer"
      search: "/haproxy/settings/searchServers"
    attrs:
      - name: Name
        type: string
        key: name
      - name: Address
        type: string
        key: address
      - name: Port
        type: string
        key: port
      - name: Enabled
        type: selectedmap
        key: enabled
```

**Affects:** API client controller layer; long-term development velocity.

### Testing Architecture

**Decision:** Vagrant locally for development, QEMU in GitHub Actions for CI. Test framework is environment-agnostic.

| Environment | Tool | OPNsense Instance | Use Case |
|---|---|---|---|
| Local development | Vagrant (`Vagrantfile` in repo) | Ephemeral VM, spun up per session | Developer acceptance testing |
| CI (GitHub Actions) | QEMU (browningluke's proven approach) | Ephemeral VM per CI run | Automated acceptance tests |
| Both | Environment variables | `OPNSENSE_URI`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET` | Test framework reads these, doesn't know how OPNsense was provisioned |

**Test framework design (`internal/acctest/`):**
- `ProtoV6ProviderFactories` — standard Plugin Framework test setup
- `PreCheck(t)` — validates env vars are set, OPNsense is reachable
- Per-test cleanup — each test creates resources with unique names and destroys them via `CheckDestroy`
- Serial execution (`-p 1`) — matches production mutex behavior; tests share one OPNsense instance

**Vagrantfile:** Included in `test/` directory. `vagrant up` produces a running OPNsense instance with API key/secret output to stdout.

**Affects:** CI/CD pipeline, developer onboarding, test reliability.

### CI/CD Pipeline

**Decision:** GitHub-primary. GitHub Actions for CI/CD. No GitLab mirror.

| Stage | Trigger | Actions |
|---|---|---|
| **Lint & Format** | Every push, every PR | `golangci-lint run`, `gofmt` check, `go vet` |
| **Unit Tests** | Every push, every PR | `go test ./pkg/... ./internal/...` (no TF_ACC, no OPNsense needed) |
| **Acceptance Tests** | PR to main, manual trigger | QEMU OPNsense VM, `TF_ACC=1 go test -p 1 ./...` |
| **Generate & Validate** | Every push | `go generate ./...`, `tfplugindocs validate`, diff check |
| **Release** | Push `v*` tag to main | GoReleaser cross-compile, GPG sign, GitHub Release, Registry auto-discovers |

**DevRail integration:** `make check` wraps lint + format + test. Makefile delegates to GitHub Actions-compatible commands (no Docker container needed for Go tooling since Go is the native language).

**Affects:** Release pipeline, developer workflow, test automation.

### Decision Impact Analysis

**Implementation Sequence:**
1. API client core (HTTP, auth, mutex, errors, types) — blocks everything
2. First hand-written API module (firewall) — validates client pattern
3. Second hand-written API module (haproxy) — validates pattern generalizes to plugins
4. Code generation pipeline — extracts pattern from hand-written modules
5. Remaining API modules generated from YAML
6. Terraform resources (hand-written, using generated API client)
7. Test framework + Vagrantfile
8. CI/CD pipeline (GitHub Actions)
9. Documentation pipeline (tfplugindocs)

**Cross-Component Dependencies:**
- Every Terraform resource depends on the API client core and its module controller
- Code generation depends on the pattern proven by hand-written modules
- Testing depends on Vagrant/QEMU OPNsense provisioning
- Release depends on GitHub Actions + GoReleaser + GPG key setup
- Documentation depends on tfplugindocs + examples + templates

## Implementation Patterns & Consistency Rules

### Pattern Categories Defined

**Critical Conflict Points Identified:** 8 areas where AI agents could make different choices when implementing different resources for this provider.

### Go Naming Patterns

**File naming:**
- All lowercase, underscore-separated: `haproxy_server_resource.go`, NOT `HaproxyServerResource.go`
- Four-file pattern per resource: `{resource}_resource.go`, `{resource}_schema.go`, `{resource}_model.go`, `{resource}_resource_test.go`
- Data sources: `{resource}_data_source.go`
- Module registration: `exports.go` per service package

**Package naming:**
- All lowercase, single word when possible: `haproxy`, `quagga`, `firewall`, `unbound`
- Match OPNsense module names (not Terraform resource names)

**Function naming:**
- Go standard: `PascalCase` for exported, `camelCase` for unexported
- Resource constructors: `newServerResource()`, `newBackendResource()`
- Model conversions: `(m *ServerModel) toAPI() *api.Server`, `(m *ServerModel) fromAPI(s *api.Server)`
- Data source constructors: `newServerDataSource()`

**Variable naming:**
- Go standard: `camelCase` for local variables
- Context always: `ctx context.Context` as first parameter
- Request/response: `req` and `resp` for Terraform CRUD method parameters
- Diagnostics: `resp.Diagnostics.Append(diags...)` pattern consistently

**Constant naming:**
- Resource type names: `const resourceTypeName = "opnsense_haproxy_server"`

### Terraform Schema Patterns

**Resource naming convention:**
`opnsense_{module}_{resource}` — maps to OPNsense API module and resource type.

| OPNsense API Path | Terraform Resource Name |
|---|---|
| `/api/haproxy/settings/...Server` | `opnsense_haproxy_server` |
| `/api/quagga/bgp/...Neighbor` | `opnsense_quagga_bgp_neighbor` |
| `/api/firewall/filter/...Rule` | `opnsense_firewall_filter_rule` |
| `/api/unbound/settings/...HostOverride` | `opnsense_unbound_host_override` |

**Attribute naming convention:**
- All snake_case in Terraform schema (Terraform standard)
- Map from OPNsense API camelCase to Terraform snake_case: `linkedServers` → `linked_servers`
- Boolean attributes use positive naming: `enabled`, not `disabled`
- UUID references named by target: `backend_id`, `server_ids`, not `linked_servers_uuids`

**Attribute description convention:**
- Start with a verb phrase or "The..."/"Whether..."
- Mention OPNsense field name if it differs from Terraform attribute name
- Note defaults for Computed attributes
- Example: `"Whether this server is enabled. Maps to OPNsense 'ssl' field. Defaults to true."`
- Example: `"The IP address or hostname of the backend server. Maps to OPNsense 'address' field."`

**Standard attribute patterns (every resource):**

```go
"id": schema.StringAttribute{
    Computed:    true,
    Description: "UUID of the resource in OPNsense.",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

**Optional + Computed with default (e.g., enabled):**

```go
"enabled": schema.BoolAttribute{
    Optional:    true,
    Computed:    true,
    Default:     booldefault.StaticBool(true),
    Description: "Whether this resource is enabled. Defaults to true.",
},
```

**UUID cross-reference:**

```go
"backend_id": schema.StringAttribute{
    Required:    true,
    Description: "UUID of the HAProxy backend.",
    Validators: []validator.String{
        validators.IsUUIDv4(),
    },
},
```

**Set of UUIDs (unordered multi-reference):**

```go
"server_ids": schema.SetAttribute{
    ElementType: types.StringType,
    Required:    true,
    Description: "Set of HAProxy server UUIDs linked to this backend.",
    Validators: []validator.Set{
        setvalidator.ValueStringsAre(validators.IsUUIDv4()),
    },
},
```

### Resource Implementation Patterns

**Resource struct — every resource follows this pattern:**

```go
type serverResource struct {
    client *opnsense.Client
}

func (r *serverResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

**File scope rules — strict separation:**

| File | Contains | Does NOT Contain |
|---|---|---|
| `{resource}_resource.go` | CRUD methods (Create/Read/Update/Delete/ImportState), Configure, Metadata, Schema reference | Helpers, conversion logic, API structs |
| `{resource}_schema.go` | Schema() method returning the Terraform schema definition | CRUD logic, model structs |
| `{resource}_model.go` | Terraform model struct, API model struct, `toAPI()`, `fromAPI()`, per-resource helper functions | CRUD logic, schema definition |
| `{resource}_resource_test.go` | Acceptance tests, test config functions | Shared test helpers (those go in `internal/acctest/`) |
| `exports.go` | `Resources()` and `DataSources()` registration functions | Any implementation logic |

**CRUD method structure — every resource follows this exact pattern:**

**Create:**
1. Read plan into Terraform model struct
2. Convert model to API struct via `toAPI()`
3. Call `api.Add[K](r.client, ctx, opts, apiStruct)`
4. Capture UUID from response
5. Read back from API via `api.Get[K](r.client, ctx, opts, uuid)`
6. Convert API response to Terraform model via `fromAPI()`
7. Set state from model (not from plan)

**Read:**
1. Read current state to get UUID
2. Call `api.Get[K](r.client, ctx, opts, uuid)`
3. If `NotFoundError` → `resp.State.RemoveResource(ctx)` and return
4. Convert API response to model via `fromAPI()`
5. Set state from model

**Update:**
1. Read plan into Terraform model struct
2. Read current state to get UUID
3. Convert model to API struct via `toAPI()`
4. Call `api.Update[K](r.client, ctx, opts, apiStruct, uuid)`
5. Read back from API via `api.Get[K](r.client, ctx, opts, uuid)`
6. Convert API response to Terraform model via `fromAPI()`
7. Set state from model (not from plan)

**Delete:**
1. Read current state to get UUID
2. Call `api.Delete(r.client, ctx, opts, uuid)`
3. State automatically cleared on return

**ImportState:**
```go
func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**Model boundary rule:** `fromAPI()` receives a clean, unwrapped Go struct. The API client's generic `Get[K]` function handles the OPNsense monad unwrapping internally. Model code never sees the response wrapper.

### Error Handling Patterns

**In resource CRUD methods — consistent error handling:**

```go
// After any API call:
if err != nil {
    var notFoundErr *errs.NotFoundError
    if errors.As(err, &notFoundErr) {
        resp.State.RemoveResource(ctx)
        return
    }
    resp.Diagnostics.AddError(
        "Error reading HAProxy server",
        fmt.Sprintf("Could not read HAProxy server %s: %s", state.ID.ValueString(), err),
    )
    return
}
```

**Never catch and ignore errors. Always surface via Diagnostics.**

### Test Patterns

**Test naming:** `TestAcc{ResourceType}_{scenario}`
- `TestAccHAProxyServer_basic`
- `TestAccHAProxyServer_update`
- `TestAccFirewallAlias_import`

**Test structure — every resource test includes these steps:**

```go
resource.Test(t, resource.TestCase{
    ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
    PreCheck:                 func() { acctest.PreCheck(t) },
    Steps: []resource.TestStep{
        // Step 1: Create and verify
        {
            Config: testConfig,
            Check:  resource.ComposeAggregateTestCheckFunc(...),
        },
        // Step 2: Import
        {
            ResourceName:      "opnsense_haproxy_server.test",
            ImportState:       true,
            ImportStateVerify: true,
        },
        // Step 3: Update and verify
        {
            Config: updatedConfig,
            Check:  resource.ComposeAggregateTestCheckFunc(...),
        },
    },
})
```

**Test config functions:** `testAccHAProxyServerConfig(name, address string, port int) string` — returns HCL as string.

**Test resource naming:** All test resources use `"test"` as the resource name: `resource "opnsense_haproxy_server" "test"`.

**CheckDestroy:** Every test must implement `CheckDestroy` to verify the resource was cleaned up.

### Documentation Patterns

**Template file per resource:** `templates/resources/haproxy_server.md.tmpl`

**Example HCL per resource:** `examples/resources/opnsense_haproxy_server/resource.tf`

**Import example per resource:** `examples/resources/opnsense_haproxy_server/import.sh`

**Composition examples:** Placed in `examples/compositions/` with descriptive directory names:
- `examples/compositions/haproxy-full-stack/main.tf`
- `examples/compositions/bgp-peering/main.tf`
- `examples/compositions/customer-onboarding/main.tf`

### Enforcement Guidelines

**All AI Agents MUST:**
1. Follow the four-file resource pattern exactly — no combining files, no extra files, no helpers in resource files
2. Use `toAPI()` / `fromAPI()` conversions in the model file — never convert inline in CRUD methods
3. Always read back from API after Create/Update — never set state from plan values
4. Use `resp.State.RemoveResource()` on NotFoundError in Read — never return error for missing resources
5. Include import step in every acceptance test — no resource ships without import verification
6. Use `types.Set` for unordered collections — never `types.List` for UUID sets
7. Name Terraform attributes in snake_case — even when OPNsense uses camelCase
8. Store `*opnsense.Client` in resource struct via `Configure` — never create clients in resource methods or use globals
9. Attribute descriptions follow the convention: verb phrase, OPNsense field mapping, defaults noted

**Pattern Enforcement (Automated — no human review dependency):**
- `golangci-lint` catches Go naming and style violations
- CI structural check validates each `internal/service/*/` directory contains the expected four files for each resource registered in `exports.go`
- Acceptance tests validate CRUD + import + drift detection per resource
- `tfplugindocs validate` ensures documentation completeness
- `make check` wraps all automated validation — must pass before any merge

### Anti-Patterns (What to Avoid)

| Anti-Pattern | Why It's Wrong | Correct Pattern |
|---|---|---|
| Setting state from plan values after Create | Breaks drift detection — state must reflect API reality | Always `Get` from API and `fromAPI()` into state |
| Returning error in Read when resource not found | Terraform can't recover — state becomes stuck | Call `resp.State.RemoveResource()` and return nil |
| Using `types.List` for unordered UUID sets | Causes perpetual plan diffs when API returns different order | Use `types.Set` |
| Inline type conversion in CRUD methods | Duplicates logic, inconsistent across resources | All conversions in `toAPI()` / `fromAPI()` model methods |
| Hardcoding OPNsense API URLs in resources | Resources should be API-agnostic | All endpoints in `ReqOpts`, resources call generic CRUD |
| Skipping import test step | Import is a day-one requirement, not optional | Every test includes `ImportState: true` step |
| Ignoring `resp.Diagnostics` errors | Silent failures corrupt state | Always check `resp.Diagnostics.HasError()` after state operations |
| Helper functions in resource file | Resource files become grab-bags, logic is unreusable | Shared helpers → `pkg/opnsense/types.go`. Per-resource → model file |
| Unwrapping API monad in model/resource code | Violates layer boundary — API client handles protocol | `fromAPI()` receives clean struct; API client handles unwrap |
| Creating `opnsense.Client` in resource methods | Multiple clients, no shared mutex, auth duplication | Store `*opnsense.Client` from `Configure`, reuse across all methods |

## Project Structure & Boundaries

### Complete Project Directory Structure

```
terraform-provider-opnsense/
├── main.go                                    # Entry point — providerserver.Serve()
├── go.mod                                     # Module: github.com/matthew-on-git/terraform-provider-opnsense
├── go.sum
├── GNUmakefile                                # Dev targets: build, install, lint, fmt, generate, test, testacc, check
├── .goreleaser.yml                            # GoReleaser v2 — cross-compile, GPG sign, ZIP
├── .golangci.yml                              # golangci-lint v2 — 15+ linters, depguard blocks SDKv2
├── .editorconfig                              # DevRail formatting rules
├── terraform-registry-manifest.json           # Protocol v6.0 declaration
├── CHANGELOG.md                               # Semver changelog (FEATURES, IMPROVEMENTS, BUG FIXES)
├── LICENSE                                    # MPL-2.0
├── README.md                                  # Project overview, quickstart, links
├── CONTRIBUTING.md                            # Contributor guide: prerequisites, patterns, PR process
│
├── .github/
│   └── workflows/
│       ├── ci.yml                             # Lint + unit tests + generate validation (every push/PR)
│       ├── acceptance.yml                     # QEMU OPNsense VM + acceptance tests (PR to main)
│       └── release.yml                        # GoReleaser + GPG sign (on v* tag)
│
├── internal/
│   ├── provider/
│   │   ├── provider.go                        # Provider struct, Schema, Configure, Resources/DataSources registry
│   │   ├── provider_test.go                   # Provider configuration tests
│   │   └── factory.go                         # ProtoV6ProviderServerFactory
│   │
│   ├── service/                               # One package per OPNsense module
│   │   ├── firewall/                          # Core firewall (Tier 0-1)
│   │   │   ├── exports.go                     # Resources() and DataSources() registration
│   │   │   ├── alias_resource.go              # CRUD: Create/Read/Update/Delete/ImportState
│   │   │   ├── alias_schema.go                # Terraform schema definition
│   │   │   ├── alias_model.go                 # API model + toAPI()/fromAPI() conversions
│   │   │   ├── alias_data_source.go           # Read-only data source
│   │   │   ├── alias_resource_test.go         # Acceptance tests
│   │   │   ├── category_resource.go
│   │   │   ├── category_schema.go
│   │   │   ├── category_model.go
│   │   │   ├── category_resource_test.go
│   │   │   ├── filter_rule_resource.go        # Uses ReconfigureFunc (savepoint)
│   │   │   ├── filter_rule_schema.go
│   │   │   ├── filter_rule_model.go
│   │   │   ├── filter_rule_resource_test.go
│   │   │   ├── nat_port_forward_resource.go
│   │   │   ├── nat_port_forward_schema.go
│   │   │   ├── nat_port_forward_model.go
│   │   │   ├── nat_port_forward_resource_test.go
│   │   │   ├── nat_outbound_resource.go
│   │   │   ├── nat_outbound_schema.go
│   │   │   ├── nat_outbound_model.go
│   │   │   └── nat_outbound_resource_test.go
│   │   │
│   │   ├── haproxy/                           # HAProxy plugin (Tier 0, 2)
│   │   │   ├── exports.go
│   │   │   ├── server_resource.go
│   │   │   ├── server_schema.go
│   │   │   ├── server_model.go
│   │   │   ├── server_resource_test.go
│   │   │   ├── backend_resource.go
│   │   │   ├── backend_schema.go
│   │   │   ├── backend_model.go
│   │   │   ├── backend_resource_test.go
│   │   │   ├── frontend_resource.go
│   │   │   ├── frontend_schema.go
│   │   │   ├── frontend_model.go
│   │   │   ├── frontend_resource_test.go
│   │   │   ├── acl_resource.go
│   │   │   ├── acl_schema.go
│   │   │   ├── acl_model.go
│   │   │   ├── acl_resource_test.go
│   │   │   ├── healthcheck_resource.go
│   │   │   ├── healthcheck_schema.go
│   │   │   ├── healthcheck_model.go
│   │   │   └── healthcheck_resource_test.go
│   │   │
│   │   ├── system/                            # Core system + interfaces + routing (Tier 3)
│   │   │   ├── exports.go
│   │   │   ├── interface_resource.go
│   │   │   ├── vlan_resource.go
│   │   │   ├── vip_resource.go                # Virtual IPs (CARP, IP Alias)
│   │   │   ├── route_resource.go
│   │   │   ├── gateway_resource.go
│   │   │   ├── gateway_group_resource.go
│   │   │   ├── general_resource.go
│   │   │   └── ... (schema, model, test files per resource)
│   │   │
│   │   ├── quagga/                            # FRR/BGP plugin (Tier 4)
│   │   │   ├── exports.go
│   │   │   ├── general_resource.go
│   │   │   ├── bgp_general_resource.go
│   │   │   ├── bgp_neighbor_resource.go
│   │   │   ├── prefix_list_resource.go
│   │   │   ├── route_map_resource.go
│   │   │   └── ... (schema, model, test files per resource)
│   │   │
│   │   ├── acme/                              # ACME plugin (Tier 5)
│   │   │   ├── exports.go
│   │   │   ├── account_resource.go
│   │   │   ├── certificate_resource.go
│   │   │   ├── challenge_resource.go
│   │   │   └── ... (schema, model, test files per resource)
│   │   │
│   │   ├── unbound/                           # Unbound DNS (Tier 6)
│   │   │   ├── exports.go
│   │   │   ├── host_override_resource.go
│   │   │   ├── domain_override_resource.go
│   │   │   ├── acl_resource.go
│   │   │   └── ... (schema, model, test files per resource)
│   │   │
│   │   ├── wireguard/                         # WireGuard VPN (Tier 7)
│   │   │   ├── exports.go
│   │   │   ├── server_resource.go
│   │   │   ├── peer_resource.go
│   │   │   └── ... (schema, model, test files per resource)
│   │   │
│   │   ├── ipsec/                             # IPsec VPN (Tier 7)
│   │   │   ├── exports.go
│   │   │   ├── phase1_resource.go
│   │   │   ├── phase2_resource.go
│   │   │   ├── psk_resource.go
│   │   │   └── ... (schema, model, test files per resource)
│   │   │
│   │   ├── dhcpv4/                            # DHCPv4 (Tier 8)
│   │   │   ├── exports.go
│   │   │   ├── pool_resource.go
│   │   │   ├── static_mapping_resource.go
│   │   │   ├── option_resource.go
│   │   │   └── ... (schema, model, test files per resource)
│   │   │
│   │   └── ddclient/                          # Dynamic DNS (Tier 8)
│   │       ├── exports.go
│   │       ├── account_resource.go
│   │       ├── provider_resource.go
│   │       └── ... (schema, model, test files per resource)
│   │
│   ├── acctest/
│   │   └── acctest.go                         # ProtoV6ProviderFactories, PreCheck, shared test helpers
│   │
│   ├── validators/
│   │   ├── uuid.go                            # IsUUIDv4() validator
│   │   ├── port.go                            # Port range validator (1-65535)
│   │   └── ip.go                              # IP address validator
│   │
│   └── generate/                              # Code generation pipeline (built after Tier 0)
│       ├── main.go                            # Generator entry point
│       └── templates/
│           ├── controller.go.tmpl             # Per-module controller template
│           └── resource.go.tmpl               # Per-resource struct + CRUD template
│
├── pkg/
│   └── opnsense/                              # API client — independent of Terraform types
│       ├── client.go                          # HTTP client, auth transport, Configure
│       ├── crud.go                            # Generic CRUD (Add[K], Get[K], Update[K], Delete, Search[K])
│       ├── mutex.go                           # Global MutexKV (config.xml integrity protection)
│       ├── errors.go                          # NotFoundError, ValidationError, AuthError, ServerError, PluginNotFoundError
│       ├── types.go                           # SelectedMap, SelectedMapList, BoolToString, StringToBool
│       ├── pagination.go                      # Transparent search result pagination
│       ├── reconfigure.go                     # Standard reconfigure + firewall savepoint flow
│       │
│       ├── firewall/                          # Hand-written first (Tier 0), template for code gen later
│       │   ├── controller.go
│       │   ├── alias.go
│       │   ├── filter.go
│       │   ├── nat.go
│       │   └── generate.go                    # go:generate directive (added after code gen pipeline)
│       │
│       ├── haproxy/                           # Hand-written first (Tier 0), template for code gen later
│       │   ├── controller.go
│       │   ├── server.go
│       │   ├── backend.go
│       │   ├── frontend.go
│       │   ├── acl.go
│       │   └── generate.go
│       │
│       ├── quagga/                            # Generated from YAML (Tier 4+)
│       ├── acme/
│       ├── system/
│       ├── unbound/
│       ├── wireguard/
│       ├── ipsec/
│       ├── dhcpv4/
│       └── ddclient/
│
├── schema/                                    # YAML schemas for code generation
│   ├── firewall.yml
│   ├── haproxy.yml
│   ├── quagga.yml
│   ├── acme.yml
│   ├── system.yml
│   ├── unbound.yml
│   ├── wireguard.yml
│   ├── ipsec.yml
│   ├── dhcpv4.yml
│   └── ddclient.yml
│
├── templates/                                 # tfplugindocs templates
│   ├── index.md.tmpl                          # Provider overview page
│   ├── resources/
│   │   ├── firewall_alias.md.tmpl
│   │   ├── haproxy_server.md.tmpl
│   │   └── ... (one per resource)
│   └── data-sources/
│       ├── firewall_alias.md.tmpl
│       └── ... (one per data source)
│
├── examples/                                  # HCL examples for documentation + testing
│   ├── provider/
│   │   └── provider.tf
│   ├── resources/
│   │   ├── opnsense_firewall_alias/
│   │   │   ├── resource.tf
│   │   │   └── import.sh
│   │   ├── opnsense_haproxy_server/
│   │   │   ├── resource.tf
│   │   │   └── import.sh
│   │   └── ... (one directory per resource)
│   ├── data-sources/
│   │   ├── opnsense_firewall_alias/
│   │   │   └── data-source.tf
│   │   └── ... (one directory per data source)
│   └── compositions/                          # Multi-resource realistic examples
│       ├── customer-onboarding/
│       │   └── main.tf
│       ├── haproxy-full-stack/
│       │   └── main.tf
│       ├── bgp-peering/
│       │   └── main.tf
│       ├── firewall-baseline/
│       │   └── main.tf
│       ├── vpn-wireguard/
│       │   └── main.tf
│       └── dns-management/
│           └── main.tf
│
├── docs/                                      # Generated by tfplugindocs (committed to repo)
│   ├── index.md
│   ├── resources/
│   └── data-sources/
│
├── tools/
│   └── tools.go                               # Build tool deps: tfplugindocs, copywrite
│
└── test/
    ├── Vagrantfile                            # Ephemeral OPNsense VM for local development
    └── scripts/
        ├── create-apikey.sh                   # Generate API key on test VM
        └── validate-structure.sh              # CI check: service directories match exports.go
```

### Architectural Boundaries

**Layer 1 → Layer 2 (Provider → Service Modules):**
Provider registers service modules via `exports.go`. Each module is self-contained — the provider doesn't know about individual resources, only that each module provides `Resources()` and `DataSources()`.

**Layer 2 → Layer 3 (Service Modules → API Client):**
Resources access the API client via `r.client` (set during `Configure`). Resources call generic CRUD functions with `ReqOpts`. Resources never construct HTTP requests or parse raw responses.

**Layer 3 → Layer 4 (API Client Core → Controllers):**
Core provides generic `Add[K]`, `Get[K]`, `Update[K]`, `Delete`, `Search[K]`. Controllers provide per-resource `ReqOpts` and typed wrapper methods. Core handles mutex, auth, retry, pagination, error parsing. Controllers handle endpoint configuration and struct definitions.

**Layer 5 (Code Generation):**
Produces Layer 4 artifacts (controllers) from YAML schemas. Generated code lives alongside hand-written code in `pkg/opnsense/{module}/`. The `generate.go` file in each module contains the `go:generate` directive.

**External Boundary (Provider ↔ OPNsense API):**
All external communication goes through `pkg/opnsense/client.go`. Single HTTP client instance, shared across all resources. Mutex-protected writes, semaphore-limited reads. All responses parsed for OPNsense-specific error patterns.

### Requirements to Structure Mapping

| FR Category | Primary Location | Test Location |
|---|---|---|
| Provider Config (FR1-5) | `internal/provider/provider.go` | `internal/provider/provider_test.go` |
| Cross-Cutting (FR6-18) | `pkg/opnsense/` (client, crud, mutex, errors) | `pkg/opnsense/*_test.go` + all resource tests |
| Firewall (FR19-23) | `internal/service/firewall/` | `internal/service/firewall/*_test.go` |
| HAProxy (FR24-29) | `internal/service/haproxy/` | `internal/service/haproxy/*_test.go` |
| FRR/BGP (FR30-34) | `internal/service/quagga/` | `internal/service/quagga/*_test.go` |
| ACME (FR35-39) | `internal/service/acme/` | `internal/service/acme/*_test.go` |
| Core Infra (FR40-46) | `internal/service/system/` | `internal/service/system/*_test.go` |
| Unbound DNS (FR47-49) | `internal/service/unbound/` | `internal/service/unbound/*_test.go` |
| VPN (FR50-54) | `internal/service/wireguard/` + `ipsec/` | respective `*_test.go` |
| DHCP (FR55-57) | `internal/service/dhcpv4/` | `internal/service/dhcpv4/*_test.go` |
| Dynamic DNS (FR58-59) | `internal/service/ddclient/` | `internal/service/ddclient/*_test.go` |
| Data Sources (FR60-61) | `*_data_source.go` in each service package | Tested alongside resources |
| Error Handling (FR62-65) | `pkg/opnsense/errors.go` + resource CRUD methods | `pkg/opnsense/errors_test.go` + resource tests |
| Documentation (FR66-68) | `templates/` + `examples/` + `tools/` | `tfplugindocs validate` in CI |

### Data Flow

```
User HCL Config
    ↓
Terraform Core (plan/apply)
    ↓ gRPC Protocol v6
Provider (internal/provider/)
    ↓ Resources()/DataSources() registry
Service Module (internal/service/{module}/)
    ↓ r.client.{Module}().AddServer(ctx, apiStruct)
API Client Core (pkg/opnsense/)
    ↓ mutex.Lock() → HTTP POST → mutex.Unlock()
OPNsense REST API
    ↓ JSON response
API Client Core → parse response, check result, extract errors
    ↓ clean Go struct (monad unwrapped)
Service Module → fromAPI() → Terraform state
    ↓
Terraform Core → plan diff / state update
```

## Architecture Validation Results

### Coherence Validation

**Decision Compatibility:** All technology decisions verified compatible — Go 1.25.0 + Framework v1.19.0 + Protocol v6.0 + go-retryablehttp + GoReleaser v2 + golangci-lint v2. No version conflicts or incompatible dependency chains.

**Pattern Consistency:** Four-file resource pattern, CRUD method structure, error handling, and type conversion patterns are consistent across all 10 service modules. Naming conventions (Go, Terraform, files, packages) are internally consistent.

**Structure Alignment:** 7-layer architecture maps cleanly to directory structure. Layer boundaries enforced by Go's package system (imports flow one direction). Code generation produces artifacts in correct locations.

### Requirements Coverage

**Functional Requirements:** 68/68 covered (100%). Every FR maps to a specific directory and architectural layer.

**Non-Functional Requirements:** 34/34 covered (100%). Performance, security, reliability, compatibility, code quality, and error quality all have architectural support.

### Implementation Readiness

**Decision Completeness:** All critical decisions documented with versions, rationale, and impact analysis. Implementation sequence explicit (9 steps with dependencies). Deferred decisions identified.

**Structure Completeness:** ~100+ files specified in complete directory tree. All 10 OPNsense modules mapped. CI/CD workflows, test infrastructure, and documentation pipeline defined.

**Pattern Completeness:** 9 enforcement rules, 10 anti-patterns, concrete Go code examples for every pattern. Automated CI enforcement (no human review dependency).

### Gap Analysis

**Critical Gaps:** 0

**Important Gaps:** 2 (both resolve during Tier 0)
1. Firewall savepoint flow implementation detail — will be defined when `firewall_filter_rule` is implemented
2. Data source code example — will emerge alongside first `firewall_alias` data source

**Nice-to-Have Gaps:** 1
- Formal ADR for code gen approach choice (rationale captured in decisions section)

### Architecture Completeness Checklist

- [x] Project context analyzed (68 FRs, 34 NFRs, 15 constraints, 12 cross-cutting concerns)
- [x] Scale assessed (High — ~38 resources + ~38 data sources across 10 modules)
- [x] All critical decisions documented with versions
- [x] Implementation patterns with code examples
- [x] Complete directory structure with file-level detail
- [x] Layer boundaries defined and enforceable
- [x] Requirements mapped to directories (100% FR coverage)
- [x] Automated enforcement via CI
- [x] Implementation sequence defined (9 steps)

### Architecture Readiness Assessment

**Overall Status:** READY FOR IMPLEMENTATION
**Confidence Level:** High
**FR Coverage:** 100% | **NFR Coverage:** 100% | **Critical Gaps:** 0

**First Implementation Priority:**
1. Clone scaffold and rebrand to `github.com/matthew-on-git/terraform-provider-opnsense`
2. Create `pkg/opnsense/` API client core
3. Implement `opnsense_firewall_alias` (Tier 0 — simplest resource, validates full stack)
4. Implement `opnsense_haproxy_server` (Tier 0 — validates plugin API pattern)
5. Extract code generation pipeline from hand-written patterns
