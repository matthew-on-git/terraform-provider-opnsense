// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *tunableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a persistent system tunable on OPNsense. Tunables can affect kernel, network, firewall, or service behavior; test changes on a disposable appliance before applying them to production.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID of the system tunable.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tunable": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "System tunable name, for example `kern.msgbuf_show_timestamp`.",
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Configured tunable value. Values are strings because OPNsense tunables may be numeric, boolean-like, or text values.",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Description stored with the tunable.",
			},
		},
	}
}
