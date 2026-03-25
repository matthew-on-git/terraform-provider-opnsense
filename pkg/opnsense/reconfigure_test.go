// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestReconfigure_StandardEndpointSuccess(t *testing.T) {
	var calledPath atomic.Value

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledPath.Store(r.URL.Path)
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newReconfigureTestClient(t, server.URL)
	opts := ReqOpts{
		ReconfigureEndpoint: "/haproxy/service/reconfigure",
	}

	err := Reconfigure(context.Background(), client, opts)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	path, _ := calledPath.Load().(string)
	if path != "/haproxy/service/reconfigure" {
		t.Errorf("expected path %q, got %q", "/haproxy/service/reconfigure", path)
	}
}

func TestReconfigure_StandardEndpointNon200(t *testing.T) {
	// Use 400 (Bad Request) — non-retryable by go-retryablehttp, unlike 5xx.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := newReconfigureTestClient(t, server.URL)
	opts := ReqOpts{
		ReconfigureEndpoint: "/test/reconfigure",
	}

	err := Reconfigure(context.Background(), client, opts)
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
	if !strings.Contains(err.Error(), "unexpected status 400") {
		t.Errorf("expected error with status 400, got: %v", err)
	}
	if !strings.Contains(err.Error(), "/test/reconfigure") {
		t.Errorf("expected error with endpoint path, got: %v", err)
	}
}

func TestReconfigure_CallsReconfigureFunc(t *testing.T) {
	var funcCalled atomic.Bool

	opts := ReqOpts{
		ReconfigureFunc: func(_ context.Context) error {
			funcCalled.Store(true)
			return nil
		},
	}

	// Client URL doesn't matter — ReconfigureFunc should be called instead.
	client := newReconfigureTestClient(t, "http://unused.example.com")

	err := Reconfigure(context.Background(), client, opts)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !funcCalled.Load() {
		t.Error("expected ReconfigureFunc to be called")
	}
}

func TestReconfigure_PropagatesReconfigureFuncError(t *testing.T) {
	expectedErr := fmt.Errorf("savepoint apply failed")

	opts := ReqOpts{
		ReconfigureFunc: func(_ context.Context) error {
			return expectedErr
		},
	}

	client := newReconfigureTestClient(t, "http://unused.example.com")

	err := Reconfigure(context.Background(), client, opts)
	if err == nil {
		t.Fatal("expected error from ReconfigureFunc")
	}
	if err != expectedErr {
		t.Errorf("expected %v, got: %v", expectedErr, err)
	}
}

func TestReconfigure_NilWhenNeitherSet(t *testing.T) {
	opts := ReqOpts{} // No ReconfigureEndpoint, no ReconfigureFunc

	client := newReconfigureTestClient(t, "http://unused.example.com")

	err := Reconfigure(context.Background(), client, opts)
	if err != nil {
		t.Fatalf("expected nil when no reconfigure configured, got: %v", err)
	}
}

func TestReconfigure_ConnectionError(t *testing.T) {
	// Use a closed server so connections are refused.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	serverURL := server.URL
	server.Close()

	client := newReconfigureTestClient(t, serverURL)
	opts := ReqOpts{
		ReconfigureEndpoint: "/test/reconfigure",
	}

	err := Reconfigure(context.Background(), client, opts)
	if err == nil {
		t.Fatal("expected connection error")
	}
	if !strings.Contains(err.Error(), "/test/reconfigure") {
		t.Errorf("expected error with endpoint path, got: %v", err)
	}
}

func TestFirewallFilterReconfigure_Success(t *testing.T) {
	var calls []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/savepoint"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"revision":"1234.5678"}`))
		case strings.Contains(r.URL.Path, "/apply/"):
			w.WriteHeader(http.StatusOK)
		case strings.Contains(r.URL.Path, "/cancelRollback/"):
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client := newReconfigureTestClient(t, server.URL)
	fn := FirewallFilterReconfigure(client)

	err := fn(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d: %v", len(calls), calls)
	}
	if calls[0] != "/api/firewall/filter/savepoint" {
		t.Errorf("step 1: expected savepoint, got %q", calls[0])
	}
	if calls[1] != "/api/firewall/filter/apply/1234.5678" {
		t.Errorf("step 2: expected apply/1234.5678, got %q", calls[1])
	}
	if calls[2] != "/api/firewall/filter/cancelRollback/1234.5678" {
		t.Errorf("step 3: expected cancelRollback/1234.5678, got %q", calls[2])
	}
}

