# Story 31.0i: Live Interface Data Sources

## Status

review — implementation and independent reviews complete; final six local gates pass. Maintainer review/merge and exact-head hosted CI are separate publication gates. Historical live proof and repaired-source replay remain distinct.

## Priority

P1 — complete before Stories 31.1–31.3.

## Evidence and Scope

Add read-only equivalents for the pinned `interface`, `interface_all`, `interfaces_overview`, and `interfaces_overview_all` data sources. These do not replace `opnsense_interface_assignment`, which owns durable assignment/IP configuration.

Exact pinned endpoint/client evidence:

- `interface` calls `Diagnostics().GetInterface(ctx, device)`, which filters `GET /api/diagnostics/interface/getInterfaceConfig`; pinned provider source: [`interface_data_source.go`](https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/internal/service/diagnostics/interface_data_source.go), pinned client source: [`diagnostics/interface.go`](https://github.com/browningluke/opnsense-go/blob/9cd77265c81e329ba6d4399defcbf884c61ac75f/pkg/diagnostics/interface.go).
- `interface_all` calls `Diagnostics().GetInterfaceAll(ctx)` against the same `GET /api/diagnostics/interface/getInterfaceConfig` endpoint without a device filter; pinned provider source: [`interface_all_data_source.go`](https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/internal/service/diagnostics/interface_all_data_source.go), pinned client source: [`diagnostics/interface.go`](https://github.com/browningluke/opnsense-go/blob/9cd77265c81e329ba6d4399defcbf884c61ac75f/pkg/diagnostics/interface.go).
- `interfaces_overview` calls `Interfaces().OverviewGet(ctx)`, then selects the requested device from `GET /api/interfaces/overview/interfaces_info`; pinned provider source: [`overview_data_source.go`](https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/internal/service/interfaces/overview_data_source.go), pinned client source: [`interfaces/overview.go`](https://github.com/browningluke/opnsense-go/blob/9cd77265c81e329ba6d4399defcbf884c61ac75f/pkg/interfaces/overview.go).
- `interfaces_overview_all` calls the same `Interfaces().OverviewGet(ctx)` client and returns every row from `GET /api/interfaces/overview/interfaces_info`; pinned provider source: [`overview_all_data_source.go`](https://github.com/browningluke/terraform-provider-opnsense/blob/33c2a03c2bc1a7339cb5aa5b611e15014e27f7ba/internal/service/interfaces/overview_all_data_source.go), pinned client source: [`interfaces/overview.go`](https://github.com/browningluke/opnsense-go/blob/9cd77265c81e329ba6d4399defcbf884c61ac75f/pkg/interfaces/overview.go).

The pinned models cover device status, flags, media, capabilities, addresses, MTU, gateways/routes, VLAN/LAGG details, and DHCP-learned values.

## Acceptance Criteria

- **Given** an exact interface device, **when** the single-item data sources read, **then** runtime and overview fields match the API without claiming configuration ownership.
- **Given** multiple interfaces, **when** aggregate data sources read, **then** ordering and nested values are deterministic and unknown/disappearing devices produce actionable diagnostics.
- **Given** `interface_assignment`, **when** both surfaces are configured, **then** only the resource mutates configuration and the data sources remain read-only.

## Verification

Interface data-source unit tests, live read-only validation on a supported appliance, docs generation, and `make check`.


## Implementation checkpoint

Four handwritten read-only data sources share a strict allowlisted decoder and
bounded GET client. Optional scalars/objects retain typed null; explicit false,
zero and empty strings survive. Address tuple ordering and device ordering are
deterministic. Overview pagination, JSON types and duplicate identities fail
closed. Source spelling is mapped explicitly, including dotted DHCP keys,
`nd6.flags`, CARP fields and LAGG statistics. Unknown configuration/wireless data
never enters state or response diagnostics. Assignment resources are unchanged.

The no-redirect canary proved that setting CheckRedirect only on the outer
standard client is insufficient because retryablehttp follows redirects inside
its own HTTP client. The additive observation client rejects redirects at both
layers while reusing authenticated TLS transport and configured retries; existing
CRUD client semantics are unchanged. Fixture/Core checks do not establish actual
appliance support and parent Trust/IPsec/FRR limitations remain open.

## Private live acceptance path

`TestAccLiveInterfacesReadOnly` requires `OPNSENSE_LIVE_INTERFACE_ACCEPTANCE=1`
and private `OPNSENSE_LIVE_INTERFACE_URL`, `OPNSENSE_LIVE_INTERFACE_KEY`,
`OPNSENSE_LIVE_INTERFACE_SECRET`, `OPNSENSE_LIVE_INTERFACE_DEVICE` variables.
Run in the pinned toolchain container with the real Terraform binary; TLS trust
can be provided with `SSL_CERT_FILE`. The harness restricts forwarding to the
three read-only provider/menu and observation routes, counts zero rejected
mutation/other requests, and exercises all four sources plus no-change planning.
It does not print raw inventories. Core temporary state/test logs still belong
in a private artifact area and must never be committed or attached publicly.
Record the actual version independently and compare all public state fields to
fresh API truth independently; this identity/routing harness is not that oracle.
Skipping this opt-in test does not count as live proof.


## Local verification receipt

The public working-tree candidate passed focused race tests for interface
fixtures and non-skipped actual Terraform Core controls, shared client retry/
cancellation tests and coverage-audit regressions in
`ghcr.io/devrail-dev/dev-toolchain:1.12.24`. The full pinned `make _check` passed
all six categories in 68.801 seconds (lint, format, test, security, scan, docs;
no skipped categories). `git diff --check` was clean. Two successive
`go generate ./...` inventories matched exactly, with SHA256
`c002c20cae7566e569f2fa618f39d7b4c57bbeb3cb3d26b31ba62c872a459f90`
for the sorted generated-file digest inventory.

This receipt is local fixture/Core/generation evidence, not a live-appliance
result. The opt-in live harness was not run by the implementer. Actual version,
independent API/state field comparison, exact tested source and zero upstream
mutations must be recorded by the live-validation owner; hosted exact-head CI,
PR publication and release remain separate gates.

## Fresh live read-only verification — 2026-11-05 UTC

The first actual live Core attempt failed because a released API response repeated
route strings and Framework set construction rejects duplicate input. That failure
was retained independently. The repair deduplicates exact members in every public
set (including nested sets), without trimming, splitting or filtering blanks;
`TestLiveInterfaceDuplicateSetMembers` covers the source-shaped counterexample
with reserved documentation values.

A fresh attempt against OPNsense 26.7.8 passed the opt-in actual Terraform Core
harness, CLI data-only apply, refresh-only apply and detailed-exit-code no-change
plan. An independent oracle derived from pinned public schemas compared every
allowlisted field for all four data sources against fresh API responses after
both apply and refresh. All comparisons passed. The audited upstream traffic
comprised 46 GET requests and zero mutation requests; denied unauthenticated/POST
canaries made no upstream call. This was not appliance fixture replay. The exact
private authenticated proxy process and listener were verified retired.

Toolchain: `ghcr.io/devrail-dev/dev-toolchain:1.12.24`, Go 1.26.10,
Terraform 1.16.5, linux/arm64. Focused race tests and non-skipped Core controls
passed after the repair; all six full-check categories passed in 71.528 seconds.
Inventory, state, credentials and raw responses remain private. This is bounded
read-only proof, not older-firmware, OpenTofu, unobserved hardware, operational
traffic, release or parent-backlog completion evidence. The live source digest
and matrix audit are recorded in the companion verification artifact.

## Delivery continuation — 2026-11-10 UTC

Three independent frozen-patch review results have been collected. The spec
records every verdict and the single deferred context-file follow-up. Repairs
preserve public fields, duplicate-set handling, generated docs/examples and the
coverage audit. An independent repair reviewer checked the affected changes.
Final source/evidence binding and the diagnosed historical scanner false positive
are recorded in `31-0i-verification.md`; do not equate recorded-response replay
with new live appliance validation or claim release readiness while gates fail.

## Final local delivery checkpoint

Final pinned `make check`: all six categories passed in 71.667 seconds, none
skipped. Focused race/Core/client/audit checks, deterministic generation twice,
and whitespace validation passed. Independent repair review and its follow-up
confirmed no remaining defects in the affected scope. A diagnosed historical
scanner false positive uses only the repository-prescribed exact fingerprint;
no credential rule or scanner is disabled. Public diff inspection excludes
private review handoffs, raw responses/state, credentials and operator inventory.

The implementation/spec is done; the story is `review`, not merged/released.
Hosted readiness must be established against the actual PR head and recorded in
the PR/maintainer handoff. This local checkpoint is not hosted CI proof. Existing
Trust/IPsec/FRR parent limitations and the fifteen lookup gaps remain open.
