// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ConnectionResourceModel is the Terraform state model for opnsense_ipsec_connection.
type ConnectionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	RemoteAddrs types.String `tfsdk:"remote_addrs"`
	Version     types.String `tfsdk:"version"`
	Proposals   types.String `tfsdk:"proposals"`
	Unique      types.String `tfsdk:"unique"`
}

// ipsecConnectionAPIResponse is the struct for unmarshaling OPNsense GET responses.
type ipsecConnectionAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Description string               `json:"description"`
	RemoteAddrs string               `json:"remote_addrs"`
	Version     opnsense.SelectedMap `json:"version"`
	Proposals   string               `json:"proposals"`
	Unique      opnsense.SelectedMap `json:"unique"`
}

// ipsecConnectionAPIRequest is the struct for marshaling OPNsense POST requests.
type ipsecConnectionAPIRequest struct {
	Enabled     string `json:"enabled"`
	Description string `json:"description"`
	RemoteAddrs string `json:"remote_addrs"`
	Version     string `json:"version"`
	Proposals   string `json:"proposals"`
	Unique      string `json:"unique"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *ConnectionResourceModel) toAPI(_ context.Context) *ipsecConnectionAPIRequest {
	return &ipsecConnectionAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Description: m.Description.ValueString(),
		RemoteAddrs: m.RemoteAddrs.ValueString(),
		Version:     m.Version.ValueString(),
		Proposals:   m.Proposals.ValueString(),
		Unique:      m.Unique.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *ConnectionResourceModel) fromAPI(_ context.Context, a *ipsecConnectionAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Description = types.StringValue(a.Description)
	m.RemoteAddrs = types.StringValue(a.RemoteAddrs)
	m.Version = types.StringValue(string(a.Version))
	m.Proposals = types.StringValue(a.Proposals)
	m.Unique = types.StringValue(string(a.Unique))
}
