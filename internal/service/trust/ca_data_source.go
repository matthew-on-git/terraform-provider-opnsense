// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &caDataSource{}

type caDataSource struct{ client *opnsense.Client }

func newCADataSource() datasource.DataSource { return &caDataSource{} }

func (d *caDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trust_ca"
}

func (d *caDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing trust certificate authority on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"refid":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "OPNsense reference ID."},
			"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the certificate authority."},
			"common_name": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Common name."},
			"country":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Country code."},
			"lifetime":    dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Certificate authority lifetime in days."},
			"certificate": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Certificate payload returned by OPNsense."},
			"valid_from":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Certificate validity start."},
			"valid_to":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Certificate validity end."},
		},
	}
}

func (d *caDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *caDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CAResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[caAPIResponse](ctx, d.client, caReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading trust CA", fmt.Sprintf("Could not read trust CA %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
