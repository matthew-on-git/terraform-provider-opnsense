// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme

import (
	"context"

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

func (m *CertificateResourceModel) fromAPI(_ context.Context, a *certificateAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.AltNames = types.StringValue(a.AltNames)
	m.Account = types.StringValue(string(a.Account))
	m.ValidationMethod = types.StringValue(string(a.ValidationMethod))
	m.KeyLength = types.StringValue(string(a.KeyLength))
	m.AutoRenewal = types.BoolValue(opnsense.StringToBool(a.AutoRenewal))
}
