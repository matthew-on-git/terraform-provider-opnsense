---
baseline_commit: fbaf085e8287ae4f00f786484cd1ab622d77716a
---

# Story 29.6: Multi-Domain Edge Composition & Migration Validation

Status: done

## Story

As an operator evaluating an Ansible → Terraform migration,
I want a composition example that wires the full multi-domain HTTPS edge (servers → backends → map file → actions → ACLs → SSL frontend with ACME cert),
So that I can prove the provider reproduces a real OPNsense edge end-to-end before cutting over a production appliance.

## Context

Stories 29.1–29.4 add the missing pieces (action, map file, frontend cert binding, ACME issuance+refid). This story is the **integration capstone**: a single composition that exercises all of them together in the exact shape of the downstream `opnsense-manager` appliance (argocd/grafana/tipsyhive + CNAME aliases, internal-only deny, HTTP→HTTPS redirect). It is both a docs artifact and the acceptance gate that says "the HAProxy edge is migratable." It also surfaces ordering/dependency issues (cert refid availability, action→mapfile→backend references) that only appear when the resources are composed.

Depends on: 29.1, 29.2, 29.3, 29.4 (and existing 4.1–4.4, 8.1–8.3).

## Acceptance Criteria

1. **Given** the provider built with Stories 29.1–29.4
   **When** the operator applies `examples/compositions/multi-domain-edge/main.tf`
   **Then** it creates: ≥2 backends (each with ≥1 server), a `domain-map` map file routing multiple hosts (incl. a CNAME alias) to those backends, a `map_use_backend` action referencing the map, a `use_backend`+ACL pair for at least one wildcard/host case, an internal-only `http-request_deny` action, an HTTP frontend with a redirect action, and an HTTPS frontend with `ssl_enabled` + an ACME `cert_ref_id` bound via `certificates`

2. **And** the composition `terraform validate` + `terraform fmt -check` clean against the real provider schema (same bar as the existing 6 compositions, per Story 12.3)

3. **And** a second `terraform plan` immediately after apply shows **"No changes"** (the whole stack is idempotent — this is the migration-readiness signal)

4. **And** a short `README.md` in the composition maps each Terraform resource to the equivalent Ansible role/task in `opnsense-manager` (a migration crosswalk), and calls out the refid binding and ACME issuance wait explicitly

5. **And** the existing `examples/compositions/haproxy-full-stack/main.tf` is updated (or this supersedes it) so no shipped composition shows the misleading inert-ACL pattern (an ACL created but never used for routing)

6. **And** the `26-3 full-appliance-migration-guide` is updated to reference this composition and to flip the HAProxy edge + ACME rows from "not yet expressible" to "expressible as of v0.x via Epic 29"

## Tasks / Subtasks

