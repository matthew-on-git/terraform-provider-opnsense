---
baseline_commit: fbaf085e8287ae4f00f786484cd1ab622d77716a
---

# Story 29.4: ACME Certificate Issuance & Refid Output

Status: done

## Story

As an operator,
I want creating an `opnsense_acme_certificate` to actually issue the certificate and expose its HAProxy refid and status,
So that a cert is real and bindable to a frontend (Story 29.3) immediately after `terraform apply`, without a manual "sign" step in the OPNsense UI.

## Context

Story 8.2 ("ACME Certificate Resource **with Issuance**") is marked `done`, but the shipped `certificate_resource.go` does **plain CRUD only**: `Add` → `Get`. It never calls `/api/acmeclient/certificates/sign/{uuid}`, never polls issuance status, and exposes no `cert_ref_id`. So the resource creates a certificate *config entry* in a pending state but does not obtain a usable certificate, and provides nothing for HAProxy to bind to. This story closes the gap the 8.2 title promised.

This is a **correctness fix to an existing resource**, not a new resource. Treat it as such (no breaking schema changes to the required inputs).

## Acceptance Criteria

1. **Given** a valid `opnsense_acme_certificate` (account + validation_method + name)
   **When** the operator runs `terraform apply`
   **Then** after the config entry is created the provider triggers issuance via `POST /api/acmeclient/certificates/sign/{uuid}` and the apply does not return success until issuance resolves

2. **And** the provider polls `/api/acmeclient/certificates/search` (or `get`) until the cert reaches issued state (`statusCode` 200 / non-empty refid) or a bounded timeout elapses; on timeout it returns an actionable error (not a silent partial success)

3. **And** the resource exposes new **Read-Only** attributes:
   - `cert_ref_id` (String) — the HAProxy legacy refid used by `opnsense_haproxy_frontend.certificates` (Story 29.3)
   - `status_code` (String) and/or `status` (String) — issuance status for observability

4. **And** issuance behavior is governed by a configurable, bounded wait: optional `issuance_timeout` (default ~180s, matching DNS-01/HTTP-01 latency) and a poll interval; document that LE rate limits make tight loops unwise

5. **And** the operation is idempotent: on `Read`, an already-issued cert does not re-sign; re-apply with no input change shows "No changes"; renewal remains OPNsense's cron responsibility (provider does not renew)

6. **And** `Delete` and `Update` continue to work; updating an input that requires re-issuance (e.g. `alt_names`) re-signs and re-polls

7. **And** an acceptance test (guarded behind an env flag if it needs real ACME) verifies the issued path populates `cert_ref_id`; a unit/mock test verifies the sign-then-poll control flow and the timeout error path

## Tasks / Subtasks

