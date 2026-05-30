// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// OSPF6GeneralResourceModel is the Terraform state model for opnsense_quagga_ospf6_general (singleton).
type OSPF6GeneralResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	RouterID        types.String `tfsdk:"router_id"`
	Originate       types.Bool   `tfsdk:"originate_default"`
	OriginateAlways types.Bool   `tfsdk:"originate_default_always"`
	OriginateMetric types.Int64  `tfsdk:"originate_default_metric"`
	CARPDemote      types.Bool   `tfsdk:"carp_demote"`
}

type ospf6GeneralAPIResponse struct {
	Enabled         string `json:"enabled"`
	RouterID        string `json:"routerid"`
	Originate       string `json:"originate"`
	OriginateAlways string `json:"originatealways"`
	OriginateMetric string `json:"originatemetric"`
	CARPDemote      string `json:"carp_demote"`
}

type ospf6GeneralAPIRequest struct {
	Enabled         string `json:"enabled"`
	RouterID        string `json:"routerid"`
	Originate       string `json:"originate"`
	OriginateAlways string `json:"originatealways"`
	OriginateMetric string `json:"originatemetric"`
	CARPDemote      string `json:"carp_demote"`
}

func (m *OSPF6GeneralResourceModel) toAPI(_ context.Context) *ospf6GeneralAPIRequest {
	return &ospf6GeneralAPIRequest{
		Enabled:         opnsense.BoolToString(m.Enabled.ValueBool()),
		RouterID:        m.RouterID.ValueString(),
		Originate:       opnsense.BoolToString(m.Originate.ValueBool()),
		OriginateAlways: opnsense.BoolToString(m.OriginateAlways.ValueBool()),
		OriginateMetric: intOrEmpty(m.OriginateMetric.ValueInt64()),
		CARPDemote:      opnsense.BoolToString(m.CARPDemote.ValueBool()),
	}
}

func (m *OSPF6GeneralResourceModel) fromAPI(_ context.Context, a *ospf6GeneralAPIResponse, id string) {
	m.ID = types.StringValue(id)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.RouterID = types.StringValue(a.RouterID)
	m.Originate = types.BoolValue(opnsense.StringToBool(a.Originate))
	m.OriginateAlways = types.BoolValue(opnsense.StringToBool(a.OriginateAlways))
	m.OriginateMetric = types.Int64Value(intOrZero(a.OriginateMetric))
	m.CARPDemote = types.BoolValue(opnsense.StringToBool(a.CARPDemote))
}
