# Reference Terraform Provider Analysis for Network Appliances

**Date**: 2026-03-13
**Purpose**: Identify architectural patterns, best practices, and anti-patterns from existing community Terraform providers that wrap REST APIs for network appliances and firewalls.

---

## 1. Provider Inventory

| Provider | Repo | Stars | Forks | Framework | API Client | License |
|----------|------|-------|-------|-----------|------------|---------|
| OPNsense (browningluke) | [browningluke/terraform-provider-opnsense](https://github.com/browningluke/terraform-provider-opnsense) | 146 | 34 | Plugin Framework (v6) | Separate repo ([opnsense-go](https://pkg.go.dev/github.com/browningluke/opnsense-go/pkg/api)) | MPL-2.0 |
| FortiOS (Fortinet) | [fortinetdev/terraform-provider-fortios](https://github.com/fortinetdev/terraform-provider-fortios) | 81 | 59 | SDKv2 | Separate repo ([forti-sdk-go](https://github.com/fortinetdev/forti-sdk-go)) | MPL-2.0 |
| PAN-OS (Palo Alto) | [PaloAltoNetworks/terraform-provider-panos](https://github.com/PaloAltoNetworks/terraform-provider-panos) | 108 | 77 | Plugin Framework (v6) | Separate repo ([pango](https://github.com/PaloAltoNetworks/pango)) | MIT |
| RouterOS (MikroTik) | [terraform-routeros/terraform-provider-routeros](https://github.com/terraform-routeros/terraform-provider-routeros) | 332 | 93 | SDKv2 | Inline (routeros pkg) | MPL-2.0 |
| MikroTik (ddelnano) | [ddelnano/terraform-provider-mikrotik](https://github.com/ddelnano/terraform-provider-mikrotik) | 139 | 30 | Framework + SDKv2 (mux) | Inline (`client/` pkg) | MIT |
| pfSense (marshallford) | [marshallford/terraform-provider-pfsense](https://github.com/marshallford/terraform-provider-pfsense) | 39 | 6 | Plugin Framework (v6) | Inline (`pkg/pfsense/`) | MIT |

---

## 2. Detailed Provider Analysis

### 2.1 browningluke/terraform-provider-opnsense (PRIMARY REFERENCE)

**Source**: https://github.com/browningluke/terraform-provider-opnsense

**Framework**: Terraform Plugin Framework (Protocol v6) — the modern, recommended framework.

**Architecture**:
```
terraform-provider-opnsense/
├── main.go                          # Entry point, tf6server.Serve()
├── internal/
│   ├── provider/
│   │   ├── provider.go              # Provider config, resource/datasource registration
│   │   └── factory.go               # ProtoV6ProviderServerFactory
│   ├── service/                     # Resources organized by OPNsense module
│   │   ├── firewall/
│   │   │   ├── exports.go           # Resources() and DataSources() registration
│   │   │   ├── alias_resource.go    # CRUD implementation
│   │   │   ├── alias_schema.go      # Schema definition
│   │   │   ├── alias_data_source.go # Data source implementation
│   │   │   └── alias_resource_test.go # Acceptance tests
│   │   ├── unbound/
│   │   ├── wireguard/
│   │   ├── ipsec/
│   │   ├── kea/
│   │   ├── quagga/
│   │   ├── interfaces/
│   │   ├── routes/
│   │   └── diagnostics/
│   ├── validators/                  # Custom validators (IP/CIDR, UUID, etc.)
│   └── tools/                       # Type utilities
├── docs/                            # Generated documentation
├── templates/                       # tfplugindocs templates
├── examples/                        # Example configurations
└── scripts/                         # Helper scripts (API key creation)
```

**API Client (separate repo)**: https://github.com/browningluke/opnsense-go
```
opnsense-go/
├── pkg/
│   ├── api/
│   │   ├── client.go                # HTTP client, Basic Auth, retryablehttp
│   │   ├── crud_utils.go            # Generic CRUD with Go generics
│   │   ├── data.go                  # Data structures
│   │   ├── mutexkv.go               # Mutex key-value for concurrent safety
│   │   └── selected_map.go          # OPNsense-specific map types
│   ├── opnsense/
│   │   └── client.go                # High-level client composing all controllers
│   ├── firewall/
│   │   ├── controller.go            # Generated controller
│   │   ├── filter.go                # Filter resource CRUD
│   │   └── generate.go              # go:generate directive
│   ├── unbound/
│   ├── wireguard/
│   └── errs/                        # Custom error types (NotFound)
├── schema/                          # YAML schema definitions
│   ├── firewall.yml                 # Declarative resource definitions
│   ├── unbound.yml
│   └── ...
└── internal/generate/               # Code generator
    └── api/
        └── templates/               # Go templates for generated code
```

**Key Patterns**:
- **Separate API client**: The `opnsense-go` library is a standalone Go module with its own tests, schemas, and code generation. The Terraform provider depends on it.
- **Schema-driven code generation**: YAML files in `schema/` define resource endpoints, attributes, and types. Go templates generate controller boilerplate.
- **Generic CRUD**: Uses Go generics (`Add[K]`, `Get[K]`, `Update[K]`, `Delete[K]`) for type-safe CRUD operations across all resources.
- **Mutex-protected reconfiguration**: After each write operation, the client calls `ReconfigureService()` to apply changes, protected by a global mutex to prevent race conditions.
- **Service-based resource organization**: Resources are grouped by OPNsense module (firewall, unbound, wireguard, etc.), each with an `exports.go` registration file.
- **Per-resource file separation**: Each resource has separate files for schema, resource CRUD, data source, and tests.
- **Authentication**: Basic Auth (API key + API secret), with retryable HTTP client (configurable backoff/retries).
- **State migration**: `UpgradeState()` method for schema versioning (e.g., flat attrs to nested blocks).

**Testing**:
- Acceptance tests run against a real OPNsense instance in CI (QEMU VM with OPNsense image).
- Tests follow standard `resource.Test()` pattern with create/import/update/delete steps.
- CI spins up OPNsense VM, generates API keys, runs `TF_ACC=1 go test -p 1` (sequential).

**Limitations**:
- Pre-v1.0: schemas subject to change.
- Limited resource coverage (no DHCP, no full core API parity yet).
- The `opnsense-go` code generation is basic (template-based, not from OpenAPI spec).
- Bug found in provider.go: min_backoff assignment uses wrong variable.

**Release**: GoReleaser + GPG signing, triggered by version tags.

---

### 2.2 fortinetdev/terraform-provider-fortios

**Source**: https://github.com/fortinetdev/terraform-provider-fortios

**Framework**: SDKv2 (older, maintenance-mode framework).

**Architecture**:
```
terraform-provider-fortios/
├── main.go
├── fortios/                         # ALL code in single flat package
│   ├── client.go                    # Client wrapper
│   ├── config.go                    # Provider configuration + utilities
│   ├── resource_firewall_policy.go  # ~200 fields per resource
│   ├── data_source_firewall_*.go    # Data sources
│   └── ... (4000+ files!)
├── vendor/                          # Vendored dependencies
└── go.mod
```

**API Client**: Separate repo (`fortinetdev/forti-sdk-go`) with:
- `fortios/auth/` — Authentication handling
- `fortios/config/` — Configuration structs
- `fortios/request/` — HTTP request builder
- `fortios/sdkcore/` — Per-resource CRUD methods

**Key Patterns**:
- **Flat package structure**: Everything in a single `fortios/` package. Over 4000 files with no sub-packages.
- **Code generation**: Resources appear to be generated from FortiOS API specs (massive, repetitive patterns).
- **SDKv2 pattern**: Uses `schema.Resource{}` with CRUD function pointers.
- **Token-based auth**: Uses FortiOS API tokens (bearer tokens).
- **Dual target**: Supports both FortiGate and FortiManager in one provider.
- **Massive schema**: Firewall policy resource has ~200 fields.

**Limitations (anti-patterns to avoid)**:
- Flat package structure is unmaintainable at scale.
- SDKv2 is in maintenance mode; Plugin Framework is recommended.
- Vendored dependencies (outdated practice).
- No clear separation of concerns.
- Travis CI (outdated, most projects use GitHub Actions).

---

### 2.3 PaloAltoNetworks/terraform-provider-panos

**Source**: https://github.com/PaloAltoNetworks/terraform-provider-panos

**Framework**: Plugin Framework (Protocol v6) — modern.

**Architecture**:
```
terraform-provider-panos/
├── main.go
├── internal/
│   ├── provider/
│   │   ├── provider.go              # Provider config, ~90 resources registered
│   │   ├── commit.go                # Explicit commit action resource
│   │   ├── commit_crud.go           # Commit CRUD logic
│   │   ├── address.go               # Generated resource files
│   │   ├── security_policy.go
│   │   ├── security_policy_rules.go # Ordered rule management
│   │   ├── errors.go                # Error handling utilities
│   │   ├── tfid.go                  # Terraform ID helpers
│   │   └── tools.go                 # Shared utilities
│   └── manager/
│       ├── manager.go               # Generic resource manager
│       ├── entry.go                 # CRUD entry points
│       ├── config.go                # Manager configuration
│       └── uuid.go                  # UUID management for rules
└── examples/
```

**API Client (separate repo)**: https://github.com/PaloAltoNetworks/pango
```
pango/
├── client.go                        # Core SDK client (XML API + REST)
├── commit/                          # Commit operations
│   ├── firewall.go                  # Firewall commits
│   └── panorama.go                  # Panorama commits
├── device/                          # Device management
│   └── adminrole/
│       ├── entry.go                 # Resource entry type
│       ├── interfaces.go            # Interface definitions
│       ├── location.go              # Location helpers
│       └── service.go               # Service operations
└── ... (follows entry/interfaces/location/service pattern for each resource)
```

**Code Generation**: https://github.com/PaloAltoNetworks/pan-os-codegen
- Generates BOTH the pango SDK AND the Terraform provider from spec files.
- Spec files are normalized versions of the PAN-OS XML schema.
- Single source of truth for all resources.
- `go generate` commands: `-t mksdk` (SDK only), `-t mktp` (provider only).

**Key Patterns**:
- **Explicit commit resource**: PAN-OS requires explicit commits to apply configuration changes. The provider exposes this as a `panos_commit` action resource rather than auto-committing per resource operation. This gives users control over when changes are committed (batching multiple changes).
- **Push to devices**: Separate `panos_push_to_devices` resource for Panorama deployments.
- **Generic manager**: `internal/manager/` provides a reusable CRUD manager that handles entry lifecycle, state management, and import operations.
- **Code generation from specs**: Both SDK and provider are generated from the same normalized XML schema specs. This is the most sophisticated approach among all providers studied.
- **Local inspection mode**: The pango SDK can operate offline by loading XML configs, enabling testing without a live device.
- **Multi-target**: Supports both direct Firewall management and Panorama (centralized management).

**Testing**:
- Acceptance tests require live PAN-OS instance with `TF_ACC=1`.
- Environment variables for hostname, API key, SSL settings.
- Comprehensive test coverage for generated resources.

---

### 2.4 terraform-routeros/terraform-provider-routeros

**Source**: https://github.com/terraform-routeros/terraform-provider-routeros

**Framework**: SDKv2 (but the most actively maintained SDKv2 provider in this analysis).

**Architecture**:
```
terraform-provider-routeros/
├── main.go
├── routeros/
│   ├── provider.go                  # Provider config, 200+ resources registered
│   ├── mikrotik.go                  # Core abstraction layer
│   ├── mikrotik_client.go           # Client interface
│   ├── mikrotik_client_api.go       # MikroTik API protocol client
│   ├── mikrotik_client_rest.go      # REST API client
│   ├── mikrotik_crud.go             # Generic CRUD operations
│   ├── mikrotik_serialize.go        # Terraform <-> MikroTik serialization
│   ├── mikrotik_resource_drift.go   # Configuration drift detection
│   ├── provider_schema_helpers.go   # Reusable schema property factories
│   ├── resource_ip_firewall_filter.go # Individual resource definitions
│   └── ... (478 files total)
└── .github/
    └── workflows/
        └── module_testing.yml       # Tests against containerized RouterOS
```

**Key Patterns**:
- **Dual transport**: Supports both MikroTik's proprietary API protocol AND REST API, abstracted behind a common `Client` interface.
- **Schema helpers (property factories)**: `PropName()`, `PropEnabled()`, `PropMacAddressRw()`, etc. — reusable functions that generate consistent schema attributes.
- **Generic serialization**: `TerraformResourceDataToMikrotik()` and `MikrotikResourceDataToTerraform()` handle bidirectional data conversion using schema metadata annotations.
- **Metadata-driven resources**: Resources carry metadata (`MetaResourcePath`, `MetaTransformSet`, etc.) that the generic CRUD layer uses to interact with the correct API endpoints.
- **Drift detection**: `mikrotik_resource_drift.go` + `mikrotik_resource_drift.yaml` manage known configuration drift scenarios.
- **DiffSuppressFunc**: Custom diff suppression for time formats, bit/byte values, and system-generated fields.
- **Containerized testing**: CI runs acceptance tests against containerized RouterOS instances across multiple versions (7.12, 7.15, 7.16).

**Strengths**:
- Highest star count (332) — most popular provider in this category.
- Very active (updated today, 2026-03-13).
- Broad resource coverage (200+ resources).
- The property factory pattern significantly reduces boilerplate.
- Multi-version testing matrix.

**Limitations**:
- SDKv2 (not the recommended Plugin Framework).
- All code in a single `routeros/` package (flat, but managed well with naming conventions).
- API client is inline rather than a separate library.

---

### 2.5 ddelnano/terraform-provider-mikrotik

**Source**: https://github.com/ddelnano/terraform-provider-mikrotik

**Framework**: Mixed — Plugin Framework + SDKv2 via terraform-plugin-mux.

**Architecture**:
```
terraform-provider-mikrotik/
├── main.go
├── client/
│   ├── client.go                    # MikroTik API client
│   ├── client_crud.go               # Generic CRUD
│   ├── console_inspect.go           # Schema discovery via console
│   ├── bgp_instance.go              # Per-resource client methods
│   └── ... (resource-specific files)
├── mikrotik/                        # SDKv2 resources (legacy)
└── mikrotik_framework/              # Plugin Framework resources (new)
```

**Key Pattern**: Migration pattern from SDKv2 to Plugin Framework using mux. New resources use Plugin Framework; existing resources remain on SDKv2 until migrated. This is a practical incremental migration approach.

---

### 2.6 marshallford/terraform-provider-pfsense

**Source**: https://github.com/marshallford/terraform-provider-pfsense

**Framework**: Plugin Framework (Protocol v6) — modern.

**Architecture**:
```
terraform-provider-pfsense/
├── main.go
├── internal/
│   └── provider/
│       ├── provider.go              # Provider config, resource registration
│       ├── provider_custom_types.go # Custom Terraform types
│       ├── provider_plan_modifiers.go # Plan modification logic
│       ├── provider_utils.go        # Shared utilities
│       ├── validators.go            # Custom validators
│       ├── dhcpv4_apply_resource.go # Apply trigger resource
│       ├── dhcpv4_staticmapping_resource.go
│       ├── dhcpv4_staticmapping_common.go
│       ├── dnsresolver_apply_resource.go
│       ├── dnsresolver_hostoverride_resource.go
│       ├── dnsresolver_hostoverride_common.go
│       ├── firewall_filter_reload_resource.go
│       ├── firewall_ip_alias_resource.go
│       └── execute_php_command_resource.go
├── pkg/
│   └── pfsense/
│       ├── client.go                # HTTP client with CSRF, session mgmt
│       ├── http.go                  # HTTP helpers
│       ├── html.go                  # HTML scraping (no REST API!)
│       ├── errors.go                # Error types
│       ├── dhcpv4_apply.go          # Service-specific operations
│       ├── dhcpv4_staticmapping.go
│       ├── firewall_alias.go
│       └── execute_php_command.go   # PHP execution for data retrieval
└── .golangci.yml                    # Linting configuration
```

**Key Patterns**:
- **Apply/reload trigger resources**: Since pfSense (like OPNsense) requires explicit service reconfiguration after changes, this provider models it as separate "apply" resources (`dhcpv4_apply`, `dnsresolver_apply`, `firewall_filter_reload`). These resources trigger service reconfiguration when created; their Read/Update/Delete are no-ops.
- **No REST API**: pfSense doesn't have a proper REST API. The client scrapes HTML, parses CSRF tokens, and submits forms. PHP commands are executed server-side for data retrieval.
- **Feature-specific mutexes**: Global write mutex prevents concurrent writes, with feature-specific mutexes (DHCPv4, DNS, Firewall) for finer-grained locking.
- **Inline API client**: `pkg/pfsense/` is in the same repo but cleanly separated.
- **Separation of concerns**: `_common.go` files contain shared logic between resources and data sources.
- **Docker-based linting**: Makefile runs linters via Docker containers (editorconfig, shellcheck, yamllint, golangci-lint).

**Relevance to OPNsense**: Very high — pfSense and OPNsense share a common ancestor (pfSense fork). The apply/reload pattern and general architecture are directly applicable.

---

## 3. Pattern Comparison Matrix

| Pattern | browningluke/opnsense | fortios | panos | routeros | mikrotik | pfsense |
|---------|----------------------|---------|-------|----------|----------|---------|
| **Framework** | Plugin Framework v6 | SDKv2 | Plugin Framework v6 | SDKv2 | Mixed (mux) | Plugin Framework v6 |
| **API client location** | Separate repo | Separate repo | Separate repo | Inline | Inline | Inline |
| **Code generation** | YAML schemas + Go templates | Appears generated | Full codegen from XML specs | No | No | No |
| **Package structure** | `internal/service/{module}/` | Flat `fortios/` | `internal/provider/` + `internal/manager/` | Flat `routeros/` | `client/` + `mikrotik/` | `internal/provider/` + `pkg/pfsense/` |
| **Generic CRUD** | Go generics | Per-resource functions | Manager pattern | Metadata-driven | Generic client methods | Per-resource |
| **Apply/commit handling** | Auto-reconfigure in SDK | Direct API apply | Explicit commit resource | Immediate (REST) | Immediate (API) | Explicit apply resources |
| **Auth method** | Basic Auth (key+secret) | Bearer token | API key header | Username/password | Username/password | CSRF + session cookies |
| **Concurrency control** | Global mutex in SDK | None visible | Transactional multi-config | None visible | None visible | Feature-specific mutexes |
| **Testing** | QEMU VM in CI | Manual/Travis | Live device | Containerized RouterOS | Containerized RouterOS | Docker-based |
| **Doc generation** | tfplugindocs templates | Website | tfplugindocs | tfplugindocs | tfplugindocs | tfplugindocs |
| **Release** | GoReleaser + GPG | GoReleaser | GoReleaser | GoReleaser + semantic-release | GoReleaser | GoReleaser |

---

## 4. Patterns to ADOPT

### 4.1 Use Terraform Plugin Framework (not SDKv2)
All modern providers (browningluke/opnsense, panos, pfsense) use Plugin Framework v6. SDKv2 is in maintenance mode. HashiCorp explicitly recommends Plugin Framework for new providers. The ddelnano/mikrotik provider demonstrates the migration path using `terraform-plugin-mux` if needed.

### 4.2 Separate API Client Library
**browningluke/opnsense**, **panos**, and **fortios** all use separate API client libraries. Benefits:
- Independent versioning and testing.
- Reusable outside Terraform (CLI tools, other integrations).
- Cleaner dependency management.
- Can be developed and stabilized independently.

The `opnsense-go` library is the direct model: it provides a Go client for the OPNsense API with generic CRUD operations, schema-based code generation, and independent test suites.

### 4.3 Service-Based Package Organization
The browningluke/opnsense pattern of `internal/service/{module}/` is the cleanest approach:
- Each OPNsense module (firewall, unbound, wireguard) gets its own package.
- Each package has `exports.go` for registration, plus per-resource files for schema, resource, data source, and tests.
- Avoids the FortiOS anti-pattern of 4000+ files in one package.

### 4.4 Generic CRUD with Go Generics
The `opnsense-go` library's use of Go generics (`Add[K]`, `Get[K]`, `Update[K]`, `Delete[K]`) is elegant and type-safe. Combined with schema-driven code generation, this minimizes boilerplate.

### 4.5 Schema-Driven Code Generation
Multiple successful providers use this pattern:
- **opnsense-go**: YAML schema files + Go templates generate controller code.
- **pango/panos**: XML spec files generate both SDK and provider code.
- This should be considered for the API client layer at minimum.

### 4.6 Explicit Service Reconfiguration Handling
OPNsense requires `reconfigure` API calls after making changes. Two valid patterns exist:

**Pattern A — Automatic reconfiguration in the SDK (browningluke's approach)**:
- CRUD operations in the API client automatically call reconfigure after mutations.
- Protected by a global mutex to prevent concurrent reconfigurations.
- Simpler for users but can be slow (reconfigure after every single change).

**Pattern B — Explicit apply/commit resources (panos and pfsense approach)**:
- Separate Terraform resources for applying changes (e.g., `opnsense_firewall_apply`).
- Users control when changes are committed.
- Enables batching multiple changes before a single reconfigure.
- More Terraform-idiomatic for devices that have commit semantics.

**Recommendation**: Start with Pattern A (automatic) for simplicity, but design the SDK to support Pattern B in the future. The panos approach of an explicit commit action is the most sophisticated and handles the "make 10 firewall rule changes, then commit once" use case cleanly.

### 4.7 Reusable Schema Helpers / Property Factories
The RouterOS provider's `PropName()`, `PropEnabled()`, etc. pattern reduces boilerplate and ensures consistency across resources. Even with Plugin Framework (which has a different schema API), this pattern of factory functions for common attribute patterns is valuable.

### 4.8 Acceptance Testing Against Real Instances
All successful providers run acceptance tests against real device instances:
- **browningluke**: QEMU VM with OPNsense image in GitHub Actions.
- **routeros**: Containerized RouterOS in GitHub Actions.
- **panos**: Requires live device (environment variables).

The QEMU-based approach from browningluke is directly applicable.

### 4.9 GoReleaser + GPG Signing for Releases
Every provider uses GoReleaser for multi-platform binary builds and GitHub releases. GPG signing is required for Terraform Registry publication.

### 4.10 tfplugindocs for Documentation
All providers use `terraform-plugin-docs` with templates for generating registry documentation. The browningluke provider demonstrates this well with `templates/` directory containing `.md.tmpl` files.

### 4.11 Custom Validators
The browningluke provider has `internal/validators/` with reusable validators (IP/CIDR, UUID, numeric ranges). Plugin Framework has built-in validators, but custom ones are essential for domain-specific validation.

### 4.12 Error Handling with NotFound Distinction
The `opnsense-go` library has dedicated `errs.NotFound` error type. The CRUD pattern checks for this: if a resource is not found during Read, it removes the resource from Terraform state (rather than erroring). This is a critical Terraform pattern.

---

## 5. Patterns to AVOID

### 5.1 Flat Package Structure (FortiOS Anti-Pattern)
FortiOS has 4000+ files in a single `fortios/` package. This is unmaintainable, makes code discovery difficult, and leads to long compile times. Use service-based packages instead.

### 5.2 SDKv2 for New Providers
SDKv2 is in maintenance mode. Starting a new provider on SDKv2 means eventual migration pain. The ddelnano/mikrotik provider shows the awkward mux transition required.

### 5.3 Vendored Dependencies
FortiOS vendors all dependencies. Modern Go modules with `go.sum` are the standard. Vendoring adds repository bloat and makes updates harder.

### 5.4 HTML Scraping for API Communication
The pfSense provider must scrape HTML because pfSense lacks a proper API. OPNsense has a REST API — always use it. Never fall back to HTML scraping when an API exists.

### 5.5 No Concurrency Control
Providers that don't handle concurrent access to the appliance API can cause data corruption. The mutex patterns in browningluke/opnsense and marshallford/pfsense are essential.

### 5.6 Overly Large Resource Schemas
The FortiOS firewall policy resource has ~200 fields in a single schema. This is overwhelming for users. Where possible, decompose into smaller, focused resources.

### 5.7 Travis CI
Outdated CI platform. Use GitHub Actions.

---

## 6. Recommended Architecture for terraform-provider_opnsense

Based on this analysis, the recommended architecture combines the best patterns:

```
terraform-provider_opnsense/
├── main.go                          # Entry point (Plugin Framework v6)
├── internal/
│   ├── provider/
│   │   ├── provider.go              # Provider configuration and registration
│   │   └── factory.go               # ProtoV6ProviderServerFactory
│   ├── service/                     # Resources by OPNsense module
│   │   ├── firewall/
│   │   │   ├── exports.go           # Resource/DataSource registration
│   │   │   ├── alias_resource.go
│   │   │   ├── alias_schema.go
│   │   │   ├── alias_data_source.go
│   │   │   └── alias_resource_test.go
│   │   ├── unbound/
│   │   ├── dhcp/
│   │   └── ...
│   ├── validators/                  # Custom Plugin Framework validators
│   └── common/                      # Shared utilities, type converters
├── docs/                            # Generated by tfplugindocs
├── templates/                       # tfplugindocs templates
├── examples/                        # Example Terraform configurations
├── .github/
│   └── workflows/
│       ├── test.yml                 # QEMU-based acceptance tests
│       └── release.yml              # GoReleaser + GPG
├── .goreleaser.yml
├── .golangci.yml
└── GNUmakefile
```

**Separate API Client** (separate Go module/repo):
```
opnsense-go/  (or internal to the provider initially)
├── pkg/
│   ├── api/
│   │   ├── client.go                # HTTP client, auth, retry
│   │   ├── crud.go                  # Generic CRUD with Go generics
│   │   └── errors.go                # NotFound, etc.
│   ├── opnsense/
│   │   └── client.go                # High-level client
│   └── {module}/                    # Per-module controllers
│       ├── controller.go
│       └── {resource}.go
├── schema/                          # YAML resource definitions (optional)
└── internal/generate/               # Code generator (optional)
```

**Key decisions**:
1. Plugin Framework v6 (no SDKv2).
2. Separate API client (initially could be `pkg/` in the same repo, extracted later).
3. Service-based package organization mirroring OPNsense API modules.
4. Automatic reconfigure in SDK (Pattern A) initially, with clean abstractions to support explicit commit resources (Pattern B) later.
5. QEMU-based acceptance tests in CI.
6. GoReleaser + GPG for releases.
7. tfplugindocs for documentation.
8. Generic CRUD using Go generics in the API client.
9. Mutex-based concurrency control for API operations.

---

## 7. Source URLs

### Provider Repositories
- https://github.com/browningluke/terraform-provider-opnsense
- https://github.com/fortinetdev/terraform-provider-fortios
- https://github.com/PaloAltoNetworks/terraform-provider-panos
- https://github.com/terraform-routeros/terraform-provider-routeros
- https://github.com/ddelnano/terraform-provider-mikrotik
- https://github.com/marshallford/terraform-provider-pfsense

### API Client / SDK Repositories
- https://github.com/browningluke/opnsense-go (opnsense-go API client)
- https://github.com/fortinetdev/forti-sdk-go (FortiOS SDK)
- https://github.com/PaloAltoNetworks/pango (PAN-OS SDK)
- https://github.com/PaloAltoNetworks/pan-os-codegen (PAN-OS code generator)

### Terraform Registry
- https://registry.terraform.io/providers/browningluke/opnsense/latest/docs

### HashiCorp Documentation
- https://developer.hashicorp.com/terraform/plugin/best-practices
- https://developer.hashicorp.com/terraform/plugin/best-practices/provider-code
- https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles
- https://developer.hashicorp.com/terraform/plugin/framework

### Go Package Documentation
- https://pkg.go.dev/github.com/browningluke/opnsense-go/pkg/api
- https://pkg.go.dev/github.com/browningluke/opnsense-go/pkg/opnsense
- https://pkg.go.dev/github.com/terraform-routeros/terraform-provider-routeros/routeros
