// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Explicit lifecycle intent opts into in-place operations. Existing legacy
// configurations retain their historical immutable-PEM/subject replacement plan.
type legacyTrustReplacement struct{}

func (legacyTrustReplacement) Description(context.Context) string {
	return "Requires replacement for changed legacy immutable inputs unless explicit lifecycle intent is configured."
}

func (m legacyTrustReplacement) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (legacyTrustReplacement) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() || req.PlanValue.IsUnknown() || req.PlanValue.Equal(req.StateValue) {
		return
	}
	var action, token types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("action"), &action)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("operation_token"), &token)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if action.IsNull() && token.IsNull() {
		resp.RequiresReplace = true
	}
}
