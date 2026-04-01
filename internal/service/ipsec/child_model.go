// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ChildResourceModel is the Terraform state model for opnsense_ipsec_child.
type ChildResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	Connection   types.String `tfsdk:"connection_id"`
	Description  types.String `tfsdk:"description"`
	Mode         types.String `tfsdk:"mode"`
	LocalTS      types.String `tfsdk:"local_ts"`
	RemoteTS     types.String `tfsdk:"remote_ts"`
	EspProposals types.String `tfsdk:"esp_proposals"`
	StartAction  types.String `tfsdk:"start_action"`
}

// ipsecChildAPIResponse is the struct for unmarshaling OPNsense GET responses.
type ipsecChildAPIResponse struct {
	Enabled      string               `json:"enabled"`
	Connection   string               `json:"connection"`
	Description  string               `json:"description"`
	Mode         opnsense.SelectedMap `json:"mode"`
	LocalTS      string               `json:"local_ts"`
	RemoteTS     string               `json:"remote_ts"`
	EspProposals string               `json:"esp_proposals"`
	StartAction  opnsense.SelectedMap `json:"start_action"`
}

// ipsecChildAPIRequest is the struct for marshaling OPNsense POST requests.
type ipsecChildAPIRequest struct {
	Enabled      string `json:"enabled"`
	Connection   string `json:"connection"`
	Description  string `json:"description"`
	Mode         string `json:"mode"`
	LocalTS      string `json:"local_ts"`
	RemoteTS     string `json:"remote_ts"`
	EspProposals string `json:"esp_proposals"`
	StartAction  string `json:"start_action"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *ChildResourceModel) toAPI(_ context.Context) *ipsecChildAPIRequest {
	return &ipsecChildAPIRequest{
		Enabled:      opnsense.BoolToString(m.Enabled.ValueBool()),
		Connection:   m.Connection.ValueString(),
		Description:  m.Description.ValueString(),
		Mode:         m.Mode.ValueString(),
		LocalTS:      m.LocalTS.ValueString(),
		RemoteTS:     m.RemoteTS.ValueString(),
		EspProposals: m.EspProposals.ValueString(),
		StartAction:  m.StartAction.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *ChildResourceModel) fromAPI(_ context.Context, a *ipsecChildAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Connection = types.StringValue(a.Connection)
	m.Description = types.StringValue(a.Description)
	m.Mode = types.StringValue(string(a.Mode))
	m.LocalTS = types.StringValue(a.LocalTS)
	m.RemoteTS = types.StringValue(a.RemoteTS)
	m.EspProposals = types.StringValue(a.EspProposals)
	m.StartAction = types.StringValue(string(a.StartAction))
}
