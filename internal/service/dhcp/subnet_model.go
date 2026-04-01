// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dhcp

import (
	"context"

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
	Subnet      string `json:"subnet"`
	Description string `json:"description"`
	Pools       string `json:"pools"`
	OptionData  string `json:"option_data"`
}

type subnetAPIRequest struct {
	Subnet      string `json:"subnet"`
	Description string `json:"description"`
	Pools       string `json:"pools"`
	OptionData  string `json:"option_data"`
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
	m.OptionData = types.StringValue(a.OptionData)
}
