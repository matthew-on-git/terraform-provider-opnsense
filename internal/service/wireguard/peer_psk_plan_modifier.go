// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package wireguard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// An imported write-only key can legitimately be null. The framework's generic
// UseStateForUnknown intentionally does not copy null attribute state, so use
// resource existence to distinguish that case from creation. Never reinterpret
// a masked API value as a confirmed blank key or override explicit key intent.
type preserveOmittedPeerPSK struct{}

func (preserveOmittedPeerPSK) Description(context.Context) string {
	return "Preserves omitted PSK state, including unreadable imported keys, on existing peers."
}

func (m preserveOmittedPeerPSK) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (preserveOmittedPeerPSK) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || !req.ConfigValue.IsNull() || !req.PlanValue.IsUnknown() || req.StateValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}
