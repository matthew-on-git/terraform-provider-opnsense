// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &vlanDataSource{}

type vlanDataSource struct{ client *opnsense.Client }

func newVlanDataSource() datasource.DataSource { return &vlanDataSource{} }

func (d *vlanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_vlan"
}

func (d *vlanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing system VLAN on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"parent_interface": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Parent interface (e.g., 'vtnet0')."},
			"tag":              dsschema.Int64Attribute{Computed: true, MarkdownDescription: "VLAN tag (1-4094)."},
			"priority":         dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Priority code point (0-7)."},
			"proto":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "VLAN protocol. Empty for auto, '802.1q', or '802.1ad'."},
			"description":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description."},
			"device":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "VLAN device name (e.g., 'vlan0100')."},
		},
	}
}

func (d *vlanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VlanResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[vlanAPIResponse](ctx, d.client, vlanReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading system VLAN", fmt.Sprintf("Could not read system VLAN %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
