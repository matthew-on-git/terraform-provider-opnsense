# Story 1.3: Global Mutex and Reconfigure Infrastructure

Status: done

## Story

As a developer,
I want a global mutex that serializes all write operations and a reconfigure dispatch mechanism that applies changes after mutations,
so that the provider protects OPNsense's config.xml integrity and activates changes correctly.

## Acceptance Criteria

1. **AC1: Write serialization** — All Create, Update, and Delete operations are serialized through a global mutex (single key) on the `Client`. Read operations are NOT blocked by the mutex and execute in parallel.
2. **AC2: Read concurrency limiting** — Read operations are limited by a configurable `semaphore.Weighted` (default: 10 concurrent reads). The max concurrency is set via `ClientConfig.MaxReadConcurrency`.
3. **AC3: ReqOpts struct** — A `ReqOpts` struct is defined with fields: `AddEndpoint`, `GetEndpoint`, `UpdateEndpoint`, `DeleteEndpoint`, `SearchEndpoint`, `ReconfigureEndpoint` (string), `ReconfigureFunc` (func override), and `Monad` (string). `ReconfigureEndpoint` and `ReconfigureFunc` are mutually exclusive.
4. **AC4: Standard reconfigure dispatch** — A `Reconfigure(ctx, client, opts)` function calls `POST {BaseURL}{ReconfigureEndpoint}` after successful mutations. On HTTP 200, it succeeds silently. On failure, it returns an error with the endpoint and status code.
5. **AC5: Function-based reconfigure dispatch** — If `ReqOpts.ReconfigureFunc` is set, `Reconfigure` calls it instead of the standard endpoint. The `ReconfigureFunc` interface is defined but no concrete implementation exists yet (firewall savepoint comes in Epic 3).
6. **AC6: Client exposes concurrency primitives** — The `Client` exposes `LockMutex(ctx)`, `UnlockMutex()`, `AcquireRead(ctx)`, and `ReleaseRead()` methods. These are called by the CRUD layer (Story 1.5), not by resources directly.
7. **AC7: Unit tests** — Tests verify: mutex serialization (concurrent writes are sequential), semaphore limiting (reads beyond limit block), reconfigure dispatch for both standard and function-based paths, reconfigure error handling. Tests use mock HTTP server.

## Tasks / Subtasks

