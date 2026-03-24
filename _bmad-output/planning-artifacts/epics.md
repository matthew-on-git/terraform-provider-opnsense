---
stepsCompleted: [1, 2, 3, 4]
status: complete
inputDocuments:
  - "prd.md"
  - "architecture.md"
---

# terraform-provider-opnsense - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for terraform-provider-opnsense, decomposing the requirements from the PRD and Architecture into implementable stories aligned with the 9-tier implementation order.

## Requirements Inventory

### Functional Requirements

- FR1: Operator can configure the provider with OPNsense appliance URI, API key, and API secret
- FR2: Operator can configure credentials via environment variables (`OPNSENSE_URI`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET`) as an alternative to HCL configuration
- FR3: Operator can disable TLS certificate verification for self-signed certificates via `insecure` attribute or `OPNSENSE_ALLOW_INSECURE` environment variable
- FR4: Provider validates credentials during configuration by making a test API call and fails fast with a clear diagnostic if authentication fails
- FR5: Provider detects the running OPNsense version during configuration and logs it for diagnostics
- FR6: Operator can create any supported OPNsense resource via `terraform apply`
- FR7: Operator can read the current state of any supported resource via `terraform plan` (refresh)
- FR8: Operator can update any supported resource in-place via `terraform apply` when attributes change
- FR9: Operator can delete any supported resource via `terraform apply` when the resource block is removed
- FR10: Operator can import any existing OPNsense resource into Terraform state via `terraform import` using its UUID
- FR11: Provider detects drift on managed resources — out-of-band changes made via OPNsense UI or API are shown in the next `terraform plan`
- FR12: Provider applies OPNsense service reconfigure after every create, update, and delete operation, routed to the affected service module's endpoint
- FR13: Provider uses the firewall filter savepoint/apply/cancelRollback mechanism for firewall rule changes to prevent operator lockout
- FR14: Provider serializes all mutation operations through a global mutex to prevent concurrent write conflicts
- FR15: Provider populates state from API read-back after every create and update — never echoes request config into state
- FR16: Provider plan output correctly signals update-in-place vs. destroy-and-recreate for each resource change, using `RequiresReplace` only on truly immutable fields
- FR17: Provider leaves state consistent after partial apply failure — successfully created/updated resources remain in state, failed resources do not
- FR18: Provider handles reconfigure failure by reporting a clear diagnostic
- FR19: Operator can manage firewall aliases (host, network, port, URL table types) with content lists
- FR20: Operator can manage firewall categories for organizing rules
- FR21: Operator can manage firewall filter rules with source/destination, ports, protocols, interfaces, and action
- FR22: Operator can manage firewall NAT port-forward rules with target address, ports, and interface
- FR23: Operator can manage firewall NAT outbound rules with source, translation, and interface
- FR24: Operator can manage HAProxy servers with address, port, weight, SSL, and health check configuration
- FR25: Operator can manage HAProxy backends with linked servers, load balancing algorithm, health check settings, and persistence
- FR26: Operator can manage HAProxy frontends with bind addresses, default backend, SSL offloading, and linked ACLs
- FR27: Operator can manage HAProxy ACL rules with match conditions for frontend routing
- FR28: Operator can manage HAProxy health checks with type, interval, and threshold configuration
- FR29: Operator can reference HAProxy resources across types via UUID (server → backend → frontend chain)
- FR30: Operator can manage FRR general settings (enable/disable FRR service, routing profile)
- FR31: Operator can manage BGP global configuration (AS number, router ID, network advertisements)
- FR32: Operator can manage BGP neighbors with remote ASN, IP address, timers, and update-source
- FR33: Operator can manage BGP prefix lists with sequence numbers, action, and network prefixes
- FR34: Operator can manage BGP route maps with match conditions and set actions
- FR35: Operator can manage ACME accounts with email, CA server, and registration
- FR36: Operator can manage ACME certificate configurations with domain, alternative names, challenge type, and auto-renewal settings
- FR37: Operator can trigger ACME certificate issuance (sign) as part of resource creation
- FR38: Operator can revoke an ACME certificate by deleting the certificate resource
- FR39: Certificate renewal is owned by OPNsense via its built-in cron — the provider manages certificate configuration, not renewal lifecycle
- FR40: Operator can manage network interfaces and their configuration
- FR41: Operator can manage VLAN assignments on interfaces
- FR42: Operator can manage virtual IPs (CARP, IP Alias) on interfaces
- FR43: Operator can manage static routes with destination network, gateway, and metric
- FR44: Operator can manage gateways with interface, address, and monitoring settings
- FR45: Operator can manage gateway groups with priority and trigger level settings
- FR46: Operator can manage system general settings (hostname, domain, DNS servers, NTP)
- FR47: Operator can manage Unbound host overrides with hostname, domain, and IP address
- FR48: Operator can manage Unbound domain overrides with domain and forwarding server
- FR49: Operator can manage Unbound access control lists
- FR50: Operator can manage WireGuard server instances with listen port, private key, and interface settings
- FR51: Operator can manage WireGuard peers with public key, allowed IPs, endpoint, and keepalive
- FR52: Operator can manage IPsec Phase 1 connections with authentication, encryption, and remote gateway
- FR53: Operator can manage IPsec Phase 2 tunnels with local/remote networks and encryption settings
- FR54: Operator can manage IPsec pre-shared keys
- FR55: Operator can manage DHCPv4 pools with network, range, gateway, and DNS settings
- FR56: Operator can manage DHCPv4 static mappings with MAC address and fixed IP
- FR57: Operator can manage DHCPv4 options (including PXE boot options 66, 67, 150)
- FR58: Operator can manage Dynamic DNS accounts with provider, hostname, and credentials
- FR59: Operator can manage Dynamic DNS provider configuration
- FR60: Every resource type has a corresponding read-only data source for reference in other configurations
- FR61: Operator can query OPNsense system information (firmware version, installed plugins) as a data source
- FR62: Provider surfaces OPNsense API validation errors as Terraform diagnostics with field-level detail
- FR63: Provider detects and reports missing resources (deleted out-of-band) by removing them from state during refresh
- FR64: Provider reports clear errors when OPNsense API is unreachable, credentials are invalid, or required plugins are not installed
- FR65: Provider reports permission-specific errors when the API key lacks required privileges for a module
- FR66: Every resource type ships with auto-generated Registry documentation including argument reference, attribute reference, and import instructions
- FR67: Documentation includes multi-resource composition examples showing interconnected resources in context
- FR68: Provider index page documents configuration, authentication options, minimum OPNsense version, required API user permissions per module, and a complete quickstart example

### NonFunctional Requirements

- NFR1: `terraform plan` (full refresh) completes within 60 seconds for 50-100 managed resources
- NFR2: Individual resource CRUD operations complete within 10 seconds including reconfigure
- NFR3: `terraform import` for a single resource completes within 5 seconds
- NFR4: Provider binary startup completes within 3 seconds
- NFR5: Provider uses connection pooling / HTTP keep-alive to minimize TCP handshake overhead
- NFR6: Provider limits concurrent API read operations to prevent overwhelming OPNsense PHP-FPM
- NFR7: API credentials marked `Sensitive: true` are never displayed in plan output, apply output, or provider logs
- NFR8: Provider uses Terraform Plugin Framework write-only attributes for credential fields where available
- NFR9: Provider supports HTTPS with configurable TLS verification — default is TLS-verified, `insecure` mode requires explicit opt-in
- NFR10: Provider never logs request or response bodies containing credential fields at any log level
- NFR11: Provider documentation warns that state files may contain sensitive values and recommends encrypted remote backends
- NFR12: CI/CD pipeline credentials use environment variables or CI-native secret injection
- NFR13: All resource operations are idempotent — running `terraform apply` twice produces "No changes" on the second run
- NFR14: Provider never corrupts Terraform state — partial failures leave state in a recoverable condition
- NFR15: Provider handles OPNsense API transient failures with automatic retry via configurable backoff
- NFR16: Provider handles OPNsense appliance reboot gracefully — surfaces as retryable error, not state corruption
- NFR17: Provider schema version upgrades work seamlessly — state migration runs automatically
- NFR18: All acceptance tests are deterministic — no flaky tests
- NFR19: Provider is compatible with Terraform CLI 1.0+ (Protocol v6)
- NFR20: Provider is compatible with OpenTofu for core provider protocol operations
- NFR21: Provider works with the GitLab HTTP state backend for remote state with locking
- NFR22: Provider binary is statically linked (`CGO_ENABLED=0`) and runs without external dependencies
- NFR23: Provider is compatible with Terraform Cloud and Terraform Enterprise
- NFR24: All code passes `make check` — golangci-lint, formatting, security scanning, and tests
- NFR25: Acceptance test coverage: >80% of resource types have acceptance tests
- NFR26: Adding a new resource type follows a documented, repeatable four-file pattern
- NFR27: Go code follows standard Go conventions enforced by CI
- NFR28: Go module hygiene enforced — `go mod tidy` clean, no `replace` directives in releases
- NFR29: Provider releases follow semantic versioning — breaking changes only on major version bumps
- NFR30: CHANGELOG maintained in standard Terraform provider format
- NFR31: Error messages include resource type, operation, API response, and suggested action
- NFR32: Validation errors include specific field name and constraint violated
- NFR33: Permission errors identify which OPNsense privilege group is required
- NFR34: Connection errors distinguish between DNS, TLS, auth, and endpoint-not-found failures

### Additional Requirements (from Architecture)

- AR1: Project initialized from HashiCorp terraform-provider-scaffolding-framework (Go 1.25.0, Framework v1.19.0, Protocol v6.0)
- AR2: Separate API client package at `pkg/opnsense/` — independent of Terraform types, owns global mutex
- AR3: Generic CRUD functions using Go generics: `Add[K]`, `Get[K]`, `Update[K]`, `Delete`, `Search[K]` with `ReqOpts` config
- AR4: Custom HTTP transport (`apiKeyTransport`) injecting Basic Auth on every request via `go-retryablehttp`
- AR5: Global MutexKV protecting config.xml integrity — all mutations serialized, reads parallel
- AR6: Read concurrency limited via `semaphore.Weighted` (configurable, default 10) to protect PHP-FPM
- AR7: Transparent API pagination in Search/List operations — client iterates all pages internally
- AR8: Five custom error types: `NotFoundError`, `ValidationError`, `AuthError`, `ServerError`, `PluginNotFoundError`
- AR9: Three-layer type conversion: OPNsense API types ↔ Go model types ↔ Terraform Framework types
- AR10: Code generation via `text/template` from YAML schemas — but hand-write first 2 modules (firewall, haproxy) before building codegen pipeline
- AR11: Four-file resource pattern: `{resource}_resource.go`, `{resource}_schema.go`, `{resource}_model.go`, `{resource}_resource_test.go` plus `exports.go`
- AR12: Firewall filter uses `ReconfigureFunc` (savepoint/apply/cancelRollback flow), all other resources use `ReconfigureEndpoint` string
- AR13: Vagrant locally for dev testing, QEMU in GitHub Actions for CI — test framework agnostic via env vars
- AR14: GitHub-primary CI/CD: lint+unit (every push), acceptance (PR to main), release (v* tag via GoReleaser+GPG)
- AR15: Plugin detection — report clear error with plugin name when OPNsense plugin is not installed
- AR16: CI structural check validates service directory four-file pattern matches exports.go
- AR17: `dev_overrides` in `~/.terraformrc` for local build-test cycle without publishing

### UX Design Requirements

N/A — CLI tool with no user interface. Users interact via HCL configuration files, `terraform` CLI commands, and plan/apply output.

### FR Coverage Map

| FR | Epic | Description |
|---|---|---|
| FR1-FR5 | Epic 1 | Provider configuration & authentication |
| FR6-FR12, FR15-FR18 | Epic 2 | Cross-cutting lifecycle (CRUD, import, drift, reconfigure, state, partial failure) |
| FR13 | Epic 3 | Firewall savepoint/rollback |
| FR14 | Epic 1 | Global mutex |
| FR19 | Epic 2 | Firewall aliases (first resource) |
| FR20-FR23 | Epic 3 | Firewall categories, rules, NAT |
| FR24 | Epic 2 | HAProxy servers (second resource) |
| FR25-FR29 | Epic 4 | HAProxy backends, frontends, ACLs, health checks |
| FR30-FR34 | Epic 7 | FRR general, BGP config, neighbors, prefix lists, route maps |
| FR35-FR39 | Epic 8 | ACME accounts, certificates, challenge, issuance |
| FR40-FR46 | Epic 5 | Interfaces, VLANs, virtual IPs, routes, gateways, system |
| FR47-FR49 | Epic 9 | Unbound host overrides, domain overrides, ACLs |
| FR50-FR54 | Epic 10 | WireGuard servers/peers, IPsec phase 1/phase 2, PSKs |
| FR55-FR57 | Epic 11 | DHCPv4 pools, static mappings, options |
| FR58-FR59 | Epic 9 | Dynamic DNS accounts, provider config |
| FR60 (alias only) | Epic 2 | First data source (validates pattern) |
| FR60-FR61 (remaining) | Epic 12 | All remaining data sources + system info |
| FR62-FR65 | Epic 1 | Error handling |
| FR66-FR68 | Epic 12 | Documentation, composition examples, provider index |

**Coverage: 68/68 FRs mapped (100%)**

## Epic List

### Epic 1: Provider Scaffold & API Client
Developer can initialize the provider project from the HashiCorp scaffold, configure OPNsense connection, and establish the API client with authentication, error handling, and mutex protection. Vagrant test environment is operational.
**FRs covered:** FR1-FR5, FR14, FR62-FR65
**ARs covered:** AR1-AR9, AR13, AR17
**NFRs addressed:** NFR4-6, NFR7-12, NFR15, NFR19, NFR22-24, NFR27-28, NFR31-34

### Epic 2: First Resources & Lifecycle Validation
Operator can manage firewall aliases and HAProxy servers with full CRUD, import, drift detection, and reconfigure — validating the complete Terraform lifecycle and proving the resource implementation pattern.
**FRs covered:** FR6-FR12, FR15-FR18, FR19, FR24, FR60 (alias data source only)
**ARs covered:** AR11
**NFRs addressed:** NFR1-3, NFR13-14, NFR25-26

### Epic 3: Firewall Management
Operator can manage their complete firewall configuration — categories, filter rules with savepoint rollback protection, and NAT rules (port forward and outbound) — through Terraform.
**FRs covered:** FR13, FR20-FR23
**ARs covered:** AR12

### Epic 4: HAProxy Load Balancing
Operator can manage complete HAProxy setups for customer onboarding — backends linked to servers, frontends with SNI-based ACL routing, and health checks — enabling the hot-path workflow of adding new customer domains.
**FRs covered:** FR25-FR29

### Epic 5: Core Infrastructure
Operator can manage foundational network infrastructure — interfaces, VLANs, virtual IPs (CARP/IP Alias), static routes, gateways, gateway groups, and system settings — completing the core appliance configuration.
**FRs covered:** FR40-FR46

### Epic 6: Code Generation Pipeline
Developer can generate API client code from YAML schemas, accelerating development of all subsequent resource modules. Pattern extracted from hand-written firewall and HAProxy modules. Non-blocking — can run in parallel with Epic 5.
**ARs covered:** AR10

### Epic 7: BGP Dynamic Routing
Operator can manage FRR/BGP configuration for dynamic routing — general settings, BGP global config, neighbors, prefix lists, and route maps — enabling MetalLB peering and route advertisement.
**FRs covered:** FR30-FR34

### Epic 8: ACME Certificate Management
Operator can manage ACME certificate lifecycle — accounts, certificates with challenge configuration, and automated issuance — enabling SSL termination for HAProxy frontends.
**FRs covered:** FR35-FR39

### Epic 9: DNS Management
Operator can manage DNS infrastructure — Unbound host overrides, domain overrides, access control lists, Dynamic DNS accounts, and provider configuration — through Terraform.
**FRs covered:** FR47-FR49, FR58-FR59

### Epic 10: VPN Management
Operator can manage VPN tunnels — WireGuard server instances and peers, IPsec Phase 1 connections, Phase 2 tunnels, and pre-shared keys — through Terraform.
**FRs covered:** FR50-FR54

### Epic 11: DHCP Management
Operator can manage DHCPv4 — pools, static mappings, and DHCP options including PXE boot configuration — through Terraform.
**FRs covered:** FR55-FR57

### Epic 12: Data Sources, Documentation & Registry Release
Operator can discover the provider on the Terraform Registry, find comprehensive documentation with composition examples, use data sources to reference all existing resources, and install the provider via `terraform init`.
**FRs covered:** FR60-FR61 (remaining data sources), FR66-FR68
**ARs covered:** AR14, AR16
**NFRs addressed:** NFR29-30

### Epic Dependency Flow

```
Epic 1 (Scaffold + API Client)
  → Epic 2 (First Resources + Lifecycle)
    → Epic 3 (Firewall) + Epic 4 (HAProxy) [parallel]
      → Epic 5 (Core Infra) + Epic 6 (Code Gen) [parallel, non-blocking]
        → Epics 7-11 (remaining resources, any order)
          → Epic 12 (Data Sources, Docs, Release)
```

---

## Epic 1: Provider Scaffold & API Client

Developer can initialize the provider project from the HashiCorp scaffold, configure OPNsense connection, and establish the API client with authentication, error handling, and mutex protection. Vagrant test environment is operational.

### Story 1.1: Initialize Project and API Client Core

As a developer,
I want to initialize the terraform-provider-opnsense project from the HashiCorp scaffold with a working API client that authenticates with OPNsense,
So that I have a buildable project that can communicate with a real OPNsense appliance.

**Acceptance Criteria:**

**Given** the HashiCorp terraform-provider-scaffolding-framework repository is available
**When** the developer clones the scaffold and rebrands to `github.com/matthew-on-git/terraform-provider-opnsense`
**Then** `go build ./...` succeeds and `golangci-lint run` passes
**And** `main.go` serves the provider at address `registry.terraform.io/matthew-on-git/opnsense`
**And** `terraform-registry-manifest.json` declares Protocol v6.0
**And** `pkg/opnsense/client.go` creates an HTTP client using `go-retryablehttp` with configurable retry count and backoff
**And** a custom `apiKeyTransport` `RoundTripper` injects HTTP Basic Auth on every request
**And** HTTP keep-alive is enabled and TLS verification is configurable via `insecure` option
**And** unit tests verify authentication, TLS configuration, and retry behavior
**And** `make check` passes

### Story 1.2: Provider Configuration with Credential Validation

As an operator,
I want to configure the provider with my OPNsense URI, API key, and API secret (via HCL or environment variables),
So that Terraform can connect to my OPNsense appliance and fail clearly if credentials are wrong.

**Acceptance Criteria:**

**Given** the operator provides `uri`, `api_key`, and `api_secret` in the provider block
**When** `terraform plan` is executed
**Then** the provider `Configure` method initializes the API client and validates credentials by calling `/api/core/firmware/status`
**And** if credentials are valid, the provider logs the detected OPNsense version
**And** if credentials are invalid, the provider returns a clear diagnostic: "Authentication failed — verify API key and secret"
**And** environment variables (`OPNSENSE_URI`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET`, `OPNSENSE_ALLOW_INSECURE`) are accepted as fallback
**And** HCL configuration takes priority over environment variables
**And** `api_key` and `api_secret` are marked `Sensitive: true`
**And** the configured `*opnsense.Client` is set on `resp.ResourceData` and `resp.DataSourceData` so resources can access it via their `Configure` method

