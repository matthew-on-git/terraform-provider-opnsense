# Epic 31 Context: Provider Completeness, Positioning, and Migration

<!-- Generated from planning artifacts. Regenerate with compile-epic-context if planning docs change. -->

## Goal

Make the provider complete enough to be the default choice for a supported appliance, then translate verified coverage into trust and adoption. Close meaningful field, lifecycle and lookup gaps before expanding positioning, migration guidance and example stacks; breadth alone is not evidence of completeness.

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

- Use the dated support matrix and coverage audit to reconcile historical backlog labels. The November 2026 baseline is published v0.11.0 with 112 resources, 98 data sources and one ephemeral resource; 16 supported resources still lack same-name lookup counterparts. Counts do not establish field or lifecycle parity.
- Do not repeat completed firewall/NAT, subnet/PXE/DHCPv6, OpenVPN or Unbound parity. WireGuard parity/settings and OpenVPN ephemeral generation shipped in v0.7.0; gateway groups shipped in v0.9.0; generic Kea options and DDNS owners/lookups shipped in v0.10.0. Older review/backlog labels are not evidence that these owners remain unimplemented.
- Remaining audited parity groups are trust lifecycle/settings, advanced IPsec connection/child controls, advanced BGP neighbor/route-map controls, live interface lookups and configurable retry/backoff. Trust includes CA/certificate actions and a settings singleton, not merely additional certificate fields. FRR 2.0 compatibility already shipped; advanced BGP parity remains distinct.
- Keep stable API availability, provider implementation, published distribution and appliance lifecycle evidence separate. Disposable 26.7.7 configuration proof does not establish active tunnel traffic, multi-WAN failover, DHCP packet behavior, DNS delivery, older-core support or OpenTofu-specific lifecycle parity. v0.11.0 IPv6 delegation adoption has contributor-reported import/no-change evidence, not independently reproduced activation proof.
- Ordered Kea parent linkage remains gated on authoritative ordered reads; label-sorted relation choice maps cannot establish persisted order. Generic-option ownership alone does not close linkage or packet-level force semantics. Gateway-group populated first-tier clearing remains upstream-limited and must fail before saving other fields.
- System general settings remain upstream-blocked; HASync configuration needs model-shape research; PPPoE remains unverified. No legacy-page scraping, direct XML edits or unsupported private-API workarounds.
- Public docs, examples and proof must be consumer-neutral, using reserved domains and documentation address ranges. Exclude credentials, operator inventory, topology and adoption plans. Provider development does not authorize production firmware, routing or ownership cutovers.

## Technical Decisions

- Manage only durable, supported OPNsense API contracts and gate features on released core/plugin versions. Base assignment/IP support requires OPNsense 26.7.6+. FRR compatibility must inspect installed plugin/model semantics; existing proof covers os-frr 2.0_1, while numeric package-revision expansion has fixture/protocol coverage rather than a new appliance run.
- Preserve truthful Terraform state: read back after mutations, detect drift and missing objects accurately, retain stable identities, and make repeated apply converge without false changes. Import, updates, explicit clears, omitted-field preservation, recovery and cleanup need evidence appropriate to each surface.
- Serialize mutations to protect shared appliance configuration and complete the affected service's apply/reconfigure lifecycle. Parse body-level API failures even on HTTP 200; an in-use HTTP500 is not successful deletion or absence.
- Represent API collections and selected-map values losslessly. Treat runtime interface observations as read-only state, separate from durable assignment configuration. Preserve existing configuration/state compatibility when introducing canonical typed fields.
- Keep private keys and other write-only secrets sensitive and resource-only; blank/omitted reads must not erase valid prior secret state. Import-first singleton ownership must be bounded to the supported subtree, with state-only destroy rather than resetting shared settings.
- Evidence must identify the tested version and scope, preserve failures as failures, and demonstrate restoration/zero leftovers. Documentation and release verification must distinguish source checks, protocol/fixture tests, live lifecycle tests, tag/release assets and Registry publication.

## UX & Interaction Patterns

- Users evaluate coverage through README and Registry docs, configure via HCL, inspect plan output, then import existing objects before applying changes. Migration should follow dependency order and reach a no-change plan before transferring ownership.
- Use one provider per appliance/object domain in steady state; never have two providers manage the same object. Explain semantic renames and meaningful unsupported behavior before migration.
- Position broad modern core/plugin coverage only where evidence supports it. Provide generated schema/import documentation, composition examples and actionable version/plugin/permission diagnostics without blanket completeness or production-readiness claims.

## Cross-Story Dependencies

- The coverage audit scopes remaining parity work; reconcile released results before implementing or updating adoption claims. Complete remaining P0/P1 parity before Stories 31.1–31.3.
- Stable-API validation informed the already-released gateway-group and Kea owners; ordered parent linkage remains a separate follow-up. Earlier completed DHCP and assignment work must be reused rather than duplicated.
- Keep Epic 30's assignment release gate current. Coordinate v1 credibility with Epic 32 acceptance-proof artifacts and stability criteria; publication alone does not satisfy that proof gate.
