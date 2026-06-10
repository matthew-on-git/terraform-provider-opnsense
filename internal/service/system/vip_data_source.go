// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &vipDataSource{}

type vipDataSource struct{ client *opnsense.Client }

type vipDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Interface   types.String `tfsdk:"interface"`
	Mode        types.String `tfsdk:"mode"`
	Address     types.String `tfsdk:"address"`
	SubnetBits  types.Int64  `tfsdk:"subnet_bits"`
	Description types.String `tfsdk:"description"`
	VHID        types.Int64  `tfsdk:"vhid"`
	AdvBase     types.Int64  `tfsdk:"adv_base"`
	AdvSkew     types.Int64  `tfsdk:"adv_skew"`
}

func newVipDataSource() datasource.DataSource { return &vipDataSource{} }

func (d *vipDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_vip"
}

func (d *vipDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing system virtual IP on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"interface":   dsschema.StringAttribute{Computed: true, MarkdownDescription: "Interface (e.g., 'lan', 'wan')."},
			"mode":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "VIP type: 'ipalias', 'carp', or 'proxyarp'."},
			"address":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "IP address."},
			"subnet_bits": dsschema.Int64Attribute{Computed: true, MarkdownDescription: "CIDR subnet mask bits (e.g., '24')."},
			"description": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description."},
			"vhid":        dsschema.Int64Attribute{Computed: true, MarkdownDescription: "VHID for CARP (1-255)."},
			"adv_base":    dsschema.Int64Attribute{Computed: true, MarkdownDescription: "CARP advertisement interval (1-254)."},
			"adv_skew":    dsschema.Int64Attribute{Computed: true, MarkdownDescription: "CARP advertisement skew (0-254)."},
			// Omitted from the data source because OPNsense does not reliably return: password.
		},
	}
}

func (d *vipDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vipDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config vipDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[vipAPIResponse](ctx, d.client, vipReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading system virtual IP", fmt.Sprintf("Could not read system virtual IP %s: %s", id, err))
		return
	}
	config.ID = types.StringValue(id)
	config.Interface = types.StringValue(string(result.Interface))
	config.Mode = types.StringValue(string(result.Mode))
	config.Address = types.StringValue(result.Address)
	config.Description = types.StringValue(result.Description)
	config.SubnetBits = stringToInt64Value(result.SubnetBits)
	config.VHID = stringToNullableInt64Value(result.VHID)
	config.AdvBase = stringToInt64Value(result.AdvBase)
	config.AdvSkew = stringToInt64Value(result.AdvSkew)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func stringToInt64Value(value string) types.Int64 {
	if value == "" {
		return types.Int64Null()
	}
	parsed, err := opnsense.StringToInt64(value)
	if err != nil {
		return types.Int64Null()
	}
	return types.Int64Value(parsed)
}

func stringToNullableInt64Value(value string) types.Int64 {
	return stringToInt64Value(value)
}
