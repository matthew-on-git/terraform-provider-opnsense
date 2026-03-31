// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// AccountResourceModel is the Terraform state model for opnsense_acme_account.
type AccountResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Email       types.String `tfsdk:"email"`
	CA          types.String `tfsdk:"ca"`
}

type accountAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Email       string               `json:"email"`
	CA          opnsense.SelectedMap `json:"ca"`
}

type accountAPIRequest struct {
	Enabled     string `json:"enabled"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Email       string `json:"email"`
	CA          string `json:"ca"`
}

func (m *AccountResourceModel) toAPI(_ context.Context) *accountAPIRequest {
	return &accountAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Email:       m.Email.ValueString(),
		CA:          m.CA.ValueString(),
	}
}

func (m *AccountResourceModel) fromAPI(_ context.Context, a *accountAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.Email = types.StringValue(a.Email)
	m.CA = types.StringValue(string(a.CA))
}
