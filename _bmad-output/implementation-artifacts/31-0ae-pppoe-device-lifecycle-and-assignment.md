# Story 31.0ae: PPPoE Device Lifecycle and Assignment Integration

Status: blocked

Priority: P1 — provider coverage gap; not ready for implementation.

Epic: 31 — Provider Completeness, Positioning, and Migration

Origin: [provider issue #115](https://github.com/matthew-on-git/terraform-provider-opnsense/issues/115).

## Story

As an operator managing API-backed network configuration with Terraform or OpenTofu,
I want to create, import, update and delete a PPPoE client device and configure its
interface assignment without losing credentials or unrelated settings,
so that point-to-point connectivity can be managed declaratively with predictable
identity, drift detection, activation and recovery.

## Dependency and entry point

This story is blocked on a supported, released upstream PPP-device lifecycle API.
An assignment mode enum or an upstream model that recognizes PPPoE is not that API.

Read [the upstream research](../planning-artifacts/pppoe-upstream-api-research.md)
first. Upstream [core #10973](https://github.com/opnsense/core/issues/10973)
explicitly connects PPP cleanup to a future MVC/API migration and is assigned to
milestone 27.1. Its milestone is a planning signal, not an API delivery guarantee.
No specific PPP-device CRUD endpoint was verified in stable core 26.7.9 or the
inspected master snapshot. Do not invent endpoint paths or enable provider writes
from a roadmap entry, a development tag, a closed GUI fix or a merged model alone.

The initial research is complete; provider implementation and acceptance are not.
The next worker starts with the upstream gate below, not provider code changes.

## Acceptance criteria

### AC1 — Upstream tracking and released contract

Given #115 and upstream #10973, when this story is resumed, then inspect their
current discussion, the 27.1 milestone, official roadmap, linked implementation
PRs and stable core tags; record exact commits and source links. Keep status
blocked until a stable release exposes authoritative PPP-device create/read/
update/delete and activation behavior. A PPPoE assignment validator does not meet
this criterion. If the milestone slips or scope changes, update the dependency
rather than substituting a legacy GUI request.

### AC2 — Contract and ownership design before writes

Given a released API candidate, when designing support, then document actual
HTTP methods, request/response envelopes, stable identity, option maps, parent
port/VLAN references, authentication fields, empty/null/unknown semantics,
activation, error recovery, in-use deletion and assignment dependencies. Verify
that PPP-device definitions and interface assignments are distinct owners.
Propose a dedicated PPPoE owner on the actual API, extending the existing
`opnsense_interface_assignment` for supported modes rather than creating a second
assignment owner. Public Terraform resource names remain provisional until design
review. Reconcile against current main, open PRs and released resources first.

### AC3 — Identity, lifecycle and backwards compatibility

Given a supported PPPoE device and existing assignment, when creating, importing,
reading, changing, refreshing or deleting them, then identity remains stable and
subsequent plans converge. Preserve unconfigured API fields and existing supported
assignment modes. Validate supported parent devices and reject unavailable
capabilities before mutation. Import must not create a duplicate device. Default
destroy must not silently remove an in-use device or an adopted assignment; define
and test the release-backed delete/detach policy explicitly.

### AC4 — Credentials and configured intent

Given secret/redacted/write-only upstream fields, when importing or refreshing,
then absence or redaction must not become a configured empty password or cause a
credential-clearing write. Mark secret schema fields sensitive, sanitize diagnostic
bodies, and document Terraform state exposure. Prove deliberate rotation and any
API-supported explicit clear separately from omission/unknown intent. Preserve
unrelated credentials and reject ambiguous operations before writes. Do not add
fixture secrets from an appliance, operator inventory or real account.

### AC5 — Deterministic regression and Core coverage

Given valid and malformed released-contract fixtures, when tests run, then cover
schema/registration, mapping, import, no-change plans, refresh, drift, parent
changes, secret redaction, explicit false/empty values, partial failures, retries,
activation failure and dependency-aware deletion. Use Terraform Core and
OpenTofu lifecycle tests, not only protocol mocks. For create-recovery, a counted
fake API must prove a failed post-create read/activation does not produce duplicate
creates or unsafe cleanup. Keep existing assignment regression coverage passing.

### AC6 — Disposable activation and restoration proof

Given a disposable appliance, isolated non-management interfaces, independent
console access and a controlled PPPoE peer, when the owner and assignment are
applied and changed, then independent API/runtime observations and packet tests
prove discovery, negotiation, activation, device/address placement, convergence,
drift repair, credential rotation and safe teardown. Capture firmware/plugin,
Terraform/OpenTofu and provider revisions. Restore exact prior configuration and
runtime state; verify no leftover devices, assignments, routes or credentials.
Do not use a management or production WAN. Failure to restore blocks completion.

### AC7 — Documentation, release and evidence separation

Given completed implementation and acceptance, when preparing delivery, then
update resource/data-source documentation, examples, support matrix and coverage
audit without overstating versions or packet proof. Run `make check`, generated
code/docs verification and exact-head hosted checks. Record independent source,
fixture, Core, appliance and contributor-reported evidence separately. Use
conventional commits and the existing Release Please flow; do not select a manual
version or insert an Unreleased heading. Human merge/publication is separate from
implementation acceptance. This planning story does not ship PPPoE support.

## Tasks / subtasks

- [x] Reconcile #115 with current provider assignment restrictions and backlog.
- [x] Search upstream planning/issues/PRs and inspect stable and master boundaries.
- [x] Link #10973 / milestone 27.1 and distinguish preparatory work from API delivery.
- [ ] Recheck tracking after upstream changes; locate actual implementation PRs.
- [ ] Verify stable API and appliance contract (AC1); record endpoint/field table.
- [ ] Review separate device/assignment ownership and migration design (AC2).
- [ ] Implement lifecycle, capability and secret safeguards (AC3–AC4).
- [ ] Add deterministic regressions and Core/OpenTofu lifecycle checks (AC5).
- [ ] Complete disposable activation, packet and exact-restoration proof (AC6).
- [ ] Complete documentation, coverage and exact-head delivery gates (AC7).

## Dev notes

Existing files to inspect: `internal/service/iface/assignment_resource.go`,
`assignment_data_source.go`, `assignment_controls.go`, their unit/protocol tests,
`docs/resources/interface_assignment.md`, and its generation template. The
resource currently rejects `pppoe`; lookup refresh of unsupported modes also
fails closed. Keep those guards until the new contract and semantics are proven.
Do not bypass them by treating PPPoE as DHCP or static IPv4.

The PPP-device record is currently written by legacy `interfaces_ppps_edit.php`;
`NetworkInterfaceField::pppType()` reads existing `ppps` configuration to validate
an assignment. Reusing its allowed type does not provision the device.

Non-goals: PPTP/L2TP server or client revival, multi-link PPP, 6rd/6to4, generic GUI
scraping, config.xml uploads, SSH-backed Terraform resources, real WAN cutovers,
and unrelated interface-field expansion. PPPoE IPv6, service-name, advanced
negotiation and multi-link behavior require explicit contracts and scope review;
do not silently claim them as part of basic PPPoE support.

## Resume and follow-up rule

Owner: provider maintainer for gate triage, then an explicitly assigned developer
for implementation. Resume on a linked upstream API implementation or new stable
release, not merely a calendar date. Before promoting to ready-for-dev, satisfy
AC1 and review AC2. If an authoritative read, secret-preservation operation,
identity, activation or restoration requirement is still absent, retain this
story as blocked and add a narrowly scoped upstream/follow-up story with exact
missing behavior and source evidence. Do not mark this feature done because the
research document or an upstream preparatory issue was closed.

## Dev Agent Record

Planning-only artifact. No provider code, appliance calls or lifecycle tests were
executed as evidence for this story. Implementation file list and runtime proof
remain to be filled by the implementing agent. Planning PR validation is reported
in that PR and does not satisfy the feature's acceptance criteria.
