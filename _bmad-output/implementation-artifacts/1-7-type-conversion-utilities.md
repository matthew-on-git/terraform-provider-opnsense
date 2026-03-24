# Story 1.7: Type Conversion Utilities

Status: done

## Story

As a developer,
I want shared type conversion utilities for OPNsense's non-standard types (string bools, SelectedMap, CSVList),
so that every resource can convert between OPNsense API types, Go model types, and Terraform types consistently.

## Acceptance Criteria

1. **AC1: String boolean conversion** — `StringToBool("1")` returns `true`, `StringToBool("0")` returns `false`. `BoolToString(true)` returns `"1"`, `BoolToString(false)` returns `"0"`.
2. **AC2: SelectedMap type** — A custom `SelectedMap` type with `UnmarshalJSON` correctly unmarshals `{"key1": {"value": "...", "selected": 1}, "key2": {"value": "...", "selected": 0}}` and extracts the selected key (`"key1"`). Returns empty string when no key is selected.
3. **AC3: SelectedMapList type** — A custom `SelectedMapList` type with `UnmarshalJSON` returns `[]string` of all selected keys from multi-select fields.
4. **AC4: CSV list handling** — `CSVToSlice("a,b,c")` returns `["a","b","c"]`. `SliceToCSV(["a","b","c"])` returns `"a,b,c"`. Empty strings return empty slices.
5. **AC5: Integer-as-string conversion** — `StringToInt64("443")` returns `int64(443)`. `Int64ToString(443)` returns `"443"`. Invalid strings return an error.
6. **AC6: Unit tests** — Tests cover all type conversions including edge cases (empty strings, missing keys, malformed JSON, no selections, multiple selections).

## Tasks / Subtasks

