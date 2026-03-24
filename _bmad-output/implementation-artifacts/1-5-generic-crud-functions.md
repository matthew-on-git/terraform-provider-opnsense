# Story 1.5: Generic CRUD Functions

Status: done

## Story

As a developer,
I want generic CRUD functions (`Add[K]`, `Get[K]`, `Update[K]`, `Delete`) that work with any resource type via `ReqOpts` configuration,
so that each resource module only needs to define its struct and endpoint config, not HTTP logic.

## Acceptance Criteria

1. **AC1: Add[K]** — `Add[K]` wraps the resource struct in the monad key (`{"server": {...}}`), POSTs to `AddEndpoint`, and returns the UUID from the response. Acquires mutex, calls reconfigure after success.
2. **AC2: Get[K]** — `Get[K]` fetches by UUID from `GetEndpoint`, unwraps the monad, and returns a clean `*K`. Returns `NotFoundError` if the response is empty/default. Acquires read semaphore.
3. **AC3: Update[K]** — `Update[K]` wraps the resource struct in the monad key and POSTs to `UpdateEndpoint/{id}`. Acquires mutex, calls reconfigure after success.
4. **AC4: Delete** — `Delete` POSTs to `DeleteEndpoint/{id}`. Acquires mutex, calls reconfigure after success.
5. **AC5: Mutex and reconfigure** — All mutation functions (`Add`, `Update`, `Delete`) acquire `LockMutex` before the HTTP call and call `Reconfigure` after success, then `UnlockMutex`. Read functions (`Get`) use `AcquireRead`/`ReleaseRead`.
6. **AC6: Error handling** — CRUD functions use `CheckHTTPError` for status codes and `ParseMutationResponse` for mutation results. Transport errors wrapped with `NewServerError`.
7. **AC7: Context propagation** — All functions propagate `context.Context` for cancellation.
8. **AC8: Unit tests** — Tests use a `testResource` struct with a mock HTTP server. Cover: Add success/validation error, Get success/not-found, Update success, Delete success, mutex acquisition verified, reconfigure called after mutation.

## Tasks / Subtasks

