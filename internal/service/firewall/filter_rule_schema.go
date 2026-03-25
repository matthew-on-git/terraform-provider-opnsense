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

// Schema defines the Terraform schema for opnsense_firewall_filter_rule.
func (r *filterRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a firewall filter rule on OPNsense with savepoint rollback protection. Bad rules auto-revert within 60 seconds.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the firewall filter rule in OPNsense.",
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
			"action": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Rule action: `pass`, `block`, or `reject`.",
				Validators: []validator.String{
					stringvalidator.OneOf("pass", "block", "reject"),
				},
			},
			"quick": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "First match wins. Defaults to `true`.",
			},
			"interface": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Network interfaces this rule applies to.",
			},
			"direction": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("in"),
				MarkdownDescription: "Traffic direction: `in`, `out`, or `any`.",
				Validators: []validator.String{
					stringvalidator.OneOf("in", "out", "any"),
				},
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
				Default:             stringdefault.StaticString("any"),
				MarkdownDescription: "Protocol (e.g., `any`, `tcp`, `udp`, `icmp`, `TCP/UDP`).",
			},
			"source_net": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("any"),
				MarkdownDescription: "Source network (`any`, interface name, CIDR, or alias).",
			},
			"source_port": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Source port or range (e.g., `80`, `80:443`). Empty for any.",
			},
			"source_not": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Invert source match. Defaults to `false`.",
			},
			"destination_net": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("any"),
				MarkdownDescription: "Destination network (`any`, interface name, CIDR, or alias).",
			},
			"destination_port": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Destination port or range (e.g., `443`, `8080:8090`). Empty for any.",
			},
			"destination_not": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Invert destination match. Defaults to `false`.",
			},
			"gateway": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Gateway for policy routing. Empty for default.",
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
