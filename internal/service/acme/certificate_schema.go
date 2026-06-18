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
			"issuance_timeout": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("180s"),
				MarkdownDescription: "Maximum time to wait for ACME issuance after create or update. Use a Go duration such as `180s`. Avoid tight values because ACME providers enforce rate limits.",
			},
			"issuance_poll_interval": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("10s"),
				MarkdownDescription: "Interval between ACME issuance status polls. Use a Go duration such as `10s`. Avoid tight loops because ACME providers enforce rate limits.",
			},
			"cert_ref_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "HAProxy legacy certificate refid populated after successful issuance. Use this value in `opnsense_haproxy_frontend.certificates`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status_code": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ACME issuance status code reported by OPNsense. `200` indicates an issued certificate.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Human-readable ACME issuance status reported by OPNsense.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
