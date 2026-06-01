// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *caResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an internally-generated (self-signed) Certificate Authority on OPNsense. Referenced by other resources (e.g. `opnsense_openvpn_instance.ca`) via the computed `refid`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID of the CA.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"refid": schema.StringAttribute{
				Computed: true, MarkdownDescription: "OPNsense reference id, used by other resources to reference this CA.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Required: true, MarkdownDescription: "Human-readable name for the CA.",
			},
			"common_name": schema.StringAttribute{
				Optional: true, Computed: true,
				MarkdownDescription: "Certificate common name. Changing forces recreation.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"country": schema.StringAttribute{
				Optional: true, Computed: true,
				MarkdownDescription: "Two-letter country code. Changing forces recreation.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"lifetime": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Validity in days (assigned by OPNsense).",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"certificate": schema.StringAttribute{
				Computed: true, MarkdownDescription: "The generated CA certificate (PEM).",
			},
			"valid_from": schema.StringAttribute{
				Computed: true, MarkdownDescription: "Certificate validity start.",
			},
			"valid_to": schema.StringAttribute{
				Computed: true, MarkdownDescription: "Certificate validity end.",
			},
		},
	}
}