- [x] Task 1: Add the sign endpoint + (if needed) status endpoint to `certificateReqOpts` or a dedicated call (AC: #1)
  - [x] 1.1 `POST /api/acmeclient/certificates/sign/{uuid}` after `Add`
- [x] Task 2: Implement bounded poll loop for issuance status (AC: #2, #4)
  - [x] 2.1 Poll `search`/`get`; success on `statusCode == 200` and non-empty refid; respect context deadline
  - [x] 2.2 Configurable `issuance_timeout` + interval; clear timeout diagnostic
- [x] Task 3: Extend `certificate_model.go` / `certificate_schema.go` with Read-Only `cert_ref_id`, `status_code`/`status` (AC: #3)
  - [x] 3.1 Map the refid out of the API response (this is the value HAProxy binds — see Story 29.3 refid note)
- [x] Task 4: Wire sign+poll into `Create` and the re-issuance path in `Update` (AC: #1, #6)
- [x] Task 5: Ensure `Read` is non-mutating and idempotent; no re-sign on steady state (AC: #5)
- [x] Task 6: Tests — mocked control-flow + timeout path; optional live-ACME acceptance behind env gate (AC: #7)
- [x] Task 7: Update docs/example — show `opnsense_acme_certificate` → `cert_ref_id` → `opnsense_haproxy_frontend.certificates` (AC: #3)
- [x] Task 8: `make check`

### Review Findings

- [x] [Review][Patch] Validate issuance wait config before remote create/update and clean up failed creates [internal/service/acme/certificate_resource.go:64]
- [x] [Review][Patch] Parse sign action response body instead of discarding HTTP 200 failures [internal/service/acme/certificate_resource.go:225]
- [x] [Review][Patch] Avoid re-signing on provider-only wait setting changes [internal/service/acme/certificate_resource.go:111]
- [x] [Review][Patch] Preserve context cancellation instead of reporting it as issuance timeout [internal/service/acme/certificate_resource.go:190]
- [x] [Review][Patch] Accept numeric statusCode API values during search/read parsing [internal/service/acme/certificate_model.go:60]
- [x] [Review][Patch] Preserve or default provider-only wait attributes after import/read [internal/service/acme/certificate_model.go:96]
- [x] [Review][Patch] Escape certificate UUID path segment in sign endpoint [internal/service/acme/certificate_resource.go:213]
- [x] [Review][Patch] Assert actionable timeout diagnostics in the unit test [internal/service/acme/certificate_issuance_test.go:77]

## Dev Notes

### Current state (what exists today)

`internal/service/acme/certificate_resource.go` `Create` = `opnsense.Add(...)` then `opnsense.Get(...)`. `certificateReqOpts` has `Add/Get/Update/Delete/Search/Reconfigure` but **no sign**. Schema fields: `enabled, name, description, alt_names, account, validation_method, key_length, auto_renewal` — **no refid/status outputs.**

### OPNsense ACME issuance API

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Add | POST | `/api/acmeclient/certificates/add` | creates config entry (exists) |
| Sign | POST | `/api/acmeclient/certificates/sign/{uuid}` | **imperative** — triggers ACME issuance; returns quickly, work is async |
| Search | GET | `/api/acmeclient/certificates/search` | poll for `statusCode`/refid |
| Get | GET | `/api/acmeclient/certificates/get/{uuid}` | full model |
| Reconfigure | POST | `/api/acmeclient/service/reconfigure` | exists |

The downstream Ansible role's proven sequence: `add` → `sign` → poll `search` until `statusCode == 200` with non-empty `certRefId` (up to ~30 × 10s). Reuse those semantics; the refid (`certRefId`) is what HAProxy binds.

### Imperative-action modeling in a declarative provider

`sign` is an imperative trigger, which is awkward in Terraform. Recommended approach: treat it as a side effect of `Create`/`Update` that the resource drives to a terminal state before returning (so state reflects reality). Make the wait bounded and configurable. Do **not** expose `sign` as a standalone `terraform apply`-replays-it operation. On `Read`, derive `status`/`cert_ref_id` from the API; never sign in `Read`.

### Coordination

- `cert_ref_id` is the bridge to Story 29.3 (`opnsense_haproxy_frontend.certificates`). Land 29.4 before/with 29.3 for the end-to-end HTTPS path.
- Renewal stays with OPNsense cron (the downstream appliance uses an ACME watchdog cron — already representable via `opnsense_cron_job`). Document that the provider manages issuance + config, not renewal.

### What NOT to build

- No renewal scheduling — OPNsense cron owns it
- No account/challenge changes — those resources exist (8.1/8.3)
- No new required inputs — keep 8.2's inputs; only add outputs + issuance behavior + optional timeout

### References

- [Source: internal/service/acme/certificate_resource.go — current plain-CRUD implementation, no sign]
- [Source: 8-2-acme-certificate-resource-with-issuance.md — title/AC claims issuance]
- [Pairs with: Story 29.3 (frontend cert binding via refid)]
- [Downstream driver: opnsense-manager ansible/roles/acme/tasks/main.yml — add/sign/poll(statusCode=200)/refid]

## Dev Agent Record

### Implementation Plan

- Add an internal `signCertificate` call for `POST /api/acmeclient/certificates/sign/{uuid}` and serialize it with the existing write mutex.
- Drive issuance only from `Create` and `Update`, then poll certificate search until `statusCode == 200` and `certRefId` is present.
- Keep `Read` non-mutating by only fetching current API state.
- Add bounded duration controls and computed observability outputs without changing required inputs.
- Document the ACME certificate `cert_ref_id` handoff into HAProxy frontend TLS binding.

### Debug Log

- Red phase: `go test ./internal/service/acme` failed because `signAndWaitForCertificateIssuance` did not exist.
- Green phase: implemented sign/poll helpers, resource/data-source schema outputs, create/update wiring, and guarded ACME acceptance scaffold.
- Verification passed: `go test ./internal/service/acme`.
- Verification passed: `go test ./pkg/opnsense ./internal/service/acme ./internal/service/haproxy`.
- Docs regenerated with containerized `go generate ./tools` through `ghcr.io/devrail-dev/dev-toolchain:v1`.
- Required project gate passed: `make check`.
- Code review found 8 patch findings; all were applied.
- Review verification passed: `go test ./pkg/opnsense ./internal/service/acme ./internal/service/haproxy`.
- Review verification passed: `make check`.

### Completion Notes

- `opnsense_acme_certificate` now signs after create/update and waits for issued state before returning success.
- Added `issuance_timeout` and `issuance_poll_interval` duration controls with documented defaults.
- Added computed `cert_ref_id`, `status_code`, and `status` to the resource and data source.
- Read/import remain non-mutating; renewal remains OPNsense cron responsibility.
- Acceptance coverage is guarded behind `OPNSENSE_ACME_ISSUE=1` plus real domain/account/validation UUID env vars.
- Review patches added pre-mutation wait validation, best-effort create cleanup, sign response body parsing, provider-only update skipping, context cancellation preservation, numeric status parsing, path escaping, wait default preservation, and expanded unit tests.

## File List

- `internal/service/acme/certificate_resource.go`
- `internal/service/acme/certificate_model.go`
- `internal/service/acme/certificate_schema.go`
- `internal/service/acme/certificate_data_source.go`
- `internal/service/acme/certificate_issuance_test.go`
- `internal/service/acme/certificate_resource_test.go`
- `examples/resources/opnsense_acme_certificate/resource.tf`
- `templates/resources/acme_certificate.md.tmpl`
- `docs/resources/acme_certificate.md`
- `docs/data-sources/acme_certificate.md`
- `_bmad-output/implementation-artifacts/29-4-acme-certificate-issuance-and-refid.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`

## Change Log

- 2026-06-10: Implemented ACME certificate issuance sign/poll behavior, refid/status outputs, docs/example updates, guarded acceptance scaffold, and validation.
- 2026-06-10: Applied code review findings and marked story done after full validation.
