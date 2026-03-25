// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *gatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a gateway on OPNsense.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Whether this gateway is enabled. Defaults to `true`.",
			},
			"name": schema.StringAttribute{
				Required: true, MarkdownDescription: "Gateway name (unique, alphanumeric + `_-`, max 32 chars).",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Description.",
			},
			"interface": schema.StringAttribute{
				Required: true, MarkdownDescription: "Interface (e.g., `wan`).",
			},
			"ip_protocol": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("inet"),
				MarkdownDescription: "IP version: `inet` (IPv4) or `inet6` (IPv6).",
				Validators:          []validator.String{stringvalidator.OneOf("inet", "inet6")},
			},
			"gateway": schema.StringAttribute{
				Required: true, MarkdownDescription: "Gateway IP address.",
			},
			"default_gateway": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Mark as default gateway. Defaults to `false`.",
			},
			"monitor_disable": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Disable gateway monitoring. Defaults to `true`.",
			},
			"weight": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(1),
				MarkdownDescription: "Weight for gateway groups (1-5).",
				Validators:          []validator.Int64{int64validator.Between(1, 5)},
			},
			"priority": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(255),
				MarkdownDescription: "Priority (0-255). Lower = higher priority.",
				Validators:          []validator.Int64{int64validator.Between(0, 255)},
			},
		},
	}
}
