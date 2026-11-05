// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *bgpPeerGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a BGP peer group on OPNsense (FRR). Requires the `os-frr` plugin.",
		Attributes: map[string]schema.Attribute{
			"linked_prefixlist_in":  schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true, MarkdownDescription: "Linked linked_prefixlist_in policy entry UUIDs. FRR 2.0 logical group semantics are validated before writes."},
			"linked_prefixlist_out": schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true, MarkdownDescription: "Linked linked_prefixlist_out policy entry UUIDs. FRR 2.0 logical group semantics are validated before writes."},
			"linked_routemap_in":    schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true, MarkdownDescription: "Linked linked_routemap_in policy entry UUIDs. FRR 2.0 logical group semantics are validated before writes."},
			"linked_routemap_out":   schema.SetAttribute{ElementType: types.StringType, Optional: true, Computed: true, MarkdownDescription: "Linked linked_routemap_out policy entry UUIDs. FRR 2.0 logical group semantics are validated before writes."},
			"listen_ranges":         schema.SetAttribute{Optional: true, Computed: true, ElementType: types.StringType, MarkdownDescription: "Strict CIDR ranges for dynamic BGP peers."},
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Whether the peer group is enabled. Defaults to `true`.",
			},
			"name": schema.StringAttribute{
				Required: true, MarkdownDescription: "Peer group name.", Validators: []validator.String{stringvalidator.RegexMatches(frrPolicyName, "Use 1-64 letters, digits, dot, underscore or hyphen.")},
			},
			"remote_as_mode": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				Validators:          []validator.String{stringvalidator.OneOf("", "internal", "external")},
				MarkdownDescription: "Remote AS mode: `internal` or `external` (leave empty to use an explicit `remote_as`).",
			},
			"remote_as": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				Validators:          []validator.Int64{int64validator.Between(0, 4294967295)},
				MarkdownDescription: "Remote AS number (0 = unset / use remote_as_mode).",
			},
			"family": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "Address families: lowercase `ipv4`, `ipv6`.",
			},
			"update_source": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Update source interface.",
			},
			"next_hop_self": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Set next-hop to self. Defaults to `false`.",
			},
			"default_originate": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Originate default route to members. Defaults to `false`.",
			},
		},
	}
}
