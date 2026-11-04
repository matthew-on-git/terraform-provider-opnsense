// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
	if !uriAttr.IsOptional() {
		t.Error("expected 'uri' to be Optional (env var fallback)")
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

	// Check retry settings exist, are optional, and carry positive validators.
	for _, name := range []string{"retries", "min_backoff", "max_backoff"} {
		retryAttr, ok := attrs[name].(schema.Int64Attribute)
		if !ok {
			t.Fatalf("expected %q to be an Int64 attribute", name)
		}
		if !retryAttr.IsOptional() {
			t.Errorf("expected %q to be Optional", name)
		}
		if len(retryAttr.Validators) == 0 {
			t.Errorf("expected %q to have validators", name)
		}
	}
}

func TestProviderSchema_RetryValidatorsRejectInvalidValues(t *testing.T) {
	p := New("test")()
	var schemaResp frameworkprovider.SchemaResponse
	p.Schema(context.Background(), frameworkprovider.SchemaRequest{}, &schemaResp)

	tests := []struct {
		name  string
		value int64
	}{
		{"retries", 0},
		{"retries", -1},
		{"retries", int64(maxRetrySetting) + 1},
		{"min_backoff", 0},
		{"min_backoff", -5},
		{"min_backoff", int64(maxRetrySetting) + 1},
		{"max_backoff", 0},
		{"max_backoff", -5},
		{"max_backoff", int64(maxRetrySetting) + 1},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%s_%d", tc.name, tc.value), func(t *testing.T) {
			attr := schemaResp.Schema.Attributes[tc.name].(schema.Int64Attribute)
			req := validator.Int64Request{Path: path.Root(tc.name), ConfigValue: types.Int64Value(tc.value)}
			var resp validator.Int64Response
			for _, v := range attr.Validators {
				v.ValidateInt64(context.Background(), req, &resp)
			}
			if !resp.Diagnostics.HasError() {
				t.Fatalf("expected %q=%d to fail validation", tc.name, tc.value)
			}
		})
	}
}

func TestProviderSchema_RetryValidatorsAcceptBounds(t *testing.T) {
	p := New("test")()
	var schemaResp frameworkprovider.SchemaResponse
	p.Schema(context.Background(), frameworkprovider.SchemaRequest{}, &schemaResp)

	for _, name := range []string{"retries", "min_backoff", "max_backoff"} {
		attr := schemaResp.Schema.Attributes[name].(schema.Int64Attribute)
		for _, value := range []int64{1, int64(maxRetrySetting)} {
			req := validator.Int64Request{Path: path.Root(name), ConfigValue: types.Int64Value(value)}
			var resp validator.Int64Response
			for _, v := range attr.Validators {
				v.ValidateInt64(context.Background(), req, &resp)
			}
			if resp.Diagnostics.HasError() {
				t.Errorf("expected %q=%d to pass validation: %s", name, value, formatDiags(resp.Diagnostics))
			}
		}
	}
}

