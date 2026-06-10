// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &clientOverwriteDataSource{}

type clientOverwriteDataSource struct{ client *opnsense.Client }

func newClientOverwriteDataSource() datasource.DataSource { return &clientOverwriteDataSource{} }

func (d *clientOverwriteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_openvpn_client_overwrite"
}

func (d *clientOverwriteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing OpenVPN client overwrite on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":         dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this overwrite is enabled. Defaults to 'true'."},
			"common_name":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Client certificate common name this overwrite applies to."},
			"description":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description."},
			"servers":         dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "OpenVPN instance UUIDs this overwrite applies to (empty = all)."},
			"block":           dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Block this client from connecting. Defaults to 'false'."},
			"push_reset":      dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Do not inherit the global push options. Defaults to 'false'."},
			"tunnel_network":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Client-specific IPv4 tunnel network (CIDR)."},
			"local_networks":  dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Local networks (CIDR) reachable by this client."},
			"remote_networks": dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Remote networks (CIDR) behind this client (iroute)."},
			"dns_servers":     dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "DNS servers pushed to this client."},
		},
	}
}

func (d *clientOverwriteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *clientOverwriteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ClientOverwriteResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[clientOverwriteAPIResponse](ctx, d.client, clientOverwriteReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading OpenVPN client overwrite", fmt.Sprintf("Could not read OpenVPN client overwrite %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
