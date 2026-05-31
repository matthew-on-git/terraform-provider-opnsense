// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *bgpGlobalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages BGP global configuration on OPNsense (FRR). Singleton resource — one BGP config per appliance; `terraform destroy` only removes it from state. Requires the `os-frr` plugin and an enabled `opnsense_quagga_general`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "Synthetic identifier for this singleton (always `bgp`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Enable BGP. Defaults to `false`.",
			},
			"as_number": schema.Int64Attribute{
				Required: true, MarkdownDescription: "Local autonomous system number (1-4294967295).",
				Validators: []validator.Int64{int64validator.Between(1, 4294967295)},
			},
			"router_id": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "BGP router ID (IPv4 address). Empty = derived automatically.",
			},
			"distance": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				MarkdownDescription: "Administrative distance (1-255, 0 = unset).",
				Validators:          []validator.Int64{int64validator.Between(0, 255)},
			},
			"graceful_restart": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Enable graceful restart. Defaults to `false`.",
			},
			"network_import_check": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Check BGP network route existence in IGP. Defaults to `true`.",
			},
			"enforce_first_as": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Enforce first AS in AS-path of eBGP updates. Defaults to `true`.",
			},
			"log_neighbor_changes": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Log neighbor up/down changes. Defaults to `false`.",
			},
			"networks": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "Networks (CIDR) to advertise into BGP.",
			},
			"maximum_paths": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				MarkdownDescription: "Maximum eBGP ECMP paths (1-128, 0 = unset).",
				Validators:          []validator.Int64{int64validator.Between(0, 128)},
			},
			"maximum_paths_ibgp": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				MarkdownDescription: "Maximum iBGP ECMP paths (1-128, 0 = unset).",
				Validators:          []validator.Int64{int64validator.Between(0, 128)},
			},
		},
	}
}