- [x] Task 1: Implement basic type converters (AC: #1, #4, #5)
  - [x] Create `pkg/opnsense/types.go`
  - [x] Implement `BoolToString(b bool) string` — returns `"1"` or `"0"`
  - [x] Implement `StringToBool(s string) bool` — returns true for `"1"`, false otherwise
  - [x] Implement `CSVToSlice(s string) []string` — splits on comma, returns empty slice for empty string
  - [x] Implement `SliceToCSV(s []string) string` — joins with comma
  - [x] Implement `StringToInt64(s string) (int64, error)` — wraps `strconv.ParseInt`
  - [x] Implement `Int64ToString(n int64) string` — wraps `strconv.FormatInt`
- [x] Task 2: Implement SelectedMap type (AC: #2)
  - [x] Define `SelectedMap` type (underlying `string`)
  - [x] Implement `UnmarshalJSON` — parse `{"key": {"value": "...", "selected": 1}}`, extract key where `selected == 1`
  - [x] Return empty string when no key is selected
  - [x] Handle the `selected` field being either `int` (1/0) or `string` ("1"/"0") — OPNsense is inconsistent
- [x] Task 3: Implement SelectedMapList type (AC: #3)
  - [x] Define `SelectedMapList` type (underlying `[]string`)
  - [x] Implement `UnmarshalJSON` — parse same structure, extract ALL keys where `selected == 1`
  - [x] Return empty slice (not nil) when no keys are selected
  - [x] Sort selected keys for deterministic output
- [x] Task 4: Write unit tests (AC: #6)
  - [x] Create `pkg/opnsense/types_test.go`
  - [x] Test: `BoolToString` for true and false
  - [x] Test: `StringToBool` for "1", "0", and empty string
  - [x] Test: `CSVToSlice` for normal, empty, single-item, and whitespace cases
  - [x] Test: `SliceToCSV` for normal and empty cases
  - [x] Test: `StringToInt64` for valid, invalid, and empty string
  - [x] Test: `Int64ToString` for positive, zero, and negative
  - [x] Test: `SelectedMap.UnmarshalJSON` — single selection, no selection, multiple candidates
  - [x] Test: `SelectedMapList.UnmarshalJSON` — multiple selected, none selected, all selected
  - [x] Test: `SelectedMap` handles `selected` as both int and string
  - [x] Verify all tests pass with `go test ./pkg/opnsense/...`
- [x] Task 5: Verify full pipeline (AC: all)
  - [x] Run `make check` — all targets pass
  - [x] Run `go build ./...` — succeeds

## Dev Notes

### Previous Story Intelligence (from Stories 1.1-1.6)

**Key learnings to apply:**
- `ctx context.Context` MUST be first parameter (revive enforces)
- `make check` passes all 6 targets with DevRail container 1.8.1
- Package comments not needed — `types.go` is in existing `package opnsense`
- Sort keys deterministically (used in `ValidationError.Error()` — apply same pattern to `SelectedMapList`)
- No Terraform types in `pkg/opnsense/` — this package is framework-independent

**Existing code patterns to follow:**
- Error type struct pattern from `errors.go` (pointer receivers, `Error()` method)
- `sort.Strings(keys)` for deterministic output (from `errors.go:29`)

### Architecture Compliance

This story implements AR4 (type conversion utilities), the type conversion cross-cutting concern, and provides the `toAPI()`/`fromAPI()` building blocks for all resource models in Epic 2+.

**Three-layer type conversion model:**
```
OPNsense API (strings) ←→ Go model (typed) ←→ Terraform Framework (types.*)
```

Resources call `toAPI()` and `fromAPI()` which use these shared converters. The converters live in `pkg/opnsense/types.go`. Per-resource `toAPI()`/`fromAPI()` methods live in `internal/service/{module}/{resource}_model.go` (Epic 2+).

### Critical Implementation Details

**BoolToString / StringToBool:**
```go
func BoolToString(b bool) string {
    if b { return "1" }
    return "0"
}

func StringToBool(s string) bool {
    return s == "1"
}
```

**SelectedMap — OPNsense response format:**
```json
{
  "enabled": {"value": "Enable this item", "selected": 1},
  "disabled": {"value": "Disable this item", "selected": 0}
}
```
The `selected` field can be either `int` (1/0) or `string` ("1"/"0") — OPNsense is inconsistent across endpoints. Use `json.Number` or check both types.

**SelectedMap implementation pattern:**
```go
type SelectedMap string

func (s *SelectedMap) UnmarshalJSON(data []byte) error {
    var raw map[string]struct {
        Value    string      `json:"value"`
        Selected json.Number `json:"selected"`
    }
    if err := json.Unmarshal(data, &raw); err != nil {
        return err
    }
    for key, entry := range raw {
        if entry.Selected.String() == "1" {
            *s = SelectedMap(key)
            return nil
        }
    }
    *s = ""
    return nil
}
```

Using `json.Number` handles both `"selected": 1` (int) and `"selected": "1"` (string) cases.

**SelectedMapList implementation pattern:**
```go
type SelectedMapList []string

func (s *SelectedMapList) UnmarshalJSON(data []byte) error {
    // Same structure, but collect ALL keys where selected == "1"
    // Sort for deterministic output
}
```

**CSVToSlice edge cases:**
- `""` → `[]string{}` (empty slice, not `[""]`)
- `"a"` → `["a"]` (single item)
- `"a,b,c"` → `["a", "b", "c"]` (normal)
- Trim whitespace from each element

**StringToInt64:**
```go
func StringToInt64(s string) (int64, error) {
    return strconv.ParseInt(s, 10, 64)
}
```

### What NOT to Build in This Story

- No `toAPI()` / `fromAPI()` resource model methods — those come in Epic 2 per-resource
- No Terraform type conversions (`types.String` ↔ `string`) — that's in the resource model layer
- No resources or data sources
- No changes to `internal/provider/`
- Do NOT import `hashicorp/terraform-plugin-framework` in `pkg/opnsense/`

### Testing Approach

**SelectedMap tests with raw JSON:**
```go
var sm SelectedMap
json.Unmarshal([]byte(`{"opt1":{"value":"A","selected":1},"opt2":{"value":"B","selected":0}}`), &sm)
// Expect sm == "opt1"
```

**SelectedMap with int vs string selected:**
```go
// Integer selected (common):
`{"k":{"value":"v","selected":1}}`

// String selected (some endpoints):
`{"k":{"value":"v","selected":"1"}}`
```

**Reuse patterns from errors_test.go** — direct function calls with expected values, no mock servers needed.

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
    ├── search.go            # UNCHANGED
    ├── search_test.go       # UNCHANGED
    ├── types.go             # NEW: BoolToString, StringToBool, SelectedMap, SelectedMapList, CSVToSlice, SliceToCSV, StringToInt64, Int64ToString
    └── types_test.go        # NEW: unit tests for all type converters
```

### References

- [Source: architecture.md#Type Conversion] — Three-layer type model, conversion table
- [Source: architecture.md#Cross-Cutting Concerns] — Type conversion as cross-cutting
- [Source: architecture.md#Resource Model Patterns] — toAPI()/fromAPI() usage
- [Source: epics.md#Story 1.7] — Acceptance criteria, BDD scenarios
- [Previous: 1-4-custom-error-types-and-response-parsing.md] — sort.Strings pattern for deterministic output

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (1M context)

### Debug Log References

- No linting issues — all 6 `make check` targets pass cleanly on first run.

### Completion Notes List

- 6 converter functions: `BoolToString`/`StringToBool`, `CSVToSlice`/`SliceToCSV`, `StringToInt64`/`Int64ToString`
- `SelectedMap` type with `UnmarshalJSON` using `json.Number` to handle both int and string `selected` fields
- `SelectedMapList` type with `UnmarshalJSON` collecting all selected keys, sorted with `sort.Strings` for deterministic output
- Shared `selectedEntry` struct reused by both SelectedMap and SelectedMapList
- `CSVToSlice` trims whitespace from each element, returns empty slice for empty input
- `SelectedMapList` returns empty slice (not nil) when no keys selected
- 18 unit tests covering: BoolToString (2), StringToBool (5 table), CSVToSlice (4 table + 1 nil check), SliceToCSV (2), StringToInt64 (3), Int64ToString (3 table), SelectedMap (4 — selection, no selection, string selected, int selected), SelectedMapList (4 — multiple, none, all, string selected)
- All tests pass, `make check` passes all 6 targets

### Change Log

- 2026-03-23: Implemented type conversion utilities (Story 1.7)

### File List

- `pkg/opnsense/types.go` — NEW: BoolToString, StringToBool, CSVToSlice, SliceToCSV, StringToInt64, Int64ToString, SelectedMap, SelectedMapList with UnmarshalJSON
- `pkg/opnsense/types_test.go` — NEW: 18 unit tests for all type converters and edge cases
