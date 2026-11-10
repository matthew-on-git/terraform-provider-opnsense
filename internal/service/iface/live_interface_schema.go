// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package iface

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

type liveInterfaceDataSource struct {
	client        *opnsense.Client
	name          string
	overview, all bool
}

func (d *liveInterfaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.name
}

func (d *liveInterfaceDataSource) shape() liveShape {
	if d.overview {
		return overviewShape
	}
	return diagnosticShape
}

func (d *liveInterfaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := d.shape().attributes()
	if d.all {
		attributes = map[string]dsschema.Attribute{"interfaces": dsschema.ListNestedAttribute{Computed: true, NestedObject: dsschema.NestedAttributeObject{Attributes: attributes}}}
	} else {
		attributes["device"] = dsschema.StringAttribute{Required: true, MarkdownDescription: "Exact kernel device name; identifiers and descriptions are not aliases."}
	}
	resp.Schema = dsschema.Schema{Attributes: attributes, MarkdownDescription: "Reads allowlisted live interface observations without saving, applying or reconfiguring. Optional scalars and objects remain null; missing collections are empty. Interfaces sort by device and addresses by their complete public tuple, not primary priority. Runtime changes are observations, not managed configuration drift. Outputs can contain addresses and are stored in Terraform state; protect state access. Unknown backend fields, configuration and wireless data are never exposed. Assignment/IP ownership remains with existing resources."}
}

func (d *liveInterfaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *liveInterfaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	device := types.StringNull()
	if !d.all {
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("device"), &device)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if device.IsNull() || device.IsUnknown() || device.ValueString() == "" {
			resp.Diagnostics.AddError("Invalid device", "Provide an exact nonempty kernel device name.")
			return
		}
	}
	rows, err := readLiveInterfaces(ctx, d.client, d.overview)
	if err != nil {
		resp.Diagnostics.AddError("Error reading live interfaces", err.Error())
		return
	}
	if d.all {
		values := make([]attr.Value, len(rows))
		for i, row := range rows {
			values[i] = row
		}
		list, diags := types.ListValue(types.ObjectType{AttrTypes: d.shape().types()}, values)
		resp.Diagnostics.Append(diags...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &struct {
			Interfaces types.List `tfsdk:"interfaces"`
		}{Interfaces: list})...)
		return
	}
	for _, row := range rows {
		if row.Attributes()["device"].(types.String).ValueString() == device.ValueString() {
			resp.Diagnostics.Append(resp.State.Set(ctx, row)...)
			return
		}
	}
	resp.Diagnostics.AddError("Interface not found", fmt.Sprintf("No live observation matches device %q; verify the kernel device name and refresh after topology changes.", device.ValueString()))
}
