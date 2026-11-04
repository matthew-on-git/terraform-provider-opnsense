// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
		Version:             1,
		MarkdownDescription: "Manages a BGP neighbor on OPNsense. Requires the `os-frr` plugin.",
		Attributes: map[string]schema.Attribute{
			"remote_as_mode":               schema.StringAttribute{Validators: []validator.String{stringvalidator.OneOf("", "internal", "external")}, Optional: true, Computed: true, MarkdownDescription: "Source-supported remote_as_mode; omitted values adopt appliance state."},
			"local_as":                     schema.Int64Attribute{Validators: []validator.Int64{int64validator.Between(1, 4294967295)}, Optional: true, Computed: true, MarkdownDescription: "Source-supported local_as; omitted values adopt appliance state."},
			"password":                     schema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "Config-owned authentication input. Omitted values are never adopted or cleared. Sensitive state is not encrypted; remote password drift cannot be proven."},
			"weight":                       schema.Int64Attribute{Validators: []validator.Int64{int64validator.Between(0, 65535)}, Optional: true, Computed: true, MarkdownDescription: "Source-supported weight; omitted values adopt appliance state."},
			"local_ip":                     schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported local_ip; omitted values adopt appliance state."},
			"link_local_interface":         schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported link_local_interface; omitted values adopt appliance state."},
			"next_hop_self_all":            schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported next_hop_self_all; omitted values adopt appliance state."},
			"multihop":                     schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported multihop; omitted values adopt appliance state."},
			"route_reflector_client":       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported route_reflector_client; omitted values adopt appliance state."},
			"soft_reconfiguration_inbound": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported soft_reconfiguration_inbound; omitted values adopt appliance state."},
			"bfd":                          schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported bfd; omitted values adopt appliance state."},
			"bfd_strict":                   schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported bfd_strict; omitted values adopt appliance state."},
			"connect_timer":                schema.Int64Attribute{Validators: []validator.Int64{int64validator.Between(1, 65000)}, Optional: true, Computed: true, MarkdownDescription: "Source-supported connect_timer; omitted values adopt appliance state."},
			"default_originate":            schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported default_originate; omitted values adopt appliance state."},
			"as_override":                  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported as_override; omitted values adopt appliance state."},
			"remove_private_as":            schema.StringAttribute{Validators: []validator.String{stringvalidator.OneOf("", "remove-private-AS", "remove-private-AS all", "remove-private-AS replace-AS", "remove-private-AS all replace-AS")}, Optional: true, Computed: true, MarkdownDescription: "Source-supported remove_private_as; omitted values adopt appliance state."},
			"allow_as_in":                  schema.StringAttribute{Validators: []validator.String{stringvalidator.OneOf("", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "origin")}, Optional: true, Computed: true, MarkdownDescription: "Source-supported allow_as_in; omitted values adopt appliance state."},
			"disable_connected_check":      schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Source-supported disable_connected_check; omitted values adopt appliance state."},
			"attribute_unchanged":          schema.StringAttribute{Validators: []validator.String{stringvalidator.OneOf("", "as-path", "next-hop", "med")}, Optional: true, Computed: true, MarkdownDescription: "Source-supported attribute_unchanged; omitted values adopt appliance state."},
			"capabilities":                 schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, MarkdownDescription: "Source-supported capabilities; omitted values adopt appliance state."},
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
				Optional: true, Computed: true, MarkdownDescription: "Remote AS number; mode takes precedence and peer groups may supply it.",
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
			"peer_group": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Existing BGP peer-group UUID. Its policies are activated per selected address family on FRR 2.0."},
			"family":     familyAttribute(),
			"multi_protocol": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Legacy alias: true activates ipv4 and ipv6; false selects the transport family. Prefer family.",
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
