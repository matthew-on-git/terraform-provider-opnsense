---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 26.2: Public Support Matrix in Registry Docs

Status: done

## Story

As an operator evaluating the provider,
I want a clear public support matrix in the Terraform Registry provider docs,
so that I can distinguish Supported, Coming, and Upstream-blocked OPNsense domains before adopting the provider.

## Acceptance Criteria

1. Provider Registry docs include a concise Supported / Coming / Upstream-blocked matrix that is visible on the provider index page.
2. Matrix reflects the current post-Epic-25B baseline: 90 supported resources, 76 supported data sources, 90 resource docs, 76 data-source docs, and 15 remaining resource-matching data-source gaps.
3. Supported sections clearly distinguish implemented resources from implemented data sources, so users do not assume data-source parity is complete.
4. Coming items identify provider-owned follow-up work without promising delivery dates or implying the features already exist.
5. Upstream-blocked items identify the OPNsense-side missing API, upstream dependency, or verification need that prevents provider implementation.
6. Generated `docs/index.md` is updated from `templates/index.md.tmpl` through the repository documentation generation path, not edited as a divergent source of truth.
7. Any release-facing docs that still mention the old 34-data-source baseline are updated or explicitly scoped as historical v0.1.0 release notes.
8. The story does not add new resources, new data sources, migration-guide content, or broad per-resource import guidance; those belong to Stories 26.3, 27.1, and 27.2.
9. `make check` passes.

## Tasks / Subtasks

- [x] Task 1: Reconfirm documentation baseline (AC: 2, 7)
  - [x] Count current registered resources and data sources from `internal/service/*/exports.go`.
  - [x] Count current generated resource and data-source docs from `docs/resources/*.md` and `docs/data-sources/*.md`.
  - [x] Confirm the expected baseline remains 90 resources, 76 data sources, 90 resource docs, 76 data-source docs, and 15 resource-matching data-source gaps.
  - [x] Search release-facing docs for stale `34 supported data sources`, `55 data sources`, `36 remaining`, or `57` data-source-gap wording and classify each hit as historical or stale.
- [x] Task 2: Harden provider index support matrix (AC: 1, 3, 4, 5, 8)
  - [x] Update `templates/index.md.tmpl` support matrix section so it is concise, public-facing, and aligned with `_bmad-output/planning-artifacts/support-matrix.md`.
  - [x] Include current counts and clear labels for Supported, Coming, and Upstream-blocked.
  - [x] Add enough domain-level detail for users to understand what is supported without duplicating the full planning artifact.
  - [x] Keep migration/import guidance brief; do not expand into the full-appliance guide in this story.
- [x] Task 3: Regenerate Registry docs (AC: 6)
  - [x] Run the repository docs generation path used by `make check` so `docs/index.md` is regenerated from `templates/index.md.tmpl`.
  - [x] Confirm generated `docs/index.md` contains the same support matrix content and no template-only placeholders.
- [x] Task 4: Update release-facing docs only where needed (AC: 7, 8)
  - [x] Update `README.md`, `RELEASE.md`, or planning docs only if wording is stale and public-facing.
  - [x] Preserve historical release-note statements where they explicitly describe v0.1.0 at release time; add clarification only if a reader would confuse them with the current baseline.
  - [x] Do not edit generated resource/data-source docs except as a direct consequence of provider index regeneration.
- [x] Task 5: Validate (AC: 9)
  - [x] Run targeted searches for support matrix terms and stale counts.
  - [x] Run `make check` and fix all failures without suppressing checks.

### Review Findings

- [x] [Review][Patch] Update PRD FR60 from the obsolete 34-data-source baseline to the current 76-data-source / 15-gap baseline [_bmad-output/planning-artifacts/prd.md:576]
- [x] [Review][Patch] Clarify `RELEASE.md` 34-data-source count as historical v0.1.0 positioning [RELEASE.md:19]
- [x] [Review][Patch] Align Kea DHCPv4 option and Kea DDNS status with no-API evidence [_bmad-output/planning-artifacts/core-config-gap-analysis.md:119]
- [x] [Review][Patch] Replace provider index internal support-matrix path with a public repository link [templates/index.md.tmpl:99]
- [x] [Review][Patch] Update post-release recommended sequence now that Epic 25B is complete [_bmad-output/planning-artifacts/post-release-epics.md:56]

