# Story 1.2: Provider Configuration with Credential Validation

Status: done

## Story

As an operator,
I want to configure the provider with my OPNsense URI, API key, and API secret (via HCL or environment variables),
so that Terraform can connect to my OPNsense appliance and fail clearly if credentials are wrong.

## Acceptance Criteria

1. **AC1: Provider schema has config attributes** — Provider schema includes `uri` (Required string), `api_key` (Optional, Sensitive string), `api_secret` (Optional, Sensitive string), and `insecure` (Optional bool, default false).
2. **AC2: Environment variable fallback** — `OPNSENSE_URI`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET`, `OPNSENSE_ALLOW_INSECURE` environment variables are accepted when HCL attributes are not set. HCL takes priority over env vars.
3. **AC3: Credential validation** — During `Configure`, the provider calls `/api/core/firmware/status` to validate credentials. If valid, the OPNsense version is logged. If invalid (401/403), a clear diagnostic is returned: "Authentication failed — verify API key and secret."
4. **AC4: Client shared with resources** — The configured `*opnsense.Client` is set on `resp.ResourceData` and `resp.DataSourceData` so resources can access it via their `Configure` method.
5. **AC5: Missing required fields** — If `uri` is missing (neither HCL nor env var), the provider returns a clear diagnostic. Same for missing `api_key` and `api_secret`.
6. **AC6: Unit tests** — Tests verify: schema attribute definitions, env var fallback logic, credential validation success/failure, client shared correctly. Tests use mock HTTP server, NOT a real OPNsense instance.

## Tasks / Subtasks

- [x] Task 1: Update provider schema with OPNsense config attributes (AC: #1)
  - [x] Add `uri` as Required StringAttribute with description
  - [x] Add `api_key` as Optional Sensitive StringAttribute with description mentioning env var
  - [x] Add `api_secret` as Optional Sensitive StringAttribute with description mentioning env var
  - [x] Add `insecure` as Optional BoolAttribute with default false and description
  - [x] Create `OpnsenseProviderModel` struct with `tfsdk` tags matching attributes
- [x] Task 2: Implement Configure method with env var fallback (AC: #2, #4, #5)
  - [x] Read config into `OpnsenseProviderModel`
  - [x] For each of `uri`, `api_key`, `api_secret`, `insecure`: if HCL value is null/unknown, check environment variable
  - [x] Validate all required fields are present (uri, api_key, api_secret) — return diagnostic if missing
  - [x] Create `opnsense.Client` using `opnsense.NewClient(opnsense.ClientConfig{...})`
  - [x] Set `resp.ResourceData = client` and `resp.DataSourceData = client`
- [x] Task 3: Implement credential validation via API call (AC: #3)
  - [x] After creating client, make GET request to `{baseURL}/api/core/firmware/status`
  - [x] If HTTP 200: parse response for version info, log with `tflog.Info`
  - [x] If HTTP 401/403: return diagnostic "Authentication failed — verify API key and secret"
  - [x] If connection error: return diagnostic "Unable to connect to OPNsense at {uri}"
  - [x] If any other error: return diagnostic with error details
- [x] Task 4: Write unit tests (AC: #6)
  - [x] Create `internal/provider/provider_test.go`
  - [x] Test: provider schema contains expected attributes with correct types and sensitivity
  - [x] Test: Configure succeeds with valid credentials (mock server returns 200)
  - [x] Test: Configure fails with invalid credentials (mock server returns 401)
  - [x] Test: Configure fails when uri is missing (no HCL, no env var)
  - [x] Test: env var fallback works when HCL attributes are null
  - [x] Test: HCL takes priority over env vars when both are set
  - [x] Verify all tests pass with `go test ./internal/provider/...`
- [x] Task 5: Verify full pipeline (AC: all)
  - [x] Run `make lint` — passes
  - [x] Run `make format` — passes
  - [x] Run `go test ./...` — all tests pass
  - [x] Run `go build ./...` — succeeds

## Dev Notes

### Previous Story Intelligence (from Story 1.1)

**Key learnings to apply:**
- `ClientConfig` struct pattern is used (not positional args) — matches the existing `opnsense.NewClient(cfg ClientConfig)` signature
- DevRail's `gofumpt` is stricter than `gofmt` — run `make fix` before checking lint
- `go.mod` targets `go 1.25.0` to match DevRail container
- Package comments are required by `revive` linter
- Unused parameters must use `_` prefix
- `errcheck` requires handling `resp.Body.Close()` — use `defer func() { _ = resp.Body.Close() }()`
- `gosec` flags field names matching "secret" patterns — use `//nolint:gosec` with explanation

