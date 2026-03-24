// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearch_SinglePage(t *testing.T) {
	var requestPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(searchResponse[testResource]{
			Rows:     []testResource{{Name: "a"}, {Name: "b"}},
			Total:    2,
			RowCount: 500,
			Current:  1,
		})
	}))
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	results, err := Search[testResource](context.Background(), client, opts, SearchParams{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if results[0].Name != "a" || results[1].Name != "b" {
		t.Errorf("unexpected results: %+v", results)
	}
	// Verify SearchEndpoint is used in URL path.
	if requestPath != "/api/test/searchItems" {
		t.Errorf("expected path '/api/test/searchItems', got: %s", requestPath)
	}
}

func TestSearch_MultiPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req searchRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		switch req.Current {
		case 1:
			_ = json.NewEncoder(w).Encode(searchResponse[testResource]{
				Rows:     []testResource{{Name: "a"}, {Name: "b"}},
				Total:    3,
				RowCount: 2,
				Current:  1,
			})
		case 2:
			_ = json.NewEncoder(w).Encode(searchResponse[testResource]{
				Rows:     []testResource{{Name: "c"}},
				Total:    3,
				RowCount: 2,
				Current:  2,
			})
		default:
			t.Errorf("unexpected page %d requested", req.Current)
		}
	}))
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	results, err := Search[testResource](context.Background(), client, opts, SearchParams{RowCount: 2})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 results across 2 pages, got %d", len(results))
	}
	if results[2].Name != "c" {
		t.Errorf("expected third result 'c', got: %s", results[2].Name)
	}
}

func TestSearch_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(searchResponse[testResource]{
			Rows:     []testResource{},
			Total:    0,
			RowCount: 500,
			Current:  1,
		})
	}))
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	results, err := Search[testResource](context.Background(), client, opts, SearchParams{})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// Must be empty slice, not nil.
	if results == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	_, err := Search[testResource](context.Background(), client, opts, SearchParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got: %T: %v", err, err)
	}
}

func TestSearch_ContextCancellation(t *testing.T) {
	firstPageDone := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req searchRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")
		// Always return partial results to force pagination.
		_ = json.NewEncoder(w).Encode(searchResponse[testResource]{
			Rows:     []testResource{{Name: "item"}},
			Total:    100,
			RowCount: 1,
			Current:  req.Current,
		})

		if req.Current == 1 {
			close(firstPageDone)
		}
	}))
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after first page is fetched.
	go func() {
		<-firstPageDone
		cancel()
	}()

	_, err := Search[testResource](ctx, client, opts, SearchParams{RowCount: 1})
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
