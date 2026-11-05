// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// BGPPeerGroupResourceModel is the Terraform state model for opnsense_quagga_bgp_peergroup.
type BGPPeerGroupResourceModel struct {
	ListenRanges        types.Set    `tfsdk:"listen_ranges"`
	LinkedRoutemapOut   types.Set    `tfsdk:"linked_routemap_out"`
	LinkedRoutemapIn    types.Set    `tfsdk:"linked_routemap_in"`
	LinkedPrefixlistOut types.Set    `tfsdk:"linked_prefixlist_out"`
	LinkedPrefixlistIn  types.Set    `tfsdk:"linked_prefixlist_in"`
	ID                  types.String `tfsdk:"id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Name                types.String `tfsdk:"name"`
	RemoteASMode        types.String `tfsdk:"remote_as_mode"`
	RemoteAS            types.Int64  `tfsdk:"remote_as"`
	Family              types.Set    `tfsdk:"family"`
	UpdateSource        types.String `tfsdk:"update_source"`
	NextHopSelf         types.Bool   `tfsdk:"next_hop_self"`
	DefaultOriginate    types.Bool   `tfsdk:"default_originate"`
}

type bgpPeerGroupAPIResponse struct {
	ListenRanges        frrSelectedMapList `json:"listenranges"`
	frrReadPresence     `json:"-"`
	LinkedRoutemapOut   frrSelectedMapList `json:"linkedRoutemapOut"`
	LinkedRoutemapIn    frrSelectedMapList `json:"linkedRoutemapIn"`
	LinkedPrefixlistOut frrSelectedMapList `json:"linkedPrefixlistOut"`
	LinkedPrefixlistIn  frrSelectedMapList `json:"linkedPrefixlistIn"`
	Enabled             string             `json:"enabled"`
	Name                string             `json:"name"`
	RemoteASMode        frrSelectedMap     `json:"remote_as_mode"`
	RemoteAS            string             `json:"remoteas"`
	Family              frrSelectedMapList `json:"family"`
	UpdateSource        frrSelectedMap     `json:"updatesource"`
	NextHopSelf         string             `json:"nexthopself"`
	DefaultOriginate    string             `json:"defaultoriginate"`
}

type bgpPeerGroupAPIRequest struct {
	ListenRanges        *string `json:"listenranges,omitempty"`
	LinkedRoutemapOut   string  `json:"linkedRoutemapOut"`
	LinkedRoutemapIn    string  `json:"linkedRoutemapIn"`
	LinkedPrefixlistOut string  `json:"linkedPrefixlistOut"`
	LinkedPrefixlistIn  string  `json:"linkedPrefixlistIn"`
	Enabled             string  `json:"enabled"`
	Name                string  `json:"name"`
	RemoteASMode        string  `json:"remote_as_mode"`
	RemoteAS            string  `json:"remoteas"`
	Family              string  `json:"family"`
	UpdateSource        string  `json:"updatesource"`
	NextHopSelf         string  `json:"nexthopself"`
	DefaultOriginate    string  `json:"defaultoriginate"`
}

func (m *BGPPeerGroupResourceModel) toAPI(ctx context.Context) *bgpPeerGroupAPIRequest {
	return &bgpPeerGroupAPIRequest{
		ListenRanges:        advancedSet(ctx, m.ListenRanges),
		LinkedRoutemapOut:   setToCSV(ctx, m.LinkedRoutemapOut),
		LinkedRoutemapIn:    setToCSV(ctx, m.LinkedRoutemapIn),
		LinkedPrefixlistOut: setToCSV(ctx, m.LinkedPrefixlistOut),
		LinkedPrefixlistIn:  setToCSV(ctx, m.LinkedPrefixlistIn),
		Enabled:             opnsense.BoolToString(m.Enabled.ValueBool()),
		Name:                m.Name.ValueString(),
		RemoteASMode:        m.RemoteASMode.ValueString(),
		RemoteAS:            intOrEmpty(m.RemoteAS.ValueInt64()),
		Family:              setToCSV(ctx, m.Family),
		UpdateSource:        m.UpdateSource.ValueString(),
		NextHopSelf:         opnsense.BoolToString(m.NextHopSelf.ValueBool()),
		DefaultOriginate:    opnsense.BoolToString(m.DefaultOriginate.ValueBool()),
	}
}

func (m *BGPPeerGroupResourceModel) fromAPI(_ context.Context, a *bgpPeerGroupAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.ListenRanges = readAdvancedSet(a.hasField("listenranges"), a.ListenRanges)
	m.LinkedRoutemapOut = sliceToSet(a.LinkedRoutemapOut)
	m.LinkedRoutemapIn = sliceToSet(a.LinkedRoutemapIn)
	m.LinkedPrefixlistOut = sliceToSet(a.LinkedPrefixlistOut)
	m.LinkedPrefixlistIn = sliceToSet(a.LinkedPrefixlistIn)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.RemoteASMode = types.StringValue(string(a.RemoteASMode))
	m.RemoteAS = types.Int64Value(intOrZero(a.RemoteAS))
	m.Family = sliceToSet(a.Family)
	m.UpdateSource = types.StringValue(string(a.UpdateSource))
	m.NextHopSelf = types.BoolValue(opnsense.StringToBool(a.NextHopSelf))
	m.DefaultOriginate = types.BoolValue(opnsense.StringToBool(a.DefaultOriginate))
}
