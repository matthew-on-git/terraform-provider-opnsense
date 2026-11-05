# Story 31.0h: Quagga/BGP Field Parity

## Status

in-progress — compatibility subgate merged in PR64 and released in v0.7.0; numeric 2.0 package-revision eligibility fixed in PR69/v0.8.1. Broader advanced parity remains backlog.

## Priority

P0 compatibility gate after active Story 31.0e; remaining field parity stays P1. Complete before Stories 31.1–31.3 and any production FRR 2.0 provider adoption.

## Evidence and Scope

Extend the existing BGP neighbor and route-map owners. Neighbor gaps include BFD, authentication, connect timer, default route, multihop, route-reflector, weight, connected-check, attribute preservation, and next-hop-self-all. Route-map gaps include AS-path/community/prefix-list references and route-map ID semantics. Use existing stable `quagga/bgp` item CRUD/toggle/reconfigure endpoints and retain the broader local FRR surface.

## FRR 2.0 Compatibility Gate (2026-10-20)

OPNsense 26.7.7 includes `os-frr` 2.0. The pinned [plugin migration notes](https://github.com/opnsense/plugins/blob/5ac2959c8514b4f60e94d65439be5f629f10521f/net/frr/pkg-descr) replace neighbor `multiprotocol` with explicit `family`, move peer-group policies into address-family configuration, group policy references, tighten validation and add automatic routing-protocol firewall rules. The compatibility implementation adapts the existing owners using installed-package/model capability checks, preserves boolean configurations and version-0 identities, and handles grouped policy references. This is not a claim that every 26.7.7 appliance has the same plugin version or that all advanced parity is complete.

Entry point: inventory core **and installed os-frr versions** on the disposable target; inspect pinned `net/frr/.../Quagga/BGP.xml`, BGP migrations, `Api/BgpController.php`, service/reconfigure and firewall templates before editing `internal/service/quagga/`. Keep existing Terraform type names and state addresses. Extend the existing neighbor, peer-group, route-map and policy owners rather than adding parallel resources.

1. Establish explicit version/capability selection for old and 2.0 plugin payloads. Do not infer plugin version from core version alone or silently drop unsupported configured fields.
2. Model address-family selection without losing IPv6-only or dual-stack intent; define upgrade/state handling for existing `multi_protocol` configurations, unknowns and imports. Unsupported combinations must fail before writes with actionable diagnostics.
3. Re-audit grouped AS-path/community/prefix-list/route-map relationships, peer-group address-family policy and reference ordering. API route existence alone is not parity proof.
4. Check generated automatic firewall rules and effective routing policy on the isolated target, including no unintended default-route export; no production routing/firewall changes are authorized by this provider story.
5. Preserve linked policy identities and strict reload failure reporting. If old-plugin acceptance is unavailable, leave that support claim explicitly unverified rather than replacing its proof with a mocked 2.0 response.

## Acceptance Criteria

- **Given** advanced neighbor/route-map objects, **when** imported, **then** all audited fields read back and matching configuration plans no change.
- **Given** linked policy objects, **when** updated or drifted, **then** UUID relationships and ordering remain stable through reconfigure.
- **Given** existing local-only FRR resources, **when** parity fields ship, **then** registrations and state addresses remain backward compatible.

- **Given** supported legacy and FRR 2.0 plugin versions, **when** create/import/read/update and external drift run, **then** the same existing Terraform resource addresses converge with correct family and policy semantics and no perpetual diff; record version-specific non-skipped proof.
- **Given** an existing `multi_protocol` state or an unsupported plugin capability, **when** planning upgrades/imports, **then** state handling is deterministic and unsupported intent cannot be silently discarded or written.
- **Given** a 2.0 configuration with linked policies and protocol auto rules, **when** reconfigure runs, **then** generated configuration, route filters, effective firewall rules, ordering and failure diagnostics are verified on a disposable appliance and exact cleanup restores baseline.

## Verification

Quagga unit/schema/generated tests, real Terraform import/no-change/drift/state-upgrade tests, version-specific disposable BGP/policy/reconfigure/cleanup proof, docs generation, and `make check`. Record exact core/plugin versions, source digest and unverified rows; do not claim production compatibility from source inspection alone.

## Compatibility implementation boundary

- Existing registrations and individual policy-entry UUID ownership are unchanged.
  Neighbor `family`/legacy alias resolution and offline schema-v1 upgrade are additive;
  peer-group policy fields remain flat; route-map AS-path/community relations retain
  `match`/`match3`, and `order` remains API sequence `id`, not resource UUID.
- Owner-local strict reads, mutation/delete results, `status=ok` service responses,
  serial discovery/write/recovery, authoritative group membership and private pending
  activation protect the owners under test. Unexposed passwords are never read into
  public state or included as empty clears in payloads.
- FRR reconfigure is a soft reload. Fresh-lab BGP configuration precedes enabling
  General so watchfrr starts bgpd. An already-running daemon-set change may require a
  separately reviewed restart. PF activation is explicitly a separate lab operation;
  disabled generated-config restoration uses a VM/TLS-guarded host-only template sync.
  Neither operation is claimed as provider behavior.
- See [live evidence](31-0h-frr2-live-evidence.md) for the exact frozen digest and
  actually executed/non-skipped versions. Legacy live, OpenTofu and broader advanced BGP parity remain unverified.
  Compatibility publication is complete; operator-specific adoption is tracked privately.
- Historical WireGuard evidence belongs to its prior candidate; these FRR changes
  do not extend that proof to this digest. No coverage-audit row is promoted to
  complete without full row-level proof.

## Advanced parity increment

The advanced candidate adds all 20 neighbor fields and peer-group listen ranges
on the existing owners, with strict source enum/range validation, config-owned
sensitive password input and password-free lookups. Route-map sequence remains
numeric and UUID relationships remain unchanged. Released source hashes are
recorded in `31-0h-advanced-source-evidence.json`. Protocol/Core fixture evidence,
new disposable lifecycle evidence, release publication and Registry availability
are distinct gates. The complete current-appliance 17-step guarded lifecycle, generated/policy and collector matrix passed; see [sanitized proof](31-0h-advanced-live-proof.json). Parent status remains open for numeric-clear design, older-plugin/appliance and OpenTofu execution, and advanced authenticated/operational IPv6 sessions. Release/registry publication is a separate verified gate, not inferred from live acceptance.

### Remaining numeric-clear interface limitation

`local_as`, `connect_timer`, and `remote_as` are Optional+Computed numeric
properties. After adopting a non-null appliance value, omission or `null`
preserves that value; neither is a configured clear. There is currently no
configured numeric clear for these properties. No zero sentinel is introduced.
Designing an explicit clear interface remains an open parent Story 31.0h
follow-up; empty string/set clears described above apply only to those types.
