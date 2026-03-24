// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// --- Schema tests ---

func TestProvider_Schema_HasExpectedAttributes(t *testing.T) {
	p := New("test")()
	resp := &frameworkprovider.SchemaResponse{}
	p.Schema(context.Background(), frameworkprovider.SchemaRequest{}, resp)

	attrs := resp.Schema.Attributes
	if attrs == nil {
		t.Fatal("expected schema attributes, got nil")
	}

	// Check uri exists and is required.
	uriAttr, ok := attrs["uri"]
	if !ok {
		t.Fatal("expected 'uri' attribute in schema")
	}
	if !uriAttr.IsRequired() {
		t.Error("expected 'uri' to be Required")
	}

	// Check api_key exists, is optional and sensitive.
	apiKeyAttr, ok := attrs["api_key"]
	if !ok {
		t.Fatal("expected 'api_key' attribute in schema")
	}
	if !apiKeyAttr.IsOptional() {
		t.Error("expected 'api_key' to be Optional")
	}
	if !apiKeyAttr.IsSensitive() {
		t.Error("expected 'api_key' to be Sensitive")
	}

	// Check api_secret exists, is optional and sensitive.
	apiSecretAttr, ok := attrs["api_secret"]
	if !ok {
		t.Fatal("expected 'api_secret' attribute in schema")
	}
	if !apiSecretAttr.IsOptional() {
		t.Error("expected 'api_secret' to be Optional")
	}
	if !apiSecretAttr.IsSensitive() {
		t.Error("expected 'api_secret' to be Sensitive")
	}

	// Check insecure exists and is optional.
	insecureAttr, ok := attrs["insecure"]
	if !ok {
		t.Fatal("expected 'insecure' attribute in schema")
	}
	if !insecureAttr.IsOptional() {
		t.Error("expected 'insecure' to be Optional")
	}
}

// --- resolveProviderConfig tests ---

func TestResolveProviderConfig_EnvVarFallback(t *testing.T) {
	t.Setenv("OPNSENSE_URI", "http://env.example.com")
	t.Setenv("OPNSENSE_API_KEY", "envkey")
	t.Setenv("OPNSENSE_API_SECRET", "envsecret")
	t.Setenv("OPNSENSE_ALLOW_INSECURE", "true")

	data := OpnsenseProviderModel{
		URI:       types.StringNull(),
		APIKey:    types.StringNull(),
		APISecret: types.StringNull(),
		Insecure:  types.BoolNull(),
	}

	cfg := resolveProviderConfig(data)

	if cfg.uri != "http://env.example.com" {
		t.Errorf("expected URI %q, got %q", "http://env.example.com", cfg.uri)
	}
	if cfg.apiKey != "envkey" {
		t.Errorf("expected API key %q, got %q", "envkey", cfg.apiKey)
	}
	if cfg.apiSecret != "envsecret" {
		t.Errorf("expected API secret %q, got %q", "envsecret", cfg.apiSecret)
	}
	if !cfg.insecure {
		t.Error("expected insecure true from env var, got false")
	}
}

func TestResolveProviderConfig_HCLPriorityOverEnvVars(t *testing.T) {
	t.Setenv("OPNSENSE_URI", "http://env.example.com")
	t.Setenv("OPNSENSE_API_KEY", "envkey")
	t.Setenv("OPNSENSE_API_SECRET", "envsecret")
	t.Setenv("OPNSENSE_ALLOW_INSECURE", "true")

	data := OpnsenseProviderModel{
		URI:       types.StringValue("http://hcl.example.com"),
		APIKey:    types.StringValue("hclkey"),
		APISecret: types.StringValue("hclsecret"),
		Insecure:  types.BoolValue(false),
	}

	cfg := resolveProviderConfig(data)

	if cfg.uri != "http://hcl.example.com" {
		t.Errorf("expected HCL URI %q, got %q", "http://hcl.example.com", cfg.uri)
	}
	if cfg.apiKey != "hclkey" {
		t.Errorf("expected HCL API key %q, got %q", "hclkey", cfg.apiKey)
	}
	if cfg.apiSecret != "hclsecret" {
		t.Errorf("expected HCL API secret %q, got %q", "hclsecret", cfg.apiSecret)
	}
	if cfg.insecure {
		t.Error("expected insecure false from HCL, got true")
	}
}

