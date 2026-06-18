// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// CertificateResourceModel is the Terraform state model for opnsense_acme_certificate.
type CertificateResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	AltNames         types.String `tfsdk:"alt_names"`
	Account          types.String `tfsdk:"account"`
	ValidationMethod types.String `tfsdk:"validation_method"`
	KeyLength        types.String `tfsdk:"key_length"`
	AutoRenewal      types.Bool   `tfsdk:"auto_renewal"`
	IssuanceTimeout  types.String `tfsdk:"issuance_timeout"`
	IssuanceInterval types.String `tfsdk:"issuance_poll_interval"`
	CertRefID        types.String `tfsdk:"cert_ref_id"`
	StatusCode       types.String `tfsdk:"status_code"`
	Status           types.String `tfsdk:"status"`
}

// CertificateDataSourceModel is the Terraform state model for the
// opnsense_acme_certificate data source.
type CertificateDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	AltNames         types.String `tfsdk:"alt_names"`
	Account          types.String `tfsdk:"account"`
	ValidationMethod types.String `tfsdk:"validation_method"`
	KeyLength        types.String `tfsdk:"key_length"`
	AutoRenewal      types.Bool   `tfsdk:"auto_renewal"`
	CertRefID        types.String `tfsdk:"cert_ref_id"`
	StatusCode       types.String `tfsdk:"status_code"`
	Status           types.String `tfsdk:"status"`
}

type certificateAPIResponse struct {
	Enabled          string               `json:"enabled"`
	Name             string               `json:"name"`
	Description      string               `json:"description"`
	AltNames         string               `json:"altNames"`
	Account          opnsense.SelectedMap `json:"account"`
	ValidationMethod opnsense.SelectedMap `json:"validationMethod"`
	KeyLength        opnsense.SelectedMap `json:"keyLength"`
	AutoRenewal      string               `json:"autoRenewal"`
	CertRefID        string               `json:"certRefId"`
	StatusCode       certificateAPIString `json:"statusCode"`
	Status           string               `json:"status"`
}

type certificateAPIRequest struct {
	Enabled          string `json:"enabled"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	AltNames         string `json:"altNames"`
	Account          string `json:"account"`
	ValidationMethod string `json:"validationMethod"`
	KeyLength        string `json:"keyLength"`
	AutoRenewal      string `json:"autoRenewal"`
}

type certificateSearchRow struct {
	UUID       string               `json:"uuid"`
	CertRefID  string               `json:"certRefId"`
	StatusCode certificateAPIString `json:"statusCode"`
	Status     string               `json:"status"`
}

type certificateAPIString string

func (s *certificateAPIString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*s = certificateAPIString(text)
		return nil
	}

	var number float64
	if err := json.Unmarshal(data, &number); err == nil {
		*s = certificateAPIString(strconv.FormatFloat(number, 'f', -1, 64))
		return nil
	}

	return fmt.Errorf("expected string, number, or null")
}

func (s certificateAPIString) String() string { return string(s) }

func (m *CertificateResourceModel) toAPI(_ context.Context) *certificateAPIRequest {
	return &certificateAPIRequest{
		Enabled:          opnsense.BoolToString(m.Enabled.ValueBool()),
		Name:             m.Name.ValueString(),
		Description:      m.Description.ValueString(),
		AltNames:         m.AltNames.ValueString(),
		Account:          m.Account.ValueString(),
		ValidationMethod: m.ValidationMethod.ValueString(),
		KeyLength:        m.KeyLength.ValueString(),
		AutoRenewal:      opnsense.BoolToString(m.AutoRenewal.ValueBool()),
	}
}

func (m *CertificateResourceModel) fromAPI(ctx context.Context, a *certificateAPIResponse, uuid string) {
	timeout := m.IssuanceTimeout
	interval := m.IssuanceInterval
	m.fromAPIResponse(ctx, a, uuid)
	m.IssuanceTimeout = stringValueOrDefault(timeout, "180s")
	m.IssuanceInterval = stringValueOrDefault(interval, "10s")
}

func (m *CertificateResourceModel) fromAPIResponse(_ context.Context, a *certificateAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.AltNames = types.StringValue(a.AltNames)
	m.Account = types.StringValue(string(a.Account))
	m.ValidationMethod = types.StringValue(string(a.ValidationMethod))
	m.KeyLength = types.StringValue(string(a.KeyLength))
	m.AutoRenewal = types.BoolValue(opnsense.StringToBool(a.AutoRenewal))
	m.CertRefID = types.StringValue(a.CertRefID)
	m.StatusCode = types.StringValue(a.StatusCode.String())
	m.Status = types.StringValue(a.Status)
}

func (m *CertificateDataSourceModel) fromAPI(_ context.Context, a *certificateAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.AltNames = types.StringValue(a.AltNames)
	m.Account = types.StringValue(string(a.Account))
	m.ValidationMethod = types.StringValue(string(a.ValidationMethod))
	m.KeyLength = types.StringValue(string(a.KeyLength))
	m.AutoRenewal = types.BoolValue(opnsense.StringToBool(a.AutoRenewal))
	m.CertRefID = types.StringValue(a.CertRefID)
	m.StatusCode = types.StringValue(a.StatusCode.String())
	m.Status = types.StringValue(a.Status)
}

func stringValueOrDefault(value types.String, fallback string) types.String {
	if value.IsNull() || value.IsUnknown() || value.ValueString() == "" {
		return types.StringValue(fallback)
	}
	return value
}

func (m *CertificateResourceModel) requiresIssuance(previous CertificateResourceModel) bool {
	return !m.Name.Equal(previous.Name) ||
		!m.AltNames.Equal(previous.AltNames) ||
		!m.Account.Equal(previous.Account) ||
		!m.ValidationMethod.Equal(previous.ValidationMethod) ||
		!m.KeyLength.Equal(previous.KeyLength)
}

func (m *CertificateResourceModel) requiresRemoteUpdate(previous CertificateResourceModel) bool {
	return m.requiresIssuance(previous) ||
		!m.Enabled.Equal(previous.Enabled) ||
		!m.Description.Equal(previous.Description) ||
		!m.AutoRenewal.Equal(previous.AutoRenewal)
}

type certificateSignResponse struct {
	Result      string            `json:"result"`
	Status      string            `json:"status"`
	Validations map[string]string `json:"validations"`
}

func parseCertificateSignResponse(body []byte) error {
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil
	}

	var response certificateSignResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse sign response: %w", err)
	}

	switch {
	case strings.EqualFold(response.Result, "saved") || strings.EqualFold(response.Result, "ok") || strings.EqualFold(response.Status, "ok"):
		return nil
	case response.Result != "":
		_, err := opnsense.ParseMutationResponse(body)
		return err
	default:
		return nil
	}
}

func (m *CertificateResourceModel) issuanceWaitConfig() (time.Duration, time.Duration, error) {
	timeout, err := time.ParseDuration(m.IssuanceTimeout.ValueString())
	if err != nil || timeout <= 0 {
		return 0, 0, fmt.Errorf("issuance_timeout must be a positive Go duration such as 180s")
	}

	interval, err := time.ParseDuration(m.IssuanceInterval.ValueString())
	if err != nil || interval <= 0 {
		return 0, 0, fmt.Errorf("issuance_poll_interval must be a positive Go duration such as 10s")
	}

	return timeout, interval, nil
}
