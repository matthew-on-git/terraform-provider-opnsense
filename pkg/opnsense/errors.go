// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package opnsense

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// ValidationError is returned when a mutation response has result != "saved".
// OPNsense returns HTTP 200 on validation errors — the API client must parse
// the response body to detect these.
type ValidationError struct {
	Fields map[string]string // field name → error message
}

func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		return "validation failed"
	}
	// Sort keys for deterministic output.
	keys := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s: %s", k, e.Fields[k]))
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// NotFoundError is returned when a resource is not found (deleted out-of-band).
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string { return e.Message }

// AuthError is returned for HTTP 401 or 403 responses.
type AuthError struct {
	StatusCode int
}

func (e *AuthError) Error() string {
	if e.StatusCode == http.StatusForbidden {
		return fmt.Sprintf("authorization failed (HTTP %d)", e.StatusCode)
	}
	return fmt.Sprintf("authentication failed (HTTP %d)", e.StatusCode)
}

// ServerError wraps transport/server failures after retries are exhausted.
type ServerError struct {
	Message string
	Cause   error
}

func (e *ServerError) Error() string { return e.Message }
func (e *ServerError) Unwrap() error { return e.Cause }

// PluginNotFoundError is returned for HTTP 404 on plugin API endpoints.
type PluginNotFoundError struct {
	PluginName string
}

func (e *PluginNotFoundError) Error() string {
	return fmt.Sprintf("plugin '%s' is not installed on OPNsense", e.PluginName)
}

// NewServerError creates a ServerError wrapping a transport or server failure.
func NewServerError(endpoint string, cause error) *ServerError {
	return &ServerError{
		Message: fmt.Sprintf("request to %s failed: %s", endpoint, cause),
		Cause:   cause,
	}
}

// mutationResponse is the JSON structure returned by OPNsense mutation endpoints.
type mutationResponse struct {
	Result      string            `json:"result"`
	UUID        string            `json:"uuid"`
	Validations map[string]string `json:"validations"`
}

// ParseMutationResponse parses a mutation API response body.
// Returns the UUID on success, or a ValidationError if result != "saved".
func ParseMutationResponse(body []byte) (string, error) {
	var resp mutationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to parse mutation response: %w", err)
	}
	if resp.Result != "saved" {
		return "", &ValidationError{Fields: resp.Validations}
	}
	return resp.UUID, nil
}

// CheckHTTPError returns an appropriate error type for non-success HTTP status codes.
// Returns nil for HTTP 200. Called before body parsing.
func CheckHTTPError(statusCode int, endpoint string) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return &AuthError{StatusCode: statusCode}
	case http.StatusNotFound:
		return &PluginNotFoundError{PluginName: extractPluginName(endpoint)}
	default:
		return nil
	}
}

// extractPluginName extracts the plugin/module name from an OPNsense API path.
// Example: "/api/haproxy/settings/getServer" → "haproxy"
func extractPluginName(endpoint string) string {
	trimmed := strings.TrimPrefix(endpoint, "/api/")
	if idx := strings.Index(trimmed, "/"); idx > 0 {
		return trimmed[:idx]
	}
	return trimmed
}
