// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *vlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VLAN on OPNsense.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID of the VLAN.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"parent_interface": schema.StringAttribute{
				Required: true, MarkdownDescription: "Parent interface (e.g., `vtnet0`).",
			},
			"tag": schema.Int64Attribute{
				Required: true, MarkdownDescription: "VLAN tag (1-4094).",
				Validators: []validator.Int64{int64validator.Between(1, 4094)},
			},
			"priority": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				MarkdownDescription: "Priority code point (0-7).",
				Validators:          []validator.Int64{int64validator.Between(0, 7)},
			},
			"proto": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "VLAN protocol. Empty for auto, `802.1q`, or `802.1ad`.",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Description.",
			},
			"device": schema.StringAttribute{
				Required: true, MarkdownDescription: "VLAN device name (e.g., `vlan0100`).",
			},
		},
	}
}
