// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// BGPRedistributionResourceModel is the Terraform state model for opnsense_quagga_bgp_redistribution.
type BGPRedistributionResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Description  types.String `tfsdk:"description"`
	Redistribute types.String `tfsdk:"redistribute"`
	RouteMap     types.String `tfsdk:"route_map"`
}

type bgpRedistributionAPIResponse struct {
	Enabled      string               `json:"enabled"`
	Description  string               `json:"description"`
	Redistribute opnsense.SelectedMap `json:"redistribute"`
	RouteMap     opnsense.SelectedMap `json:"linkedRoutemap"`
}

type bgpRedistributionAPIRequest struct {
	Enabled      string `json:"enabled"`
	Description  string `json:"description"`
	Redistribute string `json:"redistribute"`
	RouteMap     string `json:"linkedRoutemap"`
}

func (m *BGPRedistributionResourceModel) toAPI(_ context.Context) *bgpRedistributionAPIRequest {
	return &bgpRedistributionAPIRequest{
		Enabled:      opnsense.BoolToString(m.Enabled.ValueBool()),
		Description:  m.Description.ValueString(),
		Redistribute: m.Redistribute.ValueString(),
		RouteMap:     m.RouteMap.ValueString(),
	}
}

func (m *BGPRedistributionResourceModel) fromAPI(_ context.Context, a *bgpRedistributionAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Description = types.StringValue(a.Description)
	m.Redistribute = types.StringValue(string(a.Redistribute))
	m.RouteMap = types.StringValue(string(a.RouteMap))
}
