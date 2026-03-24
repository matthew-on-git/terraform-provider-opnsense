# Story 1.4: Custom Error Types and Response Parsing

Status: done

## Story

As a developer,
I want custom error types that parse OPNsense's non-standard API responses into structured errors,
so that resources can handle validation failures, missing resources, and permission errors correctly.

## Acceptance Criteria

1. **AC1: ValidationError** — When a mutation response has `result != "saved"` (HTTP 200), a `ValidationError` is returned containing field names and error messages from the `validations` map.
2. **AC2: NotFoundError** — When a GET response indicates a missing resource (blank/default record or JSON unmarshal type error), a `NotFoundError` is returned.
3. **AC3: AuthError** — When the response is HTTP 401 or 403, an `AuthError` is returned with the status code.
4. **AC4: PluginNotFoundError** — When the response is HTTP 404 on a plugin endpoint, a `PluginNotFoundError` is returned with the plugin name extracted from the URL path.
5. **AC5: ServerError** — When all retries are exhausted on HTTP 500+/timeout, a `ServerError` is returned wrapping the underlying error.
6. **AC6: errors.As compatibility** — All error types work with Go's `errors.As()` for type-safe error handling in resource code.
7. **AC7: Unit tests** — Tests verify each error type is correctly constructed and that response parsing helper functions return the right error type from sample API responses.

## Tasks / Subtasks

