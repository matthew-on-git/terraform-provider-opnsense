// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &dhcpv6SubnetDataSource{}

type dhcpv6SubnetDataSource struct{ client *opnsense.Client }

func newDHCPv6SubnetDataSource() datasource.DataSource { return &dhcpv6SubnetDataSource{} }

func (d *dhcpv6SubnetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kea_dhcpv6_subnet"
}

func (d *dhcpv6SubnetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing Kea DHCPv6 subnet on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id":           dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"subnet":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "IPv6 subnet in CIDR notation."},
			"interface":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Interface for the subnet."},
			"allocator":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Address allocator."},
			"pd_allocator": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Prefix delegation allocator."},
			"pools":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "DHCPv6 pool definitions."},
			"description":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the subnet."},
		},
	}
}

func (d *dhcpv6SubnetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dhcpv6SubnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DHCPv6SubnetResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[dhcpv6SubnetAPIResponse](ctx, d.client, dhcpv6SubnetReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Kea DHCPv6 subnet", fmt.Sprintf("Could not read Kea DHCPv6 subnet %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
