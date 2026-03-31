// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// AccountResourceModel is the Terraform state model for opnsense_ddclient_account.
type AccountResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Service     types.String `tfsdk:"service"`
	Hostnames   types.String `tfsdk:"hostnames"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	Description types.String `tfsdk:"description"`
}

// ddclientAccountAPIResponse is the struct for unmarshaling OPNsense GET responses.
type ddclientAccountAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Service     opnsense.SelectedMap `json:"service"`
	Hostnames   string               `json:"hostnames"`
	Username    string               `json:"username"`
	Password    string               `json:"password"`
	Description string               `json:"description"`
}

// ddclientAccountAPIRequest is the struct for marshaling OPNsense POST requests.
type ddclientAccountAPIRequest struct {
	Enabled     string `json:"enabled"`
	Service     string `json:"service"`
	Hostnames   string `json:"hostnames"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Description string `json:"description"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *AccountResourceModel) toAPI(_ context.Context) *ddclientAccountAPIRequest {
	return &ddclientAccountAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Service:     m.Service.ValueString(),
		Hostnames:   m.Hostnames.ValueString(),
		Username:    m.Username.ValueString(),
		Password:    m.Password.ValueString(),
		Description: m.Description.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
// Password is write-only and not populated from API responses.
func (m *AccountResourceModel) fromAPI(_ context.Context, a *ddclientAccountAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Service = types.StringValue(string(a.Service))
	m.Hostnames = types.StringValue(a.Hostnames)
	m.Username = types.StringValue(a.Username)
	m.Description = types.StringValue(a.Description)
}