### Story 1.3: Global Mutex and Reconfigure Infrastructure

As a developer,
I want a global mutex that serializes all write operations and a reconfigure dispatch mechanism that applies changes after mutations,
So that the provider protects OPNsense's config.xml integrity and activates changes correctly.

**Acceptance Criteria:**

**Given** the API client is initialized
**When** multiple CRUD operations are dispatched concurrently by Terraform
**Then** all Create, Update, and Delete operations are serialized through a global mutex (single key)
**And** Read operations are NOT blocked by the mutex and execute in parallel
**And** read concurrency is limited by a configurable semaphore (default: 10 concurrent reads)
**And** after every successful mutation, the client checks `ReqOpts`: if `ReconfigureFunc` is set, it is called; otherwise the standard `ReconfigureEndpoint` is called via POST
**And** the `ReconfigureFunc` interface is defined but no concrete implementation exists yet (firewall savepoint logic comes in Epic 3)
**And** unit tests verify mutex serialization, semaphore limiting, and reconfigure dispatch for both standard and function-based paths

### Story 1.4: Custom Error Types and Response Parsing

As a developer,
I want custom error types that parse OPNsense's non-standard API responses into structured errors,
So that resources can handle validation failures, missing resources, and permission errors correctly.

**Acceptance Criteria:**

