// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TunableResourceModel is the Terraform state model for opnsense_system_tunable.
type TunableResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Tunable     types.String `tfsdk:"tunable"`
	Value       types.String `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
}

type tunableAPIResponse struct {
	Tunable      string `json:"tunable"`
	Value        string `json:"value"`
	Description  string `json:"descr"`
	DefaultValue string `json:"default_value"`
	Type         string `json:"type"`
}

type tunableAPIRequest struct {
	Tunable     string `json:"tunable"`
	Value       string `json:"value"`
	Description string `json:"descr"`
}

func (m *TunableResourceModel) toAPI(_ context.Context) *tunableAPIRequest {
	return &tunableAPIRequest{
		Tunable:     m.Tunable.ValueString(),
		Value:       m.Value.ValueString(),
		Description: m.Description.ValueString(),
	}
}

func (m *TunableResourceModel) fromAPI(_ context.Context, a *tunableAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Tunable = types.StringValue(a.Tunable)
	m.Value = types.StringValue(a.Value)
	m.Description = types.StringValue(a.Description)
}
