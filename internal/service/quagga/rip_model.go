// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// RIPResourceModel is the Terraform state model for opnsense_quagga_rip (singleton).
type RIPResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Version       types.Int64  `tfsdk:"version"`
	Networks      types.Set    `tfsdk:"networks"`
	Redistribute  types.Set    `tfsdk:"redistribute"`
	DefaultMetric types.Int64  `tfsdk:"default_metric"`
}

type ripAPIResponse struct {
	Enabled       string                   `json:"enabled"`
	Version       string                   `json:"version"`
	Networks      opnsense.SelectedMapList `json:"networks"`
	Redistribute  opnsense.SelectedMapList `json:"redistribute"`
	DefaultMetric string                   `json:"defaultmetric"`
}

type ripAPIRequest struct {
	Enabled       string `json:"enabled"`
	Version       string `json:"version"`
	Networks      string `json:"networks"`
	Redistribute  string `json:"redistribute"`
	DefaultMetric string `json:"defaultmetric"`
}

func (m *RIPResourceModel) toAPI(ctx context.Context) *ripAPIRequest {
	return &ripAPIRequest{
		Enabled:       opnsense.BoolToString(m.Enabled.ValueBool()),
		Version:       opnsense.Int64ToString(m.Version.ValueInt64()),
		Networks:      setToCSV(ctx, m.Networks),
		Redistribute:  setToCSV(ctx, m.Redistribute),
		DefaultMetric: intOrEmpty(m.DefaultMetric.ValueInt64()),
	}
}

func (m *RIPResourceModel) fromAPI(_ context.Context, a *ripAPIResponse, id string) {
	m.ID = types.StringValue(id)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Version = types.Int64Value(intOrZero(a.Version))
	m.Networks = sliceToSet(a.Networks)
	m.Redistribute = sliceToSet(a.Redistribute)
	m.DefaultMetric = types.Int64Value(intOrZero(a.DefaultMetric))
}
