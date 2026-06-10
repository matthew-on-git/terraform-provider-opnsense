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

var _ datasource.DataSource = &frontendDataSource{}

type frontendDataSource struct{ client *opnsense.Client }

func newFrontendDataSource() datasource.DataSource { return &frontendDataSource{} }

func (d *frontendDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_frontend"
}

func (d *frontendDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing HAProxy frontend on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":         dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this frontend is enabled. Defaults to 'true'."},
			"name":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the frontend."},
			"description":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the frontend."},
			"bind":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "Listen address and port (e.g., '0.0.0.0:443', '192.168.1.1:80', ':8080')."},
			"mode":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "Frontend mode: 'http', 'ssl', or 'tcp'."},
			"default_backend": dsschema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the default HAProxy backend for this frontend."},
			"ssl_enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Enable SSL offloading. Defaults to 'false'."},
			"linked_actions":  dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Set of HAProxy action UUIDs linked to this frontend for ACL-based routing."},
			"forward_for":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Add X-Forwarded-For header. Defaults to 'false'."},
		},
	}
}

func (d *frontendDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *frontendDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FrontendResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[frontendAPIResponse](ctx, d.client, frontendReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy frontend", fmt.Sprintf("Could not read HAProxy frontend %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
