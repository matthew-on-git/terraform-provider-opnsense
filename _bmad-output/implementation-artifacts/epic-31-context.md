# Epic 31 Context: Provider Completeness, Positioning, and Migration

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Make the provider complete enough to be the default choice for an appliance by closing meaningful field, lifecycle, and lookup gaps, then translate verified completeness into trustworthy positioning, migration guidance, and adoption examples. Success means operators can standardize on one provider for supported object domains without hidden second-tool dependencies or unsupported coverage claims.

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

- Coverage must be established at field and lifecycle depth, not inferred from resource counts or matching names. Compare against the pinned ecosystem baseline and rerun the deterministic audit after registration or schema changes. Every meaningful gap needs implementation, an intentional exception, research with missing evidence, or an explicit upstream blocker. Carry remaining singleton/sensitive lookup gaps as explicit decisions.
- Resource behavior must support import, accurate plan/apply, out-of-band drift detection and repair, idempotent no-change plans, and recoverable partial failures. Read-only lookups must not introduce a second owner or expose unsafe sensitive values. Verify lifecycle and cleanup on a disposable appliance; unit/protocol tests alone do not prove real service activation.
- Support requires stable, released target APIs and durable semantics, not merely upstream master code or an action endpoint. Base interface assignment/IP requires OPNsense 26.7.6+. PPPoE remains unverified, system-general management remains upstream-blocked, and HASync model semantics remain research-only.
- Keep gateway-group populated first-tier clearing limitations explicit. Generic Kea option CRUD and DDNS ownership are already shipped; ordered parent linkage requires an authoritative order-preserving read API. No direct XML workaround is authorized. This epic does not authorize production firmware, routing, DHCP ownership, or other operational cutovers.
- Document supported versions, plugin prerequisites, permissions, limitations, import behavior, and realistic composition examples with each public surface. Position broad modern appliance management and brownfield friendliness only as strongly as the support matrix and evidence allow; avoid competitive drama and unqualified completeness claims.
- Retry controls must make transient-failure retries and backoff configurable without compromising state safety. Preserve bounded read concurrency to protect PHP-FPM and serialize mutations to protect appliance configuration integrity.
- Protect API credentials and resource secrets with sensitive/write-only semantics where supported; never log credential-bearing bodies. TLS verification is the default, with insecure mode explicit. Examples must use injected secrets and recommend protected remote state.

## Technical Decisions

- Use the Terraform Plugin Framework and supported OPNsense REST APIs rather than legacy-page scraping, direct configuration edits, or external-script resources. Plugins must already be installed; firmware/plugin installation and backup/restore are outside provider ownership.
- Preserve the appliance's two-phase mutation contract: persist changes, then invoke the affected service's reconfigure action. Firewall filter changes require savepoint/apply/rollback protection. When persistence succeeds but activation fails, retain recoverable state and explain that changes may not be active.
- Populate state from authoritative API read-back rather than echoing requested configuration. Mark replacement only for genuinely immutable attributes, preserve schema-upgrade compatibility, and keep representations faithful to API semantics, including selections, references, ordering, and sensitive values.
- Diagnostics must identify operation and resource, expose field-level validation and required privileges, and distinguish authentication, TLS, connectivity, missing endpoint/plugin, and missing-object conditions. Errors should suggest a useful corrective action.
- Migration aims at one provider per appliance and object domain in steady state; never allow two providers to manage the same object. Close or explicitly document gaps before transferring ownership. Multi-appliance usage relies on native provider aliases, not custom orchestration.

## UX & Interaction Patterns

Users evaluate coverage and prerequisites, configure credentials, import existing objects, review the resulting plan, and then use reviewed plan/apply workflows. Documentation should make unsupported domains and migration limitations visible before adoption. Examples should demonstrate linked resources and realistic appliance compositions, not isolated snippets alone. Preserve readable, predictable plans for reviewers and provide actionable diagnostics when a workflow cannot proceed.

## Cross-Story Dependencies

The coverage audit directs parity work and must remain current as capabilities change. Prioritize remaining Trust lifecycle work before advanced IPsec, remaining FRR/BGP, live-interface lookup, and retry work; released FRR compatibility is not the advanced-parity gate. Trust settings alone do not close CA/certificate lifecycle or disposable-proof requirements. Stable-API validation precedes gateway-group and Kea owners; their shipped owners must not be reimplemented as missing work. Research ordered Kea linkage separately. Positioning, migration guidance, and adoption stacks follow verified parity. Keep completed assignment support's release probe current, and keep Epic 32 acceptance artifacts, preflight diagnostics, and v1 criteria separate from this epic's coverage claims.
