// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Schema defines the Terraform schema for opnsense_unbound_domain_override.
func (r *domainOverrideResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Unbound DNS domain override (forward) on OPNsense. Domain overrides forward queries for a specific domain to a designated DNS server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the domain override in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether this domain override is enabled. Defaults to `true`.",
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Domain to override (e.g., `example.com`). Use an empty string for global forwarding.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("query"),
				MarkdownDescription: "Forward type: `query` for ordinary DNS or `dot` for DNS over TLS.",
				Validators:          []validator.String{stringvalidator.OneOf("query", "dot")},
			},
			"server": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IP address of the DNS server to forward queries to.",
			},
			"server_port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Destination server port. When omitted, query forwarding defaults to 53 and DoT defaults to 853; an explicit value is preserved.",
				PlanModifiers:       []planmodifier.Int64{domainOverridePortDefaultPlanModifier{}},
				Validators:          []validator.Int64{int64validator.Between(1, 65535)},
			},
			"verify_cn": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "TLS certificate name for DoT. A blank value intentionally permits an unverified or self-signed server.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the domain override.",
			},
		},
	}
}
