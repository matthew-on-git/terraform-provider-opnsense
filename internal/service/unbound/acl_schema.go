// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

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

// Schema defines the Terraform schema for opnsense_unbound_acl.
func (r *aclResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Unbound DNS access control list (ACL) on OPNsense. ACLs control which clients can query the DNS resolver.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "UUID of the ACL in OPNsense.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether this ACL is enabled. Defaults to `true`.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the ACL.",
			},
			"action": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("allow"),
				MarkdownDescription: "ACL action: `allow`, `deny`, `refuse`, `allow_snoop`, `deny_non_local`, or `refuse_non_local`. Defaults to `allow`.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"allow", "deny", "refuse",
						"allow_snoop", "deny_non_local", "refuse_non_local",
					),
				},
			},
			"networks": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Networks to apply this ACL to (CIDR notation).",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Description of the ACL.",
			},
		},
	}
}
