// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &accountDataSource{}

type accountDataSource struct{ client *opnsense.Client }

type accountDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Service     types.String `tfsdk:"service"`
	Hostnames   types.String `tfsdk:"hostnames"`
	Username    types.String `tfsdk:"username"`
	Description types.String `tfsdk:"description"`
}

func newAccountDataSource() datasource.DataSource { return &accountDataSource{} }

func (d *accountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ddclient_account"
}

func (d *accountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing dynamic DNS client account on OPNsense by UUID. The password is omitted because it is sensitive/write-only and is not populated from OPNsense read responses.",
		Attributes: map[string]dsschema.Attribute{
			"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this account is enabled."},
			"service":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Dynamic DNS service provider."},
			"hostnames":   dsschema.StringAttribute{Computed: true, MarkdownDescription: "Hostnames to update."},
			"username":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Username for the dynamic DNS service."},
			"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the account."},
			// Omitted from the data source because OPNsense does not return the write-only password.
		},
	}
}

func (d *accountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *accountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config accountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[ddclientAccountAPIResponse](ctx, d.client, accountReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ddclient account", fmt.Sprintf("Could not read ddclient account %s: %s", id, err))
		return
	}
	var account AccountResourceModel
	account.fromAPI(ctx, result, id)
	config.ID = account.ID
	config.Enabled = account.Enabled
	config.Service = account.Service
	config.Hostnames = account.Hostnames
	config.Username = account.Username
	config.Description = account.Description
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
