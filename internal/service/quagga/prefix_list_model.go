// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// PrefixListResourceModel is the Terraform state model for opnsense_quagga_prefix_list.
type PrefixListResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
	Version     types.String `tfsdk:"version"`
	SeqNumber   types.Int64  `tfsdk:"sequence"`
	Action      types.String `tfsdk:"action"`
	Network     types.String `tfsdk:"network"`
}

type prefixListAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Description string               `json:"description"`
	Name        string               `json:"name"`
	Version     opnsense.SelectedMap `json:"version"`
	SeqNumber   string               `json:"seqnumber"`
	Action      opnsense.SelectedMap `json:"action"`
	Network     string               `json:"network"`
}

type prefixListAPIRequest struct {
	Enabled     string `json:"enabled"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	SeqNumber   string `json:"seqnumber"`
	Action      string `json:"action"`
	Network     string `json:"network"`
}

func (m *PrefixListResourceModel) toAPI(_ context.Context) *prefixListAPIRequest {
	return &prefixListAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Description: m.Description.ValueString(),
		Name:        m.Name.ValueString(),
		Version:     m.Version.ValueString(),
		SeqNumber:   opnsense.Int64ToString(m.SeqNumber.ValueInt64()),
		Action:      m.Action.ValueString(),
		Network:     m.Network.ValueString(),
	}
}

func (m *PrefixListResourceModel) fromAPI(_ context.Context, a *prefixListAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Description = types.StringValue(a.Description)
	m.Name = types.StringValue(a.Name)
	m.Version = types.StringValue(string(a.Version))
	m.Action = types.StringValue(string(a.Action))
	m.Network = types.StringValue(a.Network)

	if a.SeqNumber != "" {
		if v, err := opnsense.StringToInt64(a.SeqNumber); err == nil {
			m.SeqNumber = types.Int64Value(v)
		}
	}
}
