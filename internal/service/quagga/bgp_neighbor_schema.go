// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *bgpNeighborResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a BGP neighbor on OPNsense. Requires the `os-frr` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Whether this neighbor is enabled.",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Description.",
			},
			"address": schema.StringAttribute{
				Required: true, MarkdownDescription: "Neighbor IP address.",
			},
			"remote_as": schema.Int64Attribute{
				Required: true, MarkdownDescription: "Remote AS number.",
				Validators: []validator.Int64{int64validator.Between(1, 4294967295)},
			},
			"update_source": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Interface or IP to use as update source.",
			},
			"next_hop_self": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Set next-hop to self for advertised routes.",
			},
			"multi_protocol": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Enable multi-protocol BGP (IPv4+IPv6).",
			},
			"keepalive": schema.Int64Attribute{
				Optional: true, Computed: true, MarkdownDescription: "Keepalive interval in seconds (1-1000).",
				Validators: []validator.Int64{int64validator.Between(1, 1000)},
			},
			"holddown": schema.Int64Attribute{
				Optional: true, Computed: true, MarkdownDescription: "Hold-down timer in seconds (3-3000).",
				Validators: []validator.Int64{int64validator.Between(3, 3000)},
			},
			"linked_prefixlist_in": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "Inbound prefix list UUIDs.",
			},
			"linked_prefixlist_out": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "Outbound prefix list UUIDs.",
			},
			"linked_routemap_in": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "Inbound route map UUIDs.",
			},
			"linked_routemap_out": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "Outbound route map UUIDs.",
			},
		},
	}
}
