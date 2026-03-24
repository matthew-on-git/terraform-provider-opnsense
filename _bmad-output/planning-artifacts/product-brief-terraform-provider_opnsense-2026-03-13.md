---
stepsCompleted: [1, 2, 3, 4, 5, 6]
status: complete
inputDocuments:
  - "opnsense-manager Ansible project (external: /home/mmellor/Work/gitlab.mfsoho.linkridge.net/hardware-infra/opnsense/opnsense-manager)"
date: 2026-03-13
author: Matthew
---

# Product Brief: terraform-provider_opnsense

## Executive Summary

terraform-provider_opnsense is a comprehensive Terraform provider delivering full Infrastructure as Code coverage for OPNsense — both core APIs and the plugin ecosystem. No existing provider offers meaningful coverage of the plugin APIs (HAProxy, FRR/BGP, ACME, Dynamic DNS, WireGuard, IPsec) that make OPNsense viable as an enterprise edge platform, nor do they cover enough of the core APIs (Unbound DNS, firewall, interfaces) to eliminate the need for a second configuration tool. This provider delivers true plan/apply semantics across the entire OPNsense appliance, enabling operators to manage their complete network edge from a single Terraform configuration with full CI/CD pipeline support.

---

## Core Vision

### Problem Statement

Organizations adopting OPNsense as a platform edge replacement for enterprise appliances (F5, Citrix ADC, Palo Alto) have no way to manage their OPNsense configuration comprehensively through Terraform. The only existing provider (browningluke/opnsense) covers a narrow subset of base APIs and zero plugin APIs. This forces operators into split-brain configuration — some resources in Terraform, the rest in Ansible or manual UI — which eliminates the core benefit of IaC: a single source of truth with plan/apply semantics.

Operators who need to manage HAProxy load balancing, FRR/BGP dynamic routing, ACME certificate lifecycle, Unbound DNS, Dynamic DNS, WireGuard and IPsec VPN tunnels, and firewall rules on the same appliance are forced to use Ansible or manual configuration for all of it, since no provider covers enough to be the sole tool.

### Problem Impact

- **Split-brain configuration:** Partial Terraform coverage forces operators to maintain two configuration systems, doubling complexity and creating drift between tools.
- **No visibility before changes:** Operators cannot see a precise diff of what will change on their OPNsense appliance before applying, creating risk in production network environments.
- **Cognitive load:** Without plan output, operators must mentally simulate the impact of changes across interconnected resources (HAProxy backends linked to servers, BGP peers advertising routes, ACME certs bound to frontends, VPN tunnels referencing interfaces). This is exhausting and error-prone.
- **No drift detection:** Configuration drift goes unnoticed until something breaks — there is no mechanism to compare desired state against actual state.
- **Broken CI/CD workflows:** Without reliable plan/apply, automated pipelines cannot implement proper approval gates, peer review of infrastructure changes, or safe rollback strategies.

### Why Existing Solutions Fall Short

| Solution | Limitation |
|---|---|
| **browningluke/opnsense** (Terraform) | No HAProxy, FRR/BGP, ACME, Dynamic DNS, WireGuard, IPsec, or any plugin API support. Limited core API coverage. Pre-v1.0, "not recommended for production." |
| **Ansible** (current workaround) | No true plan/apply. `--check` mode leaks API calls. No state file, no drift detection. No transactional safety across linked resources. |
| **Manual UI configuration** | No version control, no auditability, no reproducibility, no multi-site scalability. |
| **Other providers** (RyanNgWH, etc.) | Abandoned or extremely limited scope. |

### Proposed Solution

A from-scratch Terraform provider for OPNsense that:

1. **Covers the complete OPNsense appliance** — core APIs (firewall, interfaces, Unbound DNS, routing, system) and plugin APIs (HAProxy, FRR/BGP, ACME, Dynamic DNS, DHCPv4, WireGuard, IPsec), eliminating the need for any second configuration tool.
2. **Delivers real plan/apply** — persistent state, accurate diffs, drift detection, and import of existing resources. Terraform semantics done right, no workarounds.
3. **Supports CI/CD workflows natively** — plan output for peer review, apply with approval gates, state locking for team collaboration.
4. **Models resources faithfully to OPNsense's API** — resource schemas derived from production operational experience managing OPNsense at scale.
5. **Built on the Terraform Plugin Framework** — modern architecture, not legacy SDKv2, ensuring long-term maintainability and access to the latest Terraform features.

