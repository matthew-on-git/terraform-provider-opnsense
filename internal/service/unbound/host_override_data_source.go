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

var _ datasource.DataSource = &hostOverrideDataSource{}

type hostOverrideDataSource struct{ client *opnsense.Client }

func newHostOverrideDataSource() datasource.DataSource { return &hostOverrideDataSource{} }

func (d *hostOverrideDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_host_override"
}

func (d *hostOverrideDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing Unbound DNS host override on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this host override is enabled."},
			"hostname":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Hostname of the override."},
			"domain":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Domain of the override."},
			"rr":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "DNS record type."},
			"server":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "IP address or value for the DNS record."},
			"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the host override."},
		},
	}
}

func (d *hostOverrideDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *hostOverrideDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config HostOverrideResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[hostOverrideAPIResponse](ctx, d.client, hostOverrideReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Unbound host override", fmt.Sprintf("Could not read Unbound host override %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
