// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dhcp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *reservationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a DHCPv4 static reservation on OPNsense via the Kea DHCP server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subnet": schema.StringAttribute{
				Required: true, MarkdownDescription: "UUID of the parent subnet.",
			},
			"ip_address": schema.StringAttribute{
				Required: true, MarkdownDescription: "Fixed IP address to assign.",
			},
			"mac_address": schema.StringAttribute{
				Required: true, MarkdownDescription: "MAC address (e.g., `00:11:22:33:44:55`).",
			},
			"hostname": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Hostname for the reservation.",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Description.",
			},
		},
	}
}
