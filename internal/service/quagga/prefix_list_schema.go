// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *prefixListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a BGP prefix list entry on OPNsense. Requires the `os-frr` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Whether this prefix list entry is enabled.",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Description.",
			},
			"name": schema.StringAttribute{
				Required: true, MarkdownDescription: "Prefix list name.",
			},
			"version": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("IPv4"),
				MarkdownDescription: "IP version: `IPv4` or `IPv6`.",
				Validators:          []validator.String{stringvalidator.OneOf("IPv4", "IPv6")},
			},
			"sequence": schema.Int64Attribute{
				Required: true, MarkdownDescription: "Sequence number.",
			},
			"action": schema.StringAttribute{
				Required: true, MarkdownDescription: "Action: `permit` or `deny`.",
				Validators: []validator.String{stringvalidator.OneOf("permit", "deny")},
			},
			"network": schema.StringAttribute{
				Required: true, MarkdownDescription: "Network prefix (e.g., `10.0.0.0/8 le 24`).",
			},
		},
	}
}
