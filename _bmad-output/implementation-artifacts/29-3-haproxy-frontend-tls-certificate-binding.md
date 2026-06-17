---
baseline_commit: fbaf085e8287ae4f00f786484cd1ab622d77716a
---

# Story 29.3: HAProxy Frontend TLS Certificate Binding

Status: done

## Story

As an operator,
I want the `opnsense_haproxy_frontend` resource to bind SSL certificates,
So that an HTTPS frontend actually serves TLS using a specific certificate (e.g. an ACME-issued cert) instead of only toggling SSL on/off.

## Context

Story 4.2 shipped the frontend with `ssl_enabled` (a boolean) and explicitly deferred certificate fields under "What NOT to Build: No SSL certificate management fields — advanced, defer (ssl_certificates, ssl_default_certificate, etc.)." As a result, **there is no way to attach a certificate to the HTTPS listener.** `ssl_enabled = true` with no cert is not a usable TLS frontend. This blocks any real HTTPS edge and is the second half of the ACME story (Story 29.4 issues the cert and exposes its refid; this story binds it).

## Acceptance Criteria

1. **Given** an existing certificate on the appliance (ACME-issued or imported) identified by its HAProxy cert **refid**
   **When** the operator sets `certificates = [<refid>]` (and optionally `default_certificate = <refid>`) on an `opnsense_haproxy_frontend` with `ssl_enabled = true`
   **Then** `setFrontend` binds those certs (`ssl_certificates`, `ssl_default_certificate`) and the frontend serves TLS with them

2. **And** `certificates` is a Set of strings (cert refids); `default_certificate` is an optional single refid; both are optional so existing frontends without certs are unaffected (backward compatible)

3. **And** binding/unbinding/reordering certificates is an **in-place** update, not destroy-and-recreate

4. **And** when `ssl_enabled = false`, cert fields are ignored (with a clear validation note), and `terraform plan` is stable (no perpetual diff) whether or not certs are set

5. **And** the resource interoperates with ACME: the `cert_ref_id` output from `opnsense_acme_certificate` (Story 29.4) can be passed directly into `certificates`

6. **And** an acceptance test (a) creates an SSL frontend bound to a cert refid and (b) verifies a no-op second plan; docs/example show ACME-cert → frontend binding

## Tasks / Subtasks

