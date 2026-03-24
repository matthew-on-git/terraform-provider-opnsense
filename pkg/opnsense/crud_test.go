// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// testResource is a minimal struct for testing CRUD generics.
type testResource struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    string `json:"port"`
}

// --- Add tests ---

func TestAdd_Success(t *testing.T) {
	var reconfigureCalled atomic.Bool
	var receivedBody string

	server := newCRUDTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/test/addItem"):
			body, _ := io.ReadAll(r.Body)
			receivedBody = string(body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"result":"saved","uuid":"test-uuid-123"}`))
		case strings.HasPrefix(r.URL.Path, "/api/test/service/reconfigure"):
			reconfigureCalled.Store(true)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()
	res := &testResource{Name: "web1", Address: "10.0.0.1", Port: "80"}

	uuid, err := Add(context.Background(), client, opts, res)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if uuid != "test-uuid-123" {
		t.Errorf("expected UUID 'test-uuid-123', got: %s", uuid)
	}

	// Verify monad wrapping.
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal([]byte(receivedBody), &parsed); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	if _, ok := parsed["item"]; !ok {
		t.Error("expected request body wrapped in monad key 'item'")
	}

	// Verify reconfigure was called.
	if !reconfigureCalled.Load() {
		t.Error("expected reconfigure to be called after Add")
	}
}

func TestAdd_ValidationError(t *testing.T) {
	server := newCRUDTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/test/addItem"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"result":"failed","validations":{"port":"invalid value"}}`))
		case strings.HasPrefix(r.URL.Path, "/api/test/service/reconfigure"):
			w.WriteHeader(http.StatusOK)
		}
	})
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()
	res := &testResource{Name: "web1"}

	_, err := Add(context.Background(), client, opts, res)
	if err == nil {
		t.Fatal("expected ValidationError, got nil")
	}
	var validErr *ValidationError
	if !errors.As(err, &validErr) {
		t.Fatalf("expected ValidationError, got: %T: %v", err, err)
	}
	if validErr.Fields["port"] != "invalid value" {
		t.Errorf("expected port validation, got: %v", validErr.Fields)
	}
}

func TestAdd_AuthError(t *testing.T) {
	server := newCRUDTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()
	res := &testResource{Name: "web1"}

	_, err := Add(context.Background(), client, opts, res)
	if err == nil {
		t.Fatal("expected AuthError, got nil")
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError, got: %T: %v", err, err)
	}
	if authErr.StatusCode != 401 {
		t.Errorf("expected status 401, got: %d", authErr.StatusCode)
	}
}

// --- Get tests ---

func TestGet_Success(t *testing.T) {
	server := newCRUDTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/test/getItem/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"item":{"name":"web1","address":"10.0.0.1","port":"80"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	result, err := Get[testResource](context.Background(), client, opts, "uuid-123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.Name != "web1" {
		t.Errorf("expected name 'web1', got: %s", result.Name)
	}
	if result.Address != "10.0.0.1" {
		t.Errorf("expected address '10.0.0.1', got: %s", result.Address)
	}
	if result.Port != "80" {
		t.Errorf("expected port '80', got: %s", result.Port)
	}
}

func TestGet_NotFound_EmptyMonad(t *testing.T) {
	server := newCRUDTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/test/getItem/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"item":{}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	_, err := Get[testResource](context.Background(), client, opts, "uuid-gone")
	if err == nil {
		t.Fatal("expected NotFoundError, got nil")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got: %T: %v", err, err)
	}
}

func TestGet_NotFound_MissingMonadKey(t *testing.T) {
	server := newCRUDTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/test/getItem/") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"other_key":{"name":"wrong"}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	_, err := Get[testResource](context.Background(), client, opts, "uuid-missing")
	if err == nil {
		t.Fatal("expected NotFoundError for missing monad key")
	}
	var notFound *NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got: %T: %v", err, err)
	}
}

