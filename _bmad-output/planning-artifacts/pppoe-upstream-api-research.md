# PPPoE / Point-to-Point API Research

Checked: 2026-11-24. Provider baseline:
`84d138d20568e959c63e4ed83a68e909f099ba25`.

Consumer request: [provider #115](https://github.com/matthew-on-git/terraform-provider-opnsense/issues/115).
Backlog story: [Story 31.4](epics.md#story-314-pppoe-device-lifecycle-and-assignment-integration).

## Upstream does track preparatory work

[OPNsense/core #10973](https://github.com/opnsense/core/issues/10973),
“interfaces: remove deprecated PPTP and L2TP support from PPP devices”, is open
and assigned to [milestone 27.1](https://github.com/opnsense/core/milestone/27).
Its body explicitly proposes removal “to be able to make a clean migration to
MVC/API via road map item”. It also tracks local/remote address migration away
from port-based/multi-link handling. The milestone currently has a 2027-01-27 due
date; this is not a promise that a PPP lifecycle API will ship then.

The [official roadmap](https://opnsense.org/roadmap/) lists “Deprecate PPTP and
L2TP support from PPP devices” under the next release, January 2027 / 27.1.
Do not turn this cleanup entry into a claim that PPPoE CRUD endpoints are already
specified, implemented or released. No separate definite PPP API delivery PR was
identified by the searches below.

[Core PR #11252](https://github.com/opnsense/core/pull/11252) merged on
2026-11-21. Its point-to-point fix corrects legacy `ppp_inuse()` checks to prevent
deleting an assigned PPP device. It is relevant dependency/teardown evidence,
not a new PPP lifecycle API. Review the released version carrying this behavior
before designing destroy semantics; do not infer it from a closed PR alone.

## Source boundaries checked

Stable core tag `26.7.9` is annotated “stable release”, peeled commit
`228ab68d54513046f79bc36d7f583936922d1951`. Development tags `27.1.a` / `27.1.b`
are not stable-release contract evidence. Master snapshot checked on this date:
`e355e8e407c5393596dffaac1911bad3163b1ac8`.

- [AssignmentController.php, stable](https://github.com/opnsense/core/blob/228ab68d54513046f79bc36d7f583936922d1951/src/opnsense/mvc/app/controllers/OPNsense/Interfaces/Api/AssignmentController.php)
  uses the `NetworkInterface` model for assignment CRUD. It is not CRUD for
  legacy PPP device definitions.
- [NetworkInterfaceField.php, stable](https://github.com/opnsense/core/blob/228ab68d54513046f79bc36d7f583936922d1951/src/opnsense/mvc/app/models/OPNsense/Interfaces/FieldTypes/NetworkInterfaceField.php)
  implements `pppType()` by reading existing `Config::getInstance()->object()->ppps`
  records. Recognition of a matching device is not device provisioning.
- [interfaces_ppps_edit.php, stable](https://github.com/opnsense/core/blob/228ab68d54513046f79bc36d7f583936922d1951/src/www/interfaces_ppps_edit.php)
  manages separate `ppps/ppp` settings, including credential handling, through the
  legacy form and configuration write path.
- The [issue's NetworkInterface.php link](https://github.com/opnsense/core/blob/4451c201806b9e086d425fa8ee9b1f5fa3a29210/src/opnsense/mvc/app/models/OPNsense/Interfaces/NetworkInterface.php#L259)
  validates PPP/PPPoE assignment types against the device; it does not create the
  underlying PPP device.
- The [master Interfaces API controller tree](https://github.com/opnsense/core/tree/e355e8e407c5393596dffaac1911bad3163b1ac8/src/opnsense/mvc/app/controllers/OPNsense/Interfaces/Api)
  still has no dedicated PPP-device controller. Stable/master Assignment and
  Settings controllers inspected do not supply that separate lifecycle. This is
  bounded source research, not proof that no future API or third-party plugin
  could provide it.

The provider's assignment schema and refresh guards currently reject PPPoE.
Neither a manually created PPPoE assignment nor an enum-only change is supported
full ownership. The existing base assignment implementation remains supported;
this gate applies to the separate point-to-point extension.

## Search record and next action

GitHub issues/PRs, all states, were searched in `opnsense/core` for `PPPoE API`,
`PPP MVC`, `PPP migration`, `point-to-point`, and PPP/PPPoE title variants;
organization-wide `point-to-point MVC` results were cross-checked. #10973 is the
explicit migration planning signal; #11252 is a relevant legacy safety fix.
Runtime PPP bugs, HA behavior and deprecated server plugins are not substitutes
for a PPPoE client configuration API contract. Search results are not an exhaustive
or permanent promise about upstream plans.

Next: re-read #10973 and linked work when new implementation or release evidence
appears. Verify device CRUD, authoritative read, secrets, identity, in-use delete
and activation on a stable tag and disposable appliance before opening provider
writes. Keep Story 31.4 blocked until then. Do not submit a duplicate upstream
request without rechecking existing tracking and maintainer discussion.

No appliance API calls, mutation, packet tests or credential processing were
performed for this research. No support or release claim is added by this file.

## Future Implementation Preparation

These notes supplement canonical Story 31.4 in `epics.md`; they are not an
implementation story or readiness approval. Generate an implementation handoff
under the configured `implementation_artifacts` only after the upstream gate and
scope/readiness review pass.

### Preparation Checklist

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

### Technical Notes

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

### Resume and Follow-up Rule

Owner: provider maintainer for gate triage, then an explicitly assigned developer
for implementation. Resume on a linked upstream API implementation or new stable
release, not merely a calendar date. Before promoting to ready-for-dev, satisfy
AC1 and review AC2. If an authoritative read, secret-preservation operation,
identity, activation or restoration requirement is still absent, retain this
story as blocked and add a narrowly scoped upstream/follow-up story with exact
missing behavior and source evidence. Do not mark this feature done because the
research document or an upstream preparatory issue was closed.
