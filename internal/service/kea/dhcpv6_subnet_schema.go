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
)

func (r *dhcpv6SubnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kea DHCPv6 subnet on OPNsense. The subnet's interface must be enabled in the Kea DHCPv6 general settings (`opnsense_kea_dhcpv6_settings`) first.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID of the subnet.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"subnet": schema.StringAttribute{
				Required: true, MarkdownDescription: "Subnet in CIDR notation (e.g. `2001:db8::/64`).",
			},
			"interface": schema.StringAttribute{
				Required: true, MarkdownDescription: "Interface this subnet is served on (must be enabled in the DHCPv6 general settings).",
			},
			"allocator": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Lease allocator strategy (empty = default, `iterative`, or `random`).",
			},
			"pd_allocator": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Prefix-delegation allocator strategy (empty = default, `iterative`, `random`, or `flq`).",
			},
			"pools": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Address pools for the subnet (newline-separated ranges).",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the subnet.",
			},
		},
	}
}
