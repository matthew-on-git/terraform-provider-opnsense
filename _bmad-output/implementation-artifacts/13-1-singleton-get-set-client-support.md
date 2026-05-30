# Story 13.1: Singleton Get/Set API Client Support

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a developer,
I want generic singleton `GetSingleton[K]` and `UpdateSingleton[K]` functions in the `pkg/opnsense` API client,
so that resources backed by OPNsense get/set settings endpoints (no UUID) — such as FRR/OSPF/RIP general settings, BGP global config, Unbound general, and Kea ctrl_agent/ddns — can read and update configuration the same way UUID-based resources use `Get[K]`/`Update[K]`.

## Context & Why

Many OPNsense MVC controllers expose a **singleton settings object** via `get`/`set` endpoints with **no UUID path segment** (e.g. `GET /api/quagga/general/get`, `POST /api/quagga/general/set`). The existing generic CRUD (`pkg/opnsense/crud.go`) cannot serve these: `Get[K]` and `Update[K]` **hard-append `/{id}`** to the endpoint (`url := c.BaseURL() + opts.GetEndpoint + "/" + id`), producing a wrong URL for singletons.

This story adds the missing client primitives. It is the **foundation gate** for ~10 downstream singleton resources in the feature-complete roadmap (Epics 19, 21, 22). Scope is **the client layer only** — no Terraform resources here.

