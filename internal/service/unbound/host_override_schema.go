// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

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

// Schema defines the Terraform schema for opnsense_unbound_host_override.
func (r *hostOverrideResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Unbound DNS host override on OPNsense. Host overrides map hostnames to IP addresses.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the host override in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether this host override is enabled. Defaults to `true`.",
			},
			"hostname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Hostname of the override (e.g., `www`).",
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Domain of the override (e.g., `example.com`).",
			},
			"rr": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("A"),
				MarkdownDescription: "DNS record type: `A`, `AAAA`, `MX`, or `TXT`. Defaults to `A`.",
				Validators: []validator.String{
					stringvalidator.OneOf("A", "AAAA", "MX", "TXT"),
				},
			},
			"server": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IP address or value for the DNS record.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the host override.",
			},
		},
	}
}
