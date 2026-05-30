// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// StaticKeyResourceModel is the Terraform state model for opnsense_openvpn_static_key.
type StaticKeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Mode        types.String `tfsdk:"mode"`
	Key         types.String `tfsdk:"key"`
	Description types.String `tfsdk:"description"`
}

type staticKeyAPIResponse struct {
	Mode        opnsense.SelectedMap `json:"mode"`
	Key         string               `json:"key"`
	Description string               `json:"description"`
}

type staticKeyAPIRequest struct {
	Mode        string `json:"mode"`
	Key         string `json:"key"`
	Description string `json:"description"`
}

func (m *StaticKeyResourceModel) toAPI(_ context.Context) *staticKeyAPIRequest {
	return &staticKeyAPIRequest{
		Mode:        m.Mode.ValueString(),
		Key:         m.Key.ValueString(),
		Description: m.Description.ValueString(),
	}
}

func (m *StaticKeyResourceModel) fromAPI(_ context.Context, a *staticKeyAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Mode = types.StringValue(string(a.Mode))
	m.Key = types.StringValue(a.Key)
	m.Description = types.StringValue(a.Description)
}
