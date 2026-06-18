// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dhcp

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SubnetResourceModel is the Terraform state model for opnsense_dhcpv4_subnet.
type SubnetResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Subnet      types.String `tfsdk:"subnet"`
	Description types.String `tfsdk:"description"`
	Pools       types.String `tfsdk:"pools"`
	OptionData  types.String `tfsdk:"option_data"`
}

type subnetAPIResponse struct {
	Subnet      string          `json:"subnet"`
	Description string          `json:"description"`
	Pools       string          `json:"pools"`
	OptionData  json.RawMessage `json:"option_data"`
}

type subnetAPIRequest struct {
	Subnet      string `json:"subnet"`
	Description string `json:"description,omitempty"`
	Pools       string `json:"pools,omitempty"`
	OptionData  string `json:"option_data,omitempty"`
}

func (m *SubnetResourceModel) toAPI(_ context.Context) *subnetAPIRequest {
	return &subnetAPIRequest{
		Subnet:      m.Subnet.ValueString(),
		Description: m.Description.ValueString(),
		Pools:       m.Pools.ValueString(),
		OptionData:  m.OptionData.ValueString(),
	}
}

func (m *SubnetResourceModel) fromAPI(_ context.Context, a *subnetAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Subnet = types.StringValue(a.Subnet)
	m.Description = types.StringValue(a.Description)
	m.Pools = types.StringValue(a.Pools)
	if len(a.OptionData) > 0 && string(a.OptionData) != "{}" && string(a.OptionData) != "[]" && string(a.OptionData) != "null" && m.OptionData.ValueString() != "" {
		m.OptionData = types.StringValue(string(a.OptionData))
	}
	if m.OptionData.IsNull() || m.OptionData.IsUnknown() {
		m.OptionData = types.StringValue("")
	}
}
