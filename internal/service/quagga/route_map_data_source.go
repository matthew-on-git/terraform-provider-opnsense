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

var _ datasource.DataSource = &routeMapDataSource{}

type routeMapDataSource struct{ client *opnsense.Client }

func newRouteMapDataSource() datasource.DataSource { return &routeMapDataSource{} }

func (d *routeMapDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_route_map"
}

func (d *routeMapDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads an existing FRR route map on OPNsense by UUID.", Attributes: map[string]dsschema.Attribute{
		"id":           dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
		"enabled":      dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this route map is enabled."},
		"description":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the route map."},
		"name":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Route map name."},
		"action":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "Route map action."},
		"order":        dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Route map order."},
		"match_prefix": dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Matched prefix lists."},
		"set":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Route map set expression."},
	}}
}

func (d *routeMapDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(c *opnsense.Client) { d.client = c })
}

func (d *routeMapDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RouteMapResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[routeMapAPIResponse](ctx, d.client, routeMapReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading route map", fmt.Sprintf("Could not read route map %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
