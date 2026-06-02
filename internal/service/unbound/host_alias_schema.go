// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Schema defines the Terraform schema for opnsense_unbound_host_alias.
func (r *hostAliasResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Unbound DNS host alias on OPNsense. A host alias adds an additional hostname that resolves to an existing host override.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the host alias in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether this host alias is enabled. Defaults to `true`.",
			},
			"host": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the host override this alias points to.",
			},
			"hostname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Hostname for the alias (the part before the domain).",
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Domain for the alias (e.g., `example.com`).",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the host alias.",
			},
		},
	}
}
