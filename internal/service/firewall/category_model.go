// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// CategoryResourceModel is the Terraform state model for opnsense_firewall_category.
type CategoryResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Auto  types.Bool   `tfsdk:"auto"`
	Color types.String `tfsdk:"color"`
}

// categoryAPIResponse is the struct for unmarshaling OPNsense GET responses.
type categoryAPIResponse struct {
	Name  string               `json:"name"`
	Auto  opnsense.SelectedMap `json:"auto"`
	Color string               `json:"color"`
}

// categoryAPIRequest is the struct for marshaling OPNsense POST requests.
type categoryAPIRequest struct {
	Name  string `json:"name"`
	Auto  string `json:"auto"`
	Color string `json:"color"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *CategoryResourceModel) toAPI(_ context.Context) *categoryAPIRequest {
	return &categoryAPIRequest{
		Name:  m.Name.ValueString(),
		Auto:  opnsense.BoolToString(m.Auto.ValueBool()),
		Color: m.Color.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *CategoryResourceModel) fromAPI(_ context.Context, a *categoryAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Name = types.StringValue(a.Name)
	m.Auto = types.BoolValue(opnsense.StringToBool(string(a.Auto)))
	m.Color = types.StringValue(a.Color)
}
