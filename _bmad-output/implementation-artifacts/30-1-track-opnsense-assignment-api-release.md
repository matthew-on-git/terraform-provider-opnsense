# Story 30.1: Track OPNsense Assignment API Release

## Status

review

## Story

As a provider maintainer,
I want an automated release/API tracker for the expanded OPNsense `interfaces/assignment` API,
so that we know exactly when interface IP configuration can be safely implemented for a supported release.

## Acceptance Criteria

1. A repeatable probe checks OPNsense `master`, active `stable/*` branches, and latest release tags for `NetworkInterface.xml` fields required by provider-managed interface settings.
2. The probe records whether `type4`, `ipaddr`, `type6`, `ipaddrv6`, gateway fields, DHCP fields, assignment `set_item`, and assignment `reconfigure` exist.
3. The probe distinguishes `master` evidence from stable release support.
4. Results update or feed `_bmad-output/planning-artifacts/support-matrix.md` and `docs/upstream-blocked.md`.
5. The tracker captures the first released tag/branch that contains the expanded API and links to the upstream commits.
6. No provider resource is marked stable until the API is present in a released OPNsense version or explicitly documented as snapshot-only.

## Dev Notes

- September 2026 finding: the duplicate upstream PR was closed because OPNsense `master` already has the assignment API expansion.
- The September check found that `stable/26.7`, `26.7.4`, and `27.1.a` did not contain the full expanded assignment API.
- OPNsense subsequently backported the expanded model/controller to `stable/26.7`; `26.7.6` is the first released tag found by the probe with the complete required surface.
- Do not create a workaround based on legacy `interfaces.php`, XML writes, or GUI scraping.

## Tasks

- [x] Add a deterministic probe for `master`, the newest stable branches, and recent release tags.
- [x] Record required IPv4, IPv6, gateway, DHCP, `set_item`, and `reconfigure` evidence.
- [x] Commit the current probe baseline and add weekly/manual drift detection.
- [x] Identify the first complete released tag and link the upstream implementation commits.
- [x] Refresh support, blocker, roadmap, migration, and sprint-status artifacts.

## Release Evidence

- First complete released tag: [`26.7.6`](https://github.com/opnsense/core/tree/26.7.6), tagged 2026-10-08.
- Stable backport commit: [`786c2e0`](https://github.com/opnsense/core/commit/786c2e0ffa2bea85f6386439e47e69709b203641).
- Original `master` expansion commit: [`92cb28a`](https://github.com/opnsense/core/commit/92cb28a6696d7ab3ffebb929560da1361b501192).
- `26.7.5` has assignment `set_item` and `reconfigure`, but not the required address, gateway, and DHCP fields.
- `26.7.6`, `26.7.7`, and `stable/26.7` have the complete probed surface.
- `27.1.a` predates the expansion and is not complete despite its higher version number.
- The committed machine-readable baseline is `_bmad-output/planning-artifacts/opnsense-assignment-api-release-probe.json`.
- The `Upstream Assignment API` workflow runs weekly, on manual dispatch, and for probe/baseline/workflow changes in pull requests or pushes to main. It fails when current upstream evidence differs from the committed baseline and uploads the fresh result for review.
- Stable API availability clears the upstream release gate, but provider support still requires Story 30.2 implementation and disposable-appliance acceptance proof. PPPoE behavior remains outside the first implementation scope until separately validated.

## Validation

- Probe output is deterministic and committed or stored as a planning artifact.
- `git diff --check` passes.
- If scripts are added, `make check` must pass before completion.

## Dev Agent Record

### Completion Notes

- Added a Go-based GitHub probe with unit tests and deterministic JSON output.
- Added scheduled/manual baseline drift detection with uploaded probe artifacts.
- Reclassified assignment/IP settings from upstream-blocked to provider-owned implementation work because OPNsense `26.7.6` is released with the required API surface.
- Kept the provider feature unsupported pending Story 30.2 implementation and acceptance validation.
- Verified repeat probe output is byte-identical to the committed baseline.
- `make docs` and full `make check` passed with DevRail `v1.12.24` (lint, format, test, security, scan, and docs).

### File List

- `.github/workflows/upstream-assignment-api.yml`
- `tools/opnsense-assignment-probe/main.go`
- `tools/opnsense-assignment-probe/main_test.go`
- `_bmad-output/planning-artifacts/opnsense-assignment-api-release-probe.json`
- `_bmad-output/implementation-artifacts/30-1-track-opnsense-assignment-api-release.md`
- `_bmad-output/implementation-artifacts/30-2-prototype-interface-assignment-ip-resource.md`
- `_bmad-output/implementation-artifacts/30-3-update-interface-support-matrix.md`
- `_bmad-output/implementation-artifacts/31-3-example-stacks-for-adoption.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `_bmad-output/planning-artifacts/provider-management-roadmap-2026-09-28.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/prd.md`
- `docs/upstream-blocked.md`
- `docs/index.md`
- `docs/migration-import.md`
- `templates/index.md.tmpl`

## Change Log

- 2026-10-17: Implemented the release tracker and recorded OPNsense `26.7.6` as the first release with the complete assignment/IP API surface.


### 2026-10-27 reviewed baseline refresh

- Reviewed the complete old/new `NetworkInterface.xml` blobs on `master` and `stable/26.7`, rather than accepting changed hashes without inspection. The change adds `<Default>0</Default>` to the `dhcp6-ia-pd-len` OptionField and requires it on stable; master relocates its existing Required marker. The provider assignment model does not expose or serialize this field.
- Reviewed model blob transitions: `master` `53d351938c60fa3334b123c9a5d13454fd60d65c` → `4fae64840308390c95857888e73dfca6136d8379`; `stable/26.7` `b43147511612454e292429d5961e90ea0ef0d122` → `85e25100479392d5d5af13b12eb72f5aa25d3290`.
- A fresh live probe differs from the old baseline only in those two model SHAs. Required-field detection, controller SHAs, actions, all tagged-release evidence and `first_released_tag = 26.7.6` are unchanged. Refresh the two reviewed hashes while preserving byte-for-byte comparison and all drift checks.
- This is monitoring evidence maintenance, not a new provider feature, firmware upgrade or appliance acceptance claim. No provider code, probe logic or release heading is changed. The workflow also runs for probe/baseline/workflow changes in PRs and on main, so reviewed baseline updates get hosted validation before and after merge without a separate manual dispatch. Weekly/manual triggers and exact drift comparison remain enabled.

### 2026-11-02 reviewed rolling-tag baseline refresh

- The rolling 12-tag window now includes OPNsense `26.7.8` (`96978001a8343669154e4492172ce0a07dac7e4b`) instead of `26.7.r`. Every common ref and all non-result report metadata remain unchanged.
- Retrieved and verified both new-tag source blobs: model `85e25100479392d5d5af13b12eb72f5aa25d3290` and controller `61193fc27a8c0fa77a0c77b5c708bb371d92c1ba`. Their complete contents match the previously reviewed `stable/26.7` blobs byte-for-byte; all 18 watched fields and both actions are present.
- Preserve explicit boundary evidence: `26.7.5` remains incomplete, `26.7.6` remains complete and the first released tag. Exact baseline comparison, rolling-window size, weekly checks and PR/main triggers are unchanged.
- Validate with probe unit tests, two independent live upstream probes matching the refreshed baseline byte-for-byte, and the full local `make check` gate. This refresh changes monitoring evidence only, not provider behavior or support claims; no release heading or live-appliance acceptance claim is added.
