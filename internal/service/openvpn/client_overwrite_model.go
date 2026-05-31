// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ClientOverwriteResourceModel is the Terraform state model for opnsense_openvpn_client_overwrite.
type ClientOverwriteResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	CommonName     types.String `tfsdk:"common_name"`
	Description    types.String `tfsdk:"description"`
	Servers        types.Set    `tfsdk:"servers"`
	Block          types.Bool   `tfsdk:"block"`
	PushReset      types.Bool   `tfsdk:"push_reset"`
	TunnelNetwork  types.String `tfsdk:"tunnel_network"`
	LocalNetworks  types.Set    `tfsdk:"local_networks"`
	RemoteNetworks types.Set    `tfsdk:"remote_networks"`
	DNSServers     types.Set    `tfsdk:"dns_servers"`
}

type clientOverwriteAPIResponse struct {
	Enabled        string                   `json:"enabled"`
	CommonName     string                   `json:"common_name"`
	Description    string                   `json:"description"`
	Servers        opnsense.SelectedMapList `json:"servers"`
	Block          string                   `json:"block"`
	PushReset      string                   `json:"push_reset"`
	TunnelNetwork  string                   `json:"tunnel_network"`
	LocalNetworks  opnsense.SelectedMapList `json:"local_networks"`
	RemoteNetworks opnsense.SelectedMapList `json:"remote_networks"`
	DNSServers     opnsense.SelectedMapList `json:"dns_servers"`
}

type clientOverwriteAPIRequest struct {
	Enabled        string `json:"enabled"`
	CommonName     string `json:"common_name"`
	Description    string `json:"description"`
	Servers        string `json:"servers"`
	Block          string `json:"block"`
	PushReset      string `json:"push_reset"`
	TunnelNetwork  string `json:"tunnel_network"`
	LocalNetworks  string `json:"local_networks"`
	RemoteNetworks string `json:"remote_networks"`
	DNSServers     string `json:"dns_servers"`
}

func (m *ClientOverwriteResourceModel) toAPI(ctx context.Context) *clientOverwriteAPIRequest {
	return &clientOverwriteAPIRequest{
		Enabled:        opnsense.BoolToString(m.Enabled.ValueBool()),
		CommonName:     m.CommonName.ValueString(),
		Description:    m.Description.ValueString(),
		Servers:        setToCSV(ctx, m.Servers),
		Block:          opnsense.BoolToString(m.Block.ValueBool()),
		PushReset:      opnsense.BoolToString(m.PushReset.ValueBool()),
		TunnelNetwork:  m.TunnelNetwork.ValueString(),
		LocalNetworks:  setToCSV(ctx, m.LocalNetworks),
		RemoteNetworks: setToCSV(ctx, m.RemoteNetworks),
		DNSServers:     setToCSV(ctx, m.DNSServers),
	}
}

func (m *ClientOverwriteResourceModel) fromAPI(_ context.Context, a *clientOverwriteAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.CommonName = types.StringValue(a.CommonName)
	m.Description = types.StringValue(a.Description)
	m.Block = types.BoolValue(opnsense.StringToBool(a.Block))
	m.PushReset = types.BoolValue(opnsense.StringToBool(a.PushReset))
	m.TunnelNetwork = types.StringValue(a.TunnelNetwork)

	m.Servers = sliceToSet(a.Servers)
	m.LocalNetworks = sliceToSet(a.LocalNetworks)
	m.RemoteNetworks = sliceToSet(a.RemoteNetworks)
	m.DNSServers = sliceToSet(a.DNSServers)
}