- [x] Task 1: Author `examples/compositions/multi-domain-edge/main.tf` (AC: #1)
  - [x] 1.1 servers + backends (http mode)
  - [x] 1.2 `opnsense_haproxy_mapfile` "domain-map" with multiple `host backend` lines incl. an alias
  - [x] 1.3 `opnsense_haproxy_action` map_use_backend → mapfile; one use_backend+acl; one http-request_deny (internal-only); one redirect; one set-header X-Forwarded-Proto
  - [x] 1.4 http-in frontend (redirect) + https-in frontend (ssl_enabled, linked_actions, certificates = acme cert_ref_id)
  - [x] 1.5 acme account + challenge + certificate (issuance) feeding `cert_ref_id`
- [x] Task 2: `README.md` migration crosswalk (Terraform resource ↔ Ansible role/task) (AC: #4)
- [x] Task 3: `terraform fmt`/`validate` clean; add to whatever CI/example-validation harness checks the other compositions (AC: #2)
- [x] Task 4: Idempotency check — document/assert no-op second plan (AC: #3)
- [x] Task 5: Update/replace `haproxy-full-stack` to remove inert-ACL pattern (AC: #5)
- [x] Task 6: Update `26-3 full-appliance-migration-guide` rows (AC: #6)
- [x] Task 7: `make check`

## Dev Notes

### Target shape (from the live appliance)

This composition should mirror `opnsense-manager`'s `ansible/inventory/mfsoho/group_vars/all.yml` topology:
- Backends: `grafana-backend`, `argocd-backend`, `tipsyhive-backend` (each → MetalLB/cluster IP:port)
- Domain map: `grafana.mfsoho.linkridge.net`, `argocd.mfsoho.linkridge.net`, `tipsyhive.mfsoho.linkridge.net`, plus aliases `thetipsyhive.com`, `tipsyhive.com` → `tipsyhive-backend`
- Internal-only deny for argocd + grafana (deny unless source ∈ RFC1918)
- One ACME cert (e.g. per-domain) bound to `https-in`
- `http-in` → redirect-to-https

Keep IPs/domains as example placeholders; the point is structural fidelity, not the real secrets/addresses.

### Migration crosswalk (for the README, AC #4)

| Ansible (opnsense-manager) | Terraform (this provider) |
|---|---|
| `roles/haproxy` servers/backends | `opnsense_haproxy_server` / `opnsense_haproxy_backend` |
| `roles/haproxy` domain map | `opnsense_haproxy_mapfile` (29.2) |
| `roles/haproxy` actions (map_use_backend, deny, redirect, set-header) | `opnsense_haproxy_action` (29.1) |
| `roles/haproxy` ACLs | `opnsense_haproxy_acl` (4.3) |
| `roles/haproxy` https-in cert binding | `opnsense_haproxy_frontend.certificates` (29.3) |
| `roles/acme` cert add/sign/poll | `opnsense_acme_certificate` + `cert_ref_id` (29.4) |
| `roles/acme_watchdog` / `roles/haproxy_watchdog` crons | `opnsense_cron_job` (existing) |
| `roles/bgp` | `opnsense_quagga_*` (Epic 7/19) |
| `roles/firewall` | `opnsense_firewall_filter_rule` (Epic 3) |
| `roles/dyndns` accounts + settings | `opnsense_ddclient_account` + `opnsense_ddclient_settings` (29.5) |
| `roles/dhcp` PXE | `opnsense_dhcpv4_*` — **still blocked** (Stories 11.3 / 21.4, Kea endpoint research) |

### Known remaining gap to call out (do not hide it)

DHCP PXE/TFTP options are **not** covered by Epic 29 — they remain tracked-but-blocked in Stories 11.3 / 21.4 pending live Kea `dhcpv4 *_option` endpoint verification (and the appliance currently uses legacy ISC dhcpd, a different engine). The migration guide and README must state that a full cutover still leaves DHCP/PXE on Ansible until that is resolved.

### What NOT to build

- No new resources here — composition + docs only
- Do not invent endpoints; if 29.1–29.4 reveal an API shape different from these notes, fix the composition to match reality and record it

### References

- [Depends on: 29.1, 29.2, 29.3, 29.4]
- [Source: 12-3 composition-examples — validation bar for compositions]
- [Source: 26-3-full-appliance-migration-guide.md — rows to update]
- [Downstream driver: opnsense-manager ansible/inventory/mfsoho/group_vars/all.yml + roles/haproxy]

## Dev Agent Record

### Implementation Plan

- Build a composition-only capstone using existing Epic 29 resources without adding new resource types.
- Validate the composition against the current local provider schema, not the released registry provider.
- Document the migration crosswalk, ACME `cert_ref_id` binding, idempotency expectation, and remaining DHCP/PXE boundary.

### Debug Log

- Loaded Story 29.6 and sprint status; selected `29-6-multi-domain-edge-composition-and-migration-validation` as the first `ready-for-dev` story.
- Added the baseline commit and moved Story 29.6 to `in-progress` in the story file and sprint status.
- Authored `examples/compositions/multi-domain-edge/main.tf` with ACME account/challenge/certificate, HAProxy servers/backends/mapfile/actions/ACLs, HTTP redirect frontend, and HTTPS frontend bound to `cert_ref_id`.
- Added `examples/compositions/multi-domain-edge/README.md` with the migration crosswalk, ACME refid note, validation commands, no-op second-plan expectation, and DHCP/PXE boundary.
- Updated `examples/compositions/haproxy-full-stack/main.tf` to explicitly document that ACLs must be linked through actions.
- Updated `docs/migration-import.md` to reference the new full edge composition and mark HAProxy edge + ACME certificate binding as expressible via Epic 29 resources.
- `terraform validate` initially failed because `opnsense_haproxy_action` config validation treated unknown computed references as empty strings.
- Fixed the HAProxy action validator to allow unknown references while still rejecting literal empty required action fields; added focused unit tests.
- Verification passed: `go test ./internal/service/haproxy`.
- Verification passed: local-provider `terraform validate` for `examples/compositions/multi-domain-edge`.
- Verification passed: `terraform fmt -check -recursive examples/compositions/multi-domain-edge`.
- Required project gate passed: `make check`.
- Code review patches applied: corrected the internal-only deny condition to deny protected hosts only from external sources, and tightened HAProxy action validation to reject null required conditional fields while allowing unknown computed references.
- Review verification passed: `go test ./internal/service/haproxy`, local-provider `terraform validate` for `examples/compositions/multi-domain-edge`, and `make check`.

### Completion Notes

- Added a full multi-domain edge composition that exercises Epic 29's HAProxy action, mapfile, ACME issuance/refid, and frontend certificate binding work together.
- Added migration documentation tying the composition back to the equivalent Ansible role/task responsibilities and calling out the no-op second-plan migration signal.
- Fixed a provider validation edge case required for composition validation with computed resource references.
- Addressed both code-review findings and completed Story 29.6.

## File List

- `examples/compositions/multi-domain-edge/main.tf`
- `examples/compositions/multi-domain-edge/README.md`
- `examples/compositions/haproxy-full-stack/main.tf`
- `docs/migration-import.md`
- `internal/service/haproxy/action_validators.go`
- `internal/service/haproxy/action_model_test.go`
- `_bmad-output/implementation-artifacts/29-6-multi-domain-edge-composition-and-migration-validation.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`

## Change Log

- 2026-06-10: Added multi-domain edge composition, migration crosswalk documentation, HAProxy migration-guide updates, and validator support for computed HAProxy action references.
- 2026-06-10: Addressed code review findings for internal-only deny semantics and HAProxy action null-field validation.

### Review Findings

- [x] [Review][Patch] Internal-only deny action denies unprotected hosts because `unless internal AND protected` fails for every public non-protected request [examples/compositions/multi-domain-edge/main.tf:125]
- [x] [Review][Patch] HAProxy action validator skips null required fields, so omitted conditional fields can pass validation [internal/service/haproxy/action_validators.go:50]
