// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *mapfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an HAProxy map file on OPNsense. Map files provide key/value routing tables for actions such as `map_use_backend`. Requires the `os-haproxy` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the HAProxy map file in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the HAProxy map file.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the HAProxy map file.",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(mapfileTypeDomain),
				MarkdownDescription: "Map match type. Supported values: `beg`, `dom`, `end`, `int`, `ip`, `reg`, `str`, `sub`. Defaults to `dom` for domain maps.",
				Validators: []validator.String{
					stringvalidator.OneOf(mapfileTypeBeg, mapfileTypeDomain, mapfileTypeEnd, mapfileTypeInt, mapfileTypeIP, mapfileTypeReg, mapfileTypeMapString, mapfileTypeSub),
				},
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Map file content as newline-separated key/value pairs, for example one `host backend-name` mapping per line.",
			},
		},
	}
}