## Dev Notes

### Current Baseline and Source of Truth

Use these current post-Epic-25B counts unless implementation truth proves otherwise:

| Capability | Count | Source |
|---|---:|---|
| Supported resources | 90 | `_bmad-output/planning-artifacts/support-matrix.md` and constructor/docs counts |
| Supported data sources | 76 | `_bmad-output/planning-artifacts/support-matrix.md` and constructor/docs counts |
| Resource docs | 90 | `docs/resources/*.md` |
| Data-source docs | 76 | `docs/data-sources/*.md` |
| Remaining resource-matching data-source gaps | 15 | `_bmad-output/planning-artifacts/data-source-parity-plan.md` |

Important count history:

- The original v0.1.0 release baseline was 90 resources and 34 data sources.
- Story 25B.2 raised the data-source/doc count to 55 and left 36 resource-matching data-source gaps.
- Story 25B.3 raised the data-source/doc count to 76 and left 15 resource-matching data-source gaps.
- Do not regress public docs back to 34 data sources unless explicitly describing the historical v0.1.0 release artifact.

Primary planning source: `_bmad-output/planning-artifacts/support-matrix.md`.

Supporting sources:

- `_bmad-output/planning-artifacts/data-source-parity-plan.md`
- `_bmad-output/planning-artifacts/registry-docs-audit.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `templates/index.md.tmpl`
- `docs/index.md`

### Scope Boundary

This is a Registry provider-index documentation story.

Do:

- Improve the support matrix section of `templates/index.md.tmpl`.
- Regenerate `docs/index.md` from the template.
- Correct stale release-facing support-count wording where it would mislead current users.
- Keep claims honest and dated only when necessary.

Do not:

- Add data sources for the remaining 15 Batch 4 singleton/sensitive special cases.
- Add resources for LAGG, Kea DDNS/options, or Dnsmasq list items; source NAT and Unbound forward are now tracked as already supported under existing provider resource names.
- Create the full migration/import guide; that is Story 26.3.
- Add broad custom templates/examples for the 33 sparse generated data-source docs; Story 26.1 recorded that as residual docs-hardening risk, but this story is about the provider index support matrix.
- Change provider code, schemas, resource behavior, data-source behavior, or generated service code.

### Previous Story Intelligence

Story 26.1 audited Registry docs and found:

- The Registry site is JavaScript-rendered, so automated fetch could not inspect the live page; local `docs/`, `templates/`, and `examples/` are the reliable source for this workflow.
- Local docs are internally consistent at 90 resource docs and 76 data-source docs.
- Provider index already covers authentication, environment variables, minimum OPNsense version, permissions, support counts, and migration/import sequencing.
- The old Story 26.2 requirement incorrectly referenced 34 data sources and was corrected to 76.
- 33 generated data-source docs lack custom templates/examples; this is a residual risk, not required scope for this story.
- 59 resource docs lack custom import guidance/templates; Story 26.3 is the correct place for broad migration/import guidance.

Story 25B.3 added the latest support-count updates to `templates/index.md.tmpl` and `docs/index.md`. Preserve those corrections while making the matrix more useful.

### Current Files To Update

Read before editing:

- `templates/index.md.tmpl`: provider index template; source of truth for provider page content before generated schema is appended.
- `docs/index.md`: generated provider index; must be regenerated after template edits.
- `_bmad-output/planning-artifacts/support-matrix.md`: planning source of truth for Supported / Coming / Upstream-blocked wording and counts.
- `_bmad-output/planning-artifacts/registry-docs-audit.md`: audit findings and residual risks from Story 26.1.

Likely files to modify:

- `templates/index.md.tmpl`
- `docs/index.md`
- `_bmad-output/implementation-artifacts/26-2-public-support-matrix-registry-docs.md`

Only modify if stale and user-facing:

- `README.md`
- `RELEASE.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`

### Suggested Provider Index Shape

The support matrix should stay concise enough for Terraform Registry readers. A good shape is:

1. Short support-model legend:
   - Supported: implemented and documented in the provider.
   - Coming: provider-owned work remains or data-source parity remains.
   - Upstream-blocked: OPNsense lacks a stable usable API or requires upstream verification/contribution.
2. Current baseline count table:
   - Resources: 90
   - Data sources: 76
   - Resource docs: 90
   - Data-source docs: 76
   - Remaining data-source gaps: 15
3. Domain summary table with three columns:
   - Supported today
   - Coming / provider-owned follow-up
   - Upstream-blocked / OPNsense dependency
4. Link or pointer to the repository planning support matrix for full detail.

Avoid exhaustive per-resource enumeration in `docs/index.md`; keep that in planning artifacts and generated resource/data-source sidebars.

### Known Supported / Coming / Upstream-Blocked Content

Supported domains and counts come from `_bmad-output/planning-artifacts/support-matrix.md`:

- Supported resources include ACME, auth, cron, DHCPv4 reservation/subnet, Dnsmasq settings, Dynamic DNS account, firewall, HAProxy, interface types except LAGG/base assignment, IPsec, Kea DHCPv6/control/HA, Monit, OpenVPN, FRR/Quagga, syslog, system gateway/route/VIP/VLAN, traffic shaper, trust, Unbound, and WireGuard.
- Supported data sources include 76 resources/data-source pages, including the completed Epic 25B Batch 1/2/3 data sources.
- Remaining data-source gaps are 15 singleton or sensitive special-case resources: `dnsmasq_settings`, `ipsec_key_pair`, `ipsec_psk`, `kea_ctrl_agent`, `kea_dhcpv6_settings`, `openvpn_static_key`, `quagga_bgp_global`, `quagga_general`, `quagga_ospf6_general`, `quagga_ospf_general`, `quagga_rip`, `quagga_static`, `trust_cert`, `unbound_dnsbl`, and `unbound_general`.
- Coming buildable provider work includes data-source parity, interface LAGG if endpoint/environment verification succeeds, documentation hardening, and release hardening. Source NAT, Unbound forward, Dnsmasq list item resources, and OSPF area are already supported under existing provider resource names.
- Needs research items include Kea DHCPv4 option/DDNS live endpoint conflicts, HASync configuration `syncitems` request/response shape, HASync status/actions, and system tunables/sysctl endpoint lifecycle.
- Upstream-blocked domains include interface base assignment/IP config/PPPoE, gateway group, and system general settings.

### Documentation Generation

Use the existing docs-generation path. Do not edit `docs/index.md` by hand and leave it divergent from `templates/index.md.tmpl`.

Known working commands from recent stories:

```bash
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go generate ./tools
```

If Docker writes generated docs as root, restore ownership with:

```bash
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 chown -R "$(id -u):$(id -g)" docs
```

Final required validation remains:

```bash
make check
```

### Targeted Verification Commands

Use these or equivalent checks before final `make check`:

```bash
rg "Data sources \| 76|Data source docs \| 76|Remaining data-source gaps \| 15" templates/index.md.tmpl docs/index.md _bmad-output/planning-artifacts/support-matrix.md
rg "34 supported data sources|55 data sources|36 remaining|57 resource-matching" README.md RELEASE.md docs templates _bmad-output/planning-artifacts _bmad-output/implementation-artifacts
rg "Supported|Coming|Upstream-blocked" templates/index.md.tmpl docs/index.md
```

Interpret historical hits carefully:

- Story files for 25B.1 and 25B.2 may legitimately describe old baselines as implementation history.
- `CHANGELOG.md` and `RELEASE.md` may legitimately describe v0.1.0 release-time counts if clearly scoped to v0.1.0.
- Current provider index and support matrix must use 76 data sources and 15 remaining data-source gaps.

### Architecture and Quality Guardrails

- Follow DevRail rules: run `make check`; do not suppress failing checks.
- Respect `.editorconfig` and existing Markdown style.
- Keep changes documentation-only unless a stale count exposes an obvious planning-story correction.
- Do not install tools on the host; use the existing containerized toolchain.
- Do not create new helper scripts for a one-off documentation audit unless there is a clear reusable need.
- Generated docs are allowed to change only through `go generate ./tools` or `make check` docs generation.

### References

- `_bmad-output/planning-artifacts/support-matrix.md` — current support matrix and counts.
- `_bmad-output/planning-artifacts/data-source-parity-plan.md` — 15 remaining data-source gaps and Batch 4 candidates.
- `_bmad-output/planning-artifacts/registry-docs-audit.md` — Story 26.1 findings and residual risks.
- `_bmad-output/implementation-artifacts/26-1-registry-docs-audit.md` — previous story record.
- `_bmad-output/planning-artifacts/post-release-epics.md` — Epic 26 sequence and goals.
- `_bmad-output/planning-artifacts/prd.md` — release positioning and user journey requirements; note its count section may describe the older 34-data-source baseline and should not override current support-matrix counts.
- `_bmad-output/planning-artifacts/architecture.md` — documentation generation layer and Terraform Registry publishing constraints.
- `templates/index.md.tmpl` — provider index source template.
- `docs/index.md` — generated provider index output.

## Project Structure Notes

Expected files touched by this story:

- `templates/index.md.tmpl`
- `docs/index.md`
- `_bmad-output/implementation-artifacts/26-2-public-support-matrix-registry-docs.md`

Possible files if stale wording is found:

- `README.md`
- `RELEASE.md`
- `_bmad-output/planning-artifacts/support-matrix.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`

Files to inspect but not modify unless necessary:

- `docs/resources/*.md`
- `docs/data-sources/*.md`
- `_bmad-output/planning-artifacts/data-source-parity-plan.md`
- `_bmad-output/planning-artifacts/registry-docs-audit.md`

## Testing Requirements

- Required final command: `make check`.
- This is a documentation story; no Go unit tests or acceptance tests should be added.
- If generated docs change unexpectedly outside `docs/index.md`, inspect the diff and explain whether that is a normal docs-generation update or an unintended change.

## Completion Criteria

The story is complete only when:

- Provider index support matrix is clear, current, and public-facing.
- Counts are current at 90 resources, 76 data sources, 90 resource docs, 76 data-source docs, and 15 remaining data-source gaps.
- Coming and Upstream-blocked items are honest and do not promise delivery dates.
- `docs/index.md` is regenerated from `templates/index.md.tmpl`.
- Stale count references in current release-facing docs are corrected or explicitly classified as historical.
- `make check` passes.

## Dev Agent Record

### Agent Model Used

gpt-5.5

### Debug Log References

- Ultimate context engine analysis completed for Story 26.2.
- Previous Story 26.1 audit loaded and incorporated.
- Current support matrix and provider index template/read-only docs loaded before story rewrite.
- Baseline verification: 90 registered resources, 76 registered data sources, 90 resource docs, 76 data-source docs, and 15 resource-matching data-source gaps.
- Docs generation: `docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go generate ./tools` regenerated `docs/index.md` from `templates/index.md.tmpl`.
- Targeted searches confirmed current provider index/support matrix use 76 data sources, 76 data-source docs, and 15 remaining data-source gaps; remaining old-count hits are historical story/release-positioning references or the story verification command itself.
- Final validation: `make check` passed.
- Review patch validation: regenerated `docs/index.md` after template link update and reran `make check` successfully.

### Completion Notes List

- Rewrote Story 26.2 from a thin draft into a comprehensive implementation guide.
- Corrected story context around current post-Epic-25B counts and excluded migration-guide/data-source implementation scope.
- Hardened provider index support matrix with explicit Supported, Coming, and Upstream-blocked meanings, current counts, and a concise domain summary.
- Updated current-facing support counts in README and planning artifacts to the 76-data-source / 15-gap baseline.
- Kept `RELEASE.md` first-release 34-data-source statement as historical and added a subsequent-release refresh note.
- `go generate ./tools` reported broad missing-template generation, but the scoped story changes are limited to provider index/support-count documentation; existing untracked data-source templates/docs remain outside this story's required file list.
- Resolved review findings for stale PRD/release counts, Kea no-API classification, Registry support-matrix link target, and completed-Epic 25B sequencing.

### Change Log

- 2026-06-02: Implemented public support matrix hardening; regenerated provider index docs; updated current-facing support counts; `make check` passed; story moved to review.
- 2026-06-02: Applied code-review patches for stale counts, support-matrix link, no-API classification, and completed work sequencing; `make check` passed.

### File List

- `_bmad-output/implementation-artifacts/26-2-public-support-matrix-registry-docs.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/core-config-gap-analysis.md`
- `_bmad-output/planning-artifacts/feature-complete-roadmap.md`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `_bmad-output/planning-artifacts/prd.md`
- `README.md`
- `RELEASE.md`
- `docs/index.md`
- `templates/index.md.tmpl`
