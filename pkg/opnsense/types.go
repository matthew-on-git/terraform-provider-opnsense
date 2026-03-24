// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
)

// BoolToString converts a Go bool to OPNsense's string boolean format.
// true → "1", false → "0".
func BoolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// StringToBool converts OPNsense's string boolean to a Go bool.
// "1" → true, anything else → false.
func StringToBool(s string) bool {
	return s == "1"
}

// CSVToSlice splits a comma-separated string into a string slice.
// Returns an empty slice (not nil) for empty strings.
func CSVToSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

// SliceToCSV joins a string slice into a comma-separated string.
func SliceToCSV(s []string) string {
	return strings.Join(s, ",")
}

// StringToInt64 converts a string to int64. Returns an error for invalid input.
func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// Int64ToString converts an int64 to its string representation.
func Int64ToString(n int64) string {
	return strconv.FormatInt(n, 10)
}

// SelectedMap is a custom type for OPNsense single-select enum fields.
// OPNsense returns these as {"key": {"value": "...", "selected": 1}}.
// UnmarshalJSON extracts the key where selected == 1.
type SelectedMap string

// selectedEntry represents a single option in the OPNsense SelectedMap.
type selectedEntry struct {
	Value    string      `json:"value"`
	Selected json.Number `json:"selected"`
}

// UnmarshalJSON extracts the selected key from an OPNsense SelectedMap response.
// Handles the "selected" field as both int (1/0) and string ("1"/"0").
func (s *SelectedMap) UnmarshalJSON(data []byte) error {
	var raw map[string]selectedEntry
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

// SelectedMapList is a custom type for OPNsense multi-select fields.
// UnmarshalJSON extracts all keys where selected == 1, sorted alphabetically.
type SelectedMapList []string

// UnmarshalJSON extracts all selected keys from an OPNsense multi-select response.
// Returns a sorted slice for deterministic output. Returns empty slice (not nil) when
// no keys are selected.
func (s *SelectedMapList) UnmarshalJSON(data []byte) error {
	var raw map[string]selectedEntry
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var selected []string
	for key, entry := range raw {
		if entry.Selected.String() == "1" {
			selected = append(selected, key)
		}
	}

	sort.Strings(selected)

	if selected == nil {
		*s = []string{}
	} else {
		*s = selected
	}
	return nil
}
