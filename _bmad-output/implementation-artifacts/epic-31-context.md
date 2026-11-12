# Epic 31 Context: Provider Completeness, Positioning, and Migration

<!-- Compiled from planning artifacts. Edit freely. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Make the provider complete enough to be the default choice for an appliance, then convert verified coverage into trustworthy positioning, migration guidance, and adoption examples. Eliminate split ownership where supported APIs permit it without presenting research, published code, or bounded appliance evidence as universal support. The active next increment is Story 31.0i's four read-only live-interface surfaces; already released sibling increments must not be rebuilt.

## Stories

- Story 31.0: Complete Provider Coverage Audit
- Story 31.0a: Firewall and NAT Field Parity
- Story 31.0b: Kea Field and Surface Parity
- Story 31.0c: OpenVPN Field and Key Parity
- Story 31.0d: Unbound Settings and Forward Parity
- Story 31.0e: WireGuard Field and Settings Parity
- Story 31.0f: Trust Lifecycle and Settings Parity
- Story 31.0g: IPsec Field Parity
- Story 31.0h: Quagga/BGP Field Parity
- Story 31.0i: Live Interface Data Sources
- Story 31.0j: Provider Retry Settings
- Story 31.0k: Gateway Group Stable API Validation
- Story 31.0l: Kea Option and DDNS Stable API Validation
- Story 31.0m: Gateway Group Resource/Data Source
- Story 31.0n: Kea Generic Option Resource/Data Source
- Story 31.0o: Kea DDNS Settings Resource/Data Source
- Story 31.0p: Kea Ordered Generic-Option Linkage
- Story 31.1: Complete Provider Positioning README and Registry Pass
- Story 31.2: Existing Provider Migration Guide
- Story 31.3: Example Stacks for Adoption

## Requirements & Constraints

- Judge completeness by semantic fields, lifecycle, import, drift, and safe lookup behavior, not resource counts or names alone. Give singleton/sensitive lookup gaps explicit implementation or exception decisions; write-only values must not become misleading read outputs.
- Support predictable plans, omission-preserving adoption, idempotent updates, recoverable partial failures, and compatible state evolution. Separate API persistence, service activation, active traffic, older-firmware support, and OpenTofu execution claims.
- Ship generated Registry documentation and realistic composition examples with each surface. Explain minimum versions, plugins, permissions, authentication, imports, support boundaries, and remaining limitations. Do not copy historical competitor-positioning claims without fresh verification.
- Protect credentials and private material in diagnostics/logs; Sensitive state is not encryption. TLS verification is the default. Respect bounded read concurrency so refresh does not overwhelm the appliance. The performance target is a 60-second refresh for a typical 50–100-resource configuration.
- Require appropriate regression/protocol tests, generated-doc consistency, and `make check`; live read validation is distinct from fixtures. Keep exact tested source, disposable evidence, merge, release, and Registry proof separately attributable. A skip is not proof.
- Implementation does not authorize production firmware changes, routing changes, or ownership cutovers. Unsupported system-general settings remain upstream-blocked; HASync model shape and PPPoE semantics remain research, not owners to invent.

## Technical Decisions

- Use supported REST semantics and the Terraform Plugin Framework/Protocol v6 contract. Configuration mutations must serialize to protect shared appliance configuration and follow their service-specific activation/recovery contract. Read-only data sources must never acquire mutation, activation, or recovery ownership.
- Preserve the distinction between durable assignment/IP configuration and operational interface observations. The four missing surfaces are `interface`, `interface_all`, `interfaces_overview`, and `interfaces_overview_all`: device/runtime observations and combined assignment/runtime overview, each with single and aggregate forms. They are not a replacement assignment resource or an assignment-configuration lookup.
- Live observations must reflect authoritative API truth, with deterministic nested/aggregate values and actionable missing/disappearing-device diagnostics. Do not fabricate missing values, infer configuration ownership, or expose write-only secrets through lookups.
- Ordered Kea generic-option linkage stays blocked. Released core 26.7.8 has no authoritative order-preserving parent read. Existing generic-option CRUD and DDNS owners do not establish linkage order. Raw XML, selected-map key order, and grid-search CSV are not authorized.

## UX & Interaction Patterns

Operators evaluate supported capabilities, configure credentials, import existing objects, and review a no-change plan before transferring ownership. Migration guidance must explain semantic renames, dependency order, state compatibility, and exceptions instead of promising drop-in equivalence. Examples should compose real appliance workflows and make the boundary between observation and management evident. Diagnostics should identify the failed operation and distinguish permissions, authentication, TLS, connectivity, and unavailable endpoints, with a useful next action.

## Cross-Story Dependencies

- Reconciled release baseline: `origin/main` at `b090de1` includes merged retry, Trust, IPsec, and FRR increments. GitHub releases confirm [v0.13.0](https://github.com/matthew-on-git/terraform-provider-opnsense/releases/tag/v0.13.0) retry controls, [v0.14.0](https://github.com/matthew-on-git/terraform-provider-opnsense/releases/tag/v0.14.0) supported Trust lifecycle/read coverage, [v0.15.0](https://github.com/matthew-on-git/terraform-provider-opnsense/releases/tag/v0.15.0) advanced IPsec, and [v0.16.0](https://github.com/matthew-on-git/terraform-provider-opnsense/releases/tag/v0.16.0) advanced FRR/BGP. The latter was published at 2026-11-05T02:25:30Z. Older backlog/unreleased labels, including the v0.16 changelog heading, are stale; do not repeat these increments.
- Release does not close the entire Trust/IPsec/FRR parent stories. Keyless-CA support, negotiated IPsec traffic, older firmware/plugin compatibility, OpenTofu execution, and remaining row-level/route-collector coverage retain their documented independent gates.
- Proceed with 31.0i before positioning, migration, and adoption Stories 31.1–31.3. Keep 31.0p blocked until a later released core exposes an authoritative ordered parent read. Gateway groups, Kea generic-option/DDNS ownership, and base assignment/IP adoption already exist and must be reused rather than duplicated.
- Historical parity inventories disagree: the current coverage audit identifies 15 resource-matching lookup gaps after the certificate lookup, while older snapshots say 16. These are separate from the four live-interface gaps; reconcile inventory before making public count claims.
- Epic 32 separately owns acceptance-proof artifacts, preflight/version diagnostics, and v1 stability criteria. Indexed planning checkpoints can lag verified repository/release facts.
