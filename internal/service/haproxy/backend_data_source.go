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

var _ datasource.DataSource = &backendDataSource{}

type backendDataSource struct{ client *opnsense.Client }

func newBackendDataSource() datasource.DataSource { return &backendDataSource{} }

func (d *backendDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_backend"
}

func (d *backendDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing HAProxy backend on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":              dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this backend is enabled. Defaults to 'true'."},
			"name":                 dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the backend pool."},
			"description":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the backend."},
			"mode":                 dsschema.StringAttribute{Computed: true, MarkdownDescription: "Backend mode: 'http' (Layer 7) or 'tcp' (Layer 4)."},
			"algorithm":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "Load balancing algorithm: 'source', 'roundrobin', 'static-rr', 'leastconn', 'uri', 'random'."},
			"linked_servers":       dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Set of HAProxy server UUIDs linked to this backend."},
			"health_check_enabled": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether health checking is enabled. Defaults to 'true'."},
			"persistence":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Session persistence mode: 'sticktable' or 'cookie'."},
			"forward_for":          dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Add X-Forwarded-For header. Defaults to 'false'."},
		},
	}
}

func (d *backendDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *backendDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config BackendResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[backendAPIResponse](ctx, d.client, backendReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy backend", fmt.Sprintf("Could not read HAProxy backend %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
