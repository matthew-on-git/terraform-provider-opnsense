---
baseline_commit: b82ada36410a12ffaacb226cfbebc7977e6cb29e
---

# Story 26.3: Full-Appliance Migration and Import Guide

Status: done

## Story

As a brownfield OPNsense operator, I want a full-appliance migration guide, so that I can import existing configuration into Terraform in a safe dependency order.

## Acceptance Criteria

1. Migration guide explains UUID import workflow and dependency-order import strategy.
2. Guide includes examples for independent resources, chained HAProxy resources, firewall rules/NAT, routing, and VPN resources.
3. Guide explains how to reach a no-change plan after each import.
4. Guide explains limitations for sensitive write-only fields and upstream-blocked domains.
5. Provider index links to the migration guide or includes enough guidance directly.
6. `make check` passes.

## Tasks / Subtasks

- [x] Expand `docs/migration-import.md` from summary guidance into a full workflow.
- [x] Add concrete command examples for representative resources.
- [x] Add dependency-order checklist by domain.
- [x] Add troubleshooting section for import drift and write-only fields.
- [x] Link from provider index or README.
- [x] Run `make check`.

### Review Findings

- [x] [Review][Patch] HAProxy example cannot reach no-change after importing only servers [`docs/migration-import.md:84`]
- [x] [Review][Patch] Firewall/NAT example cannot reach no-change after importing only the alias [`docs/migration-import.md:137`]
- [x] [Review][Patch] Routing example cannot reach no-change after importing only the gateway [`docs/migration-import.md:180`]
- [x] [Review][Patch] Provider index implies source NAT is not supported despite outbound NAT resource docs [`templates/index.md.tmpl:97`]
- [x] [Review][Patch] Write-only secret guidance should say existing secret values must be retained or rotated [`docs/migration-import.md:273`]
- [x] [Review][Defer] Infeasible/no-API stories are marked backlog and may be reselected [`_bmad-output/implementation-artifacts/sprint-status.yaml:138`] — deferred, pre-existing
- [x] [Review][Defer] README dev-toolchain tag differs from Makefile pin [`README.md:53`] — deferred, pre-existing

## Dev Notes

Keep examples generic and avoid appliance-specific UUIDs except placeholders. The guide should help users migrate safely without implying every OPNsense domain is currently API-supported.

## Dev Agent Record

### Agent Model Used

OpenAI GPT-5.5 via OpenCode

### Debug Log References

- `docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go generate ./tools`
- `make security` initially failed on Go standard-library CVEs GO-2026-5039 and GO-2026-5037 in go1.25.10.
- Added `toolchain go1.25.11` to `go.mod`; reran `make security` successfully.
- `make check` passed after the docs and toolchain updates.
- Code review found 5 patch findings, 2 deferred pre-existing findings, and 2 dismissed findings.
- Applied all 5 patch findings and reran `make check` successfully.

### Completion Notes List

- Expanded `docs/migration-import.md` into a full brownfield migration workflow with UUID import guidance, dependency ordering, representative command examples, no-change plan troubleshooting, write-only secret limitations, and upstream-blocked domain boundaries.
- Added concrete examples for independent resources, chained HAProxy resources, firewall rules/NAT, routing, WireGuard, and IPsec.
- Linked the guide from the provider index template, regenerated `docs/index.md`, and added a README documentation pointer.
- Updated the Go toolchain requirement to `go1.25.11` so `govulncheck` uses the fixed standard library required by the project security gate.
- Applied review patches to remove contradictory intermediate plans from multi-resource import examples, clarify NAT follow-up wording, and clarify retained-or-rotated write-only secret handling.

### File List

- `README.md`
- `_bmad-output/implementation-artifacts/26-3-full-appliance-migration-guide.md`
- `_bmad-output/implementation-artifacts/deferred-work.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `docs/index.md`
- `docs/migration-import.md`
- `go.mod`
- `templates/index.md.tmpl`
