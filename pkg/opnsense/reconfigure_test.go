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
