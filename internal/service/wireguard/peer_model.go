// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// PeerResourceModel is the Terraform state model for opnsense_wireguard_peer.
type PeerResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Name          types.String `tfsdk:"name"`
	PublicKey     types.String `tfsdk:"public_key"`
	TunnelAddress types.String `tfsdk:"tunnel_address"`
	ServerAddress types.String `tfsdk:"server_address"`
	ServerPort    types.String `tfsdk:"server_port"`
	Keepalive     types.Int64  `tfsdk:"keepalive"`
}

// wireguardPeerAPIResponse is the struct for unmarshaling OPNsense GET responses.
type wireguardPeerAPIResponse struct {
	Enabled       string `json:"enabled"`
	Name          string `json:"name"`
	PublicKey     string `json:"pubkey"`
	TunnelAddress string `json:"tunneladdress"`
	ServerAddress string `json:"serveraddress"`
	ServerPort    string `json:"serverport"`
	Keepalive     string `json:"keepalive"`
}

// wireguardPeerAPIRequest is the struct for marshaling OPNsense POST requests.
type wireguardPeerAPIRequest struct {
	Enabled       string `json:"enabled"`
	Name          string `json:"name"`
	PublicKey     string `json:"pubkey"`
	TunnelAddress string `json:"tunneladdress"`
	ServerAddress string `json:"serveraddress"`
	ServerPort    string `json:"serverport"`
	Keepalive     string `json:"keepalive"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *PeerResourceModel) toAPI(_ context.Context) *wireguardPeerAPIRequest {
	return &wireguardPeerAPIRequest{
		Enabled:       opnsense.BoolToString(m.Enabled.ValueBool()),
		Name:          m.Name.ValueString(),
		PublicKey:     m.PublicKey.ValueString(),
		TunnelAddress: m.TunnelAddress.ValueString(),
		ServerAddress: m.ServerAddress.ValueString(),
		ServerPort:    m.ServerPort.ValueString(),
		Keepalive:     opnsense.Int64ToString(m.Keepalive.ValueInt64()),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *PeerResourceModel) fromAPI(_ context.Context, a *wireguardPeerAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.PublicKey = types.StringValue(a.PublicKey)
	m.TunnelAddress = types.StringValue(a.TunnelAddress)
	m.ServerAddress = types.StringValue(a.ServerAddress)
	m.ServerPort = types.StringValue(a.ServerPort)

	if a.Keepalive != "" {
		if v, err := opnsense.StringToInt64(a.Keepalive); err == nil {
			m.Keepalive = types.Int64Value(v)
		}
	}
}
