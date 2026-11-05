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
	RemoteASMode               types.String `tfsdk:"remote_as_mode"`
	LocalAS                    types.Int64  `tfsdk:"local_as"`
	Password                   types.String `tfsdk:"password"`
	Weight                     types.Int64  `tfsdk:"weight"`
	LocalIP                    types.String `tfsdk:"local_ip"`
	LinkLocalInterface         types.String `tfsdk:"link_local_interface"`
	NextHopSelfAll             types.Bool   `tfsdk:"next_hop_self_all"`
	Multihop                   types.Bool   `tfsdk:"multihop"`
	RouteReflectorClient       types.Bool   `tfsdk:"route_reflector_client"`
	SoftReconfigurationInbound types.Bool   `tfsdk:"soft_reconfiguration_inbound"`
	BFD                        types.Bool   `tfsdk:"bfd"`
	BFDStrict                  types.Bool   `tfsdk:"bfd_strict"`
	ConnectTimer               types.Int64  `tfsdk:"connect_timer"`
	DefaultOriginate           types.Bool   `tfsdk:"default_originate"`
	ASOverride                 types.Bool   `tfsdk:"as_override"`
	RemovePrivateAS            types.String `tfsdk:"remove_private_as"`
	AllowASIn                  types.String `tfsdk:"allow_as_in"`
	DisableConnectedCheck      types.Bool   `tfsdk:"disable_connected_check"`
	AttributeUnchanged         types.String `tfsdk:"attribute_unchanged"`
	Capabilities               types.Set    `tfsdk:"capabilities"`

	PeerGroup           types.String `tfsdk:"peer_group"`
	ID                  types.String `tfsdk:"id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	Description         types.String `tfsdk:"description"`
	Address             types.String `tfsdk:"address"`
	RemoteAS            types.Int64  `tfsdk:"remote_as"`
	UpdateSource        types.String `tfsdk:"update_source"`
	NextHopSelf         types.Bool   `tfsdk:"next_hop_self"`
	Family              types.Set    `tfsdk:"family"`
	MultiProtocol       types.Bool   `tfsdk:"multi_protocol"`
	Keepalive           types.Int64  `tfsdk:"keepalive"`
	Holddown            types.Int64  `tfsdk:"holddown"`
	LinkedPrefixlistIn  types.Set    `tfsdk:"linked_prefixlist_in"`
	LinkedPrefixlistOut types.Set    `tfsdk:"linked_prefixlist_out"`
	LinkedRoutemapIn    types.Set    `tfsdk:"linked_routemap_in"`
	LinkedRoutemapOut   types.Set    `tfsdk:"linked_routemap_out"`
}

type bgpNeighborAPIResponse struct {
	RemoteASMode               frrSelectedMap     `json:"remote_as_mode"`
	LocalAS                    string             `json:"localas"`
	Weight                     string             `json:"weight"`
	LocalIP                    string             `json:"localip"`
	LinkLocalInterface         frrSelectedMap     `json:"linklocalinterface"`
	NextHopSelfAll             string             `json:"nexthopselfall"`
	Multihop                   string             `json:"multihop"`
	RouteReflectorClient       string             `json:"rrclient"`
	SoftReconfigurationInbound string             `json:"soft_reconfiguration_inbound"`
	BFD                        string             `json:"bfd"`
	BFDStrict                  string             `json:"bfd_strict"`
	ConnectTimer               string             `json:"connecttimer"`
	DefaultOriginate           string             `json:"defaultoriginate"`
	ASOverride                 string             `json:"asoverride"`
	RemovePrivateAS            frrSelectedMap     `json:"removeprivateas"`
	AllowASIn                  frrSelectedMap     `json:"allowas_in"`
	DisableConnectedCheck      string             `json:"disable_connected_check"`
	AttributeUnchanged         frrSelectedMap     `json:"attributeunchanged"`
	Capabilities               frrSelectedMapList `json:"capabilities"`

	frrReadPresence     `json:"-"`
	PeerGroup           frrSelectedMap     `json:"peergroup"`
	Enabled             string             `json:"enabled"`
	Description         string             `json:"description"`
	Address             string             `json:"address"`
	RemoteAS            string             `json:"remoteas"`
	UpdateSource        frrSelectedMap     `json:"updatesource"`
	NextHopSelf         string             `json:"nexthopself"`
	Family              frrSelectedMapList `json:"family"`
	MultiProtocol       string             `json:"multiprotocol"`
	Keepalive           string             `json:"keepalive"`
	Holddown            string             `json:"holddown"`
	LinkedPrefixlistIn  frrSelectedMapList `json:"linkedPrefixlistIn"`
	LinkedPrefixlistOut frrSelectedMapList `json:"linkedPrefixlistOut"`
	LinkedRoutemapIn    frrSelectedMapList `json:"linkedRoutemapIn"`
	LinkedRoutemapOut   frrSelectedMapList `json:"linkedRoutemapOut"`
}

type bgpNeighborAPIRequest struct {
	RemoteASMode               *string `json:"remote_as_mode,omitempty"`
	LocalAS                    *string `json:"localas,omitempty"`
	Password                   *string `json:"password,omitempty"`
	Weight                     *string `json:"weight,omitempty"`
	LocalIP                    *string `json:"localip,omitempty"`
	LinkLocalInterface         *string `json:"linklocalinterface,omitempty"`
	NextHopSelfAll             *string `json:"nexthopselfall,omitempty"`
	Multihop                   *string `json:"multihop,omitempty"`
	RouteReflectorClient       *string `json:"rrclient,omitempty"`
	SoftReconfigurationInbound *string `json:"soft_reconfiguration_inbound,omitempty"`
	BFD                        *string `json:"bfd,omitempty"`
	BFDStrict                  *string `json:"bfd_strict,omitempty"`
	ConnectTimer               *string `json:"connecttimer,omitempty"`
	DefaultOriginate           *string `json:"defaultoriginate,omitempty"`
	ASOverride                 *string `json:"asoverride,omitempty"`
	RemovePrivateAS            *string `json:"removeprivateas,omitempty"`
	AllowASIn                  *string `json:"allowas_in,omitempty"`
	DisableConnectedCheck      *string `json:"disable_connected_check,omitempty"`
	AttributeUnchanged         *string `json:"attributeunchanged,omitempty"`
	Capabilities               *string `json:"capabilities,omitempty"`

	PeerGroup           string `json:"peergroup"`
	Enabled             string `json:"enabled"`
	Description         string `json:"description"`
	Address             string `json:"address"`
	RemoteAS            string `json:"remoteas"`
	UpdateSource        string `json:"updatesource"`
	NextHopSelf         string `json:"nexthopself"`
	Family              string `json:"family,omitempty"`
	MultiProtocol       string `json:"multiprotocol,omitempty"`
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
		RemoteASMode:               advancedString(ctx, m.RemoteASMode),
		LocalAS:                    advancedInt64(ctx, m.LocalAS),
		Password:                   advancedString(ctx, m.Password),
		Weight:                     advancedInt64(ctx, m.Weight),
		LocalIP:                    advancedString(ctx, m.LocalIP),
		LinkLocalInterface:         advancedString(ctx, m.LinkLocalInterface),
		NextHopSelfAll:             advancedBool(ctx, m.NextHopSelfAll),
		Multihop:                   advancedBool(ctx, m.Multihop),
		RouteReflectorClient:       advancedBool(ctx, m.RouteReflectorClient),
		SoftReconfigurationInbound: advancedBool(ctx, m.SoftReconfigurationInbound),
		BFD:                        advancedBool(ctx, m.BFD),
		BFDStrict:                  advancedBool(ctx, m.BFDStrict),
		ConnectTimer:               advancedInt64(ctx, m.ConnectTimer),
		DefaultOriginate:           advancedBool(ctx, m.DefaultOriginate),
		ASOverride:                 advancedBool(ctx, m.ASOverride),
		RemovePrivateAS:            advancedString(ctx, m.RemovePrivateAS),
		AllowASIn:                  advancedString(ctx, m.AllowASIn),
		DisableConnectedCheck:      advancedBool(ctx, m.DisableConnectedCheck),
		AttributeUnchanged:         advancedString(ctx, m.AttributeUnchanged),
		Capabilities:               advancedSet(ctx, m.Capabilities),

		PeerGroup:           m.PeerGroup.ValueString(),
		Enabled:             opnsense.BoolToString(m.Enabled.ValueBool()),
		Description:         m.Description.ValueString(),
		Address:             m.Address.ValueString(),
		RemoteAS:            optionalFRRInt(m.RemoteAS),
		UpdateSource:        m.UpdateSource.ValueString(),
		NextHopSelf:         opnsense.BoolToString(m.NextHopSelf.ValueBool()),
		MultiProtocol:       opnsense.BoolToString(m.MultiProtocol.ValueBool()),
		Family:              setToCSV(ctx, m.Family),
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
	m.RemoteASMode = readAdvancedString(a.hasField("remote_as_mode"), string(a.RemoteASMode))
	m.LocalAS = readAdvancedInt64(a.hasField("localas"), a.LocalAS)
	m.Weight = readAdvancedInt64(a.hasField("weight"), a.Weight)
	m.LocalIP = readAdvancedString(a.hasField("localip"), a.LocalIP)
	m.LinkLocalInterface = readAdvancedString(a.hasField("linklocalinterface"), string(a.LinkLocalInterface))
	m.NextHopSelfAll = readAdvancedBool(a.hasField("nexthopselfall"), a.NextHopSelfAll)
	m.Multihop = readAdvancedBool(a.hasField("multihop"), a.Multihop)
	m.RouteReflectorClient = readAdvancedBool(a.hasField("rrclient"), a.RouteReflectorClient)
	m.SoftReconfigurationInbound = readAdvancedBool(a.hasField("soft_reconfiguration_inbound"), a.SoftReconfigurationInbound)
	m.BFD = readAdvancedBool(a.hasField("bfd"), a.BFD)
	m.BFDStrict = readAdvancedBool(a.hasField("bfd_strict"), a.BFDStrict)
	m.ConnectTimer = readAdvancedInt64(a.hasField("connecttimer"), a.ConnectTimer)
	m.DefaultOriginate = readAdvancedBool(a.hasField("defaultoriginate"), a.DefaultOriginate)
	m.ASOverride = readAdvancedBool(a.hasField("asoverride"), a.ASOverride)
	m.RemovePrivateAS = readAdvancedString(a.hasField("removeprivateas"), string(a.RemovePrivateAS))
	m.AllowASIn = readAdvancedString(a.hasField("allowas_in"), string(a.AllowASIn))
	m.DisableConnectedCheck = readAdvancedBool(a.hasField("disable_connected_check"), a.DisableConnectedCheck)
	m.AttributeUnchanged = readAdvancedString(a.hasField("attributeunchanged"), string(a.AttributeUnchanged))
	m.Capabilities = readAdvancedSet(a.hasField("capabilities"), a.Capabilities)

	m.PeerGroup = types.StringValue(string(a.PeerGroup))
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Description = types.StringValue(a.Description)
	m.Address = types.StringValue(a.Address)
	m.UpdateSource = types.StringValue(string(a.UpdateSource))
	m.NextHopSelf = types.BoolValue(opnsense.StringToBool(a.NextHopSelf))
	m.MultiProtocol = types.BoolValue(opnsense.StringToBool(a.MultiProtocol))
	if len(a.Family) > 0 {
		m.Family = sliceToSet(a.Family)
		m.MultiProtocol = types.BoolValue(len(a.Family) == 2)
	} else if families, err := legacyFamilies(a.Address, m.MultiProtocol); err == nil {
		m.Family = families
	}

	m.RemoteAS = types.Int64Null()
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

	fromSML := func(sml frrSelectedMapList) types.Set {
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