### Key Differentiators

- **Complete appliance coverage:** The only provider aiming to manage the full OPNsense appliance — core and plugins — as a single source of truth. No split-brain configuration.
- **Plugin ecosystem support:** First and only provider covering HAProxy, FRR/BGP, ACME, Dynamic DNS, WireGuard, IPsec — the APIs that make OPNsense an enterprise edge platform.
- **Battle-tested resource models:** Resource schemas informed by production Ansible automation covering 21+ API endpoints across 8 plugin/service areas with real-world configuration patterns.
- **Modern architecture:** Built from scratch on the Terraform Plugin Framework, not constrained by legacy design decisions or partial implementations.
- **Open source portfolio piece:** Professionally maintained at matthew.mellor.earth, designed to become the community standard for OPNsense IaC.

## Target Users

### Primary Users

#### Persona 1: "The Professional Homelabber" — Alex
**Role:** Senior DevOps/Platform Engineer by day, homelab/side-business operator by night
**Context:** Uses Terraform, CI/CD, and IaC daily in their professional work. Runs OPNsense as the network edge for a homelab or small side business because enterprise appliances (F5, Palo Alto) are cost-prohibitive outside of work. Expects the same professional workflow at home that they use at the office.
**Motivation:** "I shouldn't have to downgrade my engineering practices just because I'm not on enterprise gear."
**Problem Experience:** Currently stuck with Ansible playbooks or manual UI clicks to manage OPNsense. Knows exactly how painful the gap is because they use Terraform plan/apply every day at work. The lack of a mature provider means their homelab infrastructure is the one thing they can't manage properly.
**Success Moment:** Running `terraform plan` against their OPNsense appliance and seeing a clean diff of HAProxy backends, BGP peers, and firewall rules — the same workflow they use at work — and thinking "finally."

#### Persona 2: "The MSP Operator" — Jordan
**Role:** Infrastructure engineer at a Managed Service Provider or small IT consultancy
**Context:** Manages multiple OPNsense appliances across client sites. Needs reproducible, auditable configurations. Can't afford enterprise licensing per client but needs enterprise-grade operational practices.
**Motivation:** "I need to deploy consistent configurations across 10+ client firewalls without manual drift."
**Problem Experience:** Copies configurations between OPNsense instances manually or with fragile scripts. No way to audit what changed, when, or why. Client onboarding means hours of manual UI work per appliance. Changes are risky because there's no preview.
**Success Moment:** Defining a base Terraform module for OPNsense, customizing per-client variables, and running `terraform plan` to see exactly what a new client deployment will create before applying.

#### Persona 3: "The Ambitious Homelabber" — Sam
**Role:** Tech enthusiast running a homelab, learning IaC practices
**Context:** Chose OPNsense over pfSense for the API and plugin ecosystem. Wants to learn and apply DevOps practices (Terraform, Git, CI/CD) to their home network. May not use Terraform professionally yet but aspires to.
**Motivation:** "I want my homelab to be my learning platform for professional skills."
**Problem Experience:** Found the existing Terraform provider but quickly hit walls — it doesn't cover HAProxy, WireGuard, or any of the plugins they actually installed OPNsense for. Fell back to clicking through the UI.
**Success Moment:** Their first successful `terraform apply` that provisions a WireGuard tunnel and HAProxy backend in one run, committed to a Git repo they can show in a job interview.

### Secondary Users

#### The PR Reviewer / Team Lead
Does not write HCL themselves but reviews `terraform plan` output in CI/CD pipelines before approving applies. Values clear, readable plan diffs and predictable resource naming. Their success is confidence that the plan output accurately represents what will change on the appliance.

#### Community Contributors
Open source contributors who want to add resources for OPNsense plugins not yet covered. They value clean architecture, tight coupling to the OPNsense API (no workarounds), comprehensive documentation, and a well-structured codebase that's easy to extend. The DevRail project standards and professional polish make this a codebase contributors are proud to work on.

### User Journey

