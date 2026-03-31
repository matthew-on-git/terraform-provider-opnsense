// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *certificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ACME certificate on OPNsense. Requires the `os-acme-client` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Whether this certificate is enabled.",
			},
			"name": schema.StringAttribute{
				Required: true, MarkdownDescription: "Primary domain name (FQDN).",
			},
			"description": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Description.",
			},
			"alt_names": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString(""),
				MarkdownDescription: "Comma-separated Subject Alternative Names.",
			},
			"account": schema.StringAttribute{
				Required: true, MarkdownDescription: "UUID of the ACME account to use.",
			},
			"validation_method": schema.StringAttribute{
				Required: true, MarkdownDescription: "UUID of the challenge/validation method.",
			},
			"key_length": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("key_4096"),
				MarkdownDescription: "Key length: `key_2048`, `key_3072`, `key_4096`, `key_ec256`, `key_ec384`.",
				Validators:          []validator.String{stringvalidator.OneOf("key_2048", "key_3072", "key_4096", "key_ec256", "key_ec384")},
			},
			"auto_renewal": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Enable automatic renewal. Defaults to `true`.",
			},
		},
	}
}
