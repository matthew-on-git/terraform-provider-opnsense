// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

func gatewayKnownString(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueString()
	return &v
}

func gatewayKnownBool(value types.Bool) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := opnsense.BoolToString(value.ValueBool())
	return &v
}

func gatewayReadString(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	return types.StringValue(*value)
}

func gatewayReadBool(value *gatewayAPIBool) types.Bool {
	if value == nil {
		return types.BoolNull()
	}
	return types.BoolValue(string(*value) == "1")
}

// gatewayAPIBool accepts the saved BooleanField wire format only. Missing and
// JSON-null pointer fields remain unknown rather than becoming writable false.
type gatewayAPIBool string

func (value *gatewayAPIBool) UnmarshalJSON(raw []byte) error {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("null gateway boolean must remain an absent pointer")
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return fmt.Errorf("invalid gateway boolean: %w", err)
	}
	switch text {
	case "", "0", "1":
		*value = gatewayAPIBool(text)
		return nil
	default:
		return fmt.Errorf("invalid gateway boolean: expected empty, 0 or 1")
	}
}
