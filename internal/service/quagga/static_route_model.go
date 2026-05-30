// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// StaticRouteResourceModel is the Terraform state model for opnsense_quagga_static_route.
type StaticRouteResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Network     types.String `tfsdk:"network"`
	Gateway     types.String `tfsdk:"gateway"`
	Interface   types.String `tfsdk:"interface"`
	BFD         types.Bool   `tfsdk:"bfd"`
	Description types.String `tfsdk:"description"`
}

type staticRouteAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Network     string               `json:"network"`
	Gateway     string               `json:"gateway"`
	Interface   opnsense.SelectedMap `json:"interfacename"`
	BFD         string               `json:"bfd"`
	Description string               `json:"description"`
}

type staticRouteAPIRequest struct {
	Enabled     string `json:"enabled"`
	Network     string `json:"network"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interfacename"`
	BFD         string `json:"bfd"`
	Description string `json:"description"`
}

func (m *StaticRouteResourceModel) toAPI(_ context.Context) *staticRouteAPIRequest {
	return &staticRouteAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Network:     m.Network.ValueString(),
		Gateway:     m.Gateway.ValueString(),
		Interface:   m.Interface.ValueString(),
		BFD:         opnsense.BoolToString(m.BFD.ValueBool()),
		Description: m.Description.ValueString(),
	}
}

func (m *StaticRouteResourceModel) fromAPI(_ context.Context, a *staticRouteAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Network = types.StringValue(a.Network)
	m.Gateway = types.StringValue(a.Gateway)
	m.Interface = types.StringValue(string(a.Interface))
	m.BFD = types.BoolValue(opnsense.StringToBool(a.BFD))
	m.Description = types.StringValue(a.Description)
}
