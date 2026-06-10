// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &connectionDataSource{}

type connectionDataSource struct{ client *opnsense.Client }

func newConnectionDataSource() datasource.DataSource { return &connectionDataSource{} }

func (d *connectionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_connection"
}

func (d *connectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing IPsec connection on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":      dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this connection is enabled. Defaults to 'true'."},
			"description":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the IPsec connection."},
			"remote_addrs": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Remote address(es) for the IPsec connection."},
			"version":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "IKE version: '0' (IKEv1+IKEv2), '1' (IKEv1), or '2' (IKEv2). Defaults to '2'."},
			"proposals":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "IKE proposals. Defaults to 'default'."},
			"unique":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "Connection uniqueness policy. Defaults to 'no'."},
		},
	}
}

func (d *connectionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *connectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ConnectionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[ipsecConnectionAPIResponse](ctx, d.client, connectionReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IPsec connection", fmt.Sprintf("Could not read IPsec connection %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
