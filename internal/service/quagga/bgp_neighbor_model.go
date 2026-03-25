// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// BGPNeighborResourceModel is the Terraform state model for opnsense_quagga_bgp_neighbor.
type BGPNeighborResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Description         types.String `tfsdk:"description"`
	Address             types.String `tfsdk:"address"`
	RemoteAS            types.Int64  `tfsdk:"remote_as"`
	UpdateSource        types.String `tfsdk:"update_source"`
	NextHopSelf         types.Bool   `tfsdk:"next_hop_self"`
	MultiProtocol       types.Bool   `tfsdk:"multi_protocol"`
	Keepalive           types.Int64  `tfsdk:"keepalive"`
	Holddown            types.Int64  `tfsdk:"holddown"`
	LinkedPrefixlistIn  types.Set    `tfsdk:"linked_prefixlist_in"`
	LinkedPrefixlistOut types.Set    `tfsdk:"linked_prefixlist_out"`
	LinkedRoutemapIn    types.Set    `tfsdk:"linked_routemap_in"`
	LinkedRoutemapOut   types.Set    `tfsdk:"linked_routemap_out"`
}

type bgpNeighborAPIResponse struct {
	Enabled             string                   `json:"enabled"`
	Description         string                   `json:"description"`
	Address             string                   `json:"address"`
	RemoteAS            string                   `json:"remoteas"`
	UpdateSource        opnsense.SelectedMap     `json:"updatesource"`
	NextHopSelf         string                   `json:"nexthopself"`
	MultiProtocol       string                   `json:"multiprotocol"`
	Keepalive           string                   `json:"keepalive"`
	Holddown            string                   `json:"holddown"`
	LinkedPrefixlistIn  opnsense.SelectedMapList `json:"linkedPrefixlistIn"`
	LinkedPrefixlistOut opnsense.SelectedMapList `json:"linkedPrefixlistOut"`
	LinkedRoutemapIn    opnsense.SelectedMapList `json:"linkedRoutemapIn"`
	LinkedRoutemapOut   opnsense.SelectedMapList `json:"linkedRoutemapOut"`
}

type bgpNeighborAPIRequest struct {
	Enabled             string `json:"enabled"`
	Description         string `json:"description"`
	Address             string `json:"address"`
	RemoteAS            string `json:"remoteas"`
	UpdateSource        string `json:"updatesource"`
	NextHopSelf         string `json:"nexthopself"`
	MultiProtocol       string `json:"multiprotocol"`
	Keepalive           string `json:"keepalive"`
	Holddown            string `json:"holddown"`
	LinkedPrefixlistIn  string `json:"linkedPrefixlistIn"`
	LinkedPrefixlistOut string `json:"linkedPrefixlistOut"`
	LinkedRoutemapIn    string `json:"linkedRoutemapIn"`
	LinkedRoutemapOut   string `json:"linkedRoutemapOut"`
}

func (m *BGPNeighborResourceModel) toAPI(ctx context.Context) *bgpNeighborAPIRequest {
	toCSV := func(s types.Set) string {
		if s.IsNull() || s.IsUnknown() {
			return ""
		}
		var elems []string
		s.ElementsAs(ctx, &elems, false)
		return strings.Join(elems, ",")
	}

	var keepaliveStr, holddownStr string
	if !m.Keepalive.IsNull() && !m.Keepalive.IsUnknown() {
		keepaliveStr = opnsense.Int64ToString(m.Keepalive.ValueInt64())
	}
	if !m.Holddown.IsNull() && !m.Holddown.IsUnknown() {
		holddownStr = opnsense.Int64ToString(m.Holddown.ValueInt64())
	}

	return &bgpNeighborAPIRequest{
		Enabled:             opnsense.BoolToString(m.Enabled.ValueBool()),
		Description:         m.Description.ValueString(),
		Address:             m.Address.ValueString(),
		RemoteAS:            opnsense.Int64ToString(m.RemoteAS.ValueInt64()),
		UpdateSource:        m.UpdateSource.ValueString(),
		NextHopSelf:         opnsense.BoolToString(m.NextHopSelf.ValueBool()),
		MultiProtocol:       opnsense.BoolToString(m.MultiProtocol.ValueBool()),
		Keepalive:           keepaliveStr,
		Holddown:            holddownStr,
		LinkedPrefixlistIn:  toCSV(m.LinkedPrefixlistIn),
		LinkedPrefixlistOut: toCSV(m.LinkedPrefixlistOut),
		LinkedRoutemapIn:    toCSV(m.LinkedRoutemapIn),
		LinkedRoutemapOut:   toCSV(m.LinkedRoutemapOut),
	}
}

func (m *BGPNeighborResourceModel) fromAPI(_ context.Context, a *bgpNeighborAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Description = types.StringValue(a.Description)
	m.Address = types.StringValue(a.Address)
	m.UpdateSource = types.StringValue(string(a.UpdateSource))
	m.NextHopSelf = types.BoolValue(opnsense.StringToBool(a.NextHopSelf))
	m.MultiProtocol = types.BoolValue(opnsense.StringToBool(a.MultiProtocol))

	if a.RemoteAS != "" {
		if v, err := opnsense.StringToInt64(a.RemoteAS); err == nil {
			m.RemoteAS = types.Int64Value(v)
		}
	}
	if a.Keepalive != "" {
		if v, err := opnsense.StringToInt64(a.Keepalive); err == nil {
			m.Keepalive = types.Int64Value(v)
		}
	} else {
		m.Keepalive = types.Int64Null()
	}
	if a.Holddown != "" {
		if v, err := opnsense.StringToInt64(a.Holddown); err == nil {
			m.Holddown = types.Int64Value(v)
		}
	} else {
		m.Holddown = types.Int64Null()
	}

	fromSML := func(sml opnsense.SelectedMapList) types.Set {
		if len(sml) == 0 {
			return types.SetValueMust(types.StringType, []attr.Value{})
		}
		vals := make([]attr.Value, len(sml))
		for i, v := range sml {
			vals[i] = types.StringValue(v)
		}
		return types.SetValueMust(types.StringType, vals)
	}

	m.LinkedPrefixlistIn = fromSML(a.LinkedPrefixlistIn)
	m.LinkedPrefixlistOut = fromSML(a.LinkedPrefixlistOut)
	m.LinkedRoutemapIn = fromSML(a.LinkedRoutemapIn)
	m.LinkedRoutemapOut = fromSML(a.LinkedRoutemapOut)
}
