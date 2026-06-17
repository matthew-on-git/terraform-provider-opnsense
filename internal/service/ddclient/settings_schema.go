// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Dynamic DNS daemon settings on OPNsense. This is a singleton resource; `terraform destroy` only removes it from state and does not reset appliance settings. Requires the `os-ddclient` plugin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true, MarkdownDescription: "Synthetic identifier for this singleton (always `ddclient-settings`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"enabled": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(true),
				MarkdownDescription: "Enable the Dynamic DNS daemon. Defaults to `true`.",
			},
			"backend": schema.StringAttribute{
				Optional: true, Computed: true, Default: stringdefault.StaticString("opnsense"),
				MarkdownDescription: "Dynamic DNS backend: `ddclient` or native OPNsense backend (`opnsense`). Defaults to `opnsense`.",
				Validators:          []validator.String{stringvalidator.OneOf("ddclient", "opnsense")},
			},
			"interval": schema.Int64Attribute{
				Optional: true, Computed: true, Default: int64default.StaticInt64(300),
				MarkdownDescription: "Daemon update interval in seconds (`daemon_delay`). Must be between `1` and `86400`. Defaults to `300`.",
				Validators:          []validator.Int64{int64validator.Between(1, 86400)},
			},
			"verbose": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Enable verbose Dynamic DNS daemon logging. Defaults to `false`.",
			},
			"allow_ipv6": schema.BoolAttribute{
				Optional: true, Computed: true, Default: booldefault.StaticBool(false),
				MarkdownDescription: "Allow IPv6 address updates. Defaults to `false`.",
			},
		},
	}
}
