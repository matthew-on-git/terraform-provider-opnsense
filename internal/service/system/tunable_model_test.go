// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTunableModel_toAPI(t *testing.T) {
	t.Parallel()

	model := TunableResourceModel{
		Tunable:     types.StringValue("kern.msgbuf_show_timestamp"),
		Value:       types.StringValue("1"),
		Description: types.StringValue("Managed by Terraform"),
	}

	api := model.toAPI(context.Background())
	if api.Tunable != "kern.msgbuf_show_timestamp" || api.Value != "1" || api.Description != "Managed by Terraform" {
		t.Fatalf("unexpected API request: %#v", api)
	}
}

func TestTunableModel_fromAPI(t *testing.T) {
	t.Parallel()

	var model TunableResourceModel
	model.fromAPI(context.Background(), &tunableAPIResponse{
		Tunable:      "kern.msgbuf_show_timestamp",
		Value:        "1",
		Description:  "Managed by Terraform",
		DefaultValue: "1",
		Type:         "w",
	}, "tunable-uuid")

	if model.ID.ValueString() != "tunable-uuid" {
		t.Fatalf("id = %q, want tunable-uuid", model.ID.ValueString())
	}
	if model.Tunable.ValueString() != "kern.msgbuf_show_timestamp" || model.Value.ValueString() != "1" || model.Description.ValueString() != "Managed by Terraform" {
		t.Fatalf("unexpected model: %#v", model)
	}
}
