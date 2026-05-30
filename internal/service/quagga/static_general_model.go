// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// StaticGeneralResourceModel is the Terraform state model for opnsense_quagga_static (singleton).
type StaticGeneralResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

type staticGeneralAPIResponse struct {
	Enabled string `json:"enabled"`
}

type staticGeneralAPIRequest struct {
	Enabled string `json:"enabled"`
}

func (m *StaticGeneralResourceModel) toAPI(_ context.Context) *staticGeneralAPIRequest {
	return &staticGeneralAPIRequest{Enabled: opnsense.BoolToString(m.Enabled.ValueBool())}
}

func (m *StaticGeneralResourceModel) fromAPI(_ context.Context, a *staticGeneralAPIResponse, id string) {
	m.ID = types.StringValue(id)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
}
