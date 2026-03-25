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
)

// Schema defines the Terraform schema for opnsense_haproxy_healthcheck.
func (r *healthcheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an HAProxy health check on OPNsense. Health checks verify server availability in backend pools. Requires the `os-haproxy` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the HAProxy health check in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the health check.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the health check.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("http"),
				MarkdownDescription: "Check type: `tcp`, `http`, `agent`, `ldap`, `mysql`, `pgsql`, `redis`, `smtp`, `esmtp`, `ssl`.",
				Validators: []validator.String{
					stringvalidator.OneOf("tcp", "http", "agent", "ldap", "mysql", "pgsql", "redis", "smtp", "esmtp", "ssl"),
				},
			},
			"interval": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("2s"),
				MarkdownDescription: "Check interval (e.g., `2s`, `500ms`, `1m`).",
			},
			"check_port": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Port to use for health checks. Empty to use the server port.",
			},
			"http_method": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("options"),
				MarkdownDescription: "HTTP method for health check: `options`, `head`, `get`, `put`, `post`, `delete`, `trace`.",
				Validators: []validator.String{
					stringvalidator.OneOf("options", "head", "get", "put", "post", "delete", "trace"),
				},
			},
			"http_uri": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/"),
				MarkdownDescription: "URI path for HTTP health check. Defaults to `/`.",
			},
			"http_version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("http10"),
				MarkdownDescription: "HTTP version: `http10`, `http11`, `http2`.",
				Validators: []validator.String{
					stringvalidator.OneOf("http10", "http11", "http2"),
				},
			},
			"force_ssl": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Force SSL for health checks. Defaults to `false`.",
			},
		},
	}
}
