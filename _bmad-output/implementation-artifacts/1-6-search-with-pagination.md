# Story 1.6: Search with Pagination

Status: done

## Story

As a developer,
I want a `Search[K]` function that transparently iterates paginated OPNsense search results,
so that data sources and list operations return complete results regardless of page size.

## Acceptance Criteria

1. **AC1: Search[K] function** — `Search[K]` POSTs to `SearchEndpoint` with `rowCount` and `current` pagination parameters, iterates all pages transparently, and returns `[]K` with all results accumulated.
2. **AC2: Pagination handling** — The function automatically increments the `current` page until all rows are collected (determined by `rowCount` in the response). Empty result sets return `[]K{}` (not nil, not error).
3. **AC3: SearchParams struct** — A `SearchParams` struct is defined with configurable `RowCount` (default 500) and optional `SearchQuery` for filtering.
4. **AC4: Read semaphore** — Search acquires the read semaphore (`AcquireRead`) since it's a read-only operation. The semaphore is held per-page-request, not for the entire pagination loop.
5. **AC5: Error handling** — HTTP errors checked via `CheckHTTPError` on each page. Transport errors wrapped with `NewServerError`. Context cancellation respected between pages.
6. **AC6: Unit tests** — Tests verify: single page, multi-page pagination, empty results, context cancellation between pages, semaphore acquisition. Tests use mock HTTP server.

## Tasks / Subtasks

