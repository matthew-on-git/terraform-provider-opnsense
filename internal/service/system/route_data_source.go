// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &routeDataSource{}

type routeDataSource struct{ client *opnsense.Client }

func newRouteDataSource() datasource.DataSource { return &routeDataSource{} }

func (d *routeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_route"
}

func (d *routeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing system static route on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this route is enabled. Defaults to 'true'."},
			"network":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Destination network in CIDR notation (e.g., '10.0.0.0/24')."},
			"gateway":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Gateway name or UUID."},
			"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description."},
		},
	}
}

func (d *routeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *routeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RouteResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[routeAPIResponse](ctx, d.client, routeReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading system static route", fmt.Sprintf("Could not read system static route %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