**Given** the API client makes a request to the OPNsense API
**When** the response indicates a validation error (`result != "saved"` with HTTP 200)
**Then** a `ValidationError` is returned containing the field names and error messages from the `validations` map
**And** when the response indicates a missing resource (blank/default record or JSON unmarshal type error), a `NotFoundError` is returned
**And** when the response is HTTP 401 or 403, an `AuthError` is returned with the status code
**And** when the response is HTTP 404 on a plugin endpoint, a `PluginNotFoundError` is returned with the plugin name
**And** when all retries are exhausted on HTTP 500+/timeout, a `ServerError` is returned
**And** unit tests verify each error type is correctly parsed from sample API responses

### Story 1.5: Generic CRUD Functions

As a developer,
I want generic CRUD functions (`Add[K]`, `Get[K]`, `Update[K]`, `Delete`) that work with any resource type via `ReqOpts` configuration,
So that each resource module only needs to define its struct and endpoint config, not HTTP logic.

**Acceptance Criteria:**

**Given** a resource struct type `K` and a `ReqOpts` config with endpoints and monad
**When** `Add[K]` is called with a resource struct
**Then** the body is wrapped in the monad key (`{"server": {...}}`), sent as POST to the add endpoint, and the UUID is returned from the response
**And** `Get[K]` fetches by UUID, unwraps the monad, and returns a clean struct (or `NotFoundError`)
**And** `Update[K]` wraps and sends to the update endpoint by UUID
**And** `Delete` sends to the delete endpoint by UUID
**And** all mutation functions acquire the global mutex and call reconfigure after success
**And** all functions propagate `context.Context` for cancellation
**And** unit tests use a `testResource` struct with a mock HTTP server — real OPNsense resources are validated in Epic 2