[Source: _bmad-output/planning-artifacts/feature-complete-roadmap.md#Epic 13]
[Source: _bmad-output/planning-artifacts/core-config-gap-analysis.md#Singletons]

## Acceptance Criteria

1. **Given** a singleton settings endpoint and a `ReqOpts` with `GetEndpoint` set and `Monad` set, **When** `GetSingleton[K](ctx, c, opts)` is called, **Then** it issues `GET {BaseURL}{GetEndpoint}` (**no `/{id}` appended**), unwraps the response from the monad key, and returns a clean `*K`.
2. **Given** the get response body `{"<monad>": {...}}`, **When** `GetSingleton` parses it, **Then** it returns the unwrapped struct (reusing the existing `unmarshalWrapped` helper).
3. **Given** the get response monad value is missing/empty/`{}`/`null`, **When** `GetSingleton` parses it, **Then** it returns a `*NotFoundError` (consistent with `Get[K]`).
4. **Given** a singleton settings endpoint, **When** `UpdateSingleton[K](ctx, c, opts, resource)` is called, **Then** it wraps the resource in the monad key and issues `POST {BaseURL}{UpdateEndpoint}` (**no `/{id}` appended**) with `Content-Type: application/json`.
5. **Given** the set response `{"result":"saved"}`, **When** `UpdateSingleton` completes, **Then** it parses the mutation response (reusing `ParseMutationResponse`) and calls `Reconfigure(ctx, c, opts)` on success.
6. **Given** the set response indicates a validation failure (`result != "saved"` with a `validations` map), **When** `UpdateSingleton` runs, **Then** it returns a `*ValidationError` with the field map (no reconfigure call must succeed-mask the error).
7. **Given** concurrency primitives, **When** `GetSingleton` runs it acquires/releases the **read semaphore** (`AcquireRead`/`ReleaseRead`) and **does NOT** call reconfigure; `UpdateSingleton` acquires/releases the **write mutex** (`LockMutex`/`UnlockMutex`) and calls reconfigure after success — mirroring `Get`/`Update`.
8. **Given** HTTP 401/403/404/5xx responses, **When** either function runs, **Then** errors are produced via the existing `CheckHTTPError`/`NewServerError` paths (same as `Get`/`Update`).
9. Unit tests in `pkg/opnsense/crud_test.go` cover: GetSingleton success (and assert the requested path has **no trailing UUID**), GetSingleton not-found, UpdateSingleton success (assert path has no trailing UUID + monad wrapping + reconfigure called), UpdateSingleton validation error, and GetSingleton-does-not-call-reconfigure.
10. `make check` passes (lint, format, unit tests). All new exported functions have Go doc comments matching the style of the existing CRUD functions.

## Tasks / Subtasks

- [ ] Task 1: Add `GetSingleton[K]` to `pkg/opnsense/crud.go` (AC: #1, #2, #3, #7, #8)
  - [ ] Signature: `func GetSingleton[K any](ctx context.Context, c *Client, opts ReqOpts) (*K, error)`
  - [ ] Acquire read semaphore via `c.AcquireRead(ctx)`; `defer c.ReleaseRead()`
  - [ ] Build URL as `c.BaseURL() + opts.GetEndpoint` — **no `/{id}`**
  - [ ] `http.NewRequestWithContext(ctx, http.MethodGet, url, nil)`
  - [ ] `c.HTTPClient().Do(req)` with `//nolint:gosec` comment as in `Get`
  - [ ] `CheckHTTPError(resp.StatusCode, opts.GetEndpoint)`, read body, return `unmarshalWrapped[K](opts.Monad, respBody)`
- [ ] Task 2: Add `UpdateSingleton[K]` to `pkg/opnsense/crud.go` (AC: #4, #5, #6, #7, #8)
  - [ ] Signature: `func UpdateSingleton[K any](ctx context.Context, c *Client, opts ReqOpts, resource *K) error`
  - [ ] Acquire write mutex via `c.LockMutex(ctx)`; `defer c.UnlockMutex()`
  - [ ] `marshalWrapped(opts.Monad, resource)`
  - [ ] Build URL as `c.BaseURL() + opts.UpdateEndpoint` — **no `/{id}`**; POST with JSON content type
  - [ ] `CheckHTTPError`, read body, `ParseMutationResponse(respBody)` (return error if validation/parse fails)
  - [ ] On success `return Reconfigure(ctx, c, opts)`
- [ ] Task 3: Unit tests in `pkg/opnsense/crud_test.go` (AC: #9)
  - [ ] `TestGetSingleton_Success` — handler at exact path `/api/test/getSettings` (no trailing segment); assert returned struct fields; capture `r.URL.Path` and assert it has **no** UUID suffix
  - [ ] `TestGetSingleton_NotFound` — `{"settings":{}}` → `*NotFoundError`
  - [ ] `TestGetSingleton_DoesNotCallReconfigure` — assert reconfigure not hit
  - [ ] `TestUpdateSingleton_Success` — assert path == set endpoint (no UUID), monad-wrapped body, reconfigure called
  - [ ] `TestUpdateSingleton_ValidationError` — `{"result":"failed","validations":{...}}` → `*ValidationError`
  - [ ] Add a `testSingletonReqOpts()` helper OR reuse fields from `testReqOpts()` with distinct singleton paths to avoid `HasPrefix` collisions with UUID-based tests
- [ ] Task 4: Run `make check` and fix any lint/format/test failures (AC: #10)

## Dev Notes

### Exact code shape (follow existing `Get`/`Update` in crud.go)

The two new functions are near-clones of `Get[K]` (lines ~67-95) and `Update[K]` (lines ~100-138) in `pkg/opnsense/crud.go`, with the **single difference** that the URL omits `+ "/" + id`. Reuse every existing helper — **do not** introduce new HTTP/JSON logic:

- Read path mirrors `Get`: `AcquireRead`/`ReleaseRead`, `CheckHTTPError`, `unmarshalWrapped[K]`. [Source: pkg/opnsense/crud.go:67-95]
- Write path mirrors `Update`: `LockMutex`/`UnlockMutex`, `marshalWrapped`, `CheckHTTPError`, `ParseMutationResponse`, `Reconfigure`. [Source: pkg/opnsense/crud.go:100-138]
- `marshalWrapped`/`unmarshalWrapped` are unexported helpers already in crud.go (lines 171-195) — reuse as-is. [Source: pkg/opnsense/crud.go:171-195]

Doc-comment style: match the existing block comments above `Add`/`Get`/`Update` (purpose + concurrency note). Note in the `GetSingleton` doc that it is read-only (semaphore, no reconfigure) and in `UpdateSingleton` that it acquires the write mutex and reconfigures after success.

### Why no UUID

`Get`/`Update` build `url := c.BaseURL() + opts.GetEndpoint + "/" + id`. Singleton endpoints (`/api/quagga/general/get`, `/set`) are addressed directly — appending `/` + id yields a 400/404. The new functions simply drop the id segment. [Source: pkg/opnsense/crud.go:73, 111]

### Test collision warning (critical)

Existing CRUD tests match handler routes with `strings.HasPrefix(r.URL.Path, "/api/test/getItem/")`. Use **distinct singleton paths** (e.g. `/api/test/getSettings`, `/api/test/setSettings`, reconfigure `/api/test/service/reconfigure`) so the new singleton tests do not collide, and so you can assert the exact path equals the endpoint with **no** trailing UUID. Reuse `newCRUDTestServer`, `newCRUDTestClient` helpers verbatim. [Source: pkg/opnsense/crud_test.go:350-380]

### Concurrency model (must preserve)

Per architecture: all mutations serialize through the global mutex; reads use the semaphore and never reconfigure. `GetSingleton` = read; `UpdateSingleton` = mutation. [Source: _bmad-output/planning-artifacts/architecture.md#API Client Design (AR5/AR6), crud.go LockMutex/AcquireRead usage]

### Forward context (NOT this story — do not implement)

Downstream singleton **resources** (Epics 19/21/22) will implement Terraform `Read` via `GetSingleton` and `Update` via `UpdateSingleton`. A singleton resource has **no real Create/Delete** on the appliance: Create is implemented as "call UpdateSingleton" (settings always exist) and Delete is a no-op `RemoveResource`/reset. The `id` is a synthetic constant (e.g. the controller name). That belongs to the resource stories — keep THIS story to the two client functions + tests only.

### Project Structure Notes

- All changes confined to `pkg/opnsense/crud.go` (2 new exported funcs) and `pkg/opnsense/crud_test.go` (new tests). No new files. No changes to `ReqOpts` (existing fields `GetEndpoint`, `UpdateEndpoint`, `Monad`, `ReconfigureEndpoint` suffice). [Source: pkg/opnsense/reqopts.go]
- No `internal/` or resource changes in this story.

### Testing standards

- Table/unit tests with `httptest.NewServer`, no `TF_ACC`. Mutation mocks must return non-5xx (retryable codes cause `go-retryablehttp` to exhaust retries → connection error, not the status). Use 400/418 for negative HTTP cases if needed. [Source: _bmad-output/implementation-artifacts/epic-1-retro-2026-03-23.md#Team Agreements]
- `make check` is the single gate (lint + format + test). Tools run in the dev-toolchain container via the Makefile — do not install host tools. [Source: CLAUDE.md#Critical Rules]

### References

- [Source: pkg/opnsense/crud.go:19-195] — existing Add/Get/Update/Delete + marshal helpers (the templates to clone)
- [Source: pkg/opnsense/crud_test.go:1-381] — existing test patterns + helpers
- [Source: pkg/opnsense/reqopts.go:8-30] — ReqOpts fields
- [Source: pkg/opnsense/reconfigure.go] — `Reconfigure(ctx, c, opts)` dispatch (standard endpoint vs ReconfigureFunc)
- [Source: _bmad-output/planning-artifacts/architecture.md#API Client Design]
- [Source: _bmad-output/planning-artifacts/core-config-gap-analysis.md] — singleton inventory (FRR/OSPF/RIP general, BGP global, Unbound general, Kea ctrl_agent/ddns, DDNS settings)

## Dev Agent Record

### Agent Model Used

claude-opus-4-8 (BMad Master / dev-story)

### Debug Log References

- `go build ./...` → clean
- `go test ./pkg/opnsense/ -run Singleton -v` → 5/5 PASS
- `make check` → lint ✓ format ✓ test ✓ docs ✓ | security ✗ scan ✗ (BOTH pre-existing on pristine `main` — verified by stashing this story's diff and re-running; failures are Go-stdlib/x-net CVEs and gitleaks false-positives in vendored bmad skill docs, neither caused by this story)

### Completion Notes List

- Added `GetSingleton[K]` and `UpdateSingleton[K]` to `pkg/opnsense/crud.go` as exact clones of `Get`/`Update` minus the `/{id}` URL segment, reusing all existing helpers (`marshalWrapped`, `unmarshalWrapped`, `CheckHTTPError`, `ParseMutationResponse`, `Reconfigure`, semaphore/mutex).
- Added 5 unit tests + `testSingletonReqOpts()` helper with distinct `/api/test/{get,set}Settings` paths to avoid prefix collisions and assert no-UUID URLs.
- Code review: PASS (no blocking findings). Optional future nice-to-haves: a 401-path test for UpdateSingleton (covered indirectly by shared `CheckHTTPError`); singleton-specific empty-object semantics deferred to resource layer.
- Story-relevant checks all green; only pre-existing baseline `make check` failures remain (see Debug Log) — these require an infra/baseline fix, not a code change in this story.

### File List

- `pkg/opnsense/crud.go` (modified — added GetSingleton, UpdateSingleton)
- `pkg/opnsense/crud_test.go` (modified — added 5 singleton tests + helper)
