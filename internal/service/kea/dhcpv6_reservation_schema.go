// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *dhcpv6ReservationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kea DHCPv6 host reservation on OPNsense, pinning an IPv6 address to a client DUID within a subnet.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID of the reservation.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subnet_id": schema.StringAttribute{
				Required: true, MarkdownDescription: "UUID of the Kea DHCPv6 subnet this reservation belongs to.",
			},
			"ip_address": schema.StringAttribute{
				Required: true, MarkdownDescription: "Reserved IPv6 address (must fall within the subnet).",
			},
			"duid": schema.StringAttribute{
				Required: true, MarkdownDescription: "Client DUID the reservation matches.",
			},
			"hostname": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Hostname assigned to the client.",
			},
			"domain_search": schema.SetAttribute{
				Optional: true, Computed: true,
				ElementType:         types.StringType,
				MarkdownDescription: "Domain search list handed to the client.",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the reservation.",
			},
		},
	}
}