### Story 1.6: Search with Pagination

As a developer,
I want a `Search[K]` function that transparently iterates paginated OPNsense search results,
So that data sources and list operations return complete results regardless of page size.

**Acceptance Criteria:**

**Given** an OPNsense search endpoint that returns paginated results with `rowCount` and `current` parameters
**When** `Search[K]` is called with search parameters
**Then** the function iterates all pages until all results are collected
**And** the function handles both search response format (list of items) and get response format (single item)
**And** read concurrency semaphore is respected (search is a read operation)
**And** unit tests verify pagination with multi-page mock responses and edge cases (empty results, single page, exactly full page)

### Story 1.7: Type Conversion Utilities

As a developer,
I want shared type conversion utilities for OPNsense's non-standard types (string bools, SelectedMap, CSVList),
So that every resource can convert between OPNsense API types, Go model types, and Terraform types consistently.

**Acceptance Criteria:**

**Given** OPNsense returns `"0"` / `"1"` for boolean values
**When** `StringToBool("1")` is called
**Then** it returns `true`, and `BoolToString(true)` returns `"1"`
**And** `SelectedMap` type correctly unmarshals `{"key1": {"value": "...", "selected": 1}}` and extracts the selected key
**And** `SelectedMapList` returns `[]string` of all selected keys from multi-select fields
**And** CSV string splitting/joining is handled for CSVList fields
**And** integer-as-string conversion (`"443"` ↔ `int64`) is handled
**And** unit tests cover all type conversions including edge cases (empty strings, missing keys, malformed JSON)

### Story 1.8: Vagrant Test Environment

*Parallelizable — can be done alongside any other Epic 1 story.*

As a developer,
I want a Vagrantfile that provisions an ephemeral OPNsense VM with API access,
So that I can run acceptance tests locally without risking my production appliance.

