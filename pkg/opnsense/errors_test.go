// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// --- ValidationError tests ---

func TestValidationError_ErrorMessage(t *testing.T) {
	err := &ValidationError{Fields: map[string]string{
		"port":    "value must be between 1 and 65535",
		"address": "invalid IPv4 address",
	}}
	msg := err.Error()
	if !strings.Contains(msg, "validation failed") {
		t.Errorf("expected 'validation failed' prefix, got: %s", msg)
	}
	if !strings.Contains(msg, "port: value must be between 1 and 65535") {
		t.Errorf("expected port field in message, got: %s", msg)
	}
	if !strings.Contains(msg, "address: invalid IPv4 address") {
		t.Errorf("expected address field in message, got: %s", msg)
	}
}

func TestValidationError_EmptyFields(t *testing.T) {
	err := &ValidationError{Fields: nil}
	if err.Error() != "validation failed" {
		t.Errorf("expected 'validation failed', got: %s", err.Error())
	}
}

func TestValidationError_ErrorsAs(t *testing.T) {
	var err error = &ValidationError{Fields: map[string]string{"port": "invalid"}}
	var target *ValidationError
	if !errors.As(err, &target) {
		t.Fatal("errors.As failed for ValidationError")
	}
	if target.Fields["port"] != "invalid" {
		t.Errorf("expected field 'port' = 'invalid', got: %v", target.Fields)
	}
}

// --- NotFoundError tests ---

func TestNotFoundError_ErrorMessage(t *testing.T) {
	err := &NotFoundError{Message: "resource not found"}
	if err.Error() != "resource not found" {
		t.Errorf("expected 'resource not found', got: %s", err.Error())
	}
}

func TestNotFoundError_ErrorsAs(t *testing.T) {
	var err error = &NotFoundError{Message: "gone"}
	var target *NotFoundError
	if !errors.As(err, &target) {
		t.Fatal("errors.As failed for NotFoundError")
	}
	if target.Message != "gone" {
		t.Errorf("expected message 'gone', got: %s", target.Message)
	}
}

// --- AuthError tests ---

func TestAuthError_Unauthorized(t *testing.T) {
	err := &AuthError{StatusCode: http.StatusUnauthorized}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("expected 'authentication failed' for 401, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected status code in message, got: %s", err.Error())
	}
}

func TestAuthError_Forbidden(t *testing.T) {
	err := &AuthError{StatusCode: http.StatusForbidden}
	if !strings.Contains(err.Error(), "authorization failed") {
		t.Errorf("expected 'authorization failed' for 403, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected status code in message, got: %s", err.Error())
	}
}

func TestAuthError_ErrorsAs(t *testing.T) {
	var err error = &AuthError{StatusCode: 401}
	var target *AuthError
	if !errors.As(err, &target) {
		t.Fatal("errors.As failed for AuthError")
	}
	if target.StatusCode != 401 {
		t.Errorf("expected StatusCode 401, got: %d", target.StatusCode)
	}
}

// --- ServerError tests ---

func TestServerError_ErrorMessage(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := NewServerError("/api/test", cause)
	if !strings.Contains(err.Error(), "/api/test") {
		t.Errorf("expected endpoint in message, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected cause in message, got: %s", err.Error())
	}
}

func TestServerError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("timeout")
	err := NewServerError("/api/test", cause)
	if err.Unwrap() != cause {
		t.Errorf("expected Unwrap to return cause, got: %v", err.Unwrap())
	}
}

func TestServerError_ErrorsAs(t *testing.T) {
	cause := fmt.Errorf("timeout")
	var err error = NewServerError("/api/test", cause)
	var target *ServerError
	if !errors.As(err, &target) {
		t.Fatal("errors.As failed for ServerError")
	}
	if target.Cause != cause {
		t.Errorf("expected original cause, got: %v", target.Cause)
	}
}

// --- PluginNotFoundError tests ---

func TestPluginNotFoundError_ErrorMessage(t *testing.T) {
	err := &PluginNotFoundError{PluginName: "haproxy"}
	if !strings.Contains(err.Error(), "haproxy") {
		t.Errorf("expected plugin name in message, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "not installed") {
		t.Errorf("expected 'not installed' in message, got: %s", err.Error())
	}
}

func TestPluginNotFoundError_ErrorsAs(t *testing.T) {
	var err error = &PluginNotFoundError{PluginName: "frr"}
	var target *PluginNotFoundError
	if !errors.As(err, &target) {
		t.Fatal("errors.As failed for PluginNotFoundError")
	}
	if target.PluginName != "frr" {
		t.Errorf("expected plugin name 'frr', got: %s", target.PluginName)
	}
}

// --- ParseMutationResponse tests ---

func TestParseMutationResponse_Success(t *testing.T) {
	body := []byte(`{"result":"saved","uuid":"550e8400-e29b-41d4-a716-446655440000"}`)
	uuid, err := ParseMutationResponse(body)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if uuid != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected UUID, got: %s", uuid)
	}
}

