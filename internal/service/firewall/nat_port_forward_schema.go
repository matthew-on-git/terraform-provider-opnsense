// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the Terraform schema for opnsense_firewall_nat_port_forward.
func (r *natPortForwardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a NAT port-forward rule on OPNsense. Redirects inbound traffic to an internal server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the NAT port-forward rule in OPNsense.",
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
			"interface": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Network interface for incoming traffic (e.g., `wan`).",
			},
			"ip_protocol": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("inet"),
				MarkdownDescription: "IP version: `inet` (IPv4), `inet6` (IPv6), or `inet46` (both).",
				Validators: []validator.String{
					stringvalidator.OneOf("inet", "inet6", "inet46"),
				},
			},
			"protocol": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("tcp"),
				MarkdownDescription: "Protocol (e.g., `tcp`, `udp`, `TCP/UDP`).",
			},
			"source_net": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("any"),
				MarkdownDescription: "Source network (`any`, CIDR, or alias).",
			},
			"source_port": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Source port or range. Empty for any.",
			},
			"source_not": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Invert source match. Defaults to `false`.",
			},
			"destination_net": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Destination network to match (typically the WAN address, e.g., `wanip`).",
			},
			"destination_port": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Destination port to match (the external port).",
			},
			"destination_not": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Invert destination match. Defaults to `false`.",
			},
			"target": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Internal target IP address or alias to redirect to.",
			},
			"local_port": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Internal target port to redirect to.",
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
