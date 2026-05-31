// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *staticKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an OpenVPN TLS static key on OPNsense, referenced by instances via `tls_key`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID of the static key.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"mode": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("crypt"),
				MarkdownDescription: "Key usage mode (e.g. `crypt`, `auth`, `crypt-v2`). Defaults to `crypt`.",
			},
			"key": schema.StringAttribute{
				Required: true, Sensitive: true,
				MarkdownDescription: "OpenVPN static key material (PEM-style key block). Sensitive.",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Description.",
			},
		},
	}
}
