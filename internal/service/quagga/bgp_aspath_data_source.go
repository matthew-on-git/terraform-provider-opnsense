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

var _ datasource.DataSource = &bgpASPathDataSource{}

type bgpASPathDataSource struct{ client *opnsense.Client }

func newBGPASPathDataSource() datasource.DataSource { return &bgpASPathDataSource{} }

func (d *bgpASPathDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_aspath"
}

func (d *bgpASPathDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads an existing BGP AS path rule on OPNsense by UUID.", Attributes: map[string]dsschema.Attribute{
		"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
		"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this AS path rule is enabled."},
		"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the AS path rule."},
		"number":      dsschema.Int64Attribute{Computed: true, MarkdownDescription: "AS path rule number."},
		"action":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "AS path rule action."},
		"as_pattern":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "AS path match pattern."},
	}}
}

func (d *bgpASPathDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(c *opnsense.Client) { d.client = c })
}

func (d *bgpASPathDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config BGPASPathResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[bgpASPathAPIResponse](ctx, d.client, bgpASPathReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP AS path", fmt.Sprintf("Could not read BGP AS path %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
