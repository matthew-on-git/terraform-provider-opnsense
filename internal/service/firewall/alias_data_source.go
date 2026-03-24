// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// Ensure aliasDataSource satisfies the datasource interface.
var _ datasource.DataSource = &aliasDataSource{}

// aliasDataSource implements the opnsense_firewall_alias data source.
type aliasDataSource struct {
	client *opnsense.Client
}

func newAliasDataSource() datasource.DataSource {
	return &aliasDataSource{}
}

// Metadata sets the data source type name.
func (d *aliasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_alias"
}

// Schema defines the data source schema. The id attribute is Required as the
// lookup key; all other attributes are Computed (read-only output).
func (d *aliasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up an existing firewall alias on OPNsense by UUID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the firewall alias to look up.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the alias.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Alias type (host, network, port, url, etc.).",
			},
			"content": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Alias entries (IPs, networks, ports, or URLs depending on type).",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the alias.",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether this alias is enabled.",
			},
			"proto": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Protocol filter (empty for any, IPv4, or IPv6).",
			},
			"categories": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of category UUIDs assigned to this alias.",
			},
			"update_freq": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Update frequency in days for URL table aliases.",
			},
		},
	}
}

// Configure extracts the OPNsense API client from provider data.
func (d *aliasDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			"Expected *opnsense.Client, got something else.",
		)
		return
	}
	d.client = client
}

// Read fetches a firewall alias by UUID from the OPNsense API.
func (d *aliasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AliasResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.ID.ValueString()

	result, err := opnsense.Get[aliasAPIResponse](ctx, d.client, aliasReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading firewall alias",
			fmt.Sprintf("Could not read firewall alias %s: %s", id, err),
		)
		return
	}

	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
