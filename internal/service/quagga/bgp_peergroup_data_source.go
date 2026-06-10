// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &bgpPeerGroupDataSource{}

type bgpPeerGroupDataSource struct{ client *opnsense.Client }

func newBGPPeerGroupDataSource() datasource.DataSource { return &bgpPeerGroupDataSource{} }

func (d *bgpPeerGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_peergroup"
}

func (d *bgpPeerGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads an existing BGP peer group on OPNsense by UUID.", Attributes: map[string]dsschema.Attribute{
		"id":                dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
		"enabled":           dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this peer group is enabled."},
		"name":              dsschema.StringAttribute{Computed: true, MarkdownDescription: "Peer group name."},
		"remote_as_mode":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Remote AS mode."},
		"remote_as":         dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Remote autonomous system number."},
		"family":            dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Address families."},
		"update_source":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Update source interface."},
		"next_hop_self":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether next-hop-self is enabled."},
		"default_originate": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether default-originate is enabled."},
	}}
}

func (d *bgpPeerGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(c *opnsense.Client) { d.client = c })
}

func (d *bgpPeerGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config BGPPeerGroupResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[bgpPeerGroupAPIResponse](ctx, d.client, bgpPeerGroupReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP peer group", fmt.Sprintf("Could not read BGP peer group %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
