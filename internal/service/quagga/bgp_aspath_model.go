// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// BGPASPathResourceModel is the Terraform state model for opnsense_quagga_bgp_aspath.
type BGPASPathResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	Number      types.Int64  `tfsdk:"number"`
	Action      types.String `tfsdk:"action"`
	AS          types.String `tfsdk:"as_pattern"`
}

type bgpASPathAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Description string               `json:"description"`
	Number      string               `json:"number"`
	Action      opnsense.SelectedMap `json:"action"`
	AS          string               `json:"as"`
}

type bgpASPathAPIRequest struct {
	Enabled     string `json:"enabled"`
	Description string `json:"description"`
	Number      string `json:"number"`
	Action      string `json:"action"`
	AS          string `json:"as"`
}

func (m *BGPASPathResourceModel) toAPI(_ context.Context) *bgpASPathAPIRequest {
	return &bgpASPathAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Description: m.Description.ValueString(),
		Number:      opnsense.Int64ToString(m.Number.ValueInt64()),
		Action:      m.Action.ValueString(),
		AS:          m.AS.ValueString(),
	}
}

func (m *BGPASPathResourceModel) fromAPI(_ context.Context, a *bgpASPathAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Description = types.StringValue(a.Description)
	m.Number = types.Int64Value(intOrZero(a.Number))
	m.Action = types.StringValue(string(a.Action))
	m.AS = types.StringValue(a.AS)
}