func TestResolveProviderConfig_EmptyWhenBothMissing(t *testing.T) {
	t.Setenv("OPNSENSE_URI", "")
	t.Setenv("OPNSENSE_API_KEY", "")
	t.Setenv("OPNSENSE_API_SECRET", "")
	t.Setenv("OPNSENSE_ALLOW_INSECURE", "")

	data := OpnsenseProviderModel{
		URI:       types.StringNull(),
		APIKey:    types.StringNull(),
		APISecret: types.StringNull(),
		Insecure:  types.BoolNull(),
	}

	cfg := resolveProviderConfig(data)

	if cfg.uri != "" {
		t.Errorf("expected empty URI, got %q", cfg.uri)
	}
	if cfg.apiKey != "" {
		t.Errorf("expected empty API key, got %q", cfg.apiKey)
	}
	if cfg.apiSecret != "" {
		t.Errorf("expected empty API secret, got %q", cfg.apiSecret)
	}
	if cfg.insecure {
		t.Error("expected insecure false when both null, got true")
	}
}

// --- validateRequiredConfig tests ---

func TestValidateRequiredConfig_AllPresent(t *testing.T) {
	cfg := resolvedConfig{uri: "http://example.com", apiKey: "key", apiSecret: "secret"}
	var diags diag.Diagnostics

	validateRequiredConfig(cfg, &diags)

	if diags.HasError() {
		t.Errorf("expected no errors, got: %s", formatDiags(diags))
	}
}

func TestValidateRequiredConfig_MissingURI(t *testing.T) {
	cfg := resolvedConfig{apiKey: "key", apiSecret: "secret"}
	var diags diag.Diagnostics

	validateRequiredConfig(cfg, &diags)

	if !diags.HasError() {
		t.Fatal("expected error for missing URI")
	}
	assertDiagsContain(t, diags, "Missing OPNsense URI")
	if len(diags.Errors()) != 1 {
		t.Errorf("expected exactly 1 error, got %d", len(diags.Errors()))
	}
}

func TestValidateRequiredConfig_MissingAPIKey(t *testing.T) {
	cfg := resolvedConfig{uri: "http://example.com", apiSecret: "secret"}
	var diags diag.Diagnostics

	validateRequiredConfig(cfg, &diags)

	if !diags.HasError() {
		t.Fatal("expected error for missing API key")
	}
	assertDiagsContain(t, diags, "Missing OPNsense API Key")
}

func TestValidateRequiredConfig_MissingAPISecret(t *testing.T) {
	cfg := resolvedConfig{uri: "http://example.com", apiKey: "key"}
	var diags diag.Diagnostics

	validateRequiredConfig(cfg, &diags)

	if !diags.HasError() {
		t.Fatal("expected error for missing API secret")
	}
	assertDiagsContain(t, diags, "Missing OPNsense API Secret")
}

func TestValidateRequiredConfig_AllMissing(t *testing.T) {
	cfg := resolvedConfig{}
	var diags diag.Diagnostics

	validateRequiredConfig(cfg, &diags)

	if !diags.HasError() {
		t.Fatal("expected errors for all missing fields")
	}
	if len(diags.Errors()) != 3 {
		t.Errorf("expected 3 errors (URI, API key, API secret), got %d", len(diags.Errors()))
	}
}

// --- validateCredentials tests ---

func TestValidateCredentials_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp := &frameworkprovider.ConfigureResponse{}

	validateCredentials(context.Background(), client, server.URL, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors, got: %s", diagErrors(resp))
	}
}

func TestValidateCredentials_SuccessLogsVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":          "ok",
			"product_version": "24.7.1",
		})
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp := &frameworkprovider.ConfigureResponse{}

	// validateCredentials should succeed — the version parsing is best-effort
	// and logged via tflog. We verify no errors are produced.
	validateCredentials(context.Background(), client, server.URL, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors when version is in response, got: %s", diagErrors(resp))
	}
}

func TestValidateCredentials_SuccessHandlesMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp := &frameworkprovider.ConfigureResponse{}

	// Should still succeed even if JSON parsing fails — version logging is best-effort.
	validateCredentials(context.Background(), client, server.URL, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors for malformed JSON, got: %s", diagErrors(resp))
	}
}

func TestValidateCredentials_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp := &frameworkprovider.ConfigureResponse{}

	validateCredentials(context.Background(), client, server.URL, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected authentication error")
	}
	assertConfigDiagContains(t, resp, "Authentication failed")
}

func TestValidateCredentials_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp := &frameworkprovider.ConfigureResponse{}

	validateCredentials(context.Background(), client, server.URL, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected authentication error for 403")
	}
	assertConfigDiagContains(t, resp, "Authentication failed")
}

func TestValidateCredentials_ConnectionError(t *testing.T) {
	// Use a closed server so connections are refused.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	serverURL := server.URL
	server.Close()

	client := newTestClient(t, serverURL)
	resp := &frameworkprovider.ConfigureResponse{}

	validateCredentials(context.Background(), client, serverURL, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected connection error")
	}
	assertConfigDiagSummaryContains(t, resp, "Unable to Connect")
}

