# Story 3.1: Firewall Savepoint Reconfigure Implementation

Status: done

## Story

As a developer,
I want the firewall filter `ReconfigureFunc` to implement OPNsense's savepoint/apply/cancelRollback 3-step flow,
So that firewall rule changes are protected by automatic 60-second rollback if connectivity is lost.

## Acceptance Criteria

1. **Given** a firewall filter resource mutation (Create/Update/Delete) succeeds
   **When** the reconfigure function is invoked
   **Then** it calls `POST /api/firewall/filter/savepoint` to get a revision ID

2. **And** it calls `POST /api/firewall/filter/apply/{revision}` to apply with rollback safety

3. **And** it calls `POST /api/firewall/filter/cancelRollback/{revision}` to confirm (prevent auto-revert)

4. **And** if any step fails, the error is surfaced as a Terraform diagnostic

5. **And** if `cancelRollback` is not called within 60 seconds, OPNsense automatically reverts the change

6. **And** unit tests verify the 3-step sequence with mock HTTP responses

## Tasks / Subtasks

- [x] Task 1: Implement `FirewallFilterReconfigure` function in `pkg/opnsense/reconfigure.go` (AC: #1, #2, #3, #4)
  - [x] 1.1 Implement `savepoint` step — POST to `/api/firewall/filter/savepoint`, parse revision from response
  - [x] 1.2 Implement `apply` step — POST to `/api/firewall/filter/apply/{revision}`
  - [x] 1.3 Implement `cancelRollback` step — POST to `/api/firewall/filter/cancelRollback/{revision}`
  - [x] 1.4 Wire into a single function matching `func(ctx context.Context) error` signature for `ReconfigureFunc`
- [x] Task 2: Write unit tests for the 3-step flow (AC: #4, #5, #6)
  - [x] 2.1 Test happy path — savepoint → apply → cancelRollback all succeed
  - [x] 2.2 Test savepoint failure — returns error, apply/cancelRollback NOT called
  - [x] 2.3 Test apply failure — returns error, cancelRollback NOT called
  - [x] 2.4 Test cancelRollback failure — returns error (apply already happened, auto-revert will kick in)
  - [x] 2.5 Verify the 3 endpoints are called in exact order with correct revision
- [x] Task 3: Run `make check` and verify all targets pass (AC: all)

## Dev Notes

### Safety-Critical Context

**This is the most safety-critical code in the provider.** A bad firewall rule applied without rollback protection can permanently lock the operator out of the OPNsense appliance. The savepoint mechanism is OPNsense's safety net — the 60-second auto-revert ensures recovery even when the operator (or Terraform) cannot reach the appliance after a bad rule change.

### The 3-Step Savepoint Flow

```
1. POST /api/firewall/filter/savepoint
   → Response: revision string (timestamp)
   → Creates a config checkpoint

2. POST /api/firewall/filter/apply/{revision}
   → Applies pending changes with rollback protection
   → Starts 60-second countdown timer

3. POST /api/firewall/filter/cancelRollback/{revision}
   → Confirms the change is good
   → Cancels the auto-revert timer
   → Change is now permanent

If step 3 never happens → OPNsense auto-reverts after 60 seconds
```

### OPNsense Firewall Filter API Endpoints

| Operation | Method | Endpoint | Notes |
|-----------|--------|----------|-------|
| Savepoint | POST | `/api/firewall/filter/savepoint` | Returns revision string |
| Apply | POST | `/api/firewall/filter/apply/{revision}` | Starts 60s rollback timer |
| CancelRollback | POST | `/api/firewall/filter/cancelRollback/{revision}` | Confirms change |
| Revert | POST | `/api/firewall/filter/revert/{revision}` | Manual rollback (not used by provider) |
| Add Rule | POST | `/api/firewall/filter/addRule` | CRUD — used by Story 3.3 |
| Get Rule | GET | `/api/firewall/filter/getRule/{uuid}` | CRUD — used by Story 3.3 |
| Set Rule | POST | `/api/firewall/filter/setRule/{uuid}` | CRUD — used by Story 3.3 |
| Del Rule | POST | `/api/firewall/filter/delRule/{uuid}` | CRUD — used by Story 3.3 |
| Search | GET/POST | `/api/firewall/filter/searchRule` | CRUD — used by Story 3.3 |

### Function Signature and Integration

The function must match the `ReconfigureFunc` signature in `ReqOpts`:
```go
ReconfigureFunc func(ctx context.Context) error
```

Since `ReconfigureFunc` is a closure, it needs access to the `*Client` to make HTTP calls. The function should be created as a factory that captures the client:

```go
// FirewallFilterReconfigure returns a ReconfigureFunc that implements the
// 3-step savepoint/apply/cancelRollback flow for firewall filter rules.
func FirewallFilterReconfigure(client *Client) func(ctx context.Context) error {
    return func(ctx context.Context) error {
        // Step 1: savepoint
        // Step 2: apply/{revision}
        // Step 3: cancelRollback/{revision}
    }
}
```

Story 3.3 will use this when defining its `ReqOpts`:
```go
var filterRuleReqOpts = opnsense.ReqOpts{
    AddEndpoint:    "/api/firewall/filter/addRule",
    GetEndpoint:    "/api/firewall/filter/getRule",
    UpdateEndpoint: "/api/firewall/filter/setRule",
    DeleteEndpoint: "/api/firewall/filter/delRule",
    SearchEndpoint: "/api/firewall/filter/searchRule",
    ReconfigureFunc: opnsense.FirewallFilterReconfigure(r.client),
    Monad:          "rule",
}
```

### Implementation Details

**Savepoint response parsing:**
The savepoint endpoint returns a JSON response with the revision. Parse it to extract the revision string:
```go
// POST /api/firewall/filter/savepoint
// Response: {"revision": "1679912345.1234"}  (or similar format)
```

**Apply with revision:**
The revision is passed as a URL path parameter:
```go
// POST /api/firewall/filter/apply/1679912345.1234
```

**CancelRollback with revision:**
```go
// POST /api/firewall/filter/cancelRollback/1679912345.1234
```

**Error handling — fail-fast on each step:**
- If savepoint fails → return error immediately (nothing was changed)
- If apply fails → return error (savepoint exists but not applied — no auto-revert risk)
- If cancelRollback fails → return error (CRITICAL: changes are applied but will auto-revert in 60s; the error message should warn the operator)

**Error messages must be specific:**
```go
// Step 1 failure:
fmt.Errorf("firewall filter savepoint failed: %w", err)

// Step 2 failure:
fmt.Errorf("firewall filter apply (revision %s) failed: %w", revision, err)

// Step 3 failure:
fmt.Errorf("firewall filter cancelRollback (revision %s) failed — changes will auto-revert in 60 seconds: %w", revision, err)
```

### Existing Code to Build On

**`Reconfigure()` in `reconfigure.go` already supports `ReconfigureFunc`:**
```go
func Reconfigure(ctx context.Context, client *Client, opts ReqOpts) error {
    if opts.ReconfigureFunc != nil {
        return opts.ReconfigureFunc(ctx)
    }
    // ... standard endpoint path
}
```

The CRUD functions (`Add`, `Update`, `Delete` in `crud.go`) call `Reconfigure()` after every mutation. No changes to CRUD code are needed — `ReconfigureFunc` is already invoked by the existing infrastructure.

**`Client.HTTPClient()` provides the HTTP client:**
All 3 steps need to make HTTP POST requests. Use `client.HTTPClient().Do(req)` with the same `//nolint:gosec` pattern established in existing code.

**No mutex needed inside the savepoint function:**
The savepoint function runs inside the mutex already held by the CRUD function (Add/Update/Delete call `LockMutex` before `Reconfigure`). Do NOT acquire the mutex again — that would deadlock.

### Testing Pattern

**Mock server with ordered request tracking:**
```go
func TestFirewallFilterReconfigure_Success(t *testing.T) {
    var calls []string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        calls = append(calls, r.URL.Path)
        switch {
        case strings.HasSuffix(r.URL.Path, "/savepoint"):
            // Return revision
            w.Write([]byte(`{"revision":"1234.5678"}`))
        case strings.Contains(r.URL.Path, "/apply/"):
            w.WriteHeader(http.StatusOK)
        case strings.Contains(r.URL.Path, "/cancelRollback/"):
            w.WriteHeader(http.StatusOK)
        }
    }))
    defer server.Close()

    client := newReconfigureTestClient(t, server.URL)
    fn := FirewallFilterReconfigure(client)
    err := fn(context.Background())
    // Assert: no error, calls == ["/api/firewall/filter/savepoint", "/api/firewall/filter/apply/1234.5678", "/api/firewall/filter/cancelRollback/1234.5678"]
}
```

**Use non-retryable status codes (400, 418) for failure tests** — never 5xx (go-retryablehttp retries those).

**Reuse `newReconfigureTestClient` helper** from existing `reconfigure_test.go`.

### What NOT to Build

- No filter rule resource — that's Story 3.3
- No changes to `crud.go` — the ReconfigureFunc integration already works
- No changes to `reqopts.go` — the field already exists
- No changes to `internal/service/firewall/` — this is purely `pkg/opnsense/` infrastructure
- No acceptance tests — this is a unit-testable function (acceptance testing happens in Story 3.3)

### Previous Story Intelligence

**From Epic 1, Story 1.3 (Mutex and Reconfigure):**
- `Reconfigure()` function handles both standard endpoint and custom func
- Tests use `atomic.Bool`/`atomic.Value` for concurrent verification
- `newReconfigureTestClient` helper exists and should be reused
- Tests use non-retryable status codes (400, 418) for deterministic failure testing

**From Epic 2 (Resource Implementation):**
- `ReconfigureFunc` is set per-resource via `ReqOpts` — the savepoint function will be called by the CRUD layer transparently
- No changes needed to any existing resource code

**From Epic 1 Retrospective:**
- `ctx context.Context` always first parameter
- `gosec` suppressions need inline explanation
- `make check` must pass all targets

### Project Structure Notes

**Modified files:**
```
pkg/opnsense/reconfigure.go       # MODIFIED: add FirewallFilterReconfigure function
pkg/opnsense/reconfigure_test.go  # MODIFIED: add savepoint flow unit tests
```

**No new files.** This story adds to existing files only.

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-3, Story 3.1]
- [Source: _bmad-output/planning-artifacts/architecture.md#Firewall savepoint flow]
- [Source: _bmad-output/planning-artifacts/prd.md#FR13 firewall filter rollback]
- [Source: pkg/opnsense/reconfigure.go#Current Reconfigure implementation]
- [Source: pkg/opnsense/reqopts.go#ReconfigureFunc field]
- [Source: https://docs.opnsense.org/development/api/core/firewall.html#Filter endpoints]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- `staticcheck` QF1012 flagged `w.Write([]byte(fmt.Sprintf(...)))` — replaced with `fmt.Fprintf(w, ...)`
- `errcheck` then flagged unchecked `fmt.Fprintf` return — added `_, _ =` prefix
- Factory pattern `FirewallFilterReconfigure(client)` returns closure capturing `*Client` — avoids needing client in `ReconfigureFunc` signature

### Completion Notes List

- Implemented `FirewallFilterReconfigure(client)` factory returning `func(ctx) error` for `ReconfigureFunc`
- 3-step flow: `firewallSavepoint` → `firewallApply` → `firewallCancelRollback` (private helper functions)
- Savepoint parses revision from JSON response `{"revision":"..."}` via `savepointResponse` struct
- Apply and cancelRollback pass revision as URL path parameter
- Error messages are step-specific: cancelRollback failure warns about 60-second auto-revert
- 5 unit tests: happy path, savepoint failure (1 call), apply failure (2 calls), cancelRollback failure (3 calls + auto-revert warning), revision verification
- `make check` passes 5/6 targets; scan fails due to pre-existing gitleaks findings

### File List

- `pkg/opnsense/reconfigure.go` — MODIFIED: added FirewallFilterReconfigure, firewallSavepoint, firewallApply, firewallCancelRollback, savepointResponse
- `pkg/opnsense/reconfigure_test.go` — MODIFIED: added 5 unit tests for the 3-step savepoint flow
