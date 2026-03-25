// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

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

// Schema defines the Terraform schema for opnsense_haproxy_backend.
func (r *backendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an HAProxy backend on OPNsense. Backends define server pools for load balancing. Requires the `os-haproxy` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the HAProxy backend in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether this backend is enabled. Defaults to `true`.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the backend pool.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the backend.",
			},
			"mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("http"),
				MarkdownDescription: "Backend mode: `http` (Layer 7) or `tcp` (Layer 4).",
				Validators: []validator.String{
					stringvalidator.OneOf("http", "tcp"),
				},
			},
			"algorithm": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("source"),
				MarkdownDescription: "Load balancing algorithm: `source`, `roundrobin`, `static-rr`, `leastconn`, `uri`, `random`.",
				Validators: []validator.String{
					stringvalidator.OneOf("source", "roundrobin", "static-rr", "leastconn", "uri", "random"),
				},
			},
			"linked_servers": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set of HAProxy server UUIDs linked to this backend.",
			},
			"health_check_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether health checking is enabled. Defaults to `true`.",
			},
			"persistence": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("sticktable"),
				MarkdownDescription: "Session persistence mode: `sticktable` or `cookie`.",
				Validators: []validator.String{
					stringvalidator.OneOf("sticktable", "cookie"),
				},
			},
			"forward_for": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Add X-Forwarded-For header. Defaults to `false`.",
			},
		},
	}
}
