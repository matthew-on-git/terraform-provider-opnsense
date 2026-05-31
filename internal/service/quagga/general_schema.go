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

func (r *generalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FRR (Quagga) general service settings on OPNsense. This is a singleton resource — there is one FRR configuration per appliance; `terraform destroy` only removes it from state and does not delete the appliance config. Requires the `os-frr` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "Synthetic identifier for this singleton (always `general`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Enable the FRR routing service. Defaults to `false`.",
			},
			"profile": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("traditional"),
				MarkdownDescription: "Routing profile: `traditional` or `datacenter`. Defaults to `traditional`.",
				Validators:          []validator.String{stringvalidator.OneOf("traditional", "datacenter")},
			},
			"enable_carp": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Enable CARP failover integration. Defaults to `false`.",
			},
			"enable_syslog": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Enable syslog logging. Defaults to `true`.",
			},
			"enable_snmp": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Enable SNMP (AgentX). Defaults to `false`.",
			},
			"syslog_level": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("notifications"),
				MarkdownDescription: "Syslog verbosity level. One of `critical`, `emergencies`, `errors`, `alerts`, `warnings`, `notifications`, `informational`, `debugging`.",
				Validators: []validator.String{stringvalidator.OneOf(
					"critical", "emergencies", "errors", "alerts", "warnings", "notifications", "informational", "debugging",
				)},
			},
			"firewall_rules": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Automatically generate firewall rules for routing protocols. Defaults to `true`.",
			},
		},
	}
}