| Stage | Experience |
|---|---|
| **Discovery** | Searches Terraform Registry for "opnsense," finds existing providers inadequate. Discovers terraform-provider_opnsense via Registry, GitHub, or matthew.mellor.earth portfolio. |
| **Evaluation** | Reads docs, sees plugin coverage (HAProxy, FRR, ACME, WireGuard, IPsec, Dynamic DNS). Confirms it covers their use case. Notes professional quality — DevRail standards, clean code, comprehensive examples. |
| **Onboarding** | Configures provider with OPNsense API key/secret. Runs `terraform import` to bring existing resources under management. Sees current state in `terraform plan` — first "aha!" moment. |
| **Core Usage** | Defines OPNsense infrastructure in HCL. Runs `terraform plan` in CI to preview changes. Reviews diffs in PRs. Runs `terraform apply` with confidence. Detects drift on next plan. |
| **Success Moment** | The first time `terraform plan` catches an unexpected change (drift) they didn't know about — proving the value of state-based management over fire-and-forget Ansible. |
| **Long-term** | OPNsense config lives in Git alongside all other infrastructure. Provider becomes the single source of truth. For MSPs, it becomes a module they reuse across clients. For contributors, it becomes a portfolio piece they're proud of. |

## Success Metrics

### User Success Metrics

| Metric | Target | Measurement |
|---|---|---|
| **Ansible replacement completeness** | 100% of current opnsense-manager functionality reproducible in Terraform | All 4 Ansible roles (BGP, HAProxy, ACME, DHCP) plus Unbound, Dynamic DNS, WireGuard, and IPsec fully covered by provider resources |
| **Plan accuracy** | Zero surprises on apply — `terraform plan` output matches actual changes every time | No reported issues where apply produced changes not shown in plan |
| **Import success** | Existing OPNsense configurations importable via `terraform import` without manual state editing | All resource types support import from live appliance |
| **Drift detection reliability** | `terraform plan` correctly detects out-of-band changes made via UI or API | Plan shows accurate diff when resources are modified outside Terraform |
| **Core API coverage** | Firewall aliases/rules, interfaces/VLANs, static routes, gateways, system settings all manageable | Resource count covering core OPNsense API surface |

### Business Objectives

| Objective | 3-Month Target | 12-Month Target |
|---|---|---|
| **Ansible replacement** | First plugin areas (HAProxy, FRR/BGP) fully functional — begin migrating opnsense-manager workloads | Complete replacement — opnsense-manager retired, all OPNsense config lives in Terraform |
| **Registry presence** | Published on Terraform Registry with documentation and examples for initial resources | Comprehensive provider with 8+ plugin/service areas covered |
| **Community adoption** | First external users beyond Matthew; initial GitHub stars and Registry downloads | Recognized as the go-to OPNsense provider; community PRs for plugins not in initial scope |
| **Portfolio value** | Provider listed on matthew.mellor.earth with architecture writeup | Demonstrates end-to-end professional open source project lifecycle (DevRail standards, CI/CD, docs, community) |

### Key Performance Indicators

| KPI | Target | Leading Indicator |
|---|---|---|
| **Resource type count** | 30+ resource types across core + 8 plugin areas | Steady cadence of new resources per sprint |
| **Test coverage** | >80% acceptance test coverage on all resources | Tests passing in CI on every PR |
| **Terraform Registry downloads** | 100+ monthly downloads within 12 months | Download trend increasing month-over-month |
| **GitHub engagement** | 50+ stars, 5+ community contributors within 12 months | Issues filed by external users (signals real adoption) |
| **Documentation completeness** | Every resource has usage examples, argument reference, and import instructions | Docs ship with every new resource (not backfilled later) |
| **Zero workarounds policy** | No resources that require manual steps or external scripts to function | All resources are self-contained CRUD with proper state management |

## MVP Scope

### Core Features

#### Provider Infrastructure
- **Provider authentication:** API key/secret configuration, HTTPS with configurable TLS verification
- **OPNsense API client:** Reusable HTTP client handling auth, error responses, and the OPNsense reconfigure pattern (most changes require a service reconfigure call after mutation)
- **Terraform Plugin Framework:** Built on the modern Framework (not SDKv2), with proper resource schemas, CRUD operations, and state management
- **Import support:** All resources support `terraform import` from existing OPNsense configurations — required for migrating live appliances under Terraform management
- **Acceptance test framework:** Real API tests against an OPNsense instance, integrated into CI/CD
- **CI/CD pipeline:** GitLab CI with lint, test, build, and release stages following DevRail standards
- **Terraform Registry publishing:** Provider published and discoverable on the Terraform Registry with full documentation
- **Documentation:** Every resource includes usage examples, argument/attribute reference, and import instructions

