---
stepsCompleted: [step-01-init, step-02-discovery, step-02b-vision, step-02c-executive-summary, step-03-success, step-04-journeys, step-05-domain, step-06-innovation, step-07-project-type, step-08-scoping, step-09-functional, step-10-nonfunctional, step-11-polish, step-12-complete]
status: complete
inputDocuments:
  - "product-brief-terraform-provider_opnsense-2026-03-13.md"
  - "research/technical-opnsense-api-terraform-provider-framework-research-2026-03-13.md"
  - "research/technical-terraform-plugin-framework-research-2026-03-13.md"
workflowType: 'prd'
documentCounts:
  briefs: 1
  research: 2
  brainstorming: 0
  projectDocs: 0
classification:
  projectType: "Developer Tool — Infrastructure Provider (Terraform)"
  domain: "Infrastructure / Network Automation"
  complexity: "High"
  projectContext: "Greenfield codebase targeting brownfield infrastructure"
---

# Product Requirements Document - terraform-provider_opnsense

**Author:** Matthew
**Date:** 2026-03-17

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Project Classification](#project-classification)
3. [Success Criteria](#success-criteria)
4. [Domain-Specific Requirements](#domain-specific-requirements)
5. [Terraform Provider Specific Requirements](#terraform-provider-specific-requirements)
6. [Product Scope & Phased Development](#product-scope--phased-development)
7. [User Journeys](#user-journeys)
8. [Functional Requirements](#functional-requirements)
9. [Non-Functional Requirements](#non-functional-requirements)

---

## Executive Summary

terraform-provider_opnsense is a comprehensive Terraform provider that delivers full Infrastructure as Code coverage for OPNsense appliances — core APIs and the plugin ecosystem that no existing provider covers. It enables operators to manage firewall rules, HAProxy load balancing, FRR/BGP dynamic routing, ACME certificate lifecycle, Unbound DNS, Dynamic DNS, WireGuard and IPsec VPN tunnels, DHCP, interfaces, and routing through Terraform's native plan/apply workflow with drift detection, state management, and import support.

The provider targets DevOps engineers, homelabbers, and MSP operators who have adopted OPNsense as a cost-effective replacement for enterprise edge appliances (F5, Palo Alto, Citrix ADC) but lack mature IaC tooling to manage it. The existing provider landscape is fragmented and incomplete — the most mature option (browningluke/opnsense, pre-v1.0) covers a narrow subset of base APIs with zero plugin support, forcing operators into Ansible or manual UI configuration. OPNsense's API has reached sufficient maturity to support a comprehensive provider, but no one has built one.

The primary success criterion is complete appliance management: when any operator can manage their entire OPNsense appliance from a single Terraform configuration with accurate plan/apply and drift detection, the provider has succeeded. The first validation is retiring the author's production Ansible automation (opnsense-manager) entirely in favor of Terraform.

### What Makes This Special

- **See every change before it happens.** Run `terraform plan` and get a complete diff of your OPNsense appliance — firewall rules, BGP peers, HAProxy backends, VPN tunnels, DNS overrides — in one output. No other tool gives you this for OPNsense.
- **First provider with enough coverage to actually be useful.** Every existing provider fails the completeness test — they cover enough to demo, not enough to deploy. This provider covers 8 plugin/service areas plus core APIs across 40+ endpoints, eliminating the need for a second configuration tool.
- **Plugins are the product, not an afterthought.** HAProxy, FRR/BGP, ACME, WireGuard, IPsec, and Dynamic DNS are what make OPNsense viable as an enterprise edge platform. Covering the core API without these is like a cloud provider that manages VPCs but not load balancers.
- **Handles OPNsense's API correctly, not just broadly.** The OPNsense API has non-standard patterns that break naive integrations: HTTP 200 on validation errors, blank defaults for missing resources, a two-phase reconfigure lifecycle, and write-only password fields. This provider handles all of them through a purpose-built API client with custom error types, mutex-protected reconfigure, and proper state management.
- **Built on production operational experience.** Resource schemas are informed by a battle-tested Ansible automation project (opnsense-manager) that runs in production CI/CD with plan/apply stages against a live OPNsense appliance — not speculative API mappings.
- **Modern architecture, professional standards.** Built from scratch on the Terraform Plugin Framework (not legacy SDKv2), following DevRail project standards, with acceptance tests against real OPNsense instances and documentation generated from schemas.

## Project Classification

This is a high-complexity infrastructure provider built from scratch in Go, targeting operators who need to import and manage existing OPNsense configurations through Terraform.

- **Project Type:** Developer Tool — Infrastructure Provider (Terraform)
- **Domain:** Infrastructure / Network Automation
- **Complexity:** High — 30+ resource types across 12+ API modules, cross-resource dependencies, safety-critical network configuration, OPNsense-specific API patterns (reconfigure lifecycle, HTTP 200 on validation errors, write-only fields)
- **Project Context:** Greenfield codebase targeting brownfield infrastructure — `terraform import` is a day-one requirement for migrating existing OPNsense configurations under Terraform management
- **Technology:** Go, Terraform Plugin Framework v6, OPNsense REST API

## Success Criteria

### User Success

| Criterion | Measurable Outcome |
|---|---|
| **Complete appliance management** | Every resource currently managed by opnsense-manager Ansible roles (BGP, HAProxy, ACME, DHCP) plus Unbound DNS, Dynamic DNS, WireGuard, IPsec, firewall, interfaces, and routing is manageable via Terraform |
| **Accurate plan/apply** | `terraform plan` output matches actual changes on apply — zero surprises. No reported cases where apply produced changes not shown in plan |
| **Smooth import** | Existing OPNsense resources importable via `terraform import` without manual state file editing or extensive tinkering. Import → plan → "No changes" on first try |
| **Drift detection** | Out-of-band changes made via OPNsense UI or API are detected on the next `terraform plan` with an accurate diff |
| **Ansible retirement** | opnsense-manager Ansible project fully retired. All OPNsense configuration lives in Terraform HCL committed to Git |

### Business Success

| Criterion | Target |
|---|---|
| **Personal workflow replacement** | Matthew's OPNsense appliance managed entirely via Terraform with GitLab CI pipeline (plan on MR, manual apply on merge) |
| **Portfolio quality** | Provider published on Terraform Registry, documented, and presentable at matthew.mellor.earth as a professional open-source project |
| **Community adoption** (secondary) | Provider is publicly available and usable by others. Community growth is welcome but not a gating success criterion |

### Technical Success

| Criterion | Measurable Outcome |
|---|---|
| **Acceptance test coverage** | >80% of resources have acceptance tests running against a real OPNsense instance |
| **CI/CD pipeline** | Lint, test, build, and release stages passing in CI. Acceptance tests automated |
| **DevRail compliance** | All code passes `make check` — linting, formatting, security scanning, tests |
| **Registry published** | Provider available on Terraform Registry with documentation and examples for all resources |
| **Import support** | Every resource type implements `ImportState` with UUID passthrough |
| **OPNsense version support** | Provider tested and working against current OPNsense stable release |

### Measurable Outcomes

| Milestone | Validation |
|---|---|
| **First resource works end-to-end** | `opnsense_firewall_alias`: create, read, update, delete, import, drift detection — all passing |
| **Ansible role parity** | Each of the 4 Ansible roles (BGP, HAProxy, ACME, DHCP) has equivalent Terraform resources with passing acceptance tests |
| **Full import** | Matthew's live OPNsense appliance fully imported into Terraform state. `terraform plan` shows "No changes" |
| **Ansible retired** | opnsense-manager repository archived. No Ansible playbooks running against OPNsense |

## Domain-Specific Requirements

### Network Configuration Safety

- **Firewall filter rollback is mandatory.** The provider MUST use OPNsense's savepoint/apply/cancelRollback mechanism for firewall filter rule changes. The three-step process: `savepoint` → `apply/{revision}` → `cancelRollback/{revision}` ensures automatic 60-second revert if the change causes loss of connectivity. This is not optional — a bad firewall rule applied without rollback protection can permanently lock the operator out of the appliance.
- **Non-destructive plan modifiers.** BGP neighbor, HAProxy frontend, and VPN tunnel resources must use `update-in-place` by default, not `destroy-and-recreate`. Unnecessarily destroying a BGP session or VPN tunnel causes traffic disruption. Only truly immutable fields should use `RequiresReplace`.
- **Reconfigure isolation.** Reconfigure calls must be scoped to the affected service module. Changing an HAProxy backend must not trigger a firewall reconfigure. Each service module has its own reconfigure endpoint and the provider must route correctly.

### OPNsense API Constraints

- **HTTP 200 on validation errors.** Every API mutation response must be parsed for `result != "saved"` regardless of HTTP status code. The API client must extract and surface `validations` field contents as Terraform diagnostics.
- **Blank defaults for missing UUIDs.** GET requests for non-existent UUIDs return HTTP 200 with a blank/default record, not 404. The provider must detect this condition (search-first pattern or default value detection) to correctly handle resource deletion and drift.
- **Write-only fields.** Password, pre-shared key, and secret fields (`UpdateOnlyTextField`) cannot be read back from the API. These must be marked `Sensitive: true` with `UseStateForUnknown` plan modifier. Drift detection is not possible on these fields — this is a known and accepted limitation.
- **Request body wrapper key.** All POST bodies must wrap fields in a resource-type key (e.g., `{"server": {...}}` not `{...}`). Omitting the wrapper causes a silent `{"result": "failed"}` with no error context.
- **Reconfigure after every mutation.** CRUD operations stage changes in config.xml but do not affect the running system until `POST /api/{module}/service/reconfigure` is called. The provider must call reconfigure after every Create, Update, and Delete, protected by a global mutex.

### OPNsense Version Compatibility

- **Target version:** OPNsense 26.1.x (current stable). Minimum supported: 24.1+ (when WireGuard and Firewall moved from plugins to core).
- **Version detection:** Provider should query `/api/core/firmware/info` during `Configure` to determine the running OPNsense version and log it. Version-specific endpoint routing may be needed for resources that changed between major releases.
- **Upgrade cadence:** OPNsense releases two majors per year (January and July). The provider must be tested against each new major release before claiming compatibility. Breaking API changes are expected on major releases, not minors.
- **Plugin-to-core migrations:** When OPNsense absorbs a plugin into core (as happened with WireGuard in 24.1), API endpoint paths may change. The provider should detect this and fail with a clear error directing the user to upgrade the provider version.

### Concurrency and State Safety

- **Global mutex for all mutations.** All Create, Update, Delete operations must be serialized through a single global mutex. OPNsense cannot safely handle parallel writes — concurrent mutations risk lost changes because reconfigure applies all pending changes globally, not per-resource.
- **Read operations are parallel.** `terraform plan` (which calls Read on all resources) must not be mutex-gated. Only mutations lock.
- **State must reflect API reality.** After every Create and Update, the provider must read back from the API and populate state from the response — never echo request config into state. This is the foundation of drift detection.

## Terraform Provider Specific Requirements

### Project-Type Overview

This is a Terraform provider — a Go binary that implements the Terraform Plugin Protocol v6 and is distributed via the Terraform Registry. Users interact with it exclusively through HCL configuration files, the `terraform` CLI, and plan/apply output. There is no UI, no API surface beyond the Terraform contract, and no direct user-facing documentation outside of Registry docs and example HCL.

### Technical Architecture Considerations

**Language and Runtime:**
- Go (required by Terraform Plugin Framework). Provider binary is a single statically-linked executable.
- Cross-compilation targets: linux, darwin, windows × amd64, arm64, 386, arm (standard goreleaser matrix). No additional platform targets needed.
- `CGO_ENABLED=0` for Terraform Cloud compatibility and static linking.

**Distribution and Installation:**
- Primary: Terraform Registry (`terraform init` auto-downloads). Registry address: `registry.terraform.io/matthew-on-git/opnsense`
- Secondary: GitHub Releases (manual binary download for air-gapped environments)
- No additional package managers (Homebrew, AUR, etc.)

**Provider Schema Contract:**

| Element | Convention |
|---|---|
| Provider config attributes | `uri`, `api_key`, `api_secret`, `insecure` |
| Environment variable prefix | `OPNSENSE_` (`OPNSENSE_URI`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET`, `OPNSENSE_ALLOW_INSECURE`) |
| Resource naming | `opnsense_{module}_{resource}` (e.g., `opnsense_haproxy_server`, `opnsense_firewall_alias`) |
| Data source naming | `opnsense_{module}_{resource}` (same as resources) |
| ID attribute | `id` — Computed, UUID from OPNsense, `UseStateForUnknown` plan modifier |
| Boolean attributes | `types.Bool` in Terraform, converted to/from OPNsense `"0"`/`"1"` strings |
| Multi-select attributes | `types.Set` of `types.String` (not List — unordered to prevent perpetual diffs) |
| Cross-resource references | `types.String` with UUID validation |
| Write-only fields | `Sensitive: true`, `UseStateForUnknown` plan modifier |
| Credential validation | `Configure` must validate credentials with a test API call (e.g., `/api/core/firmware/status`) and fail fast with a clear diagnostic if authentication fails |

**API Client Architecture:**

| Component | Design |
|---|---|
| Package location | `pkg/opnsense/` — separate from Terraform types, owns the global mutex |
| HTTP client | `go-retryablehttp` with custom `RoundTripper` for Basic Auth |
| CRUD pattern | Go generics: `Add[K]`, `Get[K]`, `Update[K]`, `Delete` with `ReqOpts` config per resource |
| Wrapper key | Automatic via `ReqOpts.Monad` field (e.g., `"server"`, `"rule"`, `"host"`) |
| Error handling | Custom error types (`NotFoundError`, `ValidationError`, `AuthError`) parsed from response body |
| Concurrency | Global mutex for all mutations — transparent to resource authors. Resources call the API client which acquires the lock internally. Reads are parallel. |
| Reconfigure | Inline after every CRUD operation. Standard: `ReqOpts.ReconfigureEndpoint` string. Special: `ReqOpts.ReconfigureFunc` function pointer for firewall filter's savepoint/apply/cancelRollback flow |
| Code generation | YAML schemas → Go structs + CRUD methods via `go generate` |

**Resource Implementation Pattern:**

Each resource follows a consistent four-file pattern per service module:

```
internal/
├── service/{module}/
│   ├── {resource}_resource.go      # CRUD implementation (Create/Read/Update/Delete/ImportState)
│   ├── {resource}_schema.go        # Terraform schema definition
│   ├── {resource}_model.go         # API model struct + toAPI()/fromAPI() conversion functions
│   ├── {resource}_resource_test.go # Acceptance tests
│   ├── {resource}_data_source.go   # Read-only data source variant
│   └── exports.go                  # Resources() and DataSources() registration
└── validators/                     # Shared validators (UUID, port range, IP address, etc.)
    └── uuid.go
```

Mutex, error handling, wrapper key marshaling, and reconfigure are all handled transparently in `pkg/opnsense/`. Resource authors write schema + model + CRUD logic without thinking about concurrency or API quirks.

### Documentation and Examples

**Provider index page (most important doc):** The Registry provider page must cover:
- Provider configuration (HCL block with all attributes)
- Authentication options (explicit config vs. environment variables, priority order)
- Minimum OPNsense version requirement (26.1.x, minimum 24.1+)
- Required API user permissions on OPNsense
- Quickstart: complete working example from provider config to first resource

**Documentation generation:** tfplugindocs auto-generates Registry docs from Go schema definitions. Each resource needs:
- Template in `templates/resources/{resource}.md.tmpl` (narrative + usage context)
- Example HCL in `examples/resources/opnsense_{resource}/resource.tf`
- Import example in `examples/resources/opnsense_{resource}/import.sh` showing full workflow: import block, resource block, expected "No changes" plan output

**Example style: Realistic compositions over standalone snippets.** Examples show resources in context — how they connect in real-world usage:

| Composition | Resources Shown Together |
|---|---|
| **Customer onboarding** | `opnsense_haproxy_server` + `opnsense_haproxy_backend` + frontend ACL update + `opnsense_acme_certificate` — the hot-path workflow |
| **HAProxy full stack** | Server + backend + frontend + ACME cert — complete load balancer setup |
| **BGP peering** | General config + neighbor + prefix list + route map — full dynamic routing |
| **Firewall baseline** | Category + alias + rule + NAT — foundational security config |
| **VPN setup** | WireGuard server + peer + firewall rule allowing tunnel traffic |
| **DNS management** | Unbound host override + domain override + Dynamic DNS account |

### Implementation Considerations

**Testing strategy:**
- **Unit tests:** API client logic, type converters, error parsing (no OPNsense instance needed)
- **Acceptance tests:** Full Terraform lifecycle per resource (create → read → import → update → delete) against real OPNsense
- **Bootstrap phase:** During initial development, acceptance tests run locally against a dev OPNsense instance. QEMU CI is added once the test framework is proven.
- **CI (post-bootstrap):** QEMU-based OPNsense VM in GitHub Actions, `-p 1` serial execution (configurable per-environment — developers with dedicated instances can run parallel)
- Every resource must have acceptance tests before merge. No untested resources in releases.

**Release pipeline:**
- GoReleaser cross-compiles, creates checksums, signs with GPG (RSA 4096-bit)
- GitHub Actions triggered on `v*` tags
- `terraform-registry-manifest.json` declares Protocol v6.0
- Releases are permanent on the Registry — test thoroughly before tagging

**Migration tooling:**
- All resources implement `ImportState` with UUID passthrough
- Documentation includes import examples showing full workflow (import block + resource block + expected plan output) and dependency ordering guide
- Composition examples double as migration reference — users can see the target state for their imported resources

## Product Scope & Phased Development

### MVP Strategy & Philosophy

**MVP Approach:** Problem-solving MVP — build the minimum that lets Matthew retire his Ansible automation and manage his entire OPNsense appliance through Terraform. There is no market validation phase because the developer is the primary user. The MVP is complete when `terraform plan` can show the full state of the appliance and `terraform apply` can make any change currently handled by Ansible.

**Resource Requirements:** Solo developer with AI-assisted development, following DevRail project standards. No team coordination overhead — all architectural decisions are made by the author.

### MVP Feature Set (Phase 1)

**Core User Journeys Supported:**
- Journey 1 (Migration) — full import of existing OPNsense configuration
- Journey 2 (Customer Onboarding) — HAProxy + ACME hot-path workflow
- Journey 3 (Drift Detection) — accurate plan diffs on managed resources
- Journey 4 (Local Development) — dev_overrides build/test cycle

**Implementation Order (within MVP):**

Resources are built in dependency and usage-priority order. Each tier delivers independently deployable resources for new configurations. For import workflows, cross-referenced resources require their dependency tiers to be complete first (e.g., importing a firewall rule that references an alias requires the alias resource type from Tier 1).

| Tier | Area | Key Resources | Est. Count | Rationale |
|---|---|---|---|---|
| **0** | Provider scaffold | Auth, API client, error handling, mutex, code gen pipeline, `opnsense_firewall_alias`, `opnsense_haproxy_server` | 2 resources | Validates full stack for both core and plugin APIs |
| **1** | Firewall | Categories, rules, NAT (port forward, outbound) | 4 resources | Foundation — everything depends on firewall config |
| **2** | HAProxy | Backends, frontends, ACLs, health checks | 4 resources | Hot path — customer onboarding workflow (servers already in Tier 0) |
| **3** | Core infrastructure | Interfaces, VLANs, virtual IPs, static routes, gateways, gateway groups, system settings | 7 resources | Foundation for cross-references — firewall rules reference interfaces, routes reference gateways |
| **4** | FRR/BGP | General settings, BGP config, neighbors, prefix lists, route maps | 5 resources | Current Ansible workload — MetalLB peering |
| **5** | ACME | Accounts, certificates, challenge config | 3 resources | Depends on HAProxy frontends for HTTP-01 |
| **6** | Unbound DNS | Host overrides, domain overrides, ACLs | 3 resources | Core DNS management |
| **7** | VPN | WireGuard servers, WireGuard peers, IPsec phase 1, IPsec phase 2, IPsec PSKs | 5 resources | VPN tunnel management |
| **8** | DHCPv4 + Dynamic DNS | DHCP pools, static mappings, DHCP options, Dynamic DNS accounts, Dynamic DNS provider config | 5 resources | Completes MVP scope |
| **—** | Data sources | All resource types as data sources + system info | ~38 data sources | Read-only lookups for all managed resources |

**Total estimated:** ~38 resources + ~38 data sources

**Must-Have Capabilities (every tier):**
- CRUD + ImportState for all resources
- Acceptance tests passing against real OPNsense
- Documentation with realistic composition examples
- Drift detection (state from API read-back, not config echo)

### Post-MVP Features

**Phase 2 (Growth):**
- Additional OPNsense plugins based on personal need or community demand (CrowdSec, Zabbix, Telegraf, BIND)
- Reusable Terraform modules for common patterns (HAProxy + ACME cert, BGP + MetalLB)
- QEMU-based acceptance tests in CI (if not completed during MVP)
- Multi-version OPNsense test matrix

**Phase 3 (Expansion):**
- Provider-defined functions and advanced cross-resource validation
- Community contribution workflow (CONTRIBUTING guide, PR templates, issue templates)
- Published modules on Terraform Registry
- Version-aware endpoint mapping for OPNsense major release compatibility
- Terraform resource skeleton generator (extend code gen to produce four-file resource boilerplate from YAML schema)

### Risk Mitigation Strategy

**Technical Risks:**

| Risk | Severity | Mitigation |
|---|---|---|
| OPNsense API returns blank defaults instead of 404 for missing resources | HIGH | Search-first pattern to confirm existence before GET. Implemented in API client, transparent to resources |
| HTTP 200 on validation errors | HIGH | Response body parsing on every mutation. Custom `ValidationError` type with field-level diagnostics |
| Firewall rule change locks out operator | HIGH | Mandatory savepoint/apply/cancelRollback for firewall filter resources. 60-second automatic rollback if cancelRollback not called |
| Concurrent mutations cause lost writes | HIGH | Global mutex serializing all CRUD operations. Transparent to resource authors |
| OPNsense major release breaks API endpoints | MEDIUM | Version detection during Configure. CI tests against current stable. Provider releases track OPNsense majors |
| Write-only fields can't detect drift | MEDIUM | Accepted limitation. Fields marked Sensitive with UseStateForUnknown. Documented clearly |
| Plugin API differs from core API pattern | MEDIUM | Tier 0 validates both core (`firewall_alias`) and plugin (`haproxy_server`) patterns before committing to architecture |
| OPNsense dev/test instance unavailability | MEDIUM | Document reproducible VM setup process. Consider Packer/Vagrant for automated OPNsense VM provisioning. Maintain snapshot of known-good test state |

**Resource Risks:**

| Risk | Mitigation |
|---|---|
| Solo developer — bus factor of 1 | Clean architecture + DevRail standards + comprehensive docs make the project transferable. Code gen reduces per-resource effort |
| Scope is large (~38 resources + ~38 data sources) | Tiered implementation order means each tier delivers usable value. No "all or nothing" dependency |
| Burnout on repetitive resource implementation | YAML code generation eliminates API client boilerplate. Four-file resource pattern is mechanical once established. Future: extend code gen to scaffold Terraform resource boilerplate too |

## User Journeys

### Journey 1: The Great Migration — Alex Imports His Existing OPNsense

**Opening Scene:** Alex has been managing his OPNsense appliance with Ansible for over a year. He has BGP peers connecting two Kubernetes clusters, HAProxy frontends routing traffic to a dozen services, ACME certificates auto-renewing, DHCP serving PXE boot configs, and firewall rules he's afraid to touch because he can't preview changes. He installs the provider, configures credentials, and stares at an empty `.tf` file knowing his entire network config needs to come in cleanly.

**Rising Action — Phase 1 (Independent Resources):** Alex starts with resources that have no dependencies — firewall aliases, system settings, Unbound DNS overrides, static routes. He runs `terraform import opnsense_firewall_alias.k8s_services <uuid>` and then `terraform plan`. It shows "No changes." Relief. He writes the matching HCL, imports the next resource. Each import follows the same rhythm: import → plan → confirm "No changes" → commit. Within an hour, all independent resources are under management.

**Rising Action — Phase 2 (Linked Resources):** Now the harder part. HAProxy has dependency chains: servers must exist before backends can reference them, backends before frontends. Alex imports in dependency order — servers first, then backends (whose `linked_servers` attribute resolves to the already-imported server UUIDs), then frontends. BGP neighbors reference prefix lists and route maps, so those come first. ACME certificates reference HAProxy frontends. Each import resolves cleanly because the dependencies are already in state.

**Climax:** After importing his last resource, Alex runs `terraform plan` across his entire configuration. The output reads: **"No changes. Your infrastructure matches the configuration."** His entire OPNsense appliance — every firewall rule, every BGP neighbor, every HAProxy backend, every VPN tunnel — is now under Terraform management.

**Resolution:** Alex archives the opnsense-manager Ansible repository. His OPNsense config lives in Git alongside his Kubernetes manifests and cloud infrastructure. For the first time, every piece of his infrastructure speaks the same language.

**Requirements revealed:** Import by UUID for all resources. Accurate state read-back (no drift on import). Dependency-aware import ordering guidance in documentation. Cross-resource UUID references resolve correctly after import. Plan must show "No changes" after clean import.

---

### Journey 2: New Customer Onboarding — Alex Adds a Backend Service

**Opening Scene:** A new customer signs up for Alex's platform. They need a domain routed through HAProxy with SSL termination via ACME, and a backend pointing to a new Kubernetes service. This used to mean SSH into OPNsense, click through the HAProxy UI, manually request a cert, update the frontend — a 30-minute error-prone process.

**Rising Action:** Alex opens his Terraform config and adds three resources: an `opnsense_haproxy_server` pointing to the MetalLB VIP, an `opnsense_haproxy_backend` linking to that server, and an `opnsense_acme_certificate` for the customer's domain. He updates the existing frontend's ACL to route the new domain. He runs `terraform plan`.

**Climax:** The plan output shows exactly four changes — the three new resources and the modified frontend ACL. No surprises, no side effects. Alex opens a merge request. The GitLab CI pipeline runs `terraform plan` and posts the output as a comment. He reviews the diff, approves, merges. The pipeline runs `terraform apply`. The customer's domain is live in under five minutes.

**Resolution:** Customer onboarding is now a Git commit. Reproducible, auditable, reviewable. When the next customer signs up, Alex copies the resource block, changes the domain and VIP, and opens another MR.

**Edge Case — ACME Challenge Failure:** The ACME HTTP-01 challenge fails because DNS hasn't propagated yet. `terraform apply` creates the server and backend successfully, but the certificate resource returns an error. Alex sees the diagnostic: "ACME challenge failed: domain not reachable." The server and backend exist; the cert does not. Alex fixes DNS, re-runs `terraform apply` — it picks up where it left off, creating only the certificate. The partial state is correct and recoverable.

**Edge Case — Customer Offboarding:** A customer churns. Alex removes the three resource blocks and the frontend ACL entry. `terraform plan` shows three deletions and one modification. He reviews, merges, applies. The customer's HAProxy backend, server, and certificate are cleanly removed. The appliance reconfigures once, removing all traces.

**Requirements revealed:** Cross-resource references via UUID (server → backend → frontend). ACME certificate creation with challenge configuration. HAProxy frontend ACL updates (in-place modification, not recreation). Partial apply failure must leave state consistent — created resources stay in state, failed resources don't. Re-apply must be idempotent and pick up from failure point. Deletions must reconfigure cleanly.

---

### Journey 3: Drift Detection — Alex Catches an Unauthorized Change

**Opening Scene:** It's Monday morning. Alex's colleague made an "emergency" change over the weekend through the OPNsense UI — modified an existing managed firewall rule to open an additional port for debugging, and changed a NAT rule's destination. Nobody documented it.

**Rising Action:** Alex's scheduled `terraform plan` runs in GitLab CI as part of a morning drift check. The pipeline reports changes detected.

**Climax:** The plan output shows two diffs. The firewall rule shows `port: "443" → "443,8080"` — the colleague added port 8080. The NAT rule shows `target: "10.131.0.103" → "10.131.0.199"` — the destination was changed. Alex sees exactly what was modified, attribute by attribute.

**Resolution:** Alex discusses with the colleague. The NAT change was a debug leftover — he runs `terraform apply` to revert it to the declared state. The firewall port change was intentional — he updates the HCL to include port 8080, runs `terraform plan` to confirm "No changes," and commits. The audit trail captures both decisions. Drift detection turned invisible manual changes into visible, reviewable events.

**Requirements revealed:** Accurate Read implementation that reflects actual appliance state. Attribute-level diffs in plan output. State populated from API read-back (not echoed from config). Apply reverts drift to declared state. No false positives — plan should show "No changes" when appliance matches config.

---

### Journey 4: Local Development — Alex Tests a Complex Change

**Opening Scene:** Alex needs to restructure his BGP configuration — changing ASN allocations and adding new prefix lists. This is a high-risk change that could blackhole traffic if done wrong. He wants to test it locally before pushing to CI.

**Rising Action:** Alex has `dev_overrides` configured in his `~/.terraformrc` pointing to his locally-built provider binary. He edits his BGP resources — modifies the ASN on two neighbors, adds a new prefix list, updates route maps. He runs `terraform plan` from his terminal.

**Climax:** The plan shows exactly which BGP neighbors will be updated, the new prefix list that will be created, and the route map changes. He can see that his existing MetalLB peering sessions will be reconfigured but not destroyed — the plan says "update in-place," not "must be replaced." Satisfied with the plan, he pushes to GitLab. CI confirms the same plan output. He merges and applies.

**Edge Case — Partial Apply Failure:** Alex is adding a new HAProxy frontend that references a backend. He accidentally typos the backend UUID in his HCL. `terraform apply` creates the new ACL successfully but the frontend fails with a validation error: "Backend UUID not found." Terraform reports the error with the OPNsense validation message. The ACL is in state (it was created), the frontend is not. Alex fixes the UUID in HCL, runs `terraform plan` — it shows the frontend will be created (the ACL shows "No changes" since it already exists). He re-applies successfully.

**Resolution:** The BGP reconfiguration completes without dropping a single session. The plan told him exactly what would happen, and exactly that happened. When the typo happened, the error was clear, the state was consistent, and recovery was straightforward.

**Requirements revealed:** Local dev workflow with `dev_overrides`. Provider binary builds cleanly with `go install`. Plan modifiers correctly distinguish update-in-place from destroy-and-recreate. OPNsense validation errors surface as clear Terraform diagnostics with field-level detail. Partial failures leave state consistent — successfully created resources remain in state.

---

### Journey 5: Community Contributor — Dana Adds a New Plugin Resource

**Opening Scene:** Dana runs OPNsense with the CrowdSec plugin. She discovers terraform-provider_opnsense on the Terraform Registry, sees it covers HAProxy and BGP but not CrowdSec. She checks the GitHub repo and finds clean architecture, comprehensive documentation, and a CONTRIBUTING guide.

**Prerequisites:** Dana knows Go basics, understands Terraform provider concepts (resources, data sources, state), and has access to an OPNsense instance with the CrowdSec plugin installed for testing.

**Rising Action:** Dana reads the existing HAProxy resource implementation as a reference. She sees the pattern: YAML schema for the API client endpoints, generated Go structs with JSON tags, ReqOpts config, then a Terraform resource with schema + CRUD. She creates a new `internal/service/crowdsec/` package, adds the YAML schema for the CrowdSec API endpoints, runs `go generate` to produce the API client code, and writes the Terraform resource with acceptance tests. Documentation is generated automatically by tfplugindocs from her schema definitions and example HCL — she doesn't write docs by hand.

**Climax:** Dana's PR passes CI — linting, formatting, unit tests, and acceptance tests against the QEMU OPNsense VM all green. The code review is smooth because the architecture is consistent and well-documented.

**Resolution:** The CrowdSec resources ship in the next minor release. Dana's contribution follows the same patterns as every other resource in the provider. The architecture made it easy to add without understanding the entire codebase.

**Requirements revealed:** Clean, documented architecture that contributors can follow. Service-based package organization (`internal/service/{module}/`). YAML schema → code generation pipeline. Acceptance test framework that contributors can run locally. CONTRIBUTING guide with prerequisites and step-by-step instructions. tfplugindocs auto-generates documentation from schema definitions.

### Journey Requirements Summary

| Journey | Key Capabilities Required |
|---|---|
| **Migration (Import)** | `terraform import` by UUID, accurate state read-back, dependency ordering docs, cross-reference resolution after import, "No changes" verification |
| **Customer Onboarding** | Cross-resource UUID references, in-place updates, atomic reconfigure, partial failure recovery, idempotent re-apply, clean deletion with reconfigure |
| **Drift Detection** | Accurate Read from live appliance, attribute-level diffs, state from API read-back not config echo, apply reverts to declared state, no false positives |
| **Local Development** | `dev_overrides` support, fast build cycle, correct plan modifiers (update vs. replace), clear validation error diagnostics, consistent partial failure state |
| **Contributor** | Service-based packages, YAML code gen, acceptance test framework, CONTRIBUTING docs, auto-generated documentation, consistent patterns |

## Functional Requirements

### Provider Configuration & Authentication

- FR1: Operator can configure the provider with OPNsense appliance URI, API key, and API secret
- FR2: Operator can configure credentials via environment variables (`OPNSENSE_URI`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET`) as an alternative to HCL configuration
- FR3: Operator can disable TLS certificate verification for self-signed certificates via `insecure` attribute or `OPNSENSE_ALLOW_INSECURE` environment variable
- FR4: Provider validates credentials during configuration by making a test API call and fails fast with a clear diagnostic if authentication fails
- FR5: Provider detects the running OPNsense version during configuration and logs it for diagnostics

### Resource Lifecycle Management (Cross-Cutting — applies to ALL resources)

- FR6: Operator can create any supported OPNsense resource via `terraform apply`
- FR7: Operator can read the current state of any supported resource via `terraform plan` (refresh)
- FR8: Operator can update any supported resource in-place via `terraform apply` when attributes change
- FR9: Operator can delete any supported resource via `terraform apply` when the resource block is removed
- FR10: Operator can import any existing OPNsense resource into Terraform state via `terraform import` using its UUID
- FR11: Provider detects drift on managed resources — out-of-band changes made via OPNsense UI or API are shown in the next `terraform plan`
- FR12: Provider applies OPNsense service reconfigure after every create, update, and delete operation, routed to the affected service module's endpoint (e.g., HAProxy changes reconfigure HAProxy, not Unbound)
- FR13: Provider uses the firewall filter savepoint/apply/cancelRollback mechanism for firewall rule changes to prevent operator lockout
- FR14: Provider serializes all mutation operations through a global mutex to prevent concurrent write conflicts
- FR15: Provider populates state from API read-back after every create and update — never echoes request config into state
- FR16: Provider plan output correctly signals update-in-place vs. destroy-and-recreate for each resource change, using `RequiresReplace` only on truly immutable fields
- FR17: Provider leaves state consistent after partial apply failure — successfully created/updated resources remain in state, failed resources do not
- FR18: Provider handles reconfigure failure by reporting a clear diagnostic. If CRUD succeeds but reconfigure fails, the resource is marked in state but the operator is warned that changes may not be active on the appliance

### Firewall Management

- FR19: Operator can manage firewall aliases (host, network, port, URL table types) with content lists
- FR20: Operator can manage firewall categories for organizing rules
- FR21: Operator can manage firewall filter rules with source/destination, ports, protocols, interfaces, and action (pass/block/reject)
- FR22: Operator can manage firewall NAT port-forward rules with target address, ports, and interface
- FR23: Operator can manage firewall NAT outbound rules with source, translation, and interface

### HAProxy Management

- FR24: Operator can manage HAProxy servers with address, port, weight, SSL, and health check configuration
- FR25: Operator can manage HAProxy backends with linked servers, load balancing algorithm, health check settings, and persistence
- FR26: Operator can manage HAProxy frontends with bind addresses, default backend, SSL offloading, and linked ACLs
- FR27: Operator can manage HAProxy ACL rules with match conditions (host header, path, SNI) for frontend routing
- FR28: Operator can manage HAProxy health checks with type, interval, and threshold configuration
- FR29: Operator can reference HAProxy resources across types via UUID (server → backend → frontend chain)

### FRR/BGP Management

- FR30: Operator can manage FRR general settings (enable/disable FRR service, routing profile)
- FR31: Operator can manage BGP global configuration (AS number, router ID, network advertisements)
- FR32: Operator can manage BGP neighbors with remote ASN, IP address, timers, and update-source
- FR33: Operator can manage BGP prefix lists with sequence numbers, action, and network prefixes
- FR34: Operator can manage BGP route maps with match conditions and set actions

### ACME Certificate Management

- FR35: Operator can manage ACME accounts with email, CA server (Let's Encrypt staging/production), and registration
- FR36: Operator can manage ACME certificate configurations with domain, alternative names, challenge type (HTTP-01/DNS-01), and auto-renewal settings
- FR37: Operator can trigger ACME certificate issuance (sign) as part of resource creation
- FR38: Operator can revoke an ACME certificate by deleting the certificate resource
- FR39: Certificate renewal is owned by OPNsense via its built-in cron — the provider manages certificate *configuration*, not renewal lifecycle. Provider does not detect certificate expiry as drift.

### Core Infrastructure Management

- FR40: Operator can manage network interfaces and their configuration
- FR41: Operator can manage VLAN assignments on interfaces
- FR42: Operator can manage virtual IPs (CARP, IP Alias) on interfaces
- FR43: Operator can manage static routes with destination network, gateway, and metric
- FR44: Operator can manage gateways with interface, address, and monitoring settings
- FR45: Operator can manage gateway groups with priority and trigger level settings
- FR46: Operator can manage system general settings (hostname, domain, DNS servers, NTP)

### Unbound DNS Management

- FR47: Operator can manage Unbound host overrides with hostname, domain, and IP address
- FR48: Operator can manage Unbound domain overrides with domain and forwarding server
- FR49: Operator can manage Unbound access control lists

### VPN Management

- FR50: Operator can manage WireGuard server instances with listen port, private key, and interface settings
- FR51: Operator can manage WireGuard peers with public key, allowed IPs, endpoint, and keepalive
- FR52: Operator can manage IPsec Phase 1 connections with authentication, encryption, and remote gateway
- FR53: Operator can manage IPsec Phase 2 tunnels with local/remote networks and encryption settings
- FR54: Operator can manage IPsec pre-shared keys

### DHCPv4 Management

- FR55: Operator can manage DHCPv4 pools with network, range, gateway, and DNS settings
- FR56: Operator can manage DHCPv4 static mappings with MAC address and fixed IP
- FR57: Operator can manage DHCPv4 options (including PXE boot options 66, 67, 150)

### Dynamic DNS Management

- FR58: Operator can manage Dynamic DNS accounts with provider, hostname, and credentials
- FR59: Operator can manage Dynamic DNS provider configuration

### Data Sources

- FR60: Every resource type has a corresponding read-only data source for reference in other configurations (~38 data sources matching ~38 resources)
- FR61: Operator can query OPNsense system information (firmware version, installed plugins) as a data source

### Error Handling & Diagnostics

- FR62: Provider surfaces OPNsense API validation errors as Terraform diagnostics with field-level detail
- FR63: Provider detects and reports missing resources (deleted out-of-band) by removing them from state during refresh
- FR64: Provider reports clear errors when OPNsense API is unreachable, credentials are invalid, or required plugins are not installed
- FR65: Provider reports permission-specific errors when the API key lacks required privileges for a module, indicating which permission group is needed

### Documentation & Examples

- FR66: Every resource type ships with auto-generated Registry documentation including argument reference, attribute reference, and import instructions — documentation is delivered as part of each resource, not backfilled
- FR67: Documentation includes realistic composition examples showing resources in context (e.g., complete HAProxy + ACME stack, BGP peering setup, firewall baseline, customer onboarding workflow)
- FR68: Provider index page documents configuration, authentication options, minimum OPNsense version, required API user permissions per module, and a complete quickstart example

## Non-Functional Requirements

### Performance

- NFR1: `terraform plan` (full refresh of all managed resources) completes within 60 seconds for a typical configuration of 50-100 managed resources
- NFR2: Individual resource CRUD operations complete within 10 seconds including reconfigure
- NFR3: `terraform import` for a single resource completes within 5 seconds
- NFR4: Provider binary startup (gRPC handshake + credential validation) completes within 3 seconds
- NFR5: Provider uses connection pooling / HTTP keep-alive to minimize TCP handshake overhead across sequential API calls
- NFR6: Provider limits concurrent API read operations to prevent overwhelming the OPNsense appliance's PHP-FPM worker pool (configurable concurrency limit)

### Security

- NFR7: API credentials marked `Sensitive: true` are never displayed in plan output, apply output, or provider logs
- NFR8: Provider uses Terraform Plugin Framework write-only attributes for credential fields where available, preventing credential storage in state entirely
- NFR9: Provider supports HTTPS with configurable TLS verification — default is TLS-verified, `insecure` mode requires explicit opt-in
- NFR10: Provider never logs request or response bodies containing credential fields at any log level
- NFR11: Provider documentation warns that Terraform state files may contain sensitive values and recommends encrypted remote state backends
- NFR12: CI/CD pipeline credentials use environment variables or CI-native secret injection — never hardcoded in HCL or committed to version control

### Reliability & State Consistency

- NFR13: All resource operations are idempotent — running `terraform apply` twice with no config changes produces "No changes" on the second run
- NFR14: Provider never corrupts Terraform state — partial failures leave state in a recoverable condition
- NFR15: Provider handles OPNsense API transient failures (network timeouts, 500 errors) with automatic retry via `go-retryablehttp` (configurable max retries and backoff)
- NFR16: Provider handles OPNsense appliance reboot gracefully — API unavailability during reboot surfaces as a retryable error, not state corruption
- NFR17: Provider schema version upgrades work seamlessly — when a user upgrades the provider and a resource schema has changed, state migration runs automatically without manual intervention
- NFR18: All acceptance tests are deterministic — no flaky tests. Tests that fail intermittently are quarantined and fixed, never disabled or skipped to unblock CI

### Integration & Compatibility

- NFR19: Provider is compatible with Terraform CLI 1.0+ (Protocol v6)
- NFR20: Provider is compatible with OpenTofu for core provider protocol operations (CRUD, import, plan, apply). Feature parity with OpenTofu-specific extensions is not guaranteed.
- NFR21: Provider works with the GitLab HTTP state backend for remote state with locking
- NFR22: Provider binary is statically linked (`CGO_ENABLED=0`) and runs without external dependencies on all supported platforms
- NFR23: Provider is compatible with Terraform Cloud and Terraform Enterprise execution environments

### Code Quality & Maintainability

- NFR24: All code passes `make check` — golangci-lint, formatting, security scanning, and tests
- NFR25: Acceptance test coverage: >80% of resource types have acceptance tests running against a real OPNsense instance
- NFR26: Adding a new resource type follows a documented, repeatable pattern (four-file template + YAML schema) that can be completed without understanding the full codebase
- NFR27: Go code follows standard Go conventions (gofmt, effective Go, go vet) enforced by CI
- NFR28: Go module hygiene enforced — `go mod tidy` produces no changes, no unused dependencies, no `replace` directives in released versions
- NFR29: Provider releases follow semantic versioning — breaking changes only on major version bumps
- NFR30: CHANGELOG maintained in standard Terraform provider format (FEATURES, IMPROVEMENTS, BUG FIXES)

### Error Message Quality

- NFR31: Error messages include the resource type, the operation that failed, the OPNsense API response, and a suggested action when possible
- NFR32: Validation errors include the specific field name and the constraint that was violated (e.g., `field "port": value must be between 1 and 65535`)
- NFR33: Permission errors identify which OPNsense privilege group is required for the attempted operation
- NFR34: Connection errors distinguish between DNS resolution failure, TLS errors, authentication failure, and API endpoint not found
