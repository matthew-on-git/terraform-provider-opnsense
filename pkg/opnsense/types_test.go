// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"encoding/json"
	"testing"
)

// --- BoolToString / StringToBool tests ---

func TestBoolToString(t *testing.T) {
	if BoolToString(true) != "1" {
		t.Errorf("expected '1' for true, got: %s", BoolToString(true))
	}
	if BoolToString(false) != "0" {
		t.Errorf("expected '0' for false, got: %s", BoolToString(false))
	}
}

func TestStringToBool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1", true},
		{"0", false},
		{"", false},
		{"true", false},
		{"yes", false},
	}
	for _, tt := range tests {
		got := StringToBool(tt.input)
		if got != tt.want {
			t.Errorf("StringToBool(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// --- CSVToSlice / SliceToCSV tests ---

func TestCSVToSlice(t *testing.T) {
	tests := []struct {
		input string
		want  int
		first string
	}{
		{"a,b,c", 3, "a"},
		{"a", 1, "a"},
		{"", 0, ""},
		{" a , b , c ", 3, "a"},
	}
	for _, tt := range tests {
		got := CSVToSlice(tt.input)
		if len(got) != tt.want {
			t.Errorf("CSVToSlice(%q) len = %d, want %d", tt.input, len(got), tt.want)
			continue
		}
		if tt.want > 0 && got[0] != tt.first {
			t.Errorf("CSVToSlice(%q)[0] = %q, want %q", tt.input, got[0], tt.first)
		}
	}
}

func TestCSVToSlice_EmptyReturnsEmptySlice(t *testing.T) {
	got := CSVToSlice("")
	if got == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("expected 0 elements, got %d", len(got))
	}
}

func TestSliceToCSV(t *testing.T) {
	if SliceToCSV([]string{"a", "b", "c"}) != "a,b,c" {
		t.Errorf("expected 'a,b,c', got: %s", SliceToCSV([]string{"a", "b", "c"}))
	}
	if SliceToCSV([]string{}) != "" {
		t.Errorf("expected empty string, got: %s", SliceToCSV([]string{}))
	}
}

// --- StringToInt64 / Int64ToString tests ---

func TestStringToInt64(t *testing.T) {
	n, err := StringToInt64("443")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 443 {
		t.Errorf("expected 443, got: %d", n)
	}
}

func TestStringToInt64_Invalid(t *testing.T) {
	_, err := StringToInt64("not-a-number")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestStringToInt64_Empty(t *testing.T) {
	_, err := StringToInt64("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestInt64ToString(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{443, "443"},
		{0, "0"},
		{-1, "-1"},
	}
	for _, tt := range tests {
		got := Int64ToString(tt.input)
		if got != tt.want {
			t.Errorf("Int64ToString(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- SelectedMap tests ---

func TestSelectedMap_SingleSelection(t *testing.T) {
	data := []byte(`{"enabled":{"value":"Enable","selected":1},"disabled":{"value":"Disable","selected":0}}`)
	var sm SelectedMap
	if err := json.Unmarshal(data, &sm); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if string(sm) != "enabled" {
		t.Errorf("expected 'enabled', got: %s", string(sm))
	}
}

func TestSelectedMap_NoSelection(t *testing.T) {
	data := []byte(`{"opt1":{"value":"A","selected":0},"opt2":{"value":"B","selected":0}}`)
	var sm SelectedMap
	if err := json.Unmarshal(data, &sm); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if string(sm) != "" {
		t.Errorf("expected empty string when nothing selected, got: %s", string(sm))
	}
}

func TestSelectedMap_SelectedAsString(t *testing.T) {
	// Some OPNsense endpoints return "selected" as a string instead of int.
	data := []byte(`{"mykey":{"value":"val","selected":"1"}}`)
	var sm SelectedMap
	if err := json.Unmarshal(data, &sm); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if string(sm) != "mykey" {
		t.Errorf("expected 'mykey' for string selected, got: %s", string(sm))
	}
}

func TestSelectedMap_SelectedAsInt(t *testing.T) {
	data := []byte(`{"mykey":{"value":"val","selected":1}}`)
	var sm SelectedMap
	if err := json.Unmarshal(data, &sm); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if string(sm) != "mykey" {
		t.Errorf("expected 'mykey' for int selected, got: %s", string(sm))
	}
}

// --- SelectedMapList tests ---

func TestSelectedMapList_MultipleSelected(t *testing.T) {
	data := []byte(`{"opt1":{"value":"A","selected":1},"opt2":{"value":"B","selected":1},"opt3":{"value":"C","selected":0}}`)
	var sml SelectedMapList
	if err := json.Unmarshal(data, &sml); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(sml) != 2 {
		t.Fatalf("expected 2 selected, got %d", len(sml))
	}
	// Sorted alphabetically.
	if sml[0] != "opt1" || sml[1] != "opt2" {
		t.Errorf("expected ['opt1','opt2'], got: %v", []string(sml))
	}
}

func TestSelectedMapList_NoneSelected(t *testing.T) {
	data := []byte(`{"opt1":{"value":"A","selected":0},"opt2":{"value":"B","selected":0}}`)
	var sml SelectedMapList
	if err := json.Unmarshal(data, &sml); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if sml == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(sml) != 0 {
		t.Errorf("expected 0 selected, got %d", len(sml))
	}
}

func TestSelectedMapList_AllSelected(t *testing.T) {
	data := []byte(`{"a":{"value":"A","selected":1},"b":{"value":"B","selected":1},"c":{"value":"C","selected":1}}`)
	var sml SelectedMapList
	if err := json.Unmarshal(data, &sml); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(sml) != 3 {
		t.Errorf("expected 3 selected, got %d", len(sml))
	}
}

func TestSelectedMapList_SelectedAsString(t *testing.T) {
	data := []byte(`{"k1":{"value":"v","selected":"1"},"k2":{"value":"v","selected":"0"}}`)
	var sml SelectedMapList
	if err := json.Unmarshal(data, &sml); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(sml) != 1 || sml[0] != "k1" {
		t.Errorf("expected ['k1'] for string selected, got: %v", []string(sml))
	}
}
