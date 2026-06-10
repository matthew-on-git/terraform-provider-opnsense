// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dnsmasq

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithConfigValidators = &optionResource{}

func (r *optionResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{optionTypeFieldsValidator{}}
}

type optionTypeFieldsValidator struct{}

func (optionTypeFieldsValidator) Description(_ context.Context) string {
	return "validates Dnsmasq DHCP option fields that are cleared by OPNsense for each option type"
}

func (v optionTypeFieldsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (optionTypeFieldsValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config OptionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.Type.IsUnknown() || config.Type.IsNull() {
		return
	}

	switch config.Type.ValueString() {
	case "set":
		if !config.SetTag.IsNull() && !config.SetTag.IsUnknown() && config.SetTag.ValueString() != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("set_tag"),
				"Invalid Dnsmasq Option Field",
				"OPNsense clears set_tag for Dnsmasq options with type set. Use type match when setting set_tag.",
			)
		}
	case "match":
		if !config.Interface.IsNull() && !config.Interface.IsUnknown() && config.Interface.ValueString() != "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("interface"),
				"Invalid Dnsmasq Option Field",
				"OPNsense clears interface for Dnsmasq options with type match. Use type set when matching on an interface.",
			)
		}
		if !config.Tag.IsNull() && !config.Tag.IsUnknown() && len(config.Tag.Elements()) > 0 {
			resp.Diagnostics.AddAttributeError(
				path.Root("tag"),
				"Invalid Dnsmasq Option Field",
				"OPNsense clears tag for Dnsmasq options with type match. Use type set when matching on tags.",
			)
		}
		if !config.Force.IsNull() && !config.Force.IsUnknown() && config.Force.ValueBool() {
			resp.Diagnostics.AddAttributeError(
				path.Root("force"),
				"Invalid Dnsmasq Option Field",
				"OPNsense clears force for Dnsmasq options with type match. Use type set when forcing an option.",
			)
		}
	}
}
