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
	ID              types.String `tfsdk:"id"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Connection      types.String `tfsdk:"connection_id"`
	Description     types.String `tfsdk:"description"`
	Mode            types.String `tfsdk:"mode"`
	LocalTS         types.String `tfsdk:"local_ts"`
	RemoteTS        types.String `tfsdk:"remote_ts"`
	EspProposals    types.String `tfsdk:"esp_proposals"`
	StartAction     types.String `tfsdk:"start_action"`
	CloseAction     types.String `tfsdk:"close_action"`
	DPDAction       types.String `tfsdk:"dpd_action"`
	InstallPolicies types.Bool   `tfsdk:"install_policies"`
	RequestID       types.String `tfsdk:"request_id"`
	SHA25696        types.Bool   `tfsdk:"sha256_96"`
	RekeyTime       types.String `tfsdk:"rekey_time"`
}

// ipsecChildAPIResponse is the struct for unmarshaling OPNsense GET responses.
type ipsecChildAPIResponse struct {
	Enabled         string         `json:"enabled"`
	Connection      ipsecEnum      `json:"connection"`
	Description     string         `json:"description"`
	Mode            ipsecEnum      `json:"mode"`
	LocalTS         ipsecSelection `json:"local_ts"`
	RemoteTS        ipsecSelection `json:"remote_ts"`
	EspProposals    ipsecSelection `json:"esp_proposals"`
	StartAction     ipsecEnum      `json:"start_action"`
	CloseAction     ipsecEnum      `json:"close_action"`
	DPDAction       ipsecEnum      `json:"dpd_action"`
	InstallPolicies string         `json:"policies"`
	RequestID       string         `json:"reqid"`
	SHA25696        string         `json:"sha256_96"`
	RekeyTime       string         `json:"rekey_time"`
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
	m.Connection = types.StringValue(string(a.Connection))
	m.Description = types.StringValue(a.Description)
	m.Mode = types.StringValue(string(a.Mode))
	m.LocalTS = stableIPsecCSV(m.LocalTS, a.LocalTS)
	m.RemoteTS = stableIPsecCSV(m.RemoteTS, a.RemoteTS)
	m.EspProposals = stableIPsecCSV(m.EspProposals, a.EspProposals)
	m.StartAction = types.StringValue(string(a.StartAction))
	m.CloseAction = types.StringValue(string(a.CloseAction))
	m.DPDAction = types.StringValue(string(a.DPDAction))
	m.InstallPolicies = types.BoolValue(opnsense.StringToBool(a.InstallPolicies))
	m.RequestID = types.StringValue(string(a.RequestID))
	m.SHA25696 = types.BoolValue(opnsense.StringToBool(a.SHA25696))
	m.RekeyTime = types.StringValue(string(a.RekeyTime))
}

func (a *ipsecChildAPIResponse) UnmarshalJSON(data []byte) error {
	type plain ipsecChildAPIResponse
	var value plain
	if err := decodeIPsecResponse(data, &value, childAdvancedFields); err != nil {
		return err
	}
	*a = ipsecChildAPIResponse(value)
	return nil
}

// A saved object must retain a usable identity even when activation/readback fails.
func (m *ChildResourceModel) clearUnknownAdvanced() {
	if m.CloseAction.IsUnknown() {
		m.CloseAction = types.StringNull()
	}
	if m.DPDAction.IsUnknown() {
		m.DPDAction = types.StringNull()
	}
	if m.InstallPolicies.IsUnknown() {
		m.InstallPolicies = types.BoolNull()
	}
	if m.RequestID.IsUnknown() {
		m.RequestID = types.StringNull()
	}
	if m.SHA25696.IsUnknown() {
		m.SHA25696 = types.BoolNull()
	}
	if m.RekeyTime.IsUnknown() {
		m.RekeyTime = types.StringNull()
	}
}
