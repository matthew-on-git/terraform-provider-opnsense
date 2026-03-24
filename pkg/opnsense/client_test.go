// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestNewClient_RequiresBaseURL(t *testing.T) {
	_, err := NewClient(ClientConfig{
		APIKey:    "key",
		APISecret: "secret",
	})
	if err == nil {
		t.Fatal("expected error for missing base URL")
	}
}

func TestNewClient_RequiresAPIKey(t *testing.T) {
	_, err := NewClient(ClientConfig{
		BaseURL:   "https://example.com",
		APISecret: "secret",
	})
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestNewClient_RequiresAPISecret(t *testing.T) {
	_, err := NewClient(ClientConfig{
		BaseURL: "https://example.com",
		APIKey:  "key",
	})
	if err == nil {
		t.Fatal("expected error for missing API secret")
	}
}

func TestNewClient_NormalizesBaseURL(t *testing.T) {
	client, err := NewClient(ClientConfig{
		BaseURL:   "https://example.com/",
		APIKey:    "key",
		APISecret: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.BaseURL() != "https://example.com" {
		t.Errorf("expected base URL without trailing slash, got %q", client.BaseURL())
	}
}

func TestAPIKeyTransport_SetsBasicAuth(t *testing.T) {
	var gotUser, gotPass string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, _ = r.BasicAuth()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL:   server.URL,
		APIKey:    "testkey",
		APISecret: "testsecret",
		Insecure:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := client.HTTPClient().Get(server.URL + "/api/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if gotUser != "testkey" {
		t.Errorf("expected Basic Auth user %q, got %q", "testkey", gotUser)
	}
	if gotPass != "testsecret" {
		t.Errorf("expected Basic Auth password %q, got %q", "testsecret", gotPass)
	}
}

func TestClient_RetriesOnServerError(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempt := attempts.Add(1)
		if attempt <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL:   server.URL,
		APIKey:    "key",
		APISecret: "secret",
		Insecure:  true,
		RetryMax:  3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := client.HTTPClient().Get(server.URL + "/api/test")
	if err != nil {
		t.Fatalf("request failed after retries: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 after retries, got %d", resp.StatusCode)
	}
	if got := attempts.Load(); got != 3 {
		t.Errorf("expected 3 attempts (2 failures + 1 success), got %d", got)
	}
}

func TestClient_DoesNotRetryClientError(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL:   server.URL,
		APIKey:    "key",
		APISecret: "secret",
		Insecure:  true,
		RetryMax:  3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := client.HTTPClient().Get(server.URL + "/api/test")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
	if got := attempts.Load(); got != 1 {
		t.Errorf("expected exactly 1 attempt (no retry on 400), got %d", got)
	}
}

func TestNewClient_TLSVerificationDefault(t *testing.T) {
	// Create an HTTPS server with a self-signed certificate.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// With Insecure=false (default), requests to a self-signed cert server should fail.
	client, err := NewClient(ClientConfig{
		BaseURL:   server.URL,
		APIKey:    "key",
		APISecret: "secret",
		RetryMax:  1, // Minimize retries for faster test failure
	})
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	_, err = client.HTTPClient().Get(server.URL + "/api/test")
	if err == nil {
		t.Fatal("expected TLS verification error for self-signed cert with Insecure=false, but request succeeded")
	}
}

func TestNewClient_TLSVerificationInsecure(t *testing.T) {
	// Create an HTTPS server with a self-signed certificate.
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// With Insecure=true, requests to a self-signed cert server should succeed.
	client, err := NewClient(ClientConfig{
		BaseURL:   server.URL,
		APIKey:    "key",
		APISecret: "secret",
		Insecure:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	resp, err := client.HTTPClient().Get(server.URL + "/api/test")
	if err != nil {
		t.Fatalf("expected request to succeed with Insecure=true, got error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
