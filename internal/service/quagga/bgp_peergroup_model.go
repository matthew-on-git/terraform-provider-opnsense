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
	ID               types.String `tfsdk:"id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Name             types.String `tfsdk:"name"`
	RemoteASMode     types.String `tfsdk:"remote_as_mode"`
	RemoteAS         types.Int64  `tfsdk:"remote_as"`
	Family           types.Set    `tfsdk:"family"`
	UpdateSource     types.String `tfsdk:"update_source"`
	NextHopSelf      types.Bool   `tfsdk:"next_hop_self"`
	DefaultOriginate types.Bool   `tfsdk:"default_originate"`
}

type bgpPeerGroupAPIResponse struct {
	Enabled          string                   `json:"enabled"`
	Name             string                   `json:"name"`
	RemoteASMode     opnsense.SelectedMap     `json:"remote_as_mode"`
	RemoteAS         string                   `json:"remoteas"`
	Family           opnsense.SelectedMapList `json:"family"`
	UpdateSource     opnsense.SelectedMap     `json:"updatesource"`
	NextHopSelf      string                   `json:"nexthopself"`
	DefaultOriginate string                   `json:"defaultoriginate"`
}

type bgpPeerGroupAPIRequest struct {
	Enabled          string `json:"enabled"`
	Name             string `json:"name"`
	RemoteASMode     string `json:"remote_as_mode"`
	RemoteAS         string `json:"remoteas"`
	Family           string `json:"family"`
	UpdateSource     string `json:"updatesource"`
	NextHopSelf      string `json:"nexthopself"`
	DefaultOriginate string `json:"defaultoriginate"`
}

func (m *BGPPeerGroupResourceModel) toAPI(ctx context.Context) *bgpPeerGroupAPIRequest {
	return &bgpPeerGroupAPIRequest{
		Enabled:          opnsense.BoolToString(m.Enabled.ValueBool()),
		Name:             m.Name.ValueString(),
		RemoteASMode:     m.RemoteASMode.ValueString(),
		RemoteAS:         intOrEmpty(m.RemoteAS.ValueInt64()),
		Family:           setToCSV(ctx, m.Family),
		UpdateSource:     m.UpdateSource.ValueString(),
		NextHopSelf:      opnsense.BoolToString(m.NextHopSelf.ValueBool()),
		DefaultOriginate: opnsense.BoolToString(m.DefaultOriginate.ValueBool()),
	}
}

func (m *BGPPeerGroupResourceModel) fromAPI(_ context.Context, a *bgpPeerGroupAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.RemoteASMode = types.StringValue(string(a.RemoteASMode))
	m.RemoteAS = types.Int64Value(intOrZero(a.RemoteAS))
	m.Family = sliceToSet(a.Family)
	m.UpdateSource = types.StringValue(string(a.UpdateSource))
	m.NextHopSelf = types.BoolValue(opnsense.StringToBool(a.NextHopSelf))
	m.DefaultOriginate = types.BoolValue(opnsense.StringToBool(a.DefaultOriginate))
}
