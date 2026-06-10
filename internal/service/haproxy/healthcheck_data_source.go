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

var _ datasource.DataSource = &healthcheckDataSource{}

type healthcheckDataSource struct{ client *opnsense.Client }

func newHealthcheckDataSource() datasource.DataSource { return &healthcheckDataSource{} }

func (d *healthcheckDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_healthcheck"
}

func (d *healthcheckDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing HAProxy health check on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"name":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the health check."},
			"description":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the health check."},
			"type":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Check type: 'tcp', 'http', 'agent', 'ldap', 'mysql', 'pgsql', 'redis', 'smtp', 'esmtp', 'ssl'."},
			"interval":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Check interval (e.g., '2s', '500ms', '1m')."},
			"check_port":   dsschema.StringAttribute{Computed: true, MarkdownDescription: "Port to use for health checks. Empty to use the server port."},
			"http_method":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "HTTP method for health check: 'options', 'head', 'get', 'put', 'post', 'delete', 'trace'."},
			"http_uri":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "URI path for HTTP health check. Defaults to '/'."},
			"http_version": dsschema.StringAttribute{Computed: true, MarkdownDescription: "HTTP version: 'http10', 'http11', 'http2'."},
			"force_ssl":    dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Force SSL for health checks. Defaults to 'false'."},
		},
	}
}

func (d *healthcheckDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *healthcheckDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config HealthcheckResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[healthcheckAPIResponse](ctx, d.client, healthcheckReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy health check", fmt.Sprintf("Could not read HAProxy health check %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
