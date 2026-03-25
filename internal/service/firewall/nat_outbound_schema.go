// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the Terraform schema for opnsense_firewall_nat_outbound.
func (r *natOutboundResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an outbound NAT (source NAT) rule on OPNsense. Controls how internal traffic is translated when leaving through an interface.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the outbound NAT rule in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether this rule is enabled. Defaults to `true`.",
			},
			"sequence": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				MarkdownDescription: "Rule sequence number (1-999999). Controls evaluation order.",
				Validators: []validator.Int64{
					int64validator.Between(1, 999999),
				},
			},
			"interface": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Outbound network interface (e.g., `wan`).",
			},
			"ip_protocol": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("inet"),
				MarkdownDescription: "IP version: `inet` (IPv4) or `inet6` (IPv6).",
				Validators: []validator.String{
					stringvalidator.OneOf("inet", "inet6"),
				},
			},
			"protocol": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("any"),
				MarkdownDescription: "Protocol (e.g., `any`, `tcp`, `udp`, `TCP/UDP`).",
			},
			"source_net": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("any"),
				MarkdownDescription: "Source network (`any`, CIDR, or alias).",
			},
			"source_not": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Invert source match. Defaults to `false`.",
			},
			"source_port": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Source port or range. Empty for any.",
			},
			"destination_net": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("any"),
				MarkdownDescription: "Destination network (`any`, CIDR, or alias).",
			},
			"destination_not": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Invert destination match. Defaults to `false`.",
			},
			"destination_port": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Destination port or range. Empty for any.",
			},
			"target": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Translation target IP or alias (e.g., `wanip`, interface IP, or specific address).",
			},
			"target_port": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Translated source port. Empty to preserve original.",
			},
			"no_nat": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Suppress NAT for matching traffic. Defaults to `false`.",
			},
			"static_nat_port": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Preserve the original source port. Defaults to `false`.",
			},
			"log": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Log matching packets. Defaults to `false`.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the rule.",
			},
			"categories": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set of category UUIDs assigned to this rule.",
			},
		},
	}
}
