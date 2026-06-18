// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *actionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an HAProxy action on OPNsense. Actions connect ACL matches to routing, deny, redirect, and header rewrite behavior. Requires the `os-haproxy` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the HAProxy action in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the HAProxy action.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the HAProxy action.",
			},
			"test_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("if"),
				MarkdownDescription: "ACL match polarity: `if` or `unless`. Defaults to `if`.",
				Validators: []validator.String{
					stringvalidator.OneOf("if", "unless"),
				},
			},
			"linked_acls": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set of HAProxy ACL UUIDs used as conditions for this action.",
			},
			"operator": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("and"),
				MarkdownDescription: "How multiple linked ACLs are combined: `and` or `or`. Defaults to `and`.",
				Validators: []validator.String{
					stringvalidator.OneOf("and", "or"),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Action type. Supported values: `use_backend`, `map_use_backend`, `http-request_deny`, `http-request_redirect`, `http-request_set-header`.",
				Validators: []validator.String{
					stringvalidator.OneOf(actionTypeUseBackend, actionTypeMapUseBackend, actionTypeHTTPRequestDeny, actionTypeHTTPRequestRedirect, actionTypeHTTPRequestSetHeader),
				},
			},
			"use_backend": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Backend UUID for `type = \"use_backend\"`.",
			},
			"mapfile": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Map file UUID for `type = \"map_use_backend\"`.",
			},
			"map_use_backend_default": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Default backend UUID for `type = \"map_use_backend\"` when no map entry matches.",
			},
			"http_request_option": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Raw HAProxy option text for HTTP request actions when OPNsense requires an action-specific option string.",
			},
			"deny_status": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "HTTP status code for `type = \"http-request_deny\"`.",
				Validators: []validator.Int64{
					int64validator.Between(100, 599),
				},
			},
			"redirect": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Redirect rule text for `type = \"http-request_redirect\"`, for example `scheme https code 301`.",
			},
			"set_header_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Header name for `type = \"http-request_set-header\"`.",
			},
			"set_header_content": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Header value or format string for `type = \"http-request_set-header\"`.",
			},
		},
	}
}
