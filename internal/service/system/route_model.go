// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// RouteResourceModel is the Terraform state model for opnsense_system_route.
type RouteResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Network     types.String `tfsdk:"network"`
	Gateway     types.String `tfsdk:"gateway"`
	Description types.String `tfsdk:"description"`
}

type routeAPIResponse struct {
	Disabled    string               `json:"disabled"`
	Network     string               `json:"network"`
	Gateway     opnsense.SelectedMap `json:"gateway"`
	Description string               `json:"descr"`
}

type routeAPIRequest struct {
	Disabled    string `json:"disabled"`
	Network     string `json:"network"`
	Gateway     string `json:"gateway"`
	Description string `json:"descr"`
}

func (m *RouteResourceModel) toAPI(_ context.Context) *routeAPIRequest {
	return &routeAPIRequest{
		Disabled:    opnsense.BoolToString(!m.Enabled.ValueBool()),
		Network:     m.Network.ValueString(),
		Gateway:     m.Gateway.ValueString(),
		Description: m.Description.ValueString(),
	}
}

func (m *RouteResourceModel) fromAPI(_ context.Context, a *routeAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(!opnsense.StringToBool(a.Disabled))
	m.Network = types.StringValue(a.Network)
	m.Gateway = types.StringValue(string(a.Gateway))
	m.Description = types.StringValue(a.Description)
}
