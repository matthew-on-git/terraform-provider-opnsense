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

var _ datasource.DataSource = &staticRouteDataSource{}

type staticRouteDataSource struct{ client *opnsense.Client }

func newStaticRouteDataSource() datasource.DataSource { return &staticRouteDataSource{} }

func (d *staticRouteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_static_route"
}

func (d *staticRouteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads an existing FRR static route on OPNsense by UUID.", Attributes: map[string]dsschema.Attribute{
		"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
		"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this static route is enabled."},
		"network":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Destination network."},
		"gateway":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Gateway address."},
		"interface":   dsschema.StringAttribute{Computed: true, MarkdownDescription: "Outgoing interface."},
		"bfd":         dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether BFD is enabled."},
		"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the static route."},
	}}
}

func (d *staticRouteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(c *opnsense.Client) { d.client = c })
}

func (d *staticRouteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StaticRouteResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[staticRouteAPIResponse](ctx, d.client, staticRouteReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading static route", fmt.Sprintf("Could not read static route %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
