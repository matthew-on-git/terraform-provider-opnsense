// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *peerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = peerSchemaV1()
}

func peerSchemaV1() schema.Schema {
	return schema.Schema{
		Version:             1,
		MarkdownDescription: "Manages a WireGuard peer on OPNsense. WireGuard is part of OPNsense core since 24.1; this contract is acceptance-validated on 26.7.7. `servers` is an API-derived backreference; configure associations through a server's `peers` attribute.",
		Attributes: map[string]schema.Attribute{
			"id":             schema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the WireGuard peer in OPNsense.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"enabled":        schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(true), MarkdownDescription: "Whether this peer is enabled."},
			"name":           schema.StringAttribute{Required: true, MarkdownDescription: "Name of the WireGuard peer."},
			"public_key":     schema.StringAttribute{Required: true, MarkdownDescription: "Base64-encoded 256-bit public key."},
			"psk":            schema.StringAttribute{Optional: true, Computed: true, Sensitive: true, MarkdownDescription: "Write-only base64-encoded 256-bit pre-shared key. Omission preserves prior state without asserting that an unreadable key is absent. Explicit blank disables PSK.", PlanModifiers: []planmodifier.String{preserveOmittedPeerPSK{}}},
			"tunnel_address": schema.SetAttribute{ElementType: types.StringType, Required: true, MarkdownDescription: "Allowed tunnel CIDRs.", Validators: []validator.Set{setvalidator.SizeAtLeast(1)}},
			"server_address": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString(""), MarkdownDescription: "Remote endpoint address."},
			"server_port":    schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Remote endpoint port. Null means unset.", Validators: []validator.Int64{int64validator.Between(1, 65535)}},
			"keepalive":      schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: "Persistent keepalive interval in seconds. Null means unset.", Validators: []validator.Int64{int64validator.Between(1, 65535)}},
			"servers":        schema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Authoritative API backreference to server UUIDs. Configure the association through server.peers."},
		},
	}
}
