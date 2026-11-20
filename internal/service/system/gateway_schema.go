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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
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
			"latency_low":                  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Low latency threshold in milliseconds. Canonical signed-64-bit decimal string (no leading zeros); empty selects the appliance default. Omit to retain the saved value.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Validators: []validator.String{gatewayIntegerValidator{min: 1, max: -1}}},
			"latency_high":                 schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "High latency threshold in milliseconds. Canonical signed-64-bit decimal string (no leading zeros); empty selects the appliance default. Omit to retain the saved value.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Validators: []validator.String{gatewayIntegerValidator{min: 1, max: -1}}},
			"loss_low":                     schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Low packet loss threshold in percent (1-99). Canonical signed-64-bit decimal string (no leading zeros); empty selects the appliance default. Omit to retain the saved value.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Validators: []validator.String{gatewayIntegerValidator{min: 1, max: 99}}},
			"loss_high":                    schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "High packet loss threshold in percent (1-100). Canonical signed-64-bit decimal string (no leading zeros); empty selects the appliance default. Omit to retain the saved value.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Validators: []validator.String{gatewayIntegerValidator{min: 1, max: 100}}},
			"interval":                     schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Probe interval in seconds. Canonical signed-64-bit decimal string (no leading zeros); empty selects the appliance default. Omit to retain the saved value.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Validators: []validator.String{gatewayIntegerValidator{min: 1, max: -1}}},
			"time_period":                  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Time period in seconds over which monitoring results are averaged. Canonical signed-64-bit decimal string (no leading zeros); empty selects the appliance default. Omit to retain the saved value.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Validators: []validator.String{gatewayIntegerValidator{min: 1, max: -1}}},
			"loss_interval":                schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Interval in seconds before a probe is considered lost. Canonical signed-64-bit decimal string (no leading zeros); empty selects the appliance default. Omit to retain the saved value.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Validators: []validator.String{gatewayIntegerValidator{min: 1, max: -1}}},
			"data_length":                  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Probe payload length in bytes; zero is an explicit value. Canonical signed-64-bit decimal string (no leading zeros); empty selects the appliance default. Omit to retain the saved value.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, Validators: []validator.String{gatewayIntegerValidator{min: 0, max: -1}}},
			"monitor":                      schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Monitor IP address. An empty string uses the gateway address. Omit to retain the saved value.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"far_gateway":                  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Allow a gateway outside the interface subnet. Omit to retain the saved value.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
			"monitor_no_route":             schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Do not add a host route for the monitor address. Omit to retain the saved value.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
			"monitor_kill_states":          schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Flush states when the gateway goes down. Omit to retain the saved value.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
			"monitor_kill_states_priority": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Flush states when a higher-priority gateway becomes available. Omit to retain the saved value.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
			"force_down":                   schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Treat the gateway as down regardless of monitoring. Omit to retain the saved value.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
			"no_sync":                      schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Exclude this gateway from configuration synchronization. Omit to retain the saved value.", PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
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
