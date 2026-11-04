# Epic 31 Context: Provider Completeness, Positioning, and Migration

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Make the provider complete enough for operators to standardize on it for an appliance, then translate verified coverage into trust and adoption. Close meaningful field, lifecycle, lookup, and configuration gaps before presenting broad coverage claims, migration guidance, and reusable adoption examples. Completeness means truthful support boundaries and dependable brownfield ownership, not merely a larger resource inventory.

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

- Compare ecosystem coverage at field, lookup, lifecycle, and sensitive-value semantics, not just names or counts. The established-provider comparison is pinned to v0.26.0. Account for differently named equivalents and intentional differences; rerun the deterministic audit after registration or schema changes.
- Close safe lookup gaps or record explicit singleton/sensitive exceptions. The current planning audit identifies 16 resource-matching data-source gaps; this is a dated inventory, not a permanent total or permission to expose write-only secrets.
- Support claims must distinguish implemented, released, real-appliance-proved, research-only, and upstream-limited behavior. Stable target-release APIs are the gate; upstream development branches and setup-wizard endpoints do not establish durable day-2 management semantics.
- Real lifecycle evidence must establish authoritative read-back, import, no-change plans, drift detection/repair, update, deletion, and restoration where applicable. Unit/protocol success, Terraform Core behavior, disposable-appliance activation, release publication, and Registry availability are separate evidence levels.
- Maintain one provider owner per appliance object/domain in steady state. Migration must not leave two providers managing the same object. Guidance must identify unsupported fields and ownership blockers before users transfer management.
- Keep credentials out of logs and published examples; use sensitive/write-only behavior where appropriate and document sensitive state storage. Do not fabricate readable secrets or claim drift detection for values the API cannot return.
- Use supported APIs only: no legacy-page scraping, direct XML edits, or API-ownership bypass. System-general settings remain upstream-blocked; HASync model shape and PPPoE semantics remain research-only. Ordered Kea linkage requires an authoritative order-preserving read API, not a raw XML workaround.
- This epic does not authorize production firmware changes, routing changes, or appliance ownership cutovers. Public artifacts must stay consumer-neutral.

## Technical Decisions

- Preserve the safety contract: serialized mutations protect appliance configuration integrity; apply/reconfigure the affected service after changes and retain the appropriate firewall rollback safeguards. Populate state from authoritative API reads rather than request echoes, preserve recoverable state on partial failure, and distinguish successful persistence from failed activation.
- Model real API semantics, including singleton versus UUID-backed objects, multivalue membership, selected values, and order where meaningful. Search-backed lookups must not silently truncate paginated results. Read-only lookup semantics must be safe independently of resource write behavior.
- Keep upgrades compatible through deliberate plan modifiers, schema migration when required, and explicit handling of unavailable write-only values. Retry controls must configure transient-failure behavior without weakening mutation/state safety or appliance read-concurrency protection.
- Resource documentation, import instructions, support boundaries, and composition examples ship with the supported surface. Provider-facing documentation must explain authentication, permissions, required plugins, minimum supported appliance versions, and encrypted state considerations.

## UX & Interaction Patterns

- Position the provider around broad, modern appliance management, brownfield import, and release-gated support. Avoid competitive rhetoric and unsupported claims of total coverage.
- Guide brownfield users through dependency-aware import, cross-resource references, and a no-change plan before ownership transfer. Examples should demonstrate realistic composition rather than isolated resources.
- Diagnostics should identify the failed operation, API validation field or permission requirement, and a useful recovery action. Distinguish connectivity, TLS, authentication, unavailable endpoints, and missing plugins.

## Cross-Story Dependencies

- Audit findings drive parity work; complete remaining Trust, IPsec, FRR/BGP, interface-lookup, and retry gaps before final positioning, migration, and adoption examples. Trust settings alone do not close certificate/CA lifecycle work.
- FRR 2.0 compatibility is already released; advanced FRR/BGP parity remains separate. Gateway-group ownership and Kea generic-option/DDNS ownership are also shipped; preserve their limitations and avoid implementing duplicate owners. Ordered option linkage remains a separate API gate.
- Completed interface-assignment adoption requires OPNsense 26.7.6+; its bounded lifecycle proof does not imply PPPoE support or blanket proof for every interface mode.
- The acceptance-proof, preflight/version-diagnostic, and v1-stability work in Epic 32 builds on this epic. Breadth or successful parity increments alone do not establish v1 readiness.
