// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.ResourceWithConfigValidators = &actionResource{}

func (r *actionResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{actionTypeFieldsValidator{}}
}

type actionTypeFieldsValidator struct{}

func (actionTypeFieldsValidator) Description(_ context.Context) string {
	return "validates required HAProxy action fields for each action type"
}

func (v actionTypeFieldsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (actionTypeFieldsValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config ActionResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.Type.IsUnknown() || config.Type.IsNull() {
		return
	}

	switch config.Type.ValueString() {
	case actionTypeUseBackend:
		requireString(resp, config.UseBackend, "use_backend", "use_backend is required when type is use_backend.")
	case actionTypeMapUseBackend:
		requireString(resp, config.MapUseBackendFile, "mapfile", "mapfile is required when type is map_use_backend.")
	case actionTypeHTTPRequestRedirect:
		requireString(resp, config.Redirect, "redirect", "redirect is required when type is http-request_redirect.")
	case actionTypeHTTPRequestSetHeader:
		requireString(resp, config.SetHeaderName, "set_header_name", "set_header_name is required when type is http-request_set-header.")
		requireString(resp, config.SetHeaderContent, "set_header_content", "set_header_content is required when type is http-request_set-header.")
	}
}

func requireString(resp *resource.ValidateConfigResponse, value types.String, attrName, detail string) {
	if value.IsUnknown() {
		return
	}
	if value.IsNull() || value.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root(attrName),
			"Missing HAProxy Action Field",
			detail,
		)
	}
}
