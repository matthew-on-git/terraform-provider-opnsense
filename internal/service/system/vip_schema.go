// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *vipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a virtual IP (CARP, IP Alias) on OPNsense.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"interface": schema.StringAttribute{
				Required: true, MarkdownDescription: "Interface (e.g., `lan`, `wan`).",
			},
			"mode": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("ipalias"),
				MarkdownDescription: "VIP type: `ipalias`, `carp`, or `proxyarp`.",
				Validators:          []validator.String{stringvalidator.OneOf("ipalias", "carp", "proxyarp")},
			},
			"address": schema.StringAttribute{
				Required: true, MarkdownDescription: "IP address.",
			},
			"subnet_bits": schema.Int64Attribute{
				Required: true, MarkdownDescription: "CIDR subnet mask bits (e.g., `24`).",
				Validators: []validator.Int64{int64validator.Between(1, 128)},
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Description.",
			},
			"vhid": schema.Int64Attribute{
				Optional: true, Computed: true, MarkdownDescription: "VHID for CARP (1-255).",
				Validators: []validator.Int64{int64validator.Between(1, 255)},
			},
			"password": schema.StringAttribute{
				Optional: true, Computed: true, Sensitive: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "CARP password.",
			},
			"adv_base": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(1),
				MarkdownDescription: "CARP advertisement interval (1-254).",
				Validators:          []validator.Int64{int64validator.Between(1, 254)},
			},
			"adv_skew": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				MarkdownDescription: "CARP advertisement skew (0-254).",
				Validators:          []validator.Int64{int64validator.Between(0, 254)},
			},
		},
	}
}
