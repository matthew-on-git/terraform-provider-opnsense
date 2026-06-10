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

var _ datasource.DataSource = &bgpNeighborDataSource{}

type bgpNeighborDataSource struct{ client *opnsense.Client }

func newBGPNeighborDataSource() datasource.DataSource { return &bgpNeighborDataSource{} }

func (d *bgpNeighborDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_neighbor"
}

func (d *bgpNeighborDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing BGP neighbor on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id":                    dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"enabled":               dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this BGP neighbor is enabled."},
			"description":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the BGP neighbor."},
			"address":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "Neighbor IP address."},
			"remote_as":             dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Remote autonomous system number."},
			"update_source":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Update source interface."},
			"next_hop_self":         dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether next-hop-self is enabled."},
			"multi_protocol":        dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether multi-protocol support is enabled."},
			"keepalive":             dsschema.Int64Attribute{Computed: true, MarkdownDescription: "BGP keepalive interval."},
			"holddown":              dsschema.Int64Attribute{Computed: true, MarkdownDescription: "BGP hold time."},
			"linked_prefixlist_in":  dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Inbound linked prefix lists."},
			"linked_prefixlist_out": dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Outbound linked prefix lists."},
			"linked_routemap_in":    dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Inbound linked route maps."},
			"linked_routemap_out":   dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Outbound linked route maps."},
		},
	}
}

func (d *bgpNeighborDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bgpNeighborDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config BGPNeighborResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[bgpNeighborAPIResponse](ctx, d.client, bgpNeighborReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP neighbor", fmt.Sprintf("Could not read BGP neighbor %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
