// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &serverDataSource{}

type serverDataSource struct{ client *opnsense.Client }

func newServerDataSource() datasource.DataSource { return &serverDataSource{} }

func (d *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_server"
}

func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing HAProxy server on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"name":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the server."},
			"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the server."},
			"address":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "IP address or hostname of the backend server."},
			"port":        dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Port number of the backend server (1-65535)."},
			"weight":      dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Load balancing weight (0-256). Higher values receive more traffic."},
			"mode":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Server mode: 'active', 'backup', or 'disabled'."},
			"ssl":         dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether to use SSL/TLS for connections to this server. Defaults to 'false'."},
			"ssl_verify":  dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether to verify the server's SSL certificate. Defaults to 'true'."},
			"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this server is enabled. Defaults to 'true'."},
		},
	}
}

func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ServerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[serverAPIResponse](ctx, d.client, serverReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy server", fmt.Sprintf("Could not read HAProxy server %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
