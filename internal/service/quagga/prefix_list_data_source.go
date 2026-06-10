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

var _ datasource.DataSource = &prefixListDataSource{}

type prefixListDataSource struct{ client *opnsense.Client }

func newPrefixListDataSource() datasource.DataSource { return &prefixListDataSource{} }

func (d *prefixListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_prefix_list"
}

func (d *prefixListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads an existing FRR prefix list on OPNsense by UUID.", Attributes: map[string]dsschema.Attribute{
		"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
		"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this prefix list is enabled."},
		"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the prefix list."},
		"name":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Prefix list name."},
		"version":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "IP version."},
		"sequence":    dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Sequence number."},
		"action":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Prefix list action."},
		"network":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Network prefix."},
	}}
}

func (d *prefixListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(c *opnsense.Client) { d.client = c })
}

func (d *prefixListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PrefixListResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[prefixListAPIResponse](ctx, d.client, prefixListReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading prefix list", fmt.Sprintf("Could not read prefix list %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
