// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &serverDataSource{}

type serverDataSource struct{ client *opnsense.Client }

type serverDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Name          types.String `tfsdk:"name"`
	Port          types.String `tfsdk:"port"`
	TunnelAddress types.String `tfsdk:"tunnel_address"`
	Description   types.String `tfsdk:"description"`
}

func newServerDataSource() datasource.DataSource { return &serverDataSource{} }

func (d *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wireguard_server"
}

func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing WireGuard server on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":        dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this server instance is enabled. Defaults to 'true'."},
			"name":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the WireGuard server instance."},
			"port":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Listen port for this WireGuard instance."},
			"tunnel_address": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Tunnel address (e.g., '10.0.0.1/24')."},
			"description":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the server instance."},
			// Omitted from the data source because OPNsense does not reliably return: private_key.
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
	var config serverDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[wireguardServerAPIResponse](ctx, d.client, serverReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading WireGuard server", fmt.Sprintf("Could not read WireGuard server %s: %s", id, err))
		return
	}
	config.ID = types.StringValue(id)
	config.Enabled = types.BoolValue(opnsense.StringToBool(result.Enabled))
	config.Name = types.StringValue(result.Name)
	config.Port = types.StringValue(result.Port)
	config.TunnelAddress = types.StringValue(string(result.TunnelAddress))
	config.Description = types.StringValue(result.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
