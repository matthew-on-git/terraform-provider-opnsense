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

var _ datasource.DataSource = &domainOverrideDataSource{}

type domainOverrideDataSource struct{ client *opnsense.Client }

func newDomainOverrideDataSource() datasource.DataSource { return &domainOverrideDataSource{} }

func (d *domainOverrideDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_unbound_domain_override"
}

func (d *domainOverrideDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing Unbound DNS domain override on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this domain override is enabled."},
			"domain":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Domain to override."},
			"server":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Destination DNS server."},
			"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the domain override."},
		},
	}
}

func (d *domainOverrideDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *domainOverrideDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DomainOverrideResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[domainOverrideAPIResponse](ctx, d.client, domainOverrideReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Unbound domain override", fmt.Sprintf("Could not read Unbound domain override %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
