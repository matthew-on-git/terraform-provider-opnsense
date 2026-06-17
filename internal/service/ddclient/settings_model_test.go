// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSettingsModelToAPI(t *testing.T) {
	t.Parallel()

	model := SettingsResourceModel{
		Enabled:   types.BoolValue(true),
		Backend:   types.StringValue("ddclient"),
		Interval:  types.Int64Value(300),
		Verbose:   types.BoolValue(false),
		AllowIPv6: types.BoolValue(true),
	}

	api := model.toAPI(context.Background())
	if api.Enabled != "1" || api.Backend != "ddclient" || api.DaemonDelay != "300" || api.Verbose != "0" || api.AllowIPv6 != "1" {
		t.Fatalf("unexpected API request: %#v", api)
	}
}

func TestSettingsModelFromAPI(t *testing.T) {
	t.Parallel()

	var model SettingsResourceModel
	model.fromAPI(context.Background(), &settingsAPIResponse{
		Enabled:     "1",
		Backend:     "opnsense",
		DaemonDelay: "600",
		Verbose:     "1",
		AllowIPv6:   "0",
	}, settingsID)

	if model.ID.ValueString() != settingsID || !model.Enabled.ValueBool() || model.Backend.ValueString() != "opnsense" || model.Interval.ValueInt64() != 600 || !model.Verbose.ValueBool() || model.AllowIPv6.ValueBool() {
		t.Fatalf("unexpected model: %#v", model)
	}
}