**Acceptance Criteria:**

**Given** the developer has Vagrant installed
**When** `vagrant up` is run from the `test/` directory
**Then** an OPNsense VM boots and is accessible via HTTPS on a forwarded port
**And** an API key and secret are generated and output to the console
**And** the developer can set `OPNSENSE_URI`, `OPNSENSE_API_KEY`, and `OPNSENSE_API_SECRET` from the output
**And** `vagrant destroy` cleanly removes the VM
**And** the Vagrantfile and test setup are documented in CONTRIBUTING.md

### Story 1.9: Basic CI Pipeline

As a developer,
I want GitHub Actions running lint, build, and unit tests on every push,
So that regressions are caught immediately from day one.

**Acceptance Criteria:**

**Given** code is pushed to the repository or a PR is opened
**When** the CI workflow triggers
**Then** `golangci-lint run` passes
**And** `go build ./...` succeeds
**And** `go test ./pkg/... ./internal/...` passes (unit tests only, no TF_ACC)
**And** `go generate ./...` followed by `git diff --exit-code` confirms no uncommitted generated code changes
**And** the workflow is defined in `.github/workflows/ci.yml`

---

## Epic 2: First Resources & Lifecycle Validation

Operator can manage firewall aliases and HAProxy servers with full CRUD, import, drift detection, and reconfigure — validating the complete Terraform lifecycle and proving the resource implementation pattern.

### Story 2.1: Firewall Alias Resource (First Resource)

As an operator,
I want to manage OPNsense firewall aliases through Terraform,
So that I can define host, network, and port aliases as code and see changes before applying.

**Acceptance Criteria:**

**Given** the provider is configured with valid OPNsense credentials
**When** the operator defines an `opnsense_firewall_alias` resource in HCL
**Then** `terraform apply` creates the alias on OPNsense via the API and returns the UUID
**And** `terraform plan` with no changes shows "No changes" (state read-back matches)
**And** modifying alias content in HCL shows the correct diff in `terraform plan` and applies the change
**And** removing the resource block deletes the alias from OPNsense
**And** `terraform import opnsense_firewall_alias.test <uuid>` imports an existing alias into state
**And** after import, `terraform plan` shows "No changes"
**And** out-of-band changes to the alias via OPNsense UI are detected in the next `terraform plan`
**And** state is populated from API read-back after Create and Update (never echoed from config)
**And** resource follows four-file pattern: `alias_resource.go`, `alias_schema.go`, `alias_model.go`, `alias_resource_test.go`
**And** acceptance test covers full lifecycle: create → verify → import → update → verify → destroy
**And** `internal/acctest/acctest.go` provides `ProtoV6ProviderFactories` and `PreCheck(t)` that validates env vars and OPNsense reachability
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 2.2: HAProxy Server Resource (Plugin API Validation)

As an operator,
I want to manage OPNsense HAProxy servers through Terraform,
So that I can define backend server targets as code and validate that the provider works with plugin APIs.

**Acceptance Criteria:**

**Given** the OPNsense appliance has the `os-haproxy` plugin installed
**When** the operator defines an `opnsense_haproxy_server` resource in HCL
**Then** the same full CRUD + import + drift detection lifecycle works as with firewall aliases
**And** if the HAProxy plugin is not installed, the provider returns a `PluginNotFoundError` with a clear message to install `os-haproxy`
**And** the resource schema includes server-specific attributes: name, address, port, weight, ssl, enabled
**And** boolean attributes convert correctly between Terraform `types.Bool` and OPNsense `"0"`/`"1"` strings
**And** acceptance test covers full lifecycle: create → verify → import → update → verify → destroy
**And** this validates that the API client pattern works for both core and plugin APIs
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 2.3: Firewall Alias Data Source (First Data Source)

As an operator,
I want to look up existing OPNsense firewall aliases as a data source,
So that I can reference aliases in other resource configurations without importing them.

**Acceptance Criteria:**

**Given** a firewall alias exists on OPNsense
**When** the operator defines a `data.opnsense_firewall_alias` block with the alias UUID
**Then** the data source reads the alias attributes from the API and makes them available for reference
**And** the data source follows the same `fromAPI()` conversion as the resource
**And** the data source is read-only (no Create, Update, Delete, or ImportState)
**And** acceptance test verifies data source reads match the resource state

---

## Epic 3: Firewall Management

Operator can manage their complete firewall configuration — categories, filter rules with savepoint rollback protection, and NAT rules (port forward and outbound) — through Terraform.

### Story 3.1: Firewall Savepoint Reconfigure Implementation

As a developer,
I want the firewall filter `ReconfigureFunc` to implement OPNsense's savepoint/apply/cancelRollback 3-step flow,
So that firewall rule changes are protected by automatic 60-second rollback if connectivity is lost.

**Acceptance Criteria:**

**Given** a firewall filter resource mutation (Create/Update/Delete) succeeds
**When** the reconfigure function is invoked
**Then** it calls `POST /api/firewall/filter/savepoint` to get a revision ID
**And** it calls `POST /api/firewall/filter/apply/{revision}` to apply with rollback safety
**And** it calls `POST /api/firewall/filter/cancelRollback/{revision}` to confirm (prevent auto-revert)
**And** if any step fails, the error is surfaced as a Terraform diagnostic
**And** if `cancelRollback` is not called within 60 seconds, OPNsense automatically reverts the change
**And** unit tests verify the 3-step sequence with mock HTTP responses

### Story 3.2: Firewall Category Resource

As an operator,
I want to manage firewall categories through Terraform,
So that I can organize my firewall rules into logical groups.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_firewall_category` resources
**Then** categories are created, updated, deleted, and importable via UUID
**And** categories use the standard `ReconfigureEndpoint` (not the savepoint flow — categories are not filter rules)
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 3.3: Firewall Filter Rule Resource

As an operator,
I want to manage firewall filter rules through Terraform with savepoint rollback protection,
So that I can safely modify firewall rules knowing that bad changes auto-revert within 60 seconds.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_firewall_filter_rule` resources
**Then** rules are created, updated, and deleted using the savepoint `ReconfigureFunc` (not standard reconfigure)
**And** the schema includes: source, destination, ports, protocol, interface, action (pass/block/reject), enabled, category reference
**And** `RequiresReplace` is NOT used on any attribute — all changes are update-in-place
**And** acceptance test verifies the full lifecycle including that reconfigure uses the savepoint flow
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 3.4: Firewall NAT Port Forward Resource

