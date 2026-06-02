// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// CAResourceModel is the Terraform state model for opnsense_trust_ca. It manages
// an internally-generated (self-signed) Certificate Authority.
type CAResourceModel struct {
	ID          types.String `tfsdk:"id"`
	RefID       types.String `tfsdk:"refid"`
	Description types.String `tfsdk:"description"`
	CommonName  types.String `tfsdk:"common_name"`
	Country     types.String `tfsdk:"country"`
	Lifetime    types.Int64  `tfsdk:"lifetime"`
	Certificate types.String `tfsdk:"certificate"`
	ValidFrom   types.String `tfsdk:"valid_from"`
	ValidTo     types.String `tfsdk:"valid_to"`
}

type caAPIResponse struct {
	RefID       string               `json:"refid"`
	Description string               `json:"descr"`
	CommonName  string               `json:"commonname"`
	Country     opnsense.SelectedMap `json:"country"`
	Lifetime    string               `json:"lifetime"`
	Certificate string               `json:"crt_payload"`
	ValidFrom   string               `json:"valid_from"`
	ValidTo     string               `json:"valid_to"`
}

type caAPIRequest struct {
	Action      string `json:"action"`
	Description string `json:"descr"`
	CommonName  string `json:"commonname"`
	Country     string `json:"country"`
}

func (m *CAResourceModel) toAPI(_ context.Context) *caAPIRequest {
	return &caAPIRequest{
		Action:      "internal",
		Description: m.Description.ValueString(),
		CommonName:  m.CommonName.ValueString(),
		Country:     m.Country.ValueString(),
	}
}

func (m *CAResourceModel) fromAPI(_ context.Context, a *caAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.RefID = types.StringValue(a.RefID)
	m.Description = types.StringValue(a.Description)
	m.CommonName = types.StringValue(a.CommonName)
	m.Country = types.StringValue(string(a.Country))
	m.Certificate = types.StringValue(a.Certificate)
	m.ValidFrom = types.StringValue(a.ValidFrom)
	m.ValidTo = types.StringValue(a.ValidTo)
	if a.Lifetime != "" {
		if v, err := opnsense.StringToInt64(a.Lifetime); err == nil {
			m.Lifetime = types.Int64Value(v)
		}
	}
}