func TestFirewallFilterReconfigure_SavepointFailure(t *testing.T) {
	var calls []string

	// Use 400 — non-retryable.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := newReconfigureTestClient(t, server.URL)
	fn := FirewallFilterReconfigure(client)

	err := fn(context.Background())
	if err == nil {
		t.Fatal("expected error on savepoint failure")
	}
	if !strings.Contains(err.Error(), "savepoint failed") {
		t.Errorf("expected savepoint error, got: %v", err)
	}
	// Only savepoint should have been called — apply/cancelRollback should NOT.
	if len(calls) != 1 {
		t.Errorf("expected 1 call (savepoint only), got %d: %v", len(calls), calls)
	}
}

func TestFirewallFilterReconfigure_ApplyFailure(t *testing.T) {
	var calls []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		switch {
		case strings.HasSuffix(r.URL.Path, "/savepoint"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"revision":"rev123"}`))
		case strings.Contains(r.URL.Path, "/apply/"):
			w.WriteHeader(http.StatusBadRequest) // Apply fails
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client := newReconfigureTestClient(t, server.URL)
	fn := FirewallFilterReconfigure(client)

	err := fn(context.Background())
	if err == nil {
		t.Fatal("expected error on apply failure")
	}
	if !strings.Contains(err.Error(), "apply") {
		t.Errorf("expected apply error, got: %v", err)
	}
	// Savepoint + apply called, but NOT cancelRollback.
	if len(calls) != 2 {
		t.Errorf("expected 2 calls (savepoint + apply), got %d: %v", len(calls), calls)
	}
}

func TestFirewallFilterReconfigure_CancelRollbackFailure(t *testing.T) {
	var calls []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		switch {
		case strings.HasSuffix(r.URL.Path, "/savepoint"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"revision":"rev456"}`))
		case strings.Contains(r.URL.Path, "/apply/"):
			w.WriteHeader(http.StatusOK)
		case strings.Contains(r.URL.Path, "/cancelRollback/"):
			w.WriteHeader(http.StatusBadRequest) // CancelRollback fails
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client := newReconfigureTestClient(t, server.URL)
	fn := FirewallFilterReconfigure(client)

	err := fn(context.Background())
	if err == nil {
		t.Fatal("expected error on cancelRollback failure")
	}
	if !strings.Contains(err.Error(), "cancelRollback") {
		t.Errorf("expected cancelRollback error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "auto-revert in 60 seconds") {
		t.Errorf("expected auto-revert warning, got: %v", err)
	}
	// All 3 calls should have been made.
	if len(calls) != 3 {
		t.Errorf("expected 3 calls, got %d: %v", len(calls), calls)
	}
}

func TestFirewallFilterReconfigure_RevisionPassedCorrectly(t *testing.T) {
	const expectedRevision = "1679912345.9876"
	var applyPath, cancelPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/savepoint"):
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintf(w, `{"revision":"%s"}`, expectedRevision)
		case strings.Contains(r.URL.Path, "/apply/"):
			applyPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
		case strings.Contains(r.URL.Path, "/cancelRollback/"):
			cancelPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := newReconfigureTestClient(t, server.URL)
	fn := FirewallFilterReconfigure(client)

	err := fn(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedApply := "/api/firewall/filter/apply/" + expectedRevision
	if applyPath != expectedApply {
		t.Errorf("apply path: expected %q, got %q", expectedApply, applyPath)
	}

	expectedCancel := "/api/firewall/filter/cancelRollback/" + expectedRevision
	if cancelPath != expectedCancel {
		t.Errorf("cancelRollback path: expected %q, got %q", expectedCancel, cancelPath)
	}
}

func newReconfigureTestClient(t *testing.T, url string) *Client {
	t.Helper()
	client, err := NewClient(ClientConfig{
		BaseURL:   url,
		APIKey:    "testkey",
		APISecret: "testsecret", //nolint:gosec // Test credentials only
		Insecure:  true,
		RetryMax:  1,
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return client
}