- [x] Task 1: Define error type structs (AC: #1-#6)
  - [x] Create `pkg/opnsense/errors.go`
  - [x] Define `ValidationError` struct with `Fields map[string]string` and `Error()` method
  - [x] Define `NotFoundError` struct with `Message string` and `Error()` method
  - [x] Define `AuthError` struct with `StatusCode int` and `Error()` method
  - [x] Define `ServerError` struct with `Message string`, `Cause error`, `Error()`, and `Unwrap()` methods
  - [x] Define `PluginNotFoundError` struct with `PluginName string` and `Error()` method
  - [x] All error types use pointer receivers so `errors.As()` works correctly
- [x] Task 2: Implement response parsing helpers (AC: #1-#5)
  - [x] Implement `ParseMutationResponse(body []byte) (string, error)` — parses JSON, returns UUID on success or `ValidationError` when `result != "saved"`
  - [x] Implement `CheckHTTPError(statusCode int, endpoint string) error` — returns `AuthError` (401/403), `PluginNotFoundError` (404 on plugin path), or `nil`
  - [x] Implement `NewServerError(endpoint string, cause error) *ServerError` — factory for wrapping transport errors
  - [x] Plugin name extraction: parse first path segment after `/api/` (e.g., `/api/haproxy/settings/getServer` → `haproxy`)
- [x] Task 3: Write unit tests (AC: #7)
  - [x] Create `pkg/opnsense/errors_test.go`
  - [x] Test: `ValidationError.Error()` includes field names and messages
  - [x] Test: `ParseMutationResponse` returns UUID on `{"result":"saved","uuid":"..."}`
  - [x] Test: `ParseMutationResponse` returns `ValidationError` on `{"result":"error","validations":{"port":"..."}}`
  - [x] Test: `ParseMutationResponse` returns `ValidationError` with multiple fields
  - [x] Test: `NotFoundError` implements error interface and works with `errors.As`
  - [x] Test: `CheckHTTPError` returns `AuthError` for 401 and 403
  - [x] Test: `CheckHTTPError` returns `PluginNotFoundError` for 404 with correct plugin name
  - [x] Test: `CheckHTTPError` returns nil for 200
  - [x] Test: `ServerError.Unwrap()` returns the wrapped cause
  - [x] Test: all five error types work with `errors.As()` type assertion
  - [x] Verify all tests pass with `go test ./pkg/opnsense/...`
- [x] Task 4: Verify full pipeline (AC: all)
  - [x] Run `make check` — all targets pass (lint, format, test, security, scan, docs)
  - [x] Run `go build ./...` — succeeds

## Dev Notes

### Previous Story Intelligence (from Stories 1.1-1.3)

**Key learnings to apply:**
- DevRail's `gofumpt` is stricter than `gofmt` — run `make fix` before checking lint
- `go.mod` targets `go 1.25.0`; DevRail container is now 1.8.1 with Go 1.25 builder
- Package comments are required by `revive` linter — `errors.go` is in existing `package opnsense`, no new package comment needed
- Unused parameters must use `_` prefix
- `errcheck` requires handling `resp.Body.Close()` — use `defer func() { _ = resp.Body.Close() }()`
- `gosec` G704 (SSRF) flags `HTTPClient().Do(req)` — suppress with `//nolint:gosec` and explanation when URL is from provider config
- `go-retryablehttp` retries 5xx status codes — tests for non-200 must use non-retryable codes (400, 418) for deterministic behavior
- Test helper pattern: use `t.Helper()` in test helper functions
- Mock server pattern: `httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { ... }))`
- `make check` now passes all 6 targets with DevRail container 1.8.1

**Files from previous stories this story does NOT modify:**
- `pkg/opnsense/client.go` — UNCHANGED (error types are standalone)
- `pkg/opnsense/reqopts.go` — UNCHANGED
- `pkg/opnsense/reconfigure.go` — UNCHANGED

**Files this story creates:**
- `pkg/opnsense/errors.go` — error type definitions and response parsing helpers
- `pkg/opnsense/errors_test.go` — unit tests

### Architecture Compliance

This story implements AR8 (five custom error types), FR62 (validation errors as diagnostics), FR63 (missing resource detection), FR64 (connection/plugin errors), FR65 (permission-specific errors), NFR31-34 (error message quality).

**Critical OPNsense API quirk:**
OPNsense returns HTTP 200 on validation errors. The API client MUST parse the response body and check `result != "saved"` — checking HTTP status alone is insufficient.

**OPNsense mutation response format:**
```json
// Success:
{"result": "saved", "uuid": "550e8400-e29b-41d4-a716-446655440000"}

// Validation failure (still HTTP 200!):
{"result": "failed", "validations": {"port": "value must be between 1 and 65535", "address": "invalid IPv4 address"}}
```

**OPNsense GET response format (monad-wrapped):**
```json
// Normal response:
{"server": {"name": "backend-01", "address": "10.0.0.1", "port": "80"}}

// Blank/default record (resource was deleted out-of-band):
{"server": {"name": "", "address": "", "port": ""}}
```

Note: The monad unwrapping and blank-record detection will be wired into `Get[K]` in Story 1.5. This story defines `NotFoundError` as a type that Story 1.5 will return.

### Critical Implementation Details

**Error type struct definitions:**
```go
// ValidationError is returned when a mutation response has result != "saved".
type ValidationError struct {
    Fields map[string]string // field name → error message
}

func (e *ValidationError) Error() string {
    // Format: "validation failed: field1: msg1; field2: msg2"
}

// NotFoundError is returned when a resource is not found (deleted out-of-band).
type NotFoundError struct {
    Message string
}

func (e *NotFoundError) Error() string { return e.Message }

// AuthError is returned for HTTP 401 or 403 responses.
type AuthError struct {
    StatusCode int
}

func (e *AuthError) Error() string {
    // Format: "authentication failed (HTTP 401)" or "authorization failed (HTTP 403)"
}

// ServerError wraps transport/server failures after retries are exhausted.
type ServerError struct {
    Message string
    Cause   error
}

func (e *ServerError) Error() string { return e.Message }
func (e *ServerError) Unwrap() error { return e.Cause }

// PluginNotFoundError is returned for HTTP 404 on plugin API endpoints.
type PluginNotFoundError struct {
    PluginName string
}

func (e *PluginNotFoundError) Error() string {
    // Format: "plugin 'haproxy' is not installed on OPNsense"
}
```

**ParseMutationResponse function:**
```go
// mutationResponse is the JSON structure returned by OPNsense mutation endpoints.
type mutationResponse struct {
    Result      string            `json:"result"`
    UUID        string            `json:"uuid"`
    Validations map[string]string `json:"validations"`
}

// ParseMutationResponse parses a mutation API response body.
// Returns the UUID on success, or a ValidationError if result != "saved".
func ParseMutationResponse(body []byte) (string, error) {
    var resp mutationResponse
    if err := json.Unmarshal(body, &resp); err != nil {
        return "", fmt.Errorf("failed to parse mutation response: %w", err)
    }
    if resp.Result != "saved" {
        return "", &ValidationError{Fields: resp.Validations}
    }
    return resp.UUID, nil
}
```

**CheckHTTPError function:**
```go
// CheckHTTPError returns an appropriate error type for non-success HTTP status codes.
// Returns nil for HTTP 200. Called before body parsing.
func CheckHTTPError(statusCode int, endpoint string) error {
    switch {
    case statusCode == http.StatusOK:
        return nil
    case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
        return &AuthError{StatusCode: statusCode}
    case statusCode == http.StatusNotFound:
        pluginName := extractPluginName(endpoint)
        return &PluginNotFoundError{PluginName: pluginName}
    default:
        return nil // Other status codes handled by caller or retryablehttp
    }
}
```

**Plugin name extraction:**
```go
// extractPluginName extracts the plugin/module name from an OPNsense API path.
// Example: "/api/haproxy/settings/getServer" → "haproxy"
func extractPluginName(endpoint string) string {
    // Strip /api/ prefix, take first segment
}
```

**errors.As compatibility:**
All error types use pointer receivers on `Error()`. This means `errors.As()` works when the target is a pointer-to-pointer:
```go
var validErr *ValidationError
if errors.As(err, &validErr) {
    // validErr.Fields is accessible
}
```

### What NOT to Build in This Story

- No CRUD functions — those come in Story 1.5
- No monad unwrapping — that comes in Story 1.5's `Get[K]`
- No blank-record detection logic — that comes in Story 1.5 (this story just defines `NotFoundError`)
- No resource-level error handling — that comes in Epic 2+
- No changes to `internal/provider/` — this is all in `pkg/opnsense/`
- No integration with `Reconfigure` or mutex — error types are standalone
- Do NOT import Terraform types in `pkg/opnsense/` — this package is independent of the Terraform framework

### Testing Approach

**ParseMutationResponse tests:**
Use raw JSON byte slices — no mock servers needed:
```go
body := []byte(`{"result":"saved","uuid":"abc-123"}`)
uuid, err := ParseMutationResponse(body)
// Expect uuid == "abc-123", err == nil

body = []byte(`{"result":"failed","validations":{"port":"invalid"}}`)
uuid, err = ParseMutationResponse(body)
// Expect uuid == "", errors.As(err, &ValidationError{})
```

**CheckHTTPError tests:**
Direct function calls with status codes — no mock servers needed:
```go
err := CheckHTTPError(http.StatusUnauthorized, "/api/haproxy/settings/getServer")
// Expect errors.As(err, &AuthError{StatusCode: 401})

err = CheckHTTPError(http.StatusNotFound, "/api/haproxy/settings/getServer")
// Expect errors.As(err, &PluginNotFoundError{PluginName: "haproxy"})
```

**errors.As verification:**
Test every error type with `errors.As()` to ensure pointer receiver pattern works:
```go
var target *ValidationError
if !errors.As(err, &target) {
    t.Fatal("errors.As failed for ValidationError")
}
```

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
    ├── errors.go            # NEW: 5 error types + ParseMutationResponse + CheckHTTPError
    └── errors_test.go       # NEW: unit tests
```

### References

- [Source: architecture.md#Error Handling] — Five error type definitions, trigger conditions, resource handling pattern
- [Source: architecture.md#Cross-Cutting Concerns] — Response parsing, error extraction
- [Source: architecture.md#API Client Design] — OPNsense response format, monad pattern
- [Source: prd.md#Non-Functional Requirements] — NFR31-34 (error message quality)
- [Source: prd.md#Functional Requirements] — FR62-65 (error handling)
- [Source: epics.md#Story 1.4] — Acceptance criteria, BDD scenarios
- [Previous: 1-3-global-mutex-and-reconfigure-infrastructure.md] — DevRail 1.8.1 update, gosec patterns

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- No linting issues encountered — all 6 `make check` targets pass cleanly on first run.

### Completion Notes List

- Five error types implemented: `ValidationError`, `NotFoundError`, `AuthError`, `ServerError`, `PluginNotFoundError`
- All use pointer receivers for `errors.As()` compatibility
- `ValidationError.Error()` sorts fields alphabetically for deterministic output
- `AuthError.Error()` distinguishes "authentication failed" (401) from "authorization failed" (403)
- `ServerError` implements `Unwrap()` for Go error chain traversal
- `ParseMutationResponse` parses OPNsense mutation JSON — returns UUID on success, `ValidationError` on `result != "saved"`
- `CheckHTTPError` maps HTTP status codes: 401/403 → `AuthError`, 404 → `PluginNotFoundError` (with plugin name extraction from URL path), 200 → nil
- `NewServerError` factory wraps transport errors with endpoint context
- `extractPluginName` parses first path segment after `/api/` from endpoint URL
- 22 unit tests covering: all 5 error type messages, `errors.As()` for all 5 types, ParseMutationResponse (success, validation single/multi field, invalid JSON), CheckHTTPError (200/401/403/404 with plugin extraction), ServerError unwrap, plugin name extraction table test
- All tests pass, `make check` passes all 6 targets

### Change Log

- 2026-03-19: Implemented custom error types and response parsing helpers (Story 1.4)

### File List

- `pkg/opnsense/errors.go` — NEW: 5 error types (ValidationError, NotFoundError, AuthError, ServerError, PluginNotFoundError), ParseMutationResponse, CheckHTTPError, NewServerError, extractPluginName
- `pkg/opnsense/errors_test.go` — NEW: 22 unit tests for error types, parsing helpers, and errors.As compatibility
