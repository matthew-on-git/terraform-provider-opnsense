// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &actionDataSource{}

type actionDataSource struct{ client *opnsense.Client }

func newActionDataSource() datasource.DataSource { return &actionDataSource{} }

func (d *actionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_action"
}

func (d *actionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing HAProxy action on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id":                      dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"name":                    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the HAProxy action."},
			"description":             dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the HAProxy action."},
			"test_type":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "ACL match polarity: 'if' or 'unless'."},
			"linked_acls":             dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Set of HAProxy ACL UUIDs used as conditions."},
			"operator":                dsschema.StringAttribute{Computed: true, MarkdownDescription: "How multiple linked ACLs are combined."},
			"type":                    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Action type."},
			"use_backend":             dsschema.StringAttribute{Computed: true, MarkdownDescription: "Backend UUID for use_backend actions."},
			"mapfile":                 dsschema.StringAttribute{Computed: true, MarkdownDescription: "Map file UUID for map_use_backend actions."},
			"map_use_backend_default": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Default backend UUID for map_use_backend actions."},
			"http_request_option":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Raw HAProxy option text for HTTP request actions."},
			"deny_status":             dsschema.Int64Attribute{Computed: true, MarkdownDescription: "HTTP status code for deny actions."},
			"redirect":                dsschema.StringAttribute{Computed: true, MarkdownDescription: "Redirect rule text."},
			"set_header_name":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Header name for set-header actions."},
			"set_header_content":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Header value for set-header actions."},
		},
	}
}

func (d *actionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *actionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ActionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[actionAPIResponse](ctx, d.client, actionReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy action", fmt.Sprintf("Could not read HAProxy action %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
