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

var _ datasource.DataSource = &bgpRedistributionDataSource{}

type bgpRedistributionDataSource struct{ client *opnsense.Client }

func newBGPRedistributionDataSource() datasource.DataSource { return &bgpRedistributionDataSource{} }

func (d *bgpRedistributionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_redistribution"
}

func (d *bgpRedistributionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads an existing BGP redistribution rule on OPNsense by UUID.", Attributes: map[string]dsschema.Attribute{
		"id":           dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
		"enabled":      dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this redistribution rule is enabled."},
		"description":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the redistribution rule."},
		"redistribute": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Redistribution source."},
		"route_map":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Linked route map."},
	}}
}

func (d *bgpRedistributionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(c *opnsense.Client) { d.client = c })
}

func (d *bgpRedistributionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config BGPRedistributionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[bgpRedistributionAPIResponse](ctx, d.client, bgpRedistributionReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP redistribution", fmt.Sprintf("Could not read BGP redistribution %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