func TestParseMutationResponse_ValidationError(t *testing.T) {
	body := []byte(`{"result":"failed","validations":{"port":"value must be between 1 and 65535"}}`)
	uuid, err := ParseMutationResponse(body)
	if err == nil {
		t.Fatal("expected ValidationError, got nil")
	}
	if uuid != "" {
		t.Errorf("expected empty UUID on error, got: %s", uuid)
	}
	var validErr *ValidationError
	if !errors.As(err, &validErr) {
		t.Fatalf("expected ValidationError, got: %T", err)
	}
	if validErr.Fields["port"] != "value must be between 1 and 65535" {
		t.Errorf("expected port validation message, got: %v", validErr.Fields)
	}
}

func TestParseMutationResponse_MultipleValidationFields(t *testing.T) {
	body := []byte(`{"result":"failed","validations":{"port":"invalid","address":"required"}}`)
	_, err := ParseMutationResponse(body)
	var validErr *ValidationError
	if !errors.As(err, &validErr) {
		t.Fatalf("expected ValidationError, got: %T", err)
	}
	if len(validErr.Fields) != 2 {
		t.Errorf("expected 2 validation fields, got: %d", len(validErr.Fields))
	}
}

func TestParseMutationResponse_InvalidJSON(t *testing.T) {
	body := []byte(`not json`)
	_, err := ParseMutationResponse(body)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	// Should NOT be a ValidationError — it's a parse error.
	var validErr *ValidationError
	if errors.As(err, &validErr) {
		t.Error("expected non-ValidationError for invalid JSON")
	}
}

// --- CheckHTTPError tests ---

func TestCheckHTTPError_200ReturnsNil(t *testing.T) {
	err := CheckHTTPError(http.StatusOK, "/api/test")
	if err != nil {
		t.Errorf("expected nil for 200, got: %v", err)
	}
}

func TestCheckHTTPError_401ReturnsAuthError(t *testing.T) {
	err := CheckHTTPError(http.StatusUnauthorized, "/api/haproxy/settings/getServer")
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError for 401, got: %T", err)
	}
	if authErr.StatusCode != 401 {
		t.Errorf("expected StatusCode 401, got: %d", authErr.StatusCode)
	}
}

func TestCheckHTTPError_403ReturnsAuthError(t *testing.T) {
	err := CheckHTTPError(http.StatusForbidden, "/api/haproxy/settings/getServer")
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected AuthError for 403, got: %T", err)
	}
	if authErr.StatusCode != 403 {
		t.Errorf("expected StatusCode 403, got: %d", authErr.StatusCode)
	}
}

func TestCheckHTTPError_404ReturnsPluginNotFound(t *testing.T) {
	err := CheckHTTPError(http.StatusNotFound, "/api/haproxy/settings/getServer")
	var pluginErr *PluginNotFoundError
	if !errors.As(err, &pluginErr) {
		t.Fatalf("expected PluginNotFoundError for 404, got: %T", err)
	}
	if pluginErr.PluginName != "haproxy" {
		t.Errorf("expected plugin name 'haproxy', got: %s", pluginErr.PluginName)
	}
}

func TestCheckHTTPError_404ExtractsPluginName(t *testing.T) {
	tests := []struct {
		endpoint   string
		wantPlugin string
	}{
		{"/api/haproxy/settings/getServer", "haproxy"},
		{"/api/frr/bgp/getNeighbor", "frr"},
		{"/api/acme/accounts/getAccount", "acme"},
		{"/api/core/firmware/status", "core"},
	}
	for _, tt := range tests {
		err := CheckHTTPError(http.StatusNotFound, tt.endpoint)
		var pluginErr *PluginNotFoundError
		if !errors.As(err, &pluginErr) {
			t.Errorf("endpoint %s: expected PluginNotFoundError", tt.endpoint)
			continue
		}
		if pluginErr.PluginName != tt.wantPlugin {
			t.Errorf("endpoint %s: expected plugin %q, got %q", tt.endpoint, tt.wantPlugin, pluginErr.PluginName)
		}
	}
}

// --- extractPluginName tests ---

func TestExtractPluginName(t *testing.T) {
	tests := []struct {
		endpoint string
		want     string
	}{
		{"/api/haproxy/settings/getServer", "haproxy"},
		{"/api/frr/bgp/getNeighbor", "frr"},
		{"/api/core/firmware/status", "core"},
		{"haproxy/settings/getServer", "haproxy"},
	}
	for _, tt := range tests {
		got := extractPluginName(tt.endpoint)
		if got != tt.want {
			t.Errorf("extractPluginName(%q) = %q, want %q", tt.endpoint, got, tt.want)
		}
	}
}