func TestValidateCredentials_UnexpectedStatusCode(t *testing.T) {
	// Use 418 (I'm a teapot) — a non-retryable status code that is neither 200, 401, nor 403.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	resp := &frameworkprovider.ConfigureResponse{}

	validateCredentials(context.Background(), client, server.URL, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error for unexpected status code")
	}
	assertConfigDiagSummaryContains(t, resp, "Unexpected")
}

// --- envOrValue helper tests ---

func TestEnvOrValue_HCLTakesPriority(t *testing.T) {
	t.Setenv("TEST_VAR", "env-value")
	val := types.StringValue("hcl-value")
	result := envOrValue(val, "TEST_VAR")
	if result != "hcl-value" {
		t.Errorf("expected HCL value %q, got %q", "hcl-value", result)
	}
}

func TestEnvOrValue_FallsBackToEnv(t *testing.T) {
	t.Setenv("TEST_VAR", "env-value")
	val := types.StringNull()
	result := envOrValue(val, "TEST_VAR")
	if result != "env-value" {
		t.Errorf("expected env value %q, got %q", "env-value", result)
	}
}

func TestEnvOrValue_ReturnsEmptyWhenBothMissing(t *testing.T) {
	t.Setenv("TEST_VAR", "")
	val := types.StringNull()
	result := envOrValue(val, "TEST_VAR")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestEnvOrValue_UnknownFallsBackToEnv(t *testing.T) {
	t.Setenv("TEST_VAR", "env-value")
	val := types.StringUnknown()
	result := envOrValue(val, "TEST_VAR")
	if result != "env-value" {
		t.Errorf("expected env value %q, got %q", "env-value", result)
	}
}

// --- envOrBoolValue helper tests ---

func TestEnvOrBoolValue_HCLTakesPriority(t *testing.T) {
	t.Setenv("TEST_BOOL", "true")
	val := types.BoolValue(false)
	result := envOrBoolValue(val, "TEST_BOOL")
	if result {
		t.Error("expected false from HCL, got true")
	}
}

func TestEnvOrBoolValue_FallsBackToEnv(t *testing.T) {
	t.Setenv("TEST_BOOL", "true")
	val := types.BoolNull()
	result := envOrBoolValue(val, "TEST_BOOL")
	if !result {
		t.Error("expected true from env var, got false")
	}
}

func TestEnvOrBoolValue_EnvAccepts1(t *testing.T) {
	t.Setenv("TEST_BOOL", "1")
	val := types.BoolNull()
	result := envOrBoolValue(val, "TEST_BOOL")
	if !result {
		t.Error("expected true from env var '1', got false")
	}
}

func TestEnvOrBoolValue_DefaultsFalse(t *testing.T) {
	t.Setenv("TEST_BOOL", "")
	val := types.BoolNull()
	result := envOrBoolValue(val, "TEST_BOOL")
	if result {
		t.Error("expected false when both null, got true")
	}
}

// --- Test helpers ---

func newTestClient(t *testing.T, url string) *opnsense.Client {
	t.Helper()
	client, err := opnsense.NewClient(opnsense.ClientConfig{
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

func diagErrors(resp *frameworkprovider.ConfigureResponse) string {
	var msgs []string
	for _, d := range resp.Diagnostics.Errors() {
		msgs = append(msgs, d.Summary()+": "+d.Detail())
	}
	return strings.Join(msgs, "; ")
}

func formatDiags(diags diag.Diagnostics) string {
	var msgs []string
	for _, d := range diags.Errors() {
		msgs = append(msgs, d.Summary()+": "+d.Detail())
	}
	return strings.Join(msgs, "; ")
}

func assertDiagsContain(t *testing.T, diags diag.Diagnostics, substr string) {
	t.Helper()
	for _, d := range diags.Errors() {
		if strings.Contains(d.Summary(), substr) || strings.Contains(d.Detail(), substr) {
			return
		}
	}
	t.Errorf("expected diagnostics to contain %q, got: %s", substr, formatDiags(diags))
}

func assertConfigDiagContains(t *testing.T, resp *frameworkprovider.ConfigureResponse, substr string) {
	t.Helper()
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Detail(), substr) || strings.Contains(d.Summary(), substr) {
			return
		}
	}
	t.Errorf("expected diagnostics to contain %q, got: %s", substr, diagErrors(resp))
}

func assertConfigDiagSummaryContains(t *testing.T, resp *frameworkprovider.ConfigureResponse, substr string) {
	t.Helper()
	for _, d := range resp.Diagnostics.Errors() {
		if strings.Contains(d.Summary(), substr) {
			return
		}
	}
	t.Errorf("expected diagnostic summary to contain %q, got: %s", substr, diagErrors(resp))
}
