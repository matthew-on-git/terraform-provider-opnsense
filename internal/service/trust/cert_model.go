// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// CertResourceModel is the Terraform state model for opnsense_trust_cert. It
// imports an existing certificate (and its private key) into the OPNsense trust
// store. The certificate and key PEM payloads are write-only: the API re-encodes
// them, so they are kept from configuration rather than refreshed from state.
type CertResourceModel struct {
	ID          types.String `tfsdk:"id"`
	RefID       types.String `tfsdk:"refid"`
	Description types.String `tfsdk:"description"`
	CARef       types.String `tfsdk:"ca_ref"`
	Certificate types.String `tfsdk:"certificate"`
	PrivateKey  types.String `tfsdk:"private_key"`
}

type certAPIResponse struct {
	RefID       string               `json:"refid"`
	Description string               `json:"descr"`
	CARef       opnsense.SelectedMap `json:"caref"`
}

type certAPIRequest struct {
	Action      string `json:"action"`
	Description string `json:"descr"`
	CARef       string `json:"caref"`
	Certificate string `json:"crt_payload"`
	PrivateKey  string `json:"prv_payload"`
}

func (m *CertResourceModel) toAPI(_ context.Context) *certAPIRequest {
	return &certAPIRequest{
		Action:      "import",
		Description: m.Description.ValueString(),
		CARef:       m.CARef.ValueString(),
		Certificate: m.Certificate.ValueString(),
		PrivateKey:  m.PrivateKey.ValueString(),
	}
}

func (m *CertResourceModel) fromAPI(_ context.Context, a *certAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.RefID = types.StringValue(a.RefID)
	m.Description = types.StringValue(a.Description)
	m.CARef = types.StringValue(string(a.CARef))
	// Certificate and PrivateKey are write-only: the API re-encodes the PEM, so
	// the configured values are preserved rather than refreshed here.
}
