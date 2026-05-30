// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// BGPCommunityListResourceModel is the Terraform state model for opnsense_quagga_bgp_communitylist.
type BGPCommunityListResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	Number      types.Int64  `tfsdk:"number"`
	SeqNumber   types.Int64  `tfsdk:"seq_number"`
	Action      types.String `tfsdk:"action"`
	Community   types.String `tfsdk:"community"`
}

type bgpCommunityListAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Description string               `json:"description"`
	Number      string               `json:"number"`
	SeqNumber   string               `json:"seqnumber"`
	Action      opnsense.SelectedMap `json:"action"`
	Community   string               `json:"community"`
}

type bgpCommunityListAPIRequest struct {
	Enabled     string `json:"enabled"`
	Description string `json:"description"`
	Number      string `json:"number"`
	SeqNumber   string `json:"seqnumber"`
	Action      string `json:"action"`
	Community   string `json:"community"`
}

func (m *BGPCommunityListResourceModel) toAPI(_ context.Context) *bgpCommunityListAPIRequest {
	return &bgpCommunityListAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Description: m.Description.ValueString(),
		Number:      opnsense.Int64ToString(m.Number.ValueInt64()),
		SeqNumber:   opnsense.Int64ToString(m.SeqNumber.ValueInt64()),
		Action:      m.Action.ValueString(),
		Community:   m.Community.ValueString(),
	}
}

func (m *BGPCommunityListResourceModel) fromAPI(_ context.Context, a *bgpCommunityListAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Description = types.StringValue(a.Description)
	m.Number = types.Int64Value(intOrZero(a.Number))
	m.SeqNumber = types.Int64Value(intOrZero(a.SeqNumber))
	m.Action = types.StringValue(string(a.Action))
	m.Community = types.StringValue(a.Community)
}
