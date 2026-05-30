// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

func (r *instanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenVPN instance (server or client) on OPNsense using the modern instances API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID of the OpenVPN instance.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Whether the instance is enabled. Defaults to `true`.",
			},
			"role": schema.StringAttribute{
				Required: true, MarkdownDescription: "Instance role: `server` or `client`.",
				Validators: []validator.String{stringvalidator.OneOf("server", "client")},
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Description.",
			},
			"dev_type": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("tun"),
				MarkdownDescription: "Device type: `tun`, `tap`, or `ovpn`. Defaults to `tun`.",
				Validators:          []validator.String{stringvalidator.OneOf("tun", "tap", "ovpn")},
			},
			"protocol": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("udp"),
				MarkdownDescription: "Protocol: `udp`, `udp4`, `udp6`, `tcp`, `tcp4`, `tcp6`. Defaults to `udp`.",
				Validators:          []validator.String{stringvalidator.OneOf("udp", "udp4", "udp6", "tcp", "tcp4", "tcp6")},
			},
			"port": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Listen/connect port.",
			},
			"local": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Local interface address to bind.",
			},
			"server": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "IPv4 tunnel network (CIDR) for server role.",
			},
			"topology": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("subnet"),
				MarkdownDescription: "Topology: `subnet`, `p2p`, or `net30`. Defaults to `subnet`.",
				Validators:          []validator.String{stringvalidator.OneOf("subnet", "p2p", "net30")},
			},
			"ca": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Certificate Authority reference (UUID/refid).",
			},
			"cert": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Server/client certificate reference (UUID/refid).",
			},
			"tls_key": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "TLS static key reference (UUID of an opnsense_openvpn_static_key).",
			},
			"data_ciphers": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "Allowed data ciphers (e.g. `AES-256-GCM`).",
			},
			"auth": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Auth digest algorithm (e.g. `SHA256`).",
			},
			"dns_servers": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "DNS servers pushed to clients.",
			},
			"push_route": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "Networks (CIDR) pushed as routes to clients.",
			},
			"redirect_gateway": schema.SetAttribute{
				ElementType: types.StringType, Optional: true, Computed: true,
				MarkdownDescription: "redirect-gateway flags (e.g. `def1`, `bypass-dhcp`).",
			},
			"max_clients": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				MarkdownDescription: "Maximum number of connected clients (0 = unset).",
			},
			"keepalive_interval": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				MarkdownDescription: "Keepalive ping interval in seconds (0 = unset).",
			},
			"keepalive_timeout": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(0),
				MarkdownDescription: "Keepalive timeout in seconds (0 = unset).",
			},
			"verb": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("3"),
				MarkdownDescription: "Log verbosity level (0-11). Defaults to `3`.",
			},
		},
	}
}
