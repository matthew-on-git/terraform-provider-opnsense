// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Schema defines the Terraform schema for opnsense_wireguard_peer.
func (r *peerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a WireGuard peer (client) on OPNsense. Requires the `os-wireguard` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the WireGuard peer in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether this peer is enabled. Defaults to `true`.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the WireGuard peer.",
			},
			"public_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Public key of the WireGuard peer.",
			},
			"tunnel_address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Allowed IPs / tunnel address for this peer (e.g., `10.0.0.2/32`).",
			},
			"server_address": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Endpoint address of the remote WireGuard server.",
			},
			"server_port": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Endpoint port of the remote WireGuard server.",
			},
			"keepalive": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Persistent keepalive interval in seconds. `0` disables.",
			},
		},
	}
}
