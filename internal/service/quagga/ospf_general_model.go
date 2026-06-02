// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// OSPFGeneralResourceModel is the Terraform state model for opnsense_quagga_ospf_general (singleton).
type OSPFGeneralResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	RouterID            types.String `tfsdk:"router_id"`
	CostReference       types.Int64  `tfsdk:"cost_reference"`
	LogAdjacencyChanges types.Bool   `tfsdk:"log_adjacency_changes"`
	Originate           types.Bool   `tfsdk:"originate_default"`
	OriginateAlways     types.Bool   `tfsdk:"originate_default_always"`
	OriginateMetric     types.Int64  `tfsdk:"originate_default_metric"`
	PassiveInterfaces   types.Set    `tfsdk:"passive_interfaces"`
	CARPDemote          types.Bool   `tfsdk:"carp_demote"`
}

type ospfGeneralAPIResponse struct {
	Enabled             string                   `json:"enabled"`
	RouterID            string                   `json:"routerid"`
	CostReference       string                   `json:"costreference"`
	LogAdjacencyChanges string                   `json:"logadjacencychanges"`
	Originate           string                   `json:"originate"`
	OriginateAlways     string                   `json:"originatealways"`
	OriginateMetric     string                   `json:"originatemetric"`
	PassiveInterfaces   opnsense.SelectedMapList `json:"passiveinterfaces"`
	CARPDemote          string                   `json:"carp_demote"`
}

type ospfGeneralAPIRequest struct {
	Enabled             string `json:"enabled"`
	RouterID            string `json:"routerid"`
	CostReference       string `json:"costreference"`
	LogAdjacencyChanges string `json:"logadjacencychanges"`
	Originate           string `json:"originate"`
	OriginateAlways     string `json:"originatealways"`
	OriginateMetric     string `json:"originatemetric"`
	PassiveInterfaces   string `json:"passiveinterfaces"`
	CARPDemote          string `json:"carp_demote"`
}

func (m *OSPFGeneralResourceModel) toAPI(ctx context.Context) *ospfGeneralAPIRequest {
	return &ospfGeneralAPIRequest{
		Enabled:             opnsense.BoolToString(m.Enabled.ValueBool()),
		RouterID:            m.RouterID.ValueString(),
		CostReference:       intOrEmpty(m.CostReference.ValueInt64()),
		LogAdjacencyChanges: opnsense.BoolToString(m.LogAdjacencyChanges.ValueBool()),
		Originate:           opnsense.BoolToString(m.Originate.ValueBool()),
		OriginateAlways:     opnsense.BoolToString(m.OriginateAlways.ValueBool()),
		OriginateMetric:     intOrEmpty(m.OriginateMetric.ValueInt64()),
		PassiveInterfaces:   setToCSV(ctx, m.PassiveInterfaces),
		CARPDemote:          opnsense.BoolToString(m.CARPDemote.ValueBool()),
	}
}

func (m *OSPFGeneralResourceModel) fromAPI(_ context.Context, a *ospfGeneralAPIResponse, id string) {
	m.ID = types.StringValue(id)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.RouterID = types.StringValue(a.RouterID)
	m.CostReference = types.Int64Value(intOrZero(a.CostReference))
	m.LogAdjacencyChanges = types.BoolValue(opnsense.StringToBool(a.LogAdjacencyChanges))
	m.Originate = types.BoolValue(opnsense.StringToBool(a.Originate))
	m.OriginateAlways = types.BoolValue(opnsense.StringToBool(a.OriginateAlways))
	m.OriginateMetric = types.Int64Value(intOrZero(a.OriginateMetric))
	m.PassiveInterfaces = sliceToSet(a.PassiveInterfaces)
	m.CARPDemote = types.BoolValue(opnsense.StringToBool(a.CARPDemote))
}
