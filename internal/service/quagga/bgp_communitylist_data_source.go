// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &bgpCommunityListDataSource{}

type bgpCommunityListDataSource struct{ client *opnsense.Client }

func newBGPCommunityListDataSource() datasource.DataSource { return &bgpCommunityListDataSource{} }

func (d *bgpCommunityListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_communitylist"
}

func (d *bgpCommunityListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads an existing BGP community list rule on OPNsense by UUID.", Attributes: map[string]dsschema.Attribute{
		"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
		"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this community list is enabled."},
		"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the community list."},
		"number":      dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Community list number."},
		"seq_number":  dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Sequence number."},
		"action":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Community list action."},
		"community":   dsschema.StringAttribute{Computed: true, MarkdownDescription: "Community match value."},
	}}
}

func (d *bgpCommunityListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(c *opnsense.Client) { d.client = c })
}

func (d *bgpCommunityListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config BGPCommunityListResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[bgpCommunityListAPIResponse](ctx, d.client, bgpCommunityListReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP community list", fmt.Sprintf("Could not read BGP community list %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