- [x] Task 1: Implement Add[K] (AC: #1, #5, #6, #7)
  - [x] Create `pkg/opnsense/crud.go`
  - [x] Implement `Add[K any](ctx context.Context, c *Client, opts ReqOpts, resource *K) (string, error)`
  - [x] Marshal resource to JSON, wrap in monad key: `map[string]interface{}{opts.Monad: resource}`
  - [x] Acquire mutex via `c.LockMutex(ctx)`, defer `c.UnlockMutex()`
  - [x] POST to `{BaseURL}{AddEndpoint}` with `http.NewRequestWithContext`
  - [x] Check HTTP status with `CheckHTTPError`
  - [x] Read body, parse with `ParseMutationResponse` to extract UUID
  - [x] Call `Reconfigure(ctx, c, opts)` after successful mutation
  - [x] Return UUID on success
- [x] Task 2: Implement Get[K] (AC: #2, #6, #7)
  - [x] Implement `Get[K any](ctx context.Context, c *Client, opts ReqOpts, id string) (*K, error)`
  - [x] Acquire read semaphore via `c.AcquireRead(ctx)`, defer `c.ReleaseRead()`
  - [x] GET `{BaseURL}{GetEndpoint}/{id}` with `http.NewRequestWithContext`
  - [x] Check HTTP status with `CheckHTTPError`
  - [x] Read body, unmarshal as `map[string]json.RawMessage` to extract monad key
  - [x] Unmarshal the inner value into `*K`
  - [x] If monad key is missing or inner value is empty, return `NotFoundError`
  - [x] Return clean `*K`
- [x] Task 3: Implement Update[K] (AC: #3, #5, #6, #7)
  - [x] Implement `Update[K any](ctx context.Context, c *Client, opts ReqOpts, resource *K, id string) error`
  - [x] Marshal resource, wrap in monad key
  - [x] Acquire mutex, defer unlock
  - [x] POST to `{BaseURL}{UpdateEndpoint}/{id}`
  - [x] Check HTTP status, parse mutation response (ignore UUID)
  - [x] Call `Reconfigure` after success
- [x] Task 4: Implement Delete (AC: #4, #5, #7)
  - [x] Implement `Delete(ctx context.Context, c *Client, opts ReqOpts, id string) error`
  - [x] Acquire mutex, defer unlock
  - [x] POST to `{BaseURL}{DeleteEndpoint}/{id}`
  - [x] Check HTTP status with `CheckHTTPError`
  - [x] Call `Reconfigure` after success
- [x] Task 5: Write unit tests (AC: #8)
  - [x] Create `pkg/opnsense/crud_test.go`
  - [x] Define `testResource` struct for test fixtures
  - [x] Test: `Add` sends monad-wrapped body, returns UUID from mock server
  - [x] Test: `Add` returns `ValidationError` when result != "saved"
  - [x] Test: `Get` returns unwrapped struct from monad-wrapped response
  - [x] Test: `Get` returns `NotFoundError` when monad key has empty content
  - [x] Test: `Update` sends monad-wrapped body to correct URL with ID
  - [x] Test: `Delete` sends POST to correct URL with ID
  - [x] Test: mutation functions call reconfigure endpoint (verify via mock)
  - [x] Test: `Get` does not call reconfigure (read-only)
  - [x] Test: `CheckHTTPError` integration — 401 returns `AuthError`
  - [x] Verify all tests pass with `go test ./pkg/opnsense/...`
- [x] Task 6: Verify full pipeline (AC: all)
  - [x] Run `make check` — all targets pass
  - [x] Run `go build ./...` — succeeds

## Dev Notes

### Previous Story Intelligence (from Stories 1.1-1.4)

**Key learnings to apply:**
- DevRail container is 1.8.1 with Go 1.25 builder — `make check` passes all 6 targets
- `gosec` G704 (SSRF) flags `HTTPClient().Do(req)` — suppress with `//nolint:gosec` and explanation
- `go-retryablehttp` retries 5xx — tests for non-200 use non-retryable codes (400, 418)
- `errcheck` requires `defer func() { _ = resp.Body.Close() }()`
- Test helper pattern: `t.Helper()`, `httptest.NewServer`
- LockMutex is context-aware with cleanup goroutine for deadlock prevention
- `ParseMutationResponse` returns UUID on `"saved"`, `ValidationError` otherwise
- `CheckHTTPError` maps 401/403 → `AuthError`, 404 → `PluginNotFoundError`
- `NewServerError` wraps transport errors with endpoint context

**Existing code this story uses (do NOT reimplement):**
- `client.go`: `Client.LockMutex(ctx)`, `Client.UnlockMutex()`, `Client.AcquireRead(ctx)`, `Client.ReleaseRead()`, `Client.HTTPClient()`, `Client.BaseURL()`
- `reqopts.go`: `ReqOpts` struct with all endpoint fields and `Monad`
- `reconfigure.go`: `Reconfigure(ctx, client, opts)` function
- `errors.go`: `ParseMutationResponse`, `CheckHTTPError`, `NewServerError`, `NotFoundError`, `ValidationError`

### Architecture Compliance

This story implements AR3 (generic CRUD via Go generics), FR6-FR12 (CRUD lifecycle), and the cross-cutting mutex-protected CRUD pattern.

**Critical rule — state read-back:** After `Add` and `Update`, the resource layer MUST call `Get[K]` to populate state from the API response. Never set state from plan values. The CRUD functions enforce this by returning only UUID from `Add` and only error from `Update`.

**Mutation flow (every Add/Update/Delete):**
1. `LockMutex(ctx)` — serialize writes
2. HTTP POST to endpoint
3. Parse response (check HTTP status, parse body)
4. `Reconfigure(ctx, client, opts)` — activate changes
5. `UnlockMutex()` — release

**Read flow (every Get):**
1. `AcquireRead(ctx)` — respect concurrency limit
2. HTTP GET from endpoint
3. Parse response, unwrap monad
4. `ReleaseRead()` — release slot

### Critical Implementation Details

**Monad wrapping for requests (Add/Update):**
```go
wrapped := map[string]interface{}{opts.Monad: resource}
body, err := json.Marshal(wrapped)
```

**Monad unwrapping for responses (Get):**
```go
// Response JSON: {"server": {"name": "web1", "port": "80"}}
var envelope map[string]json.RawMessage
json.Unmarshal(body, &envelope)
inner, ok := envelope[opts.Monad]
if !ok || len(inner) == 0 || string(inner) == "null" || string(inner) == "{}" {
    return nil, &NotFoundError{Message: "resource not found"}
}
var result K
json.Unmarshal(inner, &result)
return &result, nil
```

**URL construction with ID:**
```go
// Add: POST {BaseURL}{AddEndpoint}         (no ID)
// Get: GET  {BaseURL}{GetEndpoint}/{id}
// Update: POST {BaseURL}{UpdateEndpoint}/{id}
// Delete: POST {BaseURL}{DeleteEndpoint}/{id}
```

**Delete does NOT parse mutation response** — it only needs HTTP status check. The delete endpoint may return `{"result":"deleted"}` which is not `"saved"` and would falsely trigger `ValidationError` if parsed with `ParseMutationResponse`.

**HTTP method for all OPNsense mutations is POST** — not PUT or DELETE. OPNsense uses POST for create, update, AND delete operations.

**gosec suppression:** All `client.HTTPClient().Do(req)` calls need `//nolint:gosec // URL from provider-configured ReqOpts`.

### What NOT to Build in This Story

- No `Search[K]` — that's Story 1.6
- No resources or data sources — those come in Epic 2
- No type conversion utilities — Story 1.7
- No changes to `internal/provider/` — this is all in `pkg/opnsense/`
- Do NOT import Terraform types — `pkg/opnsense/` is framework-independent

### Testing Approach

**Define a test resource struct:**
```go
type testResource struct {
    Name    string `json:"name"`
    Address string `json:"address"`
    Port    string `json:"port"`
}
```

**Mock server pattern for CRUD tests:**
The mock server needs to handle multiple endpoints. Use `r.URL.Path` to route:
```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    switch {
    case strings.HasPrefix(r.URL.Path, "/api/test/addItem"):
        // Add handler — return {"result":"saved","uuid":"..."}
    case strings.HasPrefix(r.URL.Path, "/api/test/getItem"):
        // Get handler — return {"item": {...}}
    case strings.HasPrefix(r.URL.Path, "/api/test/service/reconfigure"):
        // Reconfigure handler — return 200
    }
}))
```

**Verify reconfigure is called:** Use an `atomic.Bool` flag in the mock's reconfigure handler to assert it was hit after mutation operations.

**Verify mutex integration:** The mutex is tested in `mutex_test.go`. CRUD tests should verify that mutations don't panic or deadlock — the serialization property is already proven.

**NotFoundError detection:** Mock server returns `{"item": {}}` (empty monad content) to trigger `NotFoundError` from `Get[K]`.

### Project Structure After This Story

```
pkg/
└── opnsense/
    ├── client.go           # UNCHANGED
    ├── client_test.go       # UNCHANGED
    ├── reqopts.go           # UNCHANGED
    ├── reconfigure.go       # UNCHANGED
    ├── reconfigure_test.go  # UNCHANGED
    ├── mutex_test.go        # UNCHANGED
    ├── errors.go            # UNCHANGED
    ├── errors_test.go       # UNCHANGED
    ├── crud.go              # NEW: Add[K], Get[K], Update[K], Delete
    └── crud_test.go         # NEW: unit tests with testResource + mock server
```

### References

- [Source: architecture.md#API Client Design] — Generic CRUD function signatures, monad pattern
- [Source: architecture.md#Cross-Cutting Concerns] — Mutex-protected CRUD, auto-reconfigure, state read-back
- [Source: architecture.md#Resource Implementation Patterns] — Create/Read/Update/Delete method structure
- [Source: architecture.md#Error Handling] — Error type table, trigger conditions
- [Source: epics.md#Story 1.5] — Acceptance criteria, BDD scenarios
- [Previous: 1-3-global-mutex-and-reconfigure-infrastructure.md] — Mutex API, Reconfigure function
- [Previous: 1-4-custom-error-types-and-response-parsing.md] — ParseMutationResponse, CheckHTTPError, error types

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- `revive` linter flagged `context.Context should be the first parameter` — swapped parameter order from `(c *Client, ctx context.Context, ...)` to `(ctx context.Context, c *Client, ...)` in all CRUD functions. Architecture spec had `c *Client` first, but Go convention (enforced by revive) requires `ctx` first.
- `gosec` G704 suppressed on all `client.HTTPClient().Do(req)` calls with explanation.

### Completion Notes List

- Implemented `Add[K]`, `Get[K]`, `Update[K]`, `Delete` as Go generic functions in `crud.go`
- `ctx context.Context` is first parameter (Go convention, revive-enforced) followed by `*Client`
- Monad wrapping via `marshalWrapped` helper: wraps resource in `{monad: resource}` JSON
- Monad unwrapping via `unmarshalWrapped` helper: extracts `*K` from `{monad: {...}}` response
- `Add`: LockMutex → POST → CheckHTTPError → ParseMutationResponse → Reconfigure → UnlockMutex → return UUID
- `Get`: AcquireRead → GET → CheckHTTPError → unmarshalWrapped → ReleaseRead → return *K
- `Update`: LockMutex → POST with ID → CheckHTTPError → ParseMutationResponse → Reconfigure → UnlockMutex
- `Delete`: LockMutex → POST with ID → CheckHTTPError → Reconfigure → UnlockMutex (no body parsing)
- `NotFoundError` returned when monad key missing, inner value empty/null/{}, or unmarshal fails
- 11 unit tests: Add success + validation error + auth error, Get success + not-found empty + not-found missing key + no-reconfigure, Update success, Delete success + no-body-parse
- All tests pass, `make check` passes all 6 targets

### Change Log

- 2026-03-19: Implemented generic CRUD functions with mutex/reconfigure integration (Story 1.5)

### File List

- `pkg/opnsense/crud.go` — NEW: Add[K], Get[K], Update[K], Delete generic functions with monad wrap/unwrap, mutex, reconfigure, error handling
- `pkg/opnsense/crud_test.go` — NEW: 11 unit tests with testResource struct, mock server routing, reconfigure verification