As an operator,
I want to manage NAT port-forward rules through Terraform,
So that I can expose internal services through the firewall as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_firewall_nat_port_forward` resources
**Then** NAT port-forward rules support target address, port, interface, and protocol configuration
**And** rules use the standard `ReconfigureEndpoint` (NAT has its own reconfigure endpoint, not the savepoint flow)
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 3.5: Firewall NAT Outbound Resource

As an operator,
I want to manage NAT outbound rules through Terraform,
So that I can control source NAT for outbound traffic as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_firewall_nat_outbound` resources
**Then** NAT outbound rules support source, translation, and interface configuration
**And** rules use the standard `ReconfigureEndpoint` (NAT has its own reconfigure endpoint, not the savepoint flow)
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

---

## Epic 4: HAProxy Load Balancing

Operator can manage complete HAProxy setups for customer onboarding — backends linked to servers, frontends with SNI-based ACL routing, and health checks — enabling the hot-path workflow of adding new customer domains.

### Story 4.1: HAProxy Backend Resource with Server Linking

As an operator,
I want to manage HAProxy backends that link to servers via UUID references,
So that I can define load balancer pools pointing to my backend servers.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_haproxy_backend` with `server_ids` referencing existing server UUIDs
**Then** the backend is created with linked servers using `types.Set` of `types.String` with UUID validation
**And** the schema includes: name, load balancing algorithm (roundrobin, leastconn, source), health check settings, persistence mode
**And** Terraform's dependency graph correctly orders server creation before backend creation
**And** acceptance test creates servers first, then a backend linking to them
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 4.2: HAProxy Frontend Resource with ACL Routing

As an operator,
I want to manage HAProxy frontends with ACL-based domain routing,
So that I can route incoming HTTPS traffic to the correct backend based on hostname.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_haproxy_frontend` with bind addresses, default backend, and linked ACLs
**Then** the frontend supports SSL offloading configuration and SNI-based routing
**And** modifying the ACL list is an in-place update, not a destroy-and-recreate
**And** the schema includes: bind address, bind port, default backend UUID, SSL offload settings, linked ACL UUIDs
**And** acceptance test creates server → backend → frontend chain and verifies full lifecycle
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 4.3: HAProxy ACL Resource

As an operator,
I want to manage HAProxy ACL rules for domain-based routing,
So that I can direct traffic to specific backends based on host headers, paths, or SNI.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_haproxy_acl` resources
**Then** ACLs support match conditions: host header (`hdr(host)`), path, SNI
**And** ACLs reference backends via UUID for routing decisions
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 4.4: HAProxy Health Check Resource

As an operator,
I want to manage HAProxy health checks,
So that I can configure how backends verify server availability.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_haproxy_healthcheck` resources
**Then** health checks support type (TCP, HTTP), interval, threshold, and HTTP path configuration
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

---

## Epic 5: Core Infrastructure

Operator can manage foundational network infrastructure — interfaces, VLANs, virtual IPs, static routes, gateways, gateway groups, and system settings — completing the core appliance configuration.

### Story 5.1: Interface Resource

As an operator,
I want to manage network interface configurations through Terraform,
So that I can define interface settings (enabled state, description, IP assignment) as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_system_interface` resources
**Then** interface configuration including enabled state, description, and IP settings is manageable
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 5.2: VLAN Resource

As an operator,
I want to manage VLAN assignments through Terraform,
So that I can define network segmentation as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_system_vlan` resources
**Then** VLANs can be assigned to interfaces with VLAN tag and description
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 5.3: Virtual IP Resource

As an operator,
I want to manage virtual IPs (CARP, IP Alias) through Terraform,
So that I can define high-availability and additional interface addresses as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_system_vip` resources
**Then** virtual IPs support CARP and IP Alias types with interface, address, and VHID configuration
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 5.4: Static Route Resource

As an operator,
I want to manage static routes through Terraform,
So that I can define routing tables as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_system_route` resources
**Then** routes support destination network, gateway reference, and metric configuration
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 5.5: Gateway Resource

As an operator,
I want to manage gateways through Terraform,
So that I can define network gateways with monitoring as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_system_gateway` resources
**Then** gateways support interface, address, and monitoring settings
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 5.6: Gateway Group Resource

As an operator,
I want to manage gateway groups through Terraform,
So that I can define failover and load-balanced gateway configurations as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_system_gateway_group` resources
**Then** gateway groups support priority and trigger level settings with member gateway references
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 5.7: System General Settings Resource

As an operator,
I want to manage system general settings through Terraform,
So that I can define hostname, domain, DNS servers, and NTP configuration as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_system_general` resource
**Then** system settings including hostname, domain, DNS servers, and NTP servers are manageable
**And** this is a singleton resource — no Create or Delete operations. Only Read, Update, and ImportState are implemented. The resource represents the single system configuration that always exists on the appliance.
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

---

## Epic 6: Code Generation Pipeline

Developer can generate API client code from YAML schemas, accelerating development of all subsequent resource modules. Pattern extracted from hand-written firewall and HAProxy modules.

### Story 6.1: YAML Schema Format Definition

As a developer,
I want a documented YAML schema format for defining OPNsense API modules,
So that each module's endpoints, resource structs, and attributes are declaratively defined.

**Acceptance Criteria:**

**Given** the hand-written firewall and HAProxy API client modules exist
**When** the developer creates `schema/firewall.yml` and `schema/haproxy.yml` that describe the existing modules
**Then** the YAML schemas capture: module name, reconfigure endpoint, and for each resource: name, monad, endpoints (add/get/update/delete/search), and attributes (name, type, JSON key)
**And** the YAML accurately represents what was hand-written in `pkg/opnsense/firewall/` and `pkg/opnsense/haproxy/`
**And** the schema format is documented in a README within the `schema/` directory

### Story 6.2: Code Generation Pipeline with text/template

As a developer,
I want a `go generate` pipeline that produces Go API client code from YAML schemas,
So that adding a new OPNsense module requires only writing a YAML schema file.

**Acceptance Criteria:**

**Given** YAML schema files exist in `schema/`
**When** `go generate ./...` is run
**Then** Go source files are generated in `pkg/opnsense/{module}/` containing: struct definitions with JSON tags, `ReqOpts` configuration, typed CRUD method wrappers
**And** the generated code for firewall and HAProxy matches the behavior of the hand-written versions
**And** the generator lives in `internal/generate/` with templates in `internal/generate/templates/`
**And** generated files include a `// Code generated ... DO NOT EDIT.` header
**And** `make check` validates that generated code compiles and passes lint

