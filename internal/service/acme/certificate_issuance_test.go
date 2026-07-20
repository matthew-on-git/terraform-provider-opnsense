// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func TestCertificateSignAndWaitPollsUntilIssued(t *testing.T) {
	t.Parallel()

	searchCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/acmeclient/certificates/sign/cert-1":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST sign request, got %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/acmeclient/certificates/search":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST search request, got %s", r.Method)
			}
			searchCalls++
			if searchCalls == 1 {
				_, _ = w.Write([]byte(`{"rows":[{"uuid":"cert-1","statusCode":"100","status":"pending","certRefId":""}],"rowCount":1,"total":1,"current":1}`))
				return
			}
			_, _ = w.Write([]byte(`{"rows":[{"uuid":"cert-1","statusCode":"200","status":"issued","certRefId":"ref-123"}],"rowCount":1,"total":1,"current":1}`))
		case "/api/acmeclient/certificates/get/cert-1":
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET certificate request, got %s", r.Method)
			}
			_, _ = w.Write([]byte(`{"certificate":{"enabled":"1","name":"www.example.com","description":"","altNames":"","account":"acct-1","validationMethod":"challenge-1","keyLength":"key_4096","autoRenewal":"1","statusCode":"200","status":"issued","certRefId":"ref-123"}}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	result, err := signAndWaitForCertificateIssuance(context.Background(), client, "cert-1", time.Second, time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected sign and wait error: %s", err)
	}
	if result.CertRefID != "ref-123" {
		t.Fatalf("expected cert refid ref-123, got %q", result.CertRefID)
	}
	if searchCalls != 2 {
		t.Fatalf("expected 2 search polls, got %d", searchCalls)
	}
}

func TestCertificateSignAndWaitTimeout(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/acmeclient/certificates/sign/cert-1":
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/acmeclient/certificates/search":
			_, _ = w.Write([]byte(`{"rows":[{"uuid":"cert-1","statusCode":"100","status":"pending","certRefId":""}],"rowCount":1,"total":1,"current":1}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := signAndWaitForCertificateIssuance(context.Background(), client, "cert-1", 5*time.Millisecond, time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "cert-1") || !strings.Contains(err.Error(), "statusCode 200") || !strings.Contains(err.Error(), "certRefId") {
		t.Fatalf("expected actionable timeout error, got %q", err.Error())
	}
}

func TestCertificateSignAndWaitCancellation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/acmeclient/certificates/sign/cert-1":
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/acmeclient/certificates/search":
			_, _ = w.Write([]byte(`{"rows":[{"uuid":"cert-1","statusCode":"100","status":"pending","certRefId":""}],"rowCount":1,"total":1,"current":1}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := newTestClient(t, server.URL)
	_, err := signAndWaitForCertificateIssuance(ctx, client, "cert-1", time.Second, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected context cancellation error, got %v", err)
	}
}

func TestSignCertificateParsesFailureBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/acmeclient/certificates/sign/cert-1" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":"failed","validations":{"name":"invalid"}}`))
	}))
	defer server.Close()

	err := signCertificate(context.Background(), newTestClient(t, server.URL), "cert-1")
	if err == nil || !strings.Contains(err.Error(), "name: invalid") {
		t.Fatalf("expected validation error from sign body, got %v", err)
	}
}

func TestSignCertificateEscapesUUIDPathSegment(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/api/acmeclient/certificates/sign/cert%2F1" {
			t.Fatalf("expected escaped uuid path, got path=%s escaped=%s", r.URL.Path, r.URL.EscapedPath())
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	if err := signCertificate(context.Background(), newTestClient(t, server.URL), "cert/1"); err != nil {
		t.Fatalf("unexpected sign error: %s", err)
	}
}

func TestCertificateIssuedAcceptsNumericStatusCode(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/acmeclient/certificates/search" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"rows":[{"uuid":"cert-1","statusCode":200,"status":"issued","certRefId":"ref-123"}],"rowCount":1,"total":1,"current":1}`))
	}))
	defer server.Close()

	issued, row, err := certificateIssued(context.Background(), newTestClient(t, server.URL), "cert-1")
	if err != nil {
		t.Fatalf("unexpected issued check error: %s", err)
	}
	if !issued || row.StatusCode.String() != "200" {
		t.Fatalf("expected issued numeric status, got issued=%t status=%q", issued, row.StatusCode.String())
	}
}