- [x] Task 1: Define SearchParams and search response types (AC: #3)
  - [x] Create `pkg/opnsense/search.go`
  - [x] Define `SearchParams` struct with `RowCount int` (default 500) and `SearchQuery string`
  - [x] Define `searchResponse[K]` struct for the OPNsense search response envelope: `Rows []K`, `RowCount int`, `Total int`, `Current int`
- [x] Task 2: Implement Search[K] with pagination (AC: #1, #2, #4, #5)
  - [x] Implement `Search[K any](ctx context.Context, c *Client, opts ReqOpts, params SearchParams) ([]K, error)`
  - [x] Apply default RowCount (500) if not set
  - [x] Loop: POST to `{BaseURL}{SearchEndpoint}` with JSON body `{"current": page, "rowCount": rowCount, "searchPhrase": query}`
  - [x] Acquire/release read semaphore per page request (not for entire loop)
  - [x] Parse response: extract `rows`, `rowCount`, `total`, `current`
  - [x] Accumulate rows from each page into results slice
  - [x] Stop when accumulated rows >= total (or page returns empty rows)
  - [x] Check `ctx.Err()` between pages for cancellation
  - [x] Return accumulated `[]K` (empty slice, not nil, for zero results)
- [x] Task 3: Write unit tests (AC: #6)
  - [x] Create `pkg/opnsense/search_test.go`
  - [x] Test: single page — all results fit in one response
  - [x] Test: multi-page — mock returns 2+ pages, verify all rows accumulated
  - [x] Test: empty results — mock returns zero rows, verify empty slice returned (not nil)
  - [x] Test: HTTP error — mock returns 401, verify `AuthError` returned
  - [x] Test: context cancellation — cancel context between pages, verify error
  - [x] Verify all tests pass with `go test ./pkg/opnsense/...`
- [x] Task 4: Verify full pipeline (AC: all)
  - [x] Run `make check` — all targets pass
  - [x] Run `go build ./...` — succeeds

## Dev Notes

### Previous Story Intelligence (from Stories 1.1-1.5)

**Key learnings to apply:**
- `ctx context.Context` MUST be the first parameter (revive linter enforces this)
- `gosec` G704 flags `HTTPClient().Do(req)` — suppress with `//nolint:gosec`
- `errcheck` requires `defer func() { _ = resp.Body.Close() }()`
- `go-retryablehttp` retries 5xx — tests for non-200 use non-retryable codes (400, 418)
- `make check` passes all 6 targets with DevRail container 1.8.1

**Existing code this story uses:**
- `client.go`: `Client.AcquireRead(ctx)`, `Client.ReleaseRead()`, `Client.HTTPClient()`, `Client.BaseURL()`
- `reqopts.go`: `ReqOpts.SearchEndpoint`
- `errors.go`: `CheckHTTPError`, `NewServerError`
- `crud.go`: Follow the same patterns for HTTP request construction and error handling

### Architecture Compliance

This story implements AR3 (Search[K] generic function), the transparent pagination cross-cutting concern, and NFR performance targets (60s full plan refresh).

**OPNsense search API pattern:**
Search endpoints accept POST with JSON body containing pagination params. The response wraps results in a `rows` array with pagination metadata.

**Request body:**
```json
{
  "current": 1,
  "rowCount": 500,
  "searchPhrase": ""
}
```

**Response format:**
```json
{
  "rows": [
    {"uuid": "abc-123", "name": "server1", "address": "10.0.0.1"},
    {"uuid": "def-456", "name": "server2", "address": "10.0.0.2"}
  ],
  "rowCount": 500,
  "total": 1247,
  "current": 1
}
```

- `rows`: array of result objects for the current page
- `rowCount`: rows per page (as requested)
- `total`: total matching rows across all pages
- `current`: current page number (1-indexed)

**Pagination loop logic:**
```
page = 1
results = []
loop:
    POST with {current: page, rowCount: 500}
    parse response
    append rows to results
    if len(results) >= total: break
    if len(rows) == 0: break  // safety: no infinite loop
    page++
```

**Semaphore per page, not per search:**
Acquire/release the read semaphore for each individual page request. This prevents a single large search from monopolizing all read slots for the entire pagination duration. Between pages, other reads can proceed.

### Critical Implementation Details

**Search is a POST, not a GET:**
OPNsense search endpoints use POST with a JSON body containing pagination params. This is different from typical REST where search would be GET with query params.

**Content-Type must be set:**
```go
req.Header.Set("Content-Type", "application/json")
```

**Response parsing — the search response is NOT monad-wrapped:**
Unlike `Get[K]` which returns `{monad: {...}}`, search returns `{rows: [...], total: N, ...}`. Do NOT use `unmarshalWrapped` — search has its own response format.

**Empty results return empty slice, not nil:**
```go
if len(allResults) == 0 {
    return []K{}, nil  // Empty slice, not nil
}
```

**searchResponse type for JSON parsing:**
```go
type searchResponse[K any] struct {
    Rows     []K `json:"rows"`
    RowCount int `json:"rowCount"`
    Total    int `json:"total"`
    Current  int `json:"current"`
}
```

Note: Go generics in struct type parameters require Go 1.18+. Since we target Go 1.25.0, this is fine. However, `json.Unmarshal` into a generic struct works correctly.

**searchRequest type for JSON body:**
```go
type searchRequest struct {
    Current      int    `json:"current"`
    RowCount     int    `json:"rowCount"`
    SearchPhrase string `json:"searchPhrase"`
}
```

### What NOT to Build in This Story

- No CRUD modifications — `Add`, `Get`, `Update`, `Delete` are unchanged (Story 1.5)
- No type conversion utilities — Story 1.7
- No resources or data sources — Epic 2+
- No changes to `internal/provider/`
- Do NOT use the write mutex — Search is a read-only operation

### Testing Approach

**Multi-page pagination test:**
Mock server tracks `current` param from request body and returns different pages:
```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    var req searchRequest
    json.NewDecoder(r.Body).Decode(&req)

    switch req.Current {
    case 1:
        // Return page 1 with total=3, rowCount=2
        json.NewEncoder(w).Encode(searchResponse[testResource]{
            Rows: []testResource{{Name: "a"}, {Name: "b"}},
            Total: 3, RowCount: 2, Current: 1,
        })
    case 2:
        // Return page 2 with remaining row
        json.NewEncoder(w).Encode(searchResponse[testResource]{
            Rows: []testResource{{Name: "c"}},
            Total: 3, RowCount: 2, Current: 2,
        })
    }
}))
```

**Context cancellation test:**
Use `context.WithCancel`, cancel between pages in the mock handler.

**Reuse `testResource` from `crud_test.go`** — it's in the same package.

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
    ├── crud.go              # UNCHANGED
    ├── crud_test.go         # UNCHANGED
    ├── search.go            # NEW: SearchParams, Search[K], pagination logic
    └── search_test.go       # NEW: pagination tests
```

### References

- [Source: architecture.md#API Client Design] — Search[K] function signature, SearchParams
- [Source: architecture.md#Cross-Cutting Concerns] — Transparent pagination, read semaphore
- [Source: architecture.md#Performance NFRs] — 60s plan refresh target
- [Source: epics.md#Story 1.6] — Acceptance criteria, BDD scenarios
- [Previous: 1-5-generic-crud-functions.md] — CRUD patterns, revive ctx-first, gosec suppression

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- `revive` flagged empty busy-wait loop in test — replaced with channel-based synchronization (`firstPageDone` channel).

### Completion Notes List

- `SearchParams` struct with `RowCount` (default 500) and `SearchQuery`
- `searchRequest`/`searchResponse[K]` types for OPNsense search API JSON format
- `Search[K]` transparently iterates pages: POST with `{current, rowCount, searchPhrase}`, accumulate rows until `len(results) >= total` or empty page
- `searchPage[K]` helper acquires/releases read semaphore per individual page request (not per entire search)
- Context cancellation checked between pages via `ctx.Err()`
- Empty results return `[]K{}` (empty slice, not nil)
- 5 unit tests: single page, multi-page (2 pages accumulating 3 results), empty results (verified non-nil), HTTP 401 error, context cancellation between pages
- All tests pass, `make check` passes all 6 targets

### Code Review Notes

**Reviewed by:** Claude Opus 4.6 (same session)
**Date:** 2026-03-23
**Findings:** 0 High, 1 Medium, 0 Low — fixed:
- M1: `testReqOpts()` didn't set `SearchEndpoint`, so search tests didn't verify URL path construction. Fixed by adding `SearchEndpoint` to shared helper and verifying path in `TestSearch_SinglePage`.

### Change Log

- 2026-03-19: Implemented Search[K] with transparent pagination (Story 1.6)
- 2026-03-23: Code review fix — added SearchEndpoint to testReqOpts, added URL path verification in test

### File List

- `pkg/opnsense/search.go` — NEW: SearchParams, Search[K] with pagination, searchPage helper, searchRequest/searchResponse types
- `pkg/opnsense/search_test.go` — NEW: 5 tests for single page (with URL path verification), multi-page, empty, HTTP error, context cancellation
- `pkg/opnsense/crud_test.go` — MODIFIED: added SearchEndpoint to testReqOpts()
