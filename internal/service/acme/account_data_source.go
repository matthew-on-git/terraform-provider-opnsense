// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &accountDataSource{}

type accountDataSource struct{ client *opnsense.Client }

func newAccountDataSource() datasource.DataSource { return &accountDataSource{} }

func (d *accountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_account"
}

func (d *accountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads an existing ACME account on OPNsense by UUID.", Attributes: map[string]dsschema.Attribute{
		"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
		"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this ACME account is enabled."},
		"name":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "ACME account name."},
		"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the ACME account."},
		"email":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "Account email address."},
		"ca":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "ACME CA endpoint."},
	}}
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
	var config AccountResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[accountAPIResponse](ctx, d.client, accountReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ACME account", fmt.Sprintf("Could not read ACME account %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