func TestProvider_EphemeralRegistrationAndConfigurationHandoff(t *testing.T) {
	t.Parallel()
	p := New("test")().(*OpnsenseProvider)
	if got := len(p.EphemeralResources(context.Background())); got != 1 {
		t.Fatalf("ephemeral registrations = %d, want 1", got)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/core/menu/search" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()
	var schemaResp frameworkprovider.SchemaResponse
	p.Schema(context.Background(), frameworkprovider.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}
	model := OpnsenseProviderModel{URI: types.StringValue(server.URL), APIKey: types.StringValue("test-key"), APISecret: types.StringValue("test-secret"), Insecure: types.BoolValue(false)}
	if diags := state.Set(context.Background(), &model); diags.HasError() {
		t.Fatal(diags)
	}
	var configureResp frameworkprovider.ConfigureResponse
	p.Configure(context.Background(), frameworkprovider.ConfigureRequest{Config: tfsdk.Config{Raw: state.Raw, Schema: schemaResp.Schema}}, &configureResp)
	if configureResp.Diagnostics.HasError() {
		t.Fatal(configureResp.Diagnostics)
	}
	if configureResp.ResourceData == nil || configureResp.DataSourceData == nil || configureResp.EphemeralResourceData == nil {
		t.Fatal("provider did not hand configuration to every owner kind")
	}
	if configureResp.ResourceData != configureResp.EphemeralResourceData || configureResp.DataSourceData != configureResp.EphemeralResourceData {
		t.Fatal("resource, data-source, and ephemeral owners received different clients")
	}
}

func TestProvider_ConfigureUsesConfiguredRetries(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p := New("test")().(*OpnsenseProvider)
	var schemaResp frameworkprovider.SchemaResponse
	p.Schema(context.Background(), frameworkprovider.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}
	model := OpnsenseProviderModel{
		URI:        types.StringValue(server.URL),
		APIKey:     types.StringValue("test-key"),
		APISecret:  types.StringValue("test-secret"),
		Insecure:   types.BoolValue(true),
		Retries:    types.Int64Value(1),
		MinBackoff: types.Int64Value(1),
		MaxBackoff: types.Int64Value(1),
	}
	if diags := state.Set(context.Background(), &model); diags.HasError() {
		t.Fatal(diags)
	}

	var configureResp frameworkprovider.ConfigureResponse
	p.Configure(context.Background(), frameworkprovider.ConfigureRequest{
		Config: tfsdk.Config{Raw: state.Raw, Schema: schemaResp.Schema},
	}, &configureResp)

	if !configureResp.Diagnostics.HasError() {
		t.Fatal("expected credential-validation error after exhausting configured retries")
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("expected 2 attempts with retries=1, got %d", got)
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

func TestResolveProviderConfig_RetryHCLOnly(t *testing.T) {
	t.Setenv("OPNSENSE_RETRIES", "")
	t.Setenv("OPNSENSE_MIN_BACKOFF", "")
	t.Setenv("OPNSENSE_MAX_BACKOFF", "")

	data := OpnsenseProviderModel{
		Retries:    types.Int64Value(5),
		MinBackoff: types.Int64Value(2),
		MaxBackoff: types.Int64Value(10),
	}

	cfg := resolveProviderConfig(data)

	if cfg.retries != 5 {
		t.Errorf("expected retries 5, got %d", cfg.retries)
	}
	if cfg.minBackoff != 2*time.Second {
		t.Errorf("expected min backoff 2s, got %s", cfg.minBackoff)
	}
	if cfg.maxBackoff != 10*time.Second {
		t.Errorf("expected max backoff 10s, got %s", cfg.maxBackoff)
	}
}

func TestResolveProviderConfig_RetryEnvFallback(t *testing.T) {
	t.Setenv("OPNSENSE_RETRIES", "5")
	t.Setenv("OPNSENSE_MIN_BACKOFF", "2")
	t.Setenv("OPNSENSE_MAX_BACKOFF", "10")

	data := OpnsenseProviderModel{
		Retries:    types.Int64Null(),
		MinBackoff: types.Int64Null(),
		MaxBackoff: types.Int64Null(),
	}

	cfg := resolveProviderConfig(data)

	if cfg.retries != 5 {
		t.Errorf("expected retries 5 from env, got %d", cfg.retries)
	}
	if cfg.minBackoff != 2*time.Second {
		t.Errorf("expected min backoff 2s from env, got %s", cfg.minBackoff)
	}
	if cfg.maxBackoff != 10*time.Second {
		t.Errorf("expected max backoff 10s from env, got %s", cfg.maxBackoff)
	}
}

func TestResolveProviderConfig_RetryHCLPriorityOverEnv(t *testing.T) {
	t.Setenv("OPNSENSE_RETRIES", "5")
	t.Setenv("OPNSENSE_MIN_BACKOFF", "2")
	t.Setenv("OPNSENSE_MAX_BACKOFF", "10")

	data := OpnsenseProviderModel{
		Retries:    types.Int64Value(6),
		MinBackoff: types.Int64Value(3),
		MaxBackoff: types.Int64Value(12),
	}

	cfg := resolveProviderConfig(data)

	if cfg.retries != 6 {
		t.Errorf("expected HCL retries 6, got %d", cfg.retries)
	}
	if cfg.minBackoff != 3*time.Second {
		t.Errorf("expected HCL min backoff 3s, got %s", cfg.minBackoff)
	}
	if cfg.maxBackoff != 12*time.Second {
		t.Errorf("expected HCL max backoff 12s, got %s", cfg.maxBackoff)
	}
}

func TestResolveProviderConfig_RetryDefaultsWhenOmitted(t *testing.T) {
	t.Setenv("OPNSENSE_RETRIES", "")
	t.Setenv("OPNSENSE_MIN_BACKOFF", "")
	t.Setenv("OPNSENSE_MAX_BACKOFF", "")

	cfg := resolveProviderConfig(OpnsenseProviderModel{
		Retries:    types.Int64Null(),
		MinBackoff: types.Int64Null(),
		MaxBackoff: types.Int64Null(),
	})

	if cfg.retries != 0 || cfg.minBackoff != 0 || cfg.maxBackoff != 0 {
		t.Fatalf("expected zero retry settings so client defaults apply, got %d / %s / %s", cfg.retries, cfg.minBackoff, cfg.maxBackoff)
	}
}

func TestResolveProviderConfig_RetryUnknownUsesEnv(t *testing.T) {
	t.Setenv("OPNSENSE_RETRIES", "4")

	cfg := resolveProviderConfig(OpnsenseProviderModel{Retries: types.Int64Unknown()})

	if cfg.retries != 4 {
		t.Errorf("expected env retries 4 for unknown HCL value, got %d", cfg.retries)
	}
}

func TestResolveProviderConfig_InvalidRetryEnvIsScopedAndIgnored(t *testing.T) {
	tests := []struct {
		name                       string
		retriesEnv, minEnv, maxEnv string
		wantRetries                int
		wantMin, wantMax           time.Duration
	}{
		{"all invalid", "", "not-a-number", "0", 0, 0, 0},
		{"invalid retries", "bad", "3", "4", 0, 3 * time.Second, 4 * time.Second},
		{"invalid min", "5", "0", "4", 5, 0, 4 * time.Second},
		{"invalid max", "5", "3", "-1", 5, 3 * time.Second, 0},
		{"out of range retries", "2147483648", "3", "4", 0, 3 * time.Second, 4 * time.Second},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("OPNSENSE_RETRIES", tc.retriesEnv)
			t.Setenv("OPNSENSE_MIN_BACKOFF", tc.minEnv)
			t.Setenv("OPNSENSE_MAX_BACKOFF", tc.maxEnv)

			cfg := resolveProviderConfig(OpnsenseProviderModel{
				Retries:    types.Int64Null(),
				MinBackoff: types.Int64Null(),
				MaxBackoff: types.Int64Null(),
			})

			if cfg.retries != tc.wantRetries || cfg.minBackoff != tc.wantMin || cfg.maxBackoff != tc.wantMax {
				t.Fatalf("got %d / %s / %s, want %d / %s / %s", cfg.retries, cfg.minBackoff, cfg.maxBackoff, tc.wantRetries, tc.wantMin, tc.wantMax)
			}
		})
	}
}

func TestResolvedConfigToClientConfig(t *testing.T) {
	cfg := resolvedConfig{
		uri:        "https://opnsense.example.com",
		apiKey:     "key",
		apiSecret:  "secret",
		insecure:   true,
		retries:    5,
		minBackoff: 2 * time.Second,
		maxBackoff: 10 * time.Second,
	}

	got := cfg.toClientConfig()

	if got.BaseURL != cfg.uri || got.APIKey != cfg.apiKey || got.APISecret != cfg.apiSecret || got.Insecure != cfg.insecure {
		t.Fatalf("credential/connection handoff mismatch: %+v", got)
	}
	if got.RetryMax != 5 {
		t.Errorf("expected RetryMax 5, got %d", got.RetryMax)
	}
	if got.RetryWaitMin != 2*time.Second {
		t.Errorf("expected RetryWaitMin 2s, got %s", got.RetryWaitMin)
	}
	if got.RetryWaitMax != 10*time.Second {
		t.Errorf("expected RetryWaitMax 10s, got %s", got.RetryWaitMax)
	}
	if got.MaxReadConcurrency != 0 {
		t.Errorf("expected MaxReadConcurrency to stay unset, got %d", got.MaxReadConcurrency)
	}
}

func TestResolvedConfigToClientConfig_MinAboveMaxPreserved(t *testing.T) {
	cfg := resolvedConfig{retries: 3, minBackoff: 30 * time.Second, maxBackoff: time.Second}

	got := cfg.toClientConfig()

	if got.RetryWaitMin != 30*time.Second || got.RetryWaitMax != time.Second {
		t.Fatalf("expected min/max to pass through unchanged, got %s / %s", got.RetryWaitMin, got.RetryWaitMax)
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/core/menu/search" {
			t.Errorf("unexpected validation path %q", r.URL.Path)
		}
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

	// validateCredentials should succeed even when the low-privilege endpoint
	// happens to include unrelated metadata.
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

// --- envOrPositiveInt64Value helper tests ---

func TestEnvOrPositiveInt64Value(t *testing.T) {
	tests := []struct {
		name string
		env  string
		val  types.Int64
		want int64
	}{
		{"HCL wins", "9", types.Int64Value(4), 4},
		{"env fallback", "4", types.Int64Null(), 4},
		{"whitespace env", " 5 ", types.Int64Null(), 5},
		{"unknown falls back", "4", types.Int64Unknown(), 4},
		{"blank env", "", types.Int64Null(), 0},
		{"invalid env", "not-a-number", types.Int64Null(), 0},
		{"zero env", "0", types.Int64Null(), 0},
		{"negative env", "-1", types.Int64Null(), 0},
		{"out of range env", "2147483648", types.Int64Null(), 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TEST_INT64", tc.env)
			if got := envOrPositiveInt64Value(tc.val, "TEST_INT64"); got != tc.want {
				t.Errorf("expected %d, got %d", tc.want, got)
			}
		})
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