- [x] Task 1: Extend `frontend_model.go` (AC: #1, #2)
  - [x] 1.1 Add `ssl_certificates` (SelectedMapList of refids) and `ssl_default_certificate` (SelectedMap) to `frontendAPIResponse`/`Request`
  - [x] 1.2 Add `Certificates types.Set` + `DefaultCertificate types.String` to `FrontendResourceModel`
  - [x] 1.3 Extend `toAPI`/`fromAPI`; handle empty/omitted gracefully (AC #4)
- [x] Task 2: Extend `frontend_schema.go` (AC: #2, #3, #4)
  - [x] 2.1 `certificates` (Optional Set(String)), `default_certificate` (Optional String); no RequiresReplace
  - [x] 2.2 Doc note: refids, not UUIDs; only meaningful when `ssl_enabled = true`
- [x] Task 3: Update `frontend_resource_test.go` — add SSL-with-cert case + idempotency assert (AC: #6)
- [x] Task 4: Update example + `templates/resources/haproxy_frontend.md.tmpl` to show cert binding (AC: #6)
- [x] Task 5: `make check`

### Review Findings

- [x] [Review][Patch] Ignore or reject certificate fields when `ssl_enabled = false` [`internal/service/haproxy/frontend_model.go`]
- [x] [Review][Patch] Remove `default_certificate` static default to preserve no-default compatibility [`internal/service/haproxy/frontend_schema.go`]
- [x] [Review][Patch] Validate `default_certificate` is included in `certificates` [`internal/service/haproxy/frontend_schema.go`]
- [x] [Review][Patch] Make certificate refid joining deterministic for Terraform set input [`internal/service/haproxy/frontend_model.go`]
- [x] [Review][Patch] Add coverage for unbind/update/import/SSL-disabled stability boundaries [`internal/service/haproxy/frontend_resource_test.go`]
- [x] [Review][Patch] Add missing planning artifact to File List [`_bmad-output/implementation-artifacts/29-3-haproxy-frontend-tls-certificate-binding.md`]

## Dev Notes

### The refid problem (critical)

OPNsense HAProxy binds certificates by a **legacy non-UUID "refid"** string (e.g. `69d4458b7a233`), NOT by the certificate's API UUID. The `getFrontend` response returns `ssl_certificates` as a map of `refid → label`. The downstream Ansible role handles this by parsing the frontend's `ssl_certificates` dict and rebuilding the refid list. Implications:

- `certificates` must accept and round-trip **refids**. Extract selected keys from the `SelectedMapList` in `fromAPI`; send a comma-joined refid string in `toAPI` (same shape as `linked_actions`).
- The cert's refid is **not** the same as its ACME-certificate UUID. Story 29.4 must expose the refid as `cert_ref_id` so users can wire `certificates = [opnsense_acme_certificate.x.cert_ref_id]`. Until 29.4 lands, users supply a literal refid.
- Document this prominently — it is the single most confusing part of HAProxy-on-OPNsense.

### API (unchanged endpoints — extend the body)

Same `frontendReqOpts` as Story 4.2. The added request fields go into the existing `setFrontend`/`addFrontend` body:

| API Field | Response type | TF attr | TF type |
|-----------|---------------|---------|---------|
| `ssl_certificates` | SelectedMapList (refid→label) | `certificates` | Set(String) (Optional) |
| `ssl_default_certificate` | SelectedMap (refid) | `default_certificate` | String (Optional) |

### Backward compatibility

Existing state for frontends created by Story 4.2 has no `certificates`/`default_certificate`. New Optional attributes with no default must not force a diff on those resources. Verify: import an SSL-less frontend → plan is clean.

### Coordination with the appliance's split-ownership pattern

In the downstream Ansible setup, the **ACME role owns cert binding** and the HAProxy role deliberately omits `ssl_certificates` to avoid clobbering it. In Terraform there is a single owner of the frontend resource, so this story makes the frontend the authoritative owner of its cert list. Note in docs: do not also manage the same frontend's certs out-of-band, or you will get drift.

### What NOT to build

- No HSTS / TLS tuning / cipher fields — out of scope, separate follow-up
- No certificate *creation* — that is `opnsense_acme_certificate` (Story 29.4) / imported certs
- Do not change `ssl_enabled` semantics

### References

- [Source: 4-2-haproxy-frontend-resource-with-acl-routing.md#What-NOT-to-Build — "defer ssl_certificates, ssl_default_certificate"]
- [Pairs with: Story 29.4 (ACME refid output)]
- [Downstream driver: opnsense-manager ansible/roles/acme/tasks/main.yml — frontend ssl_certificates refid binding]

## Dev Agent Record

### Debug Log

- Resolved BMad dev-story workflow customization with no prepend/append steps.
- Loaded sprint status and selected first ready story: `29-3-haproxy-frontend-tls-certificate-binding`.
- Captured baseline commit: `fbaf085e8287ae4f00f786484cd1ab622d77716a`.
- Added red-phase frontend model tests for certificate refid request/response conversion; confirmed they failed before implementation.

### Completion Notes

- Extended `opnsense_haproxy_frontend` with `certificates` and `default_certificate` attributes using OPNsense HAProxy certificate refids.
- Added API request/response mapping for `ssl_certificates` and `ssl_default_certificate`.
- Extended the frontend data source so certificate refids are visible from reads/imported state.
- Added unit tests for certificate refid conversion and an acceptance scaffold gated by `OPNSENSE_HAPROXY_CERT_REFID` with a no-op plan step.
- Updated frontend examples/templates and regenerated docs to describe refids and planned ACME `cert_ref_id` wiring.
- Resolved code-review findings by ignoring certificate fields in API payloads when SSL is disabled, plan-normalizing disabled cert fields, validating default certificate membership, sorting certificate refids before API joins, removing the explicit default on `default_certificate`, and expanding tests/import coverage.

### Validation

- Red phase: `go test ./internal/service/haproxy` failed before frontend certificate fields existed.
- `go test ./pkg/opnsense ./internal/service/haproxy` passed.
- `docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go generate ./tools` passed.
- `make check` passed.
- Post-review `go test ./pkg/opnsense ./internal/service/haproxy` passed.
- Post-review `docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/devrail-dev/dev-toolchain:1.12.0 go generate ./tools` passed.
- Post-review `make check` passed.

### File List

- `_bmad-output/implementation-artifacts/29-3-haproxy-frontend-tls-certificate-binding.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/planning-artifacts/post-release-epics.md`
- `docs/data-sources/haproxy_frontend.md`
- `docs/resources/haproxy_frontend.md`
- `examples/resources/opnsense_haproxy_frontend/resource.tf`
- `internal/service/haproxy/frontend_data_source.go`
- `internal/service/haproxy/frontend_model.go`
- `internal/service/haproxy/frontend_model_test.go`
- `internal/service/haproxy/frontend_resource_test.go`
- `internal/service/haproxy/frontend_schema.go`
- `internal/service/haproxy/frontend_validators.go`
- `templates/resources/haproxy_frontend.md.tmpl`

### Change Log

- 2026-06-10: Added HAProxy frontend certificate binding support and marked story ready for review.
- 2026-06-10: Addressed code review findings; status set to done.
