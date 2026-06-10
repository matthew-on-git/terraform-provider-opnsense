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

var _ datasource.DataSource = &challengeDataSource{}

type challengeDataSource struct{ client *opnsense.Client }

func newChallengeDataSource() datasource.DataSource { return &challengeDataSource{} }

func (d *challengeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_challenge"
}

func (d *challengeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads an existing ACME challenge on OPNsense by UUID.", Attributes: map[string]dsschema.Attribute{
		"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
		"enabled":     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this challenge is enabled."},
		"name":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Challenge name."},
		"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the challenge."},
		"method":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Challenge method."},
		"dns_service": dsschema.StringAttribute{Computed: true, MarkdownDescription: "DNS service provider."},
		"dns_sleep":   dsschema.Int64Attribute{Computed: true, MarkdownDescription: "DNS propagation wait time in seconds."},
	}}
}

func (d *challengeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *challengeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ChallengeResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[challengeAPIResponse](ctx, d.client, challengeReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ACME challenge", fmt.Sprintf("Could not read ACME challenge %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