### Story 6.3: Generate Remaining Module Skeletons

As a developer,
I want YAML schemas for all remaining modules (quagga, acme, system, unbound, wireguard, ipsec, dhcpv4, ddclient),
So that their API client code is generated and ready for Terraform resource implementation.

**Acceptance Criteria:**

**Given** the code generation pipeline works for firewall and HAProxy
**When** YAML schemas are created for all 8 remaining modules
**Then** `go generate ./...` produces API client code for all 10 modules
**And** all generated code compiles and passes lint
**And** each module has a `generate.go` file with the `go:generate` directive

---

## Epic 7: BGP Dynamic Routing

Operator can manage FRR/BGP configuration for dynamic routing — general settings, BGP global config, neighbors, prefix lists, and route maps — enabling MetalLB peering and route advertisement.

### Story 7.1: FRR General Settings Resource

As an operator,
I want to manage FRR general settings through Terraform,
So that I can enable/disable the FRR service and set the routing profile as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_quagga_general` resource
**Then** FRR service enabled state and routing profile are manageable
**And** if the `os-frr` plugin is not installed, `PluginNotFoundError` is returned
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 7.2: BGP Global Configuration Resource

As an operator,
I want to manage BGP global configuration through Terraform,
So that I can define my AS number, router ID, and network advertisements as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_quagga_bgp_general` resource
**Then** BGP AS number, router ID, and network advertisements are manageable
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 7.3: BGP Neighbor Resource

As an operator,
I want to manage BGP neighbors through Terraform,
So that I can define peering sessions with MetalLB speakers and other routers as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_quagga_bgp_neighbor` resources
**Then** neighbors support remote ASN, IP address, keepalive/holddown timers, update-source, and enabled state
**And** modifying neighbor attributes is update-in-place (not destroy-and-recreate) to avoid dropping BGP sessions
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 7.4: BGP Prefix List Resource

As an operator,
I want to manage BGP prefix lists through Terraform,
So that I can control route filtering as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_quagga_prefix_list` resources
**Then** prefix lists support sequence number, action (permit/deny), and network prefix
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 7.5: BGP Route Map Resource

As an operator,
I want to manage BGP route maps through Terraform,
So that I can define route manipulation policies as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_quagga_route_map` resources
**Then** route maps support match conditions and set actions
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

---

## Epic 8: ACME Certificate Management

Operator can manage ACME certificate lifecycle — accounts, certificates with challenge configuration, and automated issuance — enabling SSL termination for HAProxy frontends.

### Story 8.1: ACME Account Resource

As an operator,
I want to manage ACME accounts through Terraform,
So that I can register with Let's Encrypt (or other CA) as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_acme_account` resources
**Then** accounts support email, CA server URL (staging/production), and registration status
**And** if the `os-acme-client` plugin is not installed, `PluginNotFoundError` is returned
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 8.2: ACME Certificate Resource with Issuance

As an operator,
I want to manage ACME certificates through Terraform with automatic issuance on creation,
So that I can provision SSL certificates for my domains as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_acme_certificate` resources
**Then** certificates support domain, alternative names, challenge type (HTTP-01/DNS-01), and auto-renewal settings
**And** creating a certificate triggers the ACME sign/issuance flow
**And** deleting a certificate revokes it
**And** certificate renewal is NOT managed by the provider (OPNsense cron handles renewal) — the provider manages configuration only
**And** provider does not detect certificate expiry as drift
**And** revocation-via-delete (FR38) is an assumption to validate during implementation — verify that the OPNsense ACME API supports certificate revocation through the delete endpoint
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 8.3: ACME Challenge Configuration Resource

As an operator,
I want to manage ACME challenge configurations through Terraform,
So that I can define how domain ownership is verified during certificate issuance.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_acme_challenge` resources
**Then** challenge configs support HTTP-01 and DNS-01 types with provider-specific settings
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

---

## Epic 9: DNS Management

Operator can manage DNS infrastructure — Unbound host overrides, domain overrides, access control lists, Dynamic DNS accounts, and provider configuration — through Terraform.

### Story 9.1: Unbound Host Override Resource

As an operator,
I want to manage Unbound DNS host overrides through Terraform,
So that I can define local DNS records as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_unbound_host_override` resources
**Then** host overrides support hostname, domain, and IP address
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 9.2: Unbound Domain Override Resource

As an operator,
I want to manage Unbound domain overrides through Terraform,
So that I can define DNS forwarding rules as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_unbound_domain_override` resources
**Then** domain overrides support domain name and forwarding server address
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 9.3: Unbound ACL Resource

As an operator,
I want to manage Unbound access control lists through Terraform,
So that I can control which networks can query DNS as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_unbound_acl` resources
**Then** ACLs support network/CIDR and action (allow/deny/refuse)
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 9.4: Dynamic DNS Account Resource

As an operator,
I want to manage Dynamic DNS accounts through Terraform,
So that I can configure DDNS hostname updates as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_ddclient_account` resources
**Then** accounts support provider, hostname, and credentials (credentials are write-only/Sensitive)
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 9.5: Dynamic DNS Provider Configuration Resource