#### Plugin/Service Resources (8 areas — replaces opnsense-manager Ansible 1:1 plus additions)

| Area | Resources | Priority |
|---|---|---|
| **HAProxy** (`os-haproxy`) | Servers, backends, frontends, ACLs, health checks, SSL offloading | Highest — complex linked resources, validates provider architecture |
| **FRR/BGP** (`os-frr`) | General settings, BGP global config, neighbors, route maps, prefix lists | Highest — current Ansible workload |
| **ACME** (`os-acme-client`) | Accounts, certificates, challenge configuration, CA registration | High — certificate lifecycle management |
| **DHCPv4** | Pools, static mappings, DHCP options (PXE boot) | High — current Ansible workload |
| **Unbound DNS** | Host overrides, domain overrides, ACLs, DNSBL settings | High — core DNS management |
| **Dynamic DNS** (`os-ddclient`) | Accounts, provider configuration | Medium — straightforward CRUD |
| **WireGuard** (`os-wireguard`) | Server instances, peers, handshake settings | Medium — VPN management |
| **IPsec** | Phase 1 (connections), Phase 2 (tunnels), pre-shared keys | Medium — VPN management |

#### Core API Resources

| Area | Resources | Priority |
|---|---|---|
| **Firewall** | Aliases, rules, NAT (port forward, outbound), categories | High — fundamental to any OPNsense config |
| **Interfaces** | VLAN assignments, interface configuration | High — network foundation |
| **Routing** | Static routes, gateways, gateway groups | High — traffic path management |
| **System** | General settings, DNS servers, NTP | Medium — baseline appliance config |

#### Data Sources
- All resource types also available as data sources for read-only lookups (e.g., `data.opnsense_haproxy_backend` to reference existing backends)
- System information data source (firmware version, plugin list) for conditional logic

### Out of Scope for MVP

| Item | Rationale |
|---|---|
| **OPNsense plugins not listed above** | Additional plugins (e.g., Zabbix agent, Telegraf, CrowdSec) deferred to community demand post-MVP |
| **Multi-appliance orchestration** | Provider manages one OPNsense instance per provider block — multi-site is handled by Terraform's native provider aliasing, not custom logic |
| **Backup/restore resources** | Config export/import is an operational concern outside Terraform's CRUD model |
| **Firmware/plugin installation management** | Installing or upgrading plugins/firmware via Terraform is outside the provider's responsibility — plugins must be pre-installed |
| **Custom Terraform functions or complex validation** | Provider-defined functions and advanced cross-resource validation deferred to post-MVP |
| **OpenAPI/Swagger auto-generation** | Resource schemas are hand-crafted from API documentation and operational experience, not auto-generated |

### MVP Success Criteria

| Criteria | Validation |
|---|---|
| **Ansible retirement** | All 4 opnsense-manager Ansible roles (BGP, HAProxy, ACME, DHCP) fully replaced by Terraform resources with equivalent or better functionality |
| **Complete appliance management** | All 8 plugin/service areas + core APIs under Terraform control — no second tool needed |
| **Import existing config** | Live OPNsense appliance fully imported into Terraform state via `terraform import` without manual state file editing |
| **Accurate plan/apply** | `terraform plan` produces correct diffs; `terraform apply` makes exactly the changes shown in plan; no surprises |
| **Drift detection** | Out-of-band changes (via UI or API) detected on next `terraform plan` |
| **CI/CD integration** | Provider works with GitLab CI pipelines using self-hosted GitLab HTTP backend for Terraform state storage |
| **Registry published** | Provider available on Terraform Registry with complete documentation and examples |
| **DevRail compliance** | All code passes `make check` — linting, formatting, security scanning, and tests |

### Future Vision

| Phase | Scope |
|---|---|
| **Post-MVP: Community plugins** | Resources for additional OPNsense plugins based on community demand (CrowdSec, Zabbix, Telegraf, BIND, etc.) |
| **Post-MVP: Advanced features** | Provider-defined functions, complex cross-resource validation, resource dependency hints |
| **Post-MVP: Terraform modules** | Published reusable modules for common patterns (e.g., "HAProxy with ACME cert" module, "BGP peering with MetalLB" module) |
| **Long-term: Ecosystem leadership** | Become the canonical OPNsense Terraform provider; comprehensive API coverage driven by community contributions following the clean, well-documented architecture |
