// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &hostAliasDataSource{}

type hostAliasDataSource struct{ client *opnsense.Client }

func newHostAliasDataSource() datasource.DataSource { return &hostAliasDataSource{} }

func (d *hostAliasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_host_alias"
}

func (d *hostAliasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing Unbound DNS host alias on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this host alias is enabled."},
			"host":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Parent host override UUID."},
			"hostname":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Alias hostname."},
			"domain":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Alias domain."},
			"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the host alias."},
		},
	}
}

func (d *hostAliasDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *hostAliasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config HostAliasResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[hostAliasAPIResponse](ctx, d.client, hostAliasReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Unbound host alias", fmt.Sprintf("Could not read Unbound host alias %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