As an operator,
I want to manage Dynamic DNS provider settings through Terraform,
So that I can configure DDNS service parameters as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_ddclient_provider` resources
**Then** provider configuration including service-specific settings is manageable
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

---

## Epic 10: VPN Management

Operator can manage VPN tunnels — WireGuard server instances and peers, IPsec Phase 1 connections, Phase 2 tunnels, and pre-shared keys — through Terraform.

### Story 10.1: WireGuard Server Resource

As an operator,
I want to manage WireGuard server instances through Terraform,
So that I can define VPN endpoints as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_wireguard_server` resources
**Then** servers support listen port, private key (write-only/Sensitive), and interface settings
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 10.2: WireGuard Peer Resource

As an operator,
I want to manage WireGuard peers through Terraform,
So that I can define VPN client connections as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_wireguard_peer` resources
**Then** peers support public key, allowed IPs, endpoint address, and keepalive interval
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 10.3: IPsec Phase 1 Connection Resource

As an operator,
I want to manage IPsec Phase 1 connections through Terraform,
So that I can define IKE security associations as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_ipsec_phase1` resources
**Then** Phase 1 connections support authentication method, encryption algorithm, and remote gateway
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 10.4: IPsec Phase 2 Tunnel Resource

As an operator,
I want to manage IPsec Phase 2 tunnels through Terraform,
So that I can define IPsec traffic selectors as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_ipsec_phase2` resources
**Then** Phase 2 tunnels support local/remote networks and encryption settings
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 10.5: IPsec Pre-Shared Key Resource

As an operator,
I want to manage IPsec pre-shared keys through Terraform,
So that I can provision VPN authentication credentials as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_ipsec_psk` resources
**Then** PSKs support identity and key value (key is write-only/Sensitive with `UseStateForUnknown`)
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

---

## Epic 11: DHCP Management

Operator can manage DHCPv4 — pools, static mappings, and DHCP options including PXE boot configuration — through Terraform.

### Story 11.1: DHCPv4 Pool Resource

As an operator,
I want to manage DHCPv4 pools through Terraform,
So that I can define DHCP address ranges and settings as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_dhcpv4_pool` resources
**Then** pools support network, address range, gateway, DNS servers, and lease time configuration
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 11.2: DHCPv4 Static Mapping Resource

As an operator,
I want to manage DHCPv4 static mappings through Terraform,
So that I can assign fixed IPs to MAC addresses as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_dhcpv4_static_mapping` resources
**Then** static mappings support MAC address, fixed IP address, and hostname
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

### Story 11.3: DHCPv4 Option Resource

As an operator,
I want to manage DHCPv4 options through Terraform,
So that I can configure PXE boot parameters (options 66, 67, 150) as code.

**Acceptance Criteria:**

**Given** standard resource lifecycle (CRUD + import + drift detection + acceptance tests + documentation)
**When** the operator defines `opnsense_dhcpv4_option` resources
**Then** DHCP options support option number, type, and value including PXE boot options 66 (TFTP server), 67 (boot file), and 150 (TFTP server IP)
**And** resource documentation template (`templates/resources/`) and example HCL (`examples/resources/`) are included

---

## Epic 12: Data Sources, Documentation & Registry Release

Operator can discover the provider on the Terraform Registry, find comprehensive documentation with composition examples, use data sources to reference all existing resources, and install the provider via `terraform init`.

### Story 12.1: Remaining Data Sources

As an operator,
I want data sources for all resource types I manage,
So that I can reference existing OPNsense resources in my Terraform configurations without importing them.

**Acceptance Criteria:**

**Given** all resource types from Epics 2-11 are implemented
**When** the developer creates `*_data_source.go` for each resource type
**Then** every resource type has a corresponding read-only data source
**And** a `data.opnsense_system_info` data source provides firmware version and installed plugin list
**And** data sources follow the same `fromAPI()` conversion as their corresponding resources
**And** acceptance tests verify data source reads match resource state

### Story 12.2: Provider Index Documentation

As an operator,
I want comprehensive provider documentation on the Terraform Registry,
So that I can learn how to configure the provider, understand authentication options, and see a quickstart example.

**Acceptance Criteria:**

**Given** all resources and data sources are implemented
**When** the developer creates `templates/index.md.tmpl`
**Then** the provider index page documents: provider configuration block, authentication options (HCL vs env vars with priority), minimum OPNsense version (26.1.x, min 24.1+), required API user permissions per module, and a complete quickstart example
**And** `tfplugindocs generate` produces the page in `docs/index.md`

### Story 12.3: Composition Examples

As an operator,
I want realistic multi-resource examples showing how to configure common OPNsense patterns,
So that I can copy and adapt proven configurations instead of building from scratch.

**Acceptance Criteria:**

**Given** all resource types are implemented
**When** the developer creates composition examples in `examples/compositions/`
**Then** examples exist for: customer onboarding (HAProxy + ACME), HAProxy full stack, BGP peering, firewall baseline, WireGuard VPN, and DNS management
**And** each example is a working `.tf` file that could be applied to a real OPNsense appliance
**And** `terraform fmt` passes on all examples

### Story 12.4: CI/CD Pipeline and Registry Release

As a developer,
I want GitHub Actions workflows for CI and release,
So that the provider is automatically tested, built, and published to the Terraform Registry on every release.

**Acceptance Criteria:**

**Given** the GitHub repository has the required secrets (GPG_PRIVATE_KEY, PASSPHRASE)
**When** a `v*` tag is pushed to main
**Then** GoReleaser cross-compiles the provider, creates checksums, signs with GPG, and creates a GitHub Release
**And** the Terraform Registry auto-discovers the new release
**And** CI workflows run lint + unit tests on every push
**And** CI workflows run acceptance tests (QEMU OPNsense VM) on PRs to main
**And** `test/scripts/validate-structure.sh` checks that service directories match exports.go
**And** `make check` is the single gate for all validations

### Story 12.5: CHANGELOG and v0.1.0 Release

As a developer,
I want to publish v0.1.0 of the provider to the Terraform Registry,
So that the provider is installable via `terraform init` and visible to the community.

**Acceptance Criteria:**

**Given** all resources, data sources, documentation, and CI/CD are complete
**When** the developer creates the v0.1.0 tag
**Then** CHANGELOG.md follows standard Terraform provider format (FEATURES, IMPROVEMENTS, BUG FIXES)
**And** the release is published on the Terraform Registry
**And** `terraform init` with the provider source downloads and installs successfully
**And** all documentation is visible on the Registry
