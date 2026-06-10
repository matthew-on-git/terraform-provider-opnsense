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

var _ datasource.DataSource = &aclDataSource{}

type aclDataSource struct{ client *opnsense.Client }

func newACLDataSource() datasource.DataSource { return &aclDataSource{} }

func (d *aclDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_acl"
}

func (d *aclDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing Unbound DNS ACL on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this ACL is enabled."},
			"name":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "ACL name."},
			"action":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "ACL action."},
			"networks":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Networks covered by this ACL."},
			"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the ACL."},
		},
	}
}

func (d *aclDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *aclDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ACLResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[unboundACLAPIResponse](ctx, d.client, aclReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Unbound ACL", fmt.Sprintf("Could not read Unbound ACL %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
