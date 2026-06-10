// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &peerDataSource{}

type peerDataSource struct{ client *opnsense.Client }

func newPeerDataSource() datasource.DataSource { return &peerDataSource{} }

func (d *peerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireguard_peer"
}

func (d *peerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing WireGuard peer on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":        dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this peer is enabled. Defaults to 'true'."},
			"name":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the WireGuard peer."},
			"public_key":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Public key of the WireGuard peer."},
			"tunnel_address": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Allowed IPs / tunnel address for this peer (e.g., '10.0.0.2/32')."},
			"server_address": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Endpoint address of the remote WireGuard server."},
			"server_port":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Endpoint port of the remote WireGuard server."},
			"keepalive":      dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Persistent keepalive interval in seconds. '0' disables."},
		},
	}
}

func (d *peerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	d.client = client
}

func (d *peerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PeerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[wireguardPeerAPIResponse](ctx, d.client, peerReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading WireGuard peer", fmt.Sprintf("Could not read WireGuard peer %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
