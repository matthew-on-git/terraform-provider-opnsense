// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ resource.ResourceWithConfigValidators = &mapfileResource{}

func (r *mapfileResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{mapfileContentValidator{}}
}

type mapfileContentValidator struct{}

func (mapfileContentValidator) Description(_ context.Context) string {
	return "validates HAProxy map file content is not empty after whitespace normalization"
}

func (v mapfileContentValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (mapfileContentValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config MapfileResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.Content.IsUnknown() || config.Content.IsNull() {
		return
	}

	if normalizeMapfileContent(config.Content.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("content"),
			"Invalid HAProxy Map File Content",
			"Map file content must contain at least one non-whitespace mapping line.",
		)
	}
}
