// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Schema defines the Terraform schema for opnsense_firewall_category.
func (r *categoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a firewall category on OPNsense. Categories organize firewall rules into logical groups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the firewall category in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the category. Must be unique. Commas are not allowed.",
			},
			"auto": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether this category is automatically applied. Defaults to `true`.",
			},
			"color": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Color as 6-digit hex (e.g., `ff0000` for red). Empty for no color.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^([0-9a-fA-F]{6})?$`),
						"must be a 6-digit hex color code or empty",
					),
				},
			},
		},
	}
}
