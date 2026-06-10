// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &categoryDataSource{}

type categoryDataSource struct{ client *opnsense.Client }

func newCategoryDataSource() datasource.DataSource { return &categoryDataSource{} }

func (d *categoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_category"
}

func (d *categoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing firewall category on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"name":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the category. Must be unique. Commas are not allowed."},
			"auto":  dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this category is automatically applied. Defaults to 'true'."},
			"color": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Color as 6-digit hex (e.g., 'ff0000' for red). Empty for no color."},
		},
	}
}

func (d *categoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *categoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CategoryResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[categoryAPIResponse](ctx, d.client, categoryReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall category", fmt.Sprintf("Could not read firewall category %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