**Files already created that this story modifies:**
- `internal/provider/provider.go` — currently has empty `Configure` method, empty schema, `OpnsenseProvider` struct
- `pkg/opnsense/client.go` — has `Client` struct with `NewClient(cfg ClientConfig)`, `HTTPClient()`, `BaseURL()`

### Architecture Compliance

This story implements FR1-FR5 (Provider Configuration & Authentication) from the PRD.

**Provider Schema Contract (from architecture.md):**

| Element | Value |
|---|---|
| Provider config attributes | `uri`, `api_key`, `api_secret`, `insecure` |
| Environment variable prefix | `OPNSENSE_` |
| Credential validation | `Configure` must validate with test API call and fail fast |

**Key architecture rule:** The `*opnsense.Client` from `pkg/opnsense/` is set on `resp.ResourceData` and `resp.DataSourceData`. Every resource's `Configure` method will type-assert this to `*opnsense.Client`. This pattern is defined in the architecture's "Resource Implementation Patterns" section.

### Critical Implementation Details

**Provider model struct (Terraform Framework pattern):**
```go
type OpnsenseProviderModel struct {
	URI       types.String `tfsdk:"uri"`
	APIKey    types.String `tfsdk:"api_key"`
	APISecret types.String `tfsdk:"api_secret"`
	Insecure  types.Bool   `tfsdk:"insecure"`
}
```

**Environment variable fallback pattern:**
```go
uri := os.Getenv("OPNSENSE_URI")
if !data.URI.IsNull() {
	uri = data.URI.ValueString()
}
```

**Credential validation API call:**
```go
resp, err := client.HTTPClient().Get(client.BaseURL() + "/api/core/firmware/status")
```

The `/api/core/firmware/status` endpoint:
- Returns HTTP 200 with JSON `{"status": "..."}` when credentials are valid
- Returns HTTP 401/403 when credentials are invalid
- Does not require any specific OPNsense permissions beyond basic API access

**Logging with terraform-plugin-log:**
```go
import "github.com/hashicorp/terraform-plugin-log/tflog"

tflog.Info(ctx, "OPNsense connection validated", map[string]interface{}{
	"url": uri,
})
```

### What NOT to Build in This Story

- No mutex or semaphore (Story 1.3)
- No error types beyond basic HTTP status checking (Story 1.4)
- No CRUD functions (Story 1.5)
- No resources or data sources — just the provider configuration
- No acceptance tests — unit tests only with mock HTTP server

### Imports to Add

```go
// internal/provider/provider.go will need:
"os"
"github.com/hashicorp/terraform-plugin-framework/types"
"github.com/hashicorp/terraform-plugin-log/tflog"
"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
```

`terraform-plugin-log` is already an indirect dependency via the framework. It should be promoted to a direct dependency when imported.

### Testing Approach

**Use `resource.Test` from `terraform-plugin-testing` for provider-level tests:**

For schema validation tests, test the provider schema directly. For Configure tests, use `httptest.NewServer` to mock the OPNsense API:
- Return 200 with `{"status": "ok"}` for valid credentials
- Return 401 for invalid credentials
- Use `t.Setenv("OPNSENSE_URI", server.URL)` for env var tests

**Important:** Provider Configure tests in the Framework are more complex than unit tests — they need the full Terraform testing harness. However, for this story, we can test the *logic* (env var resolution, validation call) with simpler unit tests of helper functions, and defer the full provider acceptance test to Story 2.1.

### Project Structure After This Story

