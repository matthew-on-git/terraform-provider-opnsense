# Epic 31 Context: Provider Completeness, Positioning, and Migration

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Make the provider complete enough to be the default choice for an appliance, replacing fragmented configuration ownership with trustworthy Terraform management across supported core and plugin APIs. Close verified coverage gaps, distinguish unsupported surfaces honestly, and turn proven completeness into adoption through clear positioning, migration guidance, and reusable examples.

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

- Completeness means supported functionality works through real plan/apply, import, and drift detection, not merely that resource names exist. Implement only verified API surfaces; expose remaining limitations and upstream blockers without inventing ownership.
- Singleton and sensitive lookup gaps require explicit implementation or exception decisions. Historical planning inventory totals are not targets; audit the current registered implementation before adding or claiming coverage.
- Every supported resource must support accurate read-back, idempotent operations, recoverable partial failures, and appropriate import semantics. A clean import and repeated unchanged apply must produce a no-change plan. Reconfigure failure must retain successfully persisted state and explain that appliance activation may be incomplete.
- Protect credentials and private keys from diagnostics and logs; use safe sensitive/write-only handling where supported. Explain sensitive-state risks and recommend encrypted remote state. Default to verified HTTPS, with insecure mode an explicit opt-in; examples obtain credentials through environment variables or secret injection.
- Retry controls must address transient API failures with configurable maximum retries and backoff without corrupting state during appliance unavailability. Protect the appliance with bounded read concurrency. Performance goals are a 60-second refresh for 50–100 resources, 10-second CRUD including reconfigure, and 5-second individual import.
- Ship argument/attribute references, import guidance, and examples with each resource. Provider entry documentation must explain authentication, permissions, supported versions, prerequisites, support boundaries, and a usable quickstart. Positioning must distinguish demonstrated support from planned coverage.
- Public material must remain consumer-neutral. Use documentation addresses and reserved example domains, never operator inventory, credentials, topology, or deployment plans. This epic does not authorize production firmware changes, routing changes, or ownership cutover.

## Technical Decisions

- Follow Terraform Plugin Framework lifecycle contracts and Protocol v6, preserving Terraform/OpenTofu core interoperability and automatic schema migration when schemas change.
- All configuration mutations must be globally serialized to protect appliance configuration integrity. Mutations require the affected service's reconfigure operation; firewall filter changes require savepoint/apply/rollback safety rather than ordinary reconfigure.
- Populate state from authoritative API reads after writes, not request echoes. Treat UUID references as dependency edges. Use replacement only for genuinely immutable attributes; preserve unreadable write-only values without claiming drift detection for them.
- API success cannot be inferred from HTTP status alone: validation failures may arrive with HTTP 200, while missing UUIDs may return blank/default records. Surface field-level validation and permission/plugin/version guidance; remove genuinely deleted managed resources from state. Search/list lookups must consume pagination completely.
- Preserve API semantics when translating booleans, selected values, multi-selects, lists, and UUIDs. Ordered Kea generic-option linkage remains gated on an authoritative order-preserving read API; raw XML is not an authorized workaround.
- Resource documentation is generated from schemas, templates, and example HCL. Validate full create/read/import/update/delete lifecycles against real OPNsense; keep unit/protocol evidence, disposable-appliance proof, release publication, and Registry availability distinct.

## UX & Interaction Patterns

- Adoption proceeds from discovering actual coverage to configuring credentials, importing existing objects, reviewing a no-change plan, and then managing changes through reviewable plan/apply.
- Migration guidance must order independent objects before linked resources and preserve cross-resource references. Each import should be followed by plan verification before advancing; do not imply an automatic production ownership handoff.
- Composition examples should demonstrate meaningful end-to-end stacks such as HAProxy with ACME, BGP peering, and firewall baselines. Diagnostics and plan output must let operators distinguish in-place updates, replacement, validation failure, and persisted-but-not-activated changes.

## Cross-Story Dependencies

- Prior data-source, Registry hardening, buildable-gap, and HAProxy/ACME tracks are completed/historical unless regression evidence requires reopening. Assignment/IP support is also complete and requires OPNsense 26.7.6+; do not repeat those implementations.
- Finish Trust lifecycle parity before advanced IPsec and remaining FRR/BGP parity, then live-interface lookups and retry settings. Trust settings shipped as a separate increment, not completion of CA/certificate lifecycle or disposable proof. Released FRR 2.0 compatibility is not the remaining advanced parity work.
- Gateway groups, generic Kea option CRUD, and Kea DDNS settings have shipped. Ordered option linkage is separate research, not a reason to rebuild those surfaces.
- Complete positioning, migration guidance, and adoption examples after parity, carrying explicit lookup exceptions forward. Release-credibility work follows separately; System-general remains upstream-blocked, while HASync model shape and PPPoE semantics remain research rather than invented provider resources.
