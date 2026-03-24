// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

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

// Schema defines the Terraform schema for opnsense_haproxy_server.
func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an HAProxy server on OPNsense. Servers define backend targets for load balancing. Requires the `os-haproxy` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the HAProxy server in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the server.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the server.",
			},
			"address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IP address or hostname of the backend server.",
			},
			"port": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Port number of the backend server (1-65535).",
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"weight": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Load balancing weight (0-256). Higher values receive more traffic.",
				Validators: []validator.Int64{
					int64validator.Between(0, 256),
				},
			},
			"mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("active"),
				MarkdownDescription: "Server mode: `active`, `backup`, or `disabled`.",
				Validators: []validator.String{
					stringvalidator.OneOf("active", "backup", "disabled"),
				},
			},
			"ssl": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to use SSL/TLS for connections to this server. Defaults to `false`.",
			},
			"ssl_verify": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether to verify the server's SSL certificate. Defaults to `true`.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether this server is enabled. Defaults to `true`.",
			},
		},
	}
}
