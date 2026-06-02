// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// Schema defines the Terraform schema for opnsense_ipsec_key_pair.
func (r *keyPairResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an IPsec key pair on OPNsense, used for public-key authentication. The public and private key PEM payloads are write-only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID of the key pair.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true, MarkdownDescription: "Unique name for the key pair.",
			},
			"key_type": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:             stringdefault.StaticString("rsa"),
				MarkdownDescription: "Key type: `rsa` or `ecdsa`. Defaults to `rsa`.",
			},
			"public_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "PEM-encoded public key. Write-only: not refreshed from the API.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"private_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "PEM-encoded private key. Write-only and sensitive: never read back from the API.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"key_size": schema.StringAttribute{
				Computed: true, MarkdownDescription: "Key size in bits (derived from the key).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"key_fingerprint": schema.StringAttribute{
				Computed: true, MarkdownDescription: "Fingerprint of the key (derived by OPNsense).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
