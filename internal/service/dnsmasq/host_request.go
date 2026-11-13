// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dnsmasq

import "github.com/hashicorp/terraform-plugin-framework/types"

// knownString distinguishes an explicit clear from an omitted value.
func knownString(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueString()
	return &v
}