func TestGet_DoesNotCallReconfigure(t *testing.T) {
	var reconfigureCalled atomic.Bool

	server := newCRUDTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/test/getItem/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"item":{"name":"web1","address":"10.0.0.1","port":"80"}}`))
		case strings.HasPrefix(r.URL.Path, "/api/test/service/reconfigure"):
			reconfigureCalled.Store(true)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	_, err := Get[testResource](context.Background(), client, opts, "uuid-123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if reconfigureCalled.Load() {
		t.Error("Get should NOT call reconfigure — it's a read-only operation")
	}
}

// --- Update tests ---

func TestUpdate_Success(t *testing.T) {
	var reconfigureCalled atomic.Bool
	var requestPath string

	server := newCRUDTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/test/setItem/"):
			requestPath = r.URL.Path
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"result":"saved"}`))
		case strings.HasPrefix(r.URL.Path, "/api/test/service/reconfigure"):
			reconfigureCalled.Store(true)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()
	res := &testResource{Name: "web1-updated", Address: "10.0.0.2", Port: "8080"}

	err := Update(context.Background(), client, opts, res, "uuid-123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify URL includes ID.
	if !strings.HasSuffix(requestPath, "/uuid-123") {
		t.Errorf("expected URL with ID, got: %s", requestPath)
	}

	if !reconfigureCalled.Load() {
		t.Error("expected reconfigure to be called after Update")
	}
}

// --- Delete tests ---

func TestDelete_Success(t *testing.T) {
	var reconfigureCalled atomic.Bool
	var requestPath string
	var requestMethod string

	server := newCRUDTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/test/delItem/"):
			requestPath = r.URL.Path
			requestMethod = r.Method
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"result":"deleted"}`))
		case strings.HasPrefix(r.URL.Path, "/api/test/service/reconfigure"):
			reconfigureCalled.Store(true)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	err := Delete(context.Background(), client, opts, "uuid-456")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify URL includes ID.
	if !strings.HasSuffix(requestPath, "/uuid-456") {
		t.Errorf("expected URL with ID, got: %s", requestPath)
	}

	// Verify POST method (OPNsense uses POST for deletes).
	if requestMethod != http.MethodPost {
		t.Errorf("expected POST for delete, got: %s", requestMethod)
	}

	if !reconfigureCalled.Load() {
		t.Error("expected reconfigure to be called after Delete")
	}
}

func TestDelete_DoesNotParseBody(t *testing.T) {
	// Delete response has result="deleted" which is not "saved".
	// This should NOT cause a ValidationError — Delete ignores the body.
	server := newCRUDTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/test/delItem/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"result":"deleted"}`))
		case strings.HasPrefix(r.URL.Path, "/api/test/service/reconfigure"):
			w.WriteHeader(http.StatusOK)
		}
	})
	defer server.Close()

	client := newCRUDTestClient(t, server.URL)
	opts := testReqOpts()

	err := Delete(context.Background(), client, opts, "uuid-789")
	if err != nil {
		var validErr *ValidationError
		if errors.As(err, &validErr) {
			t.Fatal("Delete should NOT parse mutation response — got ValidationError")
		}
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Test helpers ---

func testReqOpts() ReqOpts {
	return ReqOpts{
		AddEndpoint:         "/api/test/addItem",
		GetEndpoint:         "/api/test/getItem",
		UpdateEndpoint:      "/api/test/setItem",
		DeleteEndpoint:      "/api/test/delItem",
		SearchEndpoint:      "/api/test/searchItems",
		ReconfigureEndpoint: "/api/test/service/reconfigure",
		Monad:               "item",
	}
}

func newCRUDTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func newCRUDTestClient(t *testing.T, url string) *Client {
	t.Helper()
	client, err := NewClient(ClientConfig{
		BaseURL:   url,
		APIKey:    "testkey",
		APISecret: fmt.Sprintf("testsecret-%s", t.Name()), //nolint:gosec // Test credentials only
		Insecure:  true,
		RetryMax:  1,
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return client
}
