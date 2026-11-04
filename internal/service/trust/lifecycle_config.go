// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Optional+computed legacy fields may be unknown in the plan when omitted, but
// an actually configured unknown must never be converted into an API default.
func trustConfigKnown(config tfsdk.Config, diagnostics *diag.Diagnostics) bool {
	if !config.Raw.IsFullyKnown() {
		diagnostics.AddError("Unknown Trust configuration", errUnknownIntent.Error())
		return false
	}
	return true
}
