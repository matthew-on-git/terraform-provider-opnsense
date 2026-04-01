// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ServerResourceModel is the Terraform state model for opnsense_wireguard_server.
type ServerResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Name          types.String `tfsdk:"name"`
	Port          types.String `tfsdk:"port"`
	PrivateKey    types.String `tfsdk:"private_key"`
	TunnelAddress types.String `tfsdk:"tunnel_address"`
	Description   types.String `tfsdk:"description"`
}

// wireguardServerAPIResponse is the struct for unmarshaling OPNsense GET responses.
type wireguardServerAPIResponse struct {
	Enabled       string `json:"enabled"`
	Name          string `json:"name"`
	Port          string `json:"port"`
	TunnelAddress string `json:"tunneladdress"`
	Description   string `json:"description"`
}

// wireguardServerAPIRequest is the struct for marshaling OPNsense POST requests.
type wireguardServerAPIRequest struct {
	Enabled       string `json:"enabled"`
	Name          string `json:"name"`
	Port          string `json:"port"`
	PrivateKey    string `json:"privkey"`
	TunnelAddress string `json:"tunneladdress"`
	Description   string `json:"description"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *ServerResourceModel) toAPI(_ context.Context) *wireguardServerAPIRequest {
	return &wireguardServerAPIRequest{
		Enabled:       opnsense.BoolToString(m.Enabled.ValueBool()),
		Name:          m.Name.ValueString(),
		Port:          m.Port.ValueString(),
		PrivateKey:    m.PrivateKey.ValueString(),
		TunnelAddress: m.TunnelAddress.ValueString(),
		Description:   m.Description.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
// PrivateKey is write-only and not populated from API responses.
func (m *ServerResourceModel) fromAPI(_ context.Context, a *wireguardServerAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.Port = types.StringValue(a.Port)
	m.TunnelAddress = types.StringValue(a.TunnelAddress)
	m.Description = types.StringValue(a.Description)
}