```
terraform-provider-opnsense/
├── internal/
│   └── provider/
│       ├── provider.go          # MODIFIED: schema, Configure, model struct
│       └── provider_test.go     # NEW: unit tests
├── pkg/
│   └── opnsense/
│       ├── client.go            # UNCHANGED
│       └── client_test.go       # UNCHANGED
└── ... (other files unchanged)
```

### References

- [Source: architecture.md#API Client Design] — Client config pattern
- [Source: architecture.md#Core Architectural Decisions] — Provider schema contract table
- [Source: architecture.md#Resource Implementation Patterns] — Resource Configure pattern (client access)
- [Source: prd.md#Provider Configuration & Authentication] — FR1-FR5
- [Source: prd.md#Non-Functional Requirements#Security] — NFR7 (Sensitive), NFR9 (TLS config)
- [Source: epics.md#Story 1.2] — Acceptance criteria
- [Previous: 1-1-initialize-project-and-api-client-core.md] — ClientConfig pattern, DevRail learnings

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- `make check` security target fails with pre-existing Go version mismatch (container Go 1.24 vs project go 1.25.0) — govulncheck cannot load packages. Not caused by this story's changes.

### Completion Notes List

- Implemented `OpnsenseProviderModel` struct with all four attributes (`uri`, `api_key`, `api_secret`, `insecure`) and proper `tfsdk` tags
- Provider schema defines `uri` as Required, `api_key`/`api_secret` as Optional+Sensitive, `insecure` as Optional
- `Configure` method reads HCL config, falls back to environment variables (`OPNSENSE_URI`, `OPNSENSE_API_KEY`, `OPNSENSE_API_SECRET`, `OPNSENSE_ALLOW_INSECURE`), validates required fields, creates `*opnsense.Client`, and shares it via `resp.ResourceData`/`resp.DataSourceData`
- Credential validation calls `/api/core/firmware/status` — handles 200 (success with tflog + version parsing), 401/403 (auth failure diagnostic), connection errors, and unexpected status codes
- Helper functions `envOrValue` and `envOrBoolValue` encapsulate the HCL-priority-over-env-var logic
- Extracted `resolvedConfig`, `resolveProviderConfig`, and `validateRequiredConfig` as testable functions
- 24 unit tests covering: schema attributes, config resolution (env var fallback, HCL priority, empty state), required field validation (missing URI/key/secret, all missing), credential validation (success, version logging, malformed JSON, 401, 403, connection error, unexpected status), env var helpers
- Tests use `httptest.NewServer` mock servers and direct function calls — no real OPNsense instance required
- All tests pass, lint passes, format passes, build succeeds
- `terraform-plugin-testing` added as test dependency; `terraform-plugin-log` promoted to direct dependency

### Code Review Notes

**Reviewed by:** Claude Opus 4.6 (same session — recommend independent re-review with different LLM)
**Date:** 2026-03-18
**Findings:** 0 High, 2 Medium, 1 Low — all Medium issues fixed:
- M1: Added version parsing from firmware status response (AC3 compliance)
- M2: Extracted `resolveProviderConfig` and `validateRequiredConfig` into testable functions with 8 new tests covering missing-field validation and env var resolution paths
- L1: `insecure` default — confirmed non-issue for provider schema (TF Framework doesn't support Computed/Default on provider attributes; runtime behavior is correct)

### Change Log

- 2026-03-18: Implemented provider configuration with credential validation (Story 1.2)
- 2026-03-18: Code review fixes — added version logging from firmware response, extracted and tested config resolution and validation functions

### File List

- `internal/provider/provider.go` — MODIFIED: schema, OpnsenseProviderModel, Configure with env var fallback, credential validation with version parsing, extracted resolveProviderConfig and validateRequiredConfig
- `internal/provider/provider_test.go` — NEW: 24 unit tests for schema, config resolution, required field validation, credential validation, and env var helpers
- `go.mod` — MODIFIED: added terraform-plugin-testing dependency, promoted terraform-plugin-log
- `go.sum` — MODIFIED: updated checksums for new/updated dependencies