func TestCertificateFromAPIPreservesWaitDefaults(t *testing.T) {
	t.Parallel()

	var model CertificateResourceModel
	model.fromAPI(context.Background(), &certificateAPIResponse{Name: "www.example.com"}, "cert-1")
	if model.IssuanceTimeout.ValueString() != "180s" || model.IssuanceInterval.ValueString() != "10s" {
		t.Fatalf("expected default wait settings, got timeout=%q interval=%q", model.IssuanceTimeout.ValueString(), model.IssuanceInterval.ValueString())
	}

	model.IssuanceTimeout = types.StringValue("300s")
	model.IssuanceInterval = types.StringValue("15s")
	model.fromAPI(context.Background(), &certificateAPIResponse{Name: "www.example.com"}, "cert-1")
	if model.IssuanceTimeout.ValueString() != "300s" || model.IssuanceInterval.ValueString() != "15s" {
		t.Fatalf("expected preserved wait settings, got timeout=%q interval=%q", model.IssuanceTimeout.ValueString(), model.IssuanceInterval.ValueString())
	}
}

func TestCertificateAltNamesUnmarshalString(t *testing.T) {
	t.Parallel()

	var altNames certificateAPIAltNames
	if err := altNames.UnmarshalJSON([]byte(`"www.example.com"`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if altNames.String() != "www.example.com" {
		t.Fatalf("expected www.example.com, got %q", altNames.String())
	}
}

func TestCertificateAltNamesUnmarshalObject(t *testing.T) {
	t.Parallel()

	var altNames certificateAPIAltNames
	if err := altNames.UnmarshalJSON([]byte(`{"key":{"value":"api.example.com","selected":1}}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := altNames.String()
	if result == "" {
		t.Fatal("expected non-empty alt names from object")
	}
}

func TestCertificateAltNamesUnmarshalNull(t *testing.T) {
	t.Parallel()

	var altNames certificateAPIAltNames
	if err := altNames.UnmarshalJSON([]byte(`null`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if altNames.String() != "" {
		t.Fatalf("expected empty string for null, got %q", altNames.String())
	}
}

func TestCertificateResponseUnmarshalObjectAltNames(t *testing.T) {
	t.Parallel()

	jsonData := `{"enabled":"1","name":"svc.example.com","description":"","altNames":{"someKey":"someValue"},"account":{"uuid":{"value":"acct-1","selected":1}},"validationMethod":{"uuid":{"value":"challenge-1","selected":1}},"keyLength":{"uuid":{"value":"key_4096","selected":1}},"autoRenewal":"1","statusCode":"200","status":"issued","certRefId":"ref-123"}`

	var resp certificateAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal response with object altNames: %v", err)
	}
	if resp.AltNames.String() == "" {
		t.Fatal("expected non-empty alt names from object")
	}
	if resp.Name != "svc.example.com" {
		t.Fatalf("expected name svc.example.com, got %q", resp.Name)
	}
}

func TestCertificateUpdateClassification(t *testing.T) {
	t.Parallel()

	state := CertificateResourceModel{
		Enabled:          types.BoolValue(true),
		Name:             types.StringValue("www.example.com"),
		Description:      types.StringValue(""),
		AltNames:         types.StringValue(""),
		Account:          types.StringValue("acct-1"),
		ValidationMethod: types.StringValue("challenge-1"),
		KeyLength:        types.StringValue("key_4096"),
		AutoRenewal:      types.BoolValue(true),
		IssuanceTimeout:  types.StringValue("180s"),
		IssuanceInterval: types.StringValue("10s"),
	}

	plan := state
	plan.IssuanceTimeout = types.StringValue("300s")
	if plan.requiresRemoteUpdate(state) || plan.requiresIssuance(state) {
		t.Fatal("provider-only wait setting changes must not update or reissue remotely")
	}

	plan = state
	plan.Description = types.StringValue("updated")
	if !plan.requiresRemoteUpdate(state) || plan.requiresIssuance(state) {
		t.Fatal("description changes should update remotely without reissuing")
	}

	plan = state
	plan.AltNames = types.StringValue("api.example.com")
	if !plan.requiresRemoteUpdate(state) || !plan.requiresIssuance(state) {
		t.Fatal("alt name changes should update remotely and reissue")
	}
}

func newTestClient(t *testing.T, baseURL string) *opnsense.Client {
	t.Helper()
	client, err := opnsense.NewClient(opnsense.ClientConfig{
		BaseURL:   baseURL,
		APIKey:    "key",
		APISecret: "secret",
	})
	if err != nil {
		t.Fatalf("failed to create test client: %s", err)
	}
	return client
}