- [x] Task 1: Add mutex and semaphore to Client (AC: #1, #2, #6)
  - [x] Add `writeMu sync.Mutex` field to `Client` struct
  - [x] Add `readSem *semaphore.Weighted` field to `Client` struct
  - [x] Add `MaxReadConcurrency int` field to `ClientConfig` (default 10)
  - [x] Initialize semaphore in `NewClient` with configured max concurrency
  - [x] Implement `LockMutex(ctx context.Context) error` — acquires write mutex (context-aware)
  - [x] Implement `UnlockMutex()` — releases write mutex
  - [x] Implement `AcquireRead(ctx context.Context) error` — acquires one semaphore slot
  - [x] Implement `ReleaseRead()` — releases one semaphore slot
- [x] Task 2: Define ReqOpts struct (AC: #3)
  - [x] Create `pkg/opnsense/reqopts.go`
  - [x] Define `ReqOpts` struct with all endpoint fields, `ReconfigureEndpoint`, `ReconfigureFunc`, and `Monad`
  - [x] `ReconfigureFunc` type: `func(ctx context.Context) error`
- [x] Task 3: Implement reconfigure dispatch (AC: #4, #5)
  - [x] Create `pkg/opnsense/reconfigure.go`
  - [x] Implement `Reconfigure(ctx context.Context, client *Client, opts ReqOpts) error`
  - [x] If `ReconfigureFunc` is set, call it and return its error
  - [x] If `ReconfigureEndpoint` is set, POST to `{BaseURL}{ReconfigureEndpoint}`
  - [x] If neither is set, return nil (no reconfigure needed)
  - [x] On HTTP 200: return nil
  - [x] On non-200: return error with endpoint and status code
  - [x] On connection error: return error with endpoint and underlying error
  - [x] Close response body correctly: `defer func() { _ = resp.Body.Close() }()`
- [x] Task 4: Write unit tests (AC: #7)
  - [x] Create `pkg/opnsense/mutex_test.go`
  - [x] Test: concurrent LockMutex calls are serialized (use goroutines + timing/ordering)
  - [x] Test: AcquireRead respects semaphore limit (acquire N+1 on limit N should block)
  - [x] Test: ReleaseRead frees semaphore slot for next reader
  - [x] Test: mutex and semaphore are independent (reads don't block on mutex)
  - [x] Create `pkg/opnsense/reconfigure_test.go`
  - [x] Test: Reconfigure calls standard endpoint on success (mock server 200)
  - [x] Test: Reconfigure returns error on non-200 status
  - [x] Test: Reconfigure calls ReconfigureFunc when set (verify func was called)
  - [x] Test: Reconfigure returns nil when neither endpoint nor func is set
  - [x] Test: Reconfigure propagates ReconfigureFunc error
  - [x] Verify all tests pass with `go test ./pkg/opnsense/...`
- [x] Task 5: Verify full pipeline (AC: all)
  - [x] Run `make lint` — passes
  - [x] Run `make format` — passes
  - [x] Run `go test ./...` — all tests pass (including existing client tests)
  - [x] Run `go build ./...` — succeeds

## Dev Notes

### Previous Story Intelligence (from Stories 1.1 and 1.2)

**Key learnings to apply:**
- DevRail's `gofumpt` is stricter than `gofmt` — run `make fix` before checking lint
- `go.mod` targets `go 1.25.0` to match DevRail container
- Package comments are required by `revive` linter — `reqopts.go` and `reconfigure.go` are in existing `package opnsense` so no new package comment needed
- Unused parameters must use `_` prefix
- `errcheck` requires handling `resp.Body.Close()` — use `defer func() { _ = resp.Body.Close() }()`
- `gosec` flags field names matching "secret" patterns — use `//nolint:gosec` with explanation if needed
- Test helper pattern: use `t.Helper()` in test helper functions
- Mock server pattern: `httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { ... }))`
- `govulncheck` fails in DevRail container (Go 1.24 vs project Go 1.25.0) — pre-existing, not a blocker

**Files from previous stories that this story modifies:**
- `pkg/opnsense/client.go` — add `writeMu` and `readSem` fields to `Client`, add `MaxReadConcurrency` to `ClientConfig`, initialize semaphore in `NewClient`, add `LockMutex`/`UnlockMutex`/`AcquireRead`/`ReleaseRead` methods

**Files this story creates:**
- `pkg/opnsense/reqopts.go` — `ReqOpts` struct definition
- `pkg/opnsense/reconfigure.go` — `Reconfigure` function
- `pkg/opnsense/mutex_test.go` — mutex and semaphore tests
- `pkg/opnsense/reconfigure_test.go` — reconfigure dispatch tests

### Architecture Compliance

This story implements FR14 (mutation serialization), FR12 (service reconfigure), FR18 (reconfigure failure diagnostics), AR2 (mutex in API client), AR5 (global MutexKV), and AR6 (read semaphore).

**Why the mutex exists (safety-critical):**
OPNsense stores ALL configuration in a single monolithic `config.xml` file. Concurrent writes risk XML corruption that can **brick the appliance**. The global mutex is not a performance optimization — it prevents data corruption.

**Why the semaphore exists:**
OPNsense runs PHP-FPM with a limited worker pool. Unbounded concurrent reads can exhaust workers, causing timeouts. The semaphore (default 10) prevents overwhelming the backend.

**Two-phase mutation lifecycle:**
Every mutation = CRUD call + service reconfigure call. The CRUD layer (Story 1.5) will call `LockMutex` → execute HTTP mutation → `Reconfigure` → `UnlockMutex`. This story provides the primitives; Story 1.5 wires them into the CRUD flow.

### Critical Implementation Details

**Client struct changes:**
```go
import (
    "sync"
    "golang.org/x/sync/semaphore"
)

type Client struct {
    httpClient *http.Client
    baseURL    string
    writeMu    sync.Mutex
    readSem    *semaphore.Weighted
}
```

**ClientConfig addition:**
```go
type ClientConfig struct {
    // ... existing fields ...
    // MaxReadConcurrency limits concurrent read operations. Default: 10.
    MaxReadConcurrency int64
}
```

**NewClient initialization (add to existing function):**
```go
maxReads := cfg.MaxReadConcurrency
if maxReads == 0 {
    maxReads = 10
}
// ... in return statement:
return &Client{
    httpClient: retryClient.StandardClient(),
    baseURL:    baseURL,
    readSem:    semaphore.NewWeighted(maxReads),
}, nil
```

Note: `writeMu` zero-value (`sync.Mutex{}`) is ready to use — no initialization needed.

**LockMutex must be context-aware** to support Terraform cancellation:
```go
func (c *Client) LockMutex(ctx context.Context) error {
    done := make(chan struct{})
    go func() {
        c.writeMu.Lock()
        close(done)
    }()
    select {
    case <-done:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

**ReqOpts struct:**
```go
type ReqOpts struct {
    AddEndpoint         string
    GetEndpoint         string
    UpdateEndpoint      string
    DeleteEndpoint      string
    SearchEndpoint      string
    ReconfigureEndpoint string                              // Standard reconfigure
    ReconfigureFunc     func(ctx context.Context) error     // Override for firewall savepoint flow
    Monad               string                              // Request body wrapper key
}
```

**Reconfigure function:**
```go
func Reconfigure(ctx context.Context, client *Client, opts ReqOpts) error {
    if opts.ReconfigureFunc != nil {
        return opts.ReconfigureFunc(ctx)
    }
    if opts.ReconfigureEndpoint == "" {
        return nil
    }
    // POST to {BaseURL}{ReconfigureEndpoint}
    // ...
}
```

The POST to the reconfigure endpoint should use the client's `HTTPClient()` to ensure auth headers are included. Create a `*http.Request` with `http.NewRequestWithContext(ctx, http.MethodPost, url, nil)`.

### Dependency to Add

```
golang.org/x/sync  (for semaphore.Weighted)
```

Run `go get golang.org/x/sync` — this is already an indirect dependency (via terraform-plugin-testing), so it may just be promoted to direct.

### What NOT to Build in This Story

- No CRUD functions — those come in Story 1.5 (they will call `LockMutex`/`UnlockMutex`/`Reconfigure`)
- No error types — those come in Story 1.4
- No firewall savepoint implementation — that comes in Epic 3 Story 3.1 (the `ReconfigureFunc` interface is defined here but no concrete implementation)
- No resources or data sources
- No changes to `internal/provider/` — this is all in `pkg/opnsense/`
- Do NOT add mutex locking to the `validateCredentials` call in provider.go — credential validation is a one-time setup call, not a CRUD operation

### Testing Approach

**Mutex serialization test pattern:**
Use goroutines with a shared counter or ordering tracker to verify that concurrent `LockMutex` calls are serialized:
```go
// Launch N goroutines that all try to LockMutex
// Inside each: record goroutine ID to an ordered slice
// After all complete: verify the slice has N entries (all ran)
// The key assertion: operations within the lock are not interleaved
```

**Semaphore limiting test pattern:**
```go
// Create client with MaxReadConcurrency = 2
// Launch 3 AcquireRead goroutines
// First 2 should succeed immediately
// Third should block until one of the first 2 calls ReleaseRead
```

Use `context.WithTimeout` to detect blocking behavior — if `AcquireRead` doesn't return within a short timeout, it's correctly blocking.

**Reconfigure test pattern:**
Use `httptest.NewServer` that:
- Returns 200 for success case
- Returns 500 for error case
- Verify the correct endpoint path was called
- For `ReconfigureFunc`: use a closure that sets a boolean flag

**Important:** Do NOT use `time.Sleep` for timing-based concurrency tests. Use channels and synchronization primitives for deterministic behavior.

### Project Structure After This Story

```
pkg/
└── opnsense/
    ├── client.go           # MODIFIED: mutex + semaphore fields, Lock/Unlock/Acquire/Release methods
    ├── client_test.go       # UNCHANGED
    ├── reqopts.go           # NEW: ReqOpts struct
    ├── reconfigure.go       # NEW: Reconfigure dispatch function
    ├── mutex_test.go        # NEW: mutex and semaphore tests
    └── reconfigure_test.go  # NEW: reconfigure dispatch tests
```

### References

- [Source: architecture.md#Core Architectural Decisions] — Global mutex and read semaphore design
- [Source: architecture.md#Cross-Cutting Concerns] — Auto-reconfigure, mutex-protected CRUD
- [Source: architecture.md#API Client Design] — ReqOpts struct, CRUD function signatures
- [Source: architecture.md#Resource Implementation Patterns] — ReconfigureEndpoint vs ReconfigureFunc
- [Source: prd.md#Provider Configuration & Authentication] — FR12, FR14, FR18
- [Source: epics.md#Story 1.3] — Acceptance criteria, BDD scenarios
- [Previous: 1-1-initialize-project-and-api-client-core.md] — Client struct, test patterns
- [Previous: 1-2-provider-configuration-with-credential-validation.md] — DevRail learnings, testing patterns

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- `gosec` G704 (SSRF via taint analysis) triggered on `client.HTTPClient().Do(req)` in reconfigure.go — false positive since URL is from provider-configured ReqOpts, not user input. Suppressed with `//nolint:gosec`.
- `go-retryablehttp` retries 5xx status codes — reconfigure non-200 test uses 400 (Bad Request) instead of 500 to get deterministic error behavior.
- `make check` security target previously failed with Go version mismatch — resolved by updating DevRail container to 1.8.1 (Go 1.25 builder).

### Completion Notes List

- Added `writeMu sync.Mutex` and `readSem *semaphore.Weighted` to `Client` struct for write serialization and read concurrency limiting
- Added `MaxReadConcurrency int64` to `ClientConfig` with default 10
- `LockMutex` is context-aware — uses goroutine + select pattern to support Terraform cancellation; cleanup goroutine prevents deadlock on context cancellation
- `AcquireRead`/`ReleaseRead` delegate to `semaphore.Weighted` from `golang.org/x/sync`
- `ReqOpts` struct defines all CRUD endpoint paths, `ReconfigureEndpoint`, `ReconfigureFunc`, and `Monad`
- `Reconfigure` dispatches to `ReconfigureFunc` (if set), standard endpoint (if set), or no-op (neither set)
- Standard reconfigure uses `http.NewRequestWithContext` for cancellation support and `client.HTTPClient().Do()` for auth headers
- 10 new tests: 4 mutex/semaphore tests (serialization, context cancellation, limit, independence), 6 reconfigure tests (standard success, non-200, func call, func error propagation, nil when neither, connection error)
- All 33 tests pass across the project (9 client, 14 provider [sic: 24], 10 new)
- `golang.org/x/sync` promoted from indirect to direct dependency

### Code Review Notes

**Reviewed by:** Claude Opus 4.6 (same session)
**Date:** 2026-03-19
**Findings:** 1 High, 0 Medium, 0 Low — fixed:
- H1: LockMutex goroutine leak on context cancellation caused permanent deadlock. Fixed by adding cleanup goroutine that acquires-then-releases orphaned lock. Added regression test verifying mutex is usable after cancellation.

### Change Log

- 2026-03-19: Implemented global mutex, read semaphore, ReqOpts, and reconfigure dispatch (Story 1.3)
- 2026-03-19: Code review fix — LockMutex deadlock prevention on context cancellation; DevRail container updated to 1.8.1

### File List

- `pkg/opnsense/client.go` — MODIFIED: added writeMu, readSem fields to Client; MaxReadConcurrency to ClientConfig; LockMutex/UnlockMutex/AcquireRead/ReleaseRead methods; semaphore initialization in NewClient; LockMutex cleanup goroutine for context cancellation
- `pkg/opnsense/reqopts.go` — NEW: ReqOpts struct with CRUD endpoints, ReconfigureEndpoint, ReconfigureFunc, Monad
- `pkg/opnsense/reconfigure.go` — NEW: Reconfigure dispatch function (standard endpoint + function override)
- `pkg/opnsense/mutex_test.go` — NEW: 4 tests for mutex serialization, context cancellation (with deadlock regression), semaphore limiting, independence
- `pkg/opnsense/reconfigure_test.go` — NEW: 6 tests for reconfigure dispatch (standard, error, func, nil, propagation, connection)
- `go.mod` — MODIFIED: promoted golang.org/x/sync to direct dependency
- `go.sum` — MODIFIED: updated checksums
- `Makefile` — MODIFIED: updated DEVRAIL_IMAGE to 1.8.1
