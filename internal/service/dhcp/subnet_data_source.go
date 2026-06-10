// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dhcp

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &subnetDataSource{}

type subnetDataSource struct{ client *opnsense.Client }

func newSubnetDataSource() datasource.DataSource { return &subnetDataSource{} }

func (d *subnetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dhcpv4_subnet"
}

func (d *subnetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing DHCPv4 subnet on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id":          dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"subnet":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Subnet in CIDR notation."},
			"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the subnet."},
			"pools":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "DHCP pool definitions."},
			"option_data": dsschema.StringAttribute{Computed: true, MarkdownDescription: "DHCP option data."},
		},
	}
}

func (d *subnetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *subnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SubnetResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[subnetAPIResponse](ctx, d.client, subnetReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DHCPv4 subnet", fmt.Sprintf("Could not read DHCPv4 subnet %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
