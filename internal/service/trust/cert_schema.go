// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *certResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Imports an existing X.509 certificate (and its private key) into the OPNsense trust store. The certificate and private key PEM payloads are write-only.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "UUID of the certificate.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"refid": schema.StringAttribute{
				Computed: true, MarkdownDescription: "OPNsense reference id, used by other resources to reference this certificate.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"description": schema.StringAttribute{
				Required: true, MarkdownDescription: "Human-readable name for the certificate.",
			},
			"ca_ref": schema.StringAttribute{
				Optional: true, Computed: true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Reference id of the issuing CA (empty for a self-contained/self-signed certificate).",
			},
			"certificate": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "PEM-encoded certificate. Write-only: not refreshed from the API.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"private_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "PEM-encoded private key. Write-only and sensitive: never read back from the API.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}
