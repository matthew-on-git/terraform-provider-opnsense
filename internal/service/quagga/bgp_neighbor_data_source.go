// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &bgpNeighborDataSource{}

type bgpNeighborDataSource struct{ client *opnsense.Client }

func newBGPNeighborDataSource() datasource.DataSource { return &bgpNeighborDataSource{} }

func (d *bgpNeighborDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_bgp_neighbor"
}

func (d *bgpNeighborDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing BGP neighbor on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"remote_as_mode":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source-supported remote_as_mode."},
			"local_as":                     dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Source-supported local_as."},
			"weight":                       dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Source-supported weight."},
			"local_ip":                     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source-supported local_ip."},
			"link_local_interface":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source-supported link_local_interface."},
			"next_hop_self_all":            dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Source-supported next_hop_self_all."},
			"multihop":                     dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Source-supported multihop."},
			"route_reflector_client":       dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Source-supported route_reflector_client."},
			"soft_reconfiguration_inbound": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Source-supported soft_reconfiguration_inbound."},
			"bfd":                          dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Source-supported bfd."},
			"bfd_strict":                   dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Source-supported bfd_strict."},
			"connect_timer":                dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Source-supported connect_timer."},
			"default_originate":            dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Source-supported default_originate."},
			"as_override":                  dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Source-supported as_override."},
			"remove_private_as":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source-supported remove_private_as."},
			"allow_as_in":                  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source-supported allow_as_in."},
			"disable_connected_check":      dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Source-supported disable_connected_check."},
			"attribute_unchanged":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source-supported attribute_unchanged."},
			"capabilities":                 dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Source-supported capabilities."},
			"id":                           dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
			"enabled":                      dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this BGP neighbor is enabled."},
			"description":                  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the BGP neighbor."},
			"address":                      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Neighbor IP address."},
			"remote_as":                    dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Remote autonomous system number."},
			"update_source":                dsschema.StringAttribute{Computed: true, MarkdownDescription: "Update source interface."},
			"next_hop_self":                dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether next-hop-self is enabled."},
			"peer_group":                   dsschema.StringAttribute{Computed: true, MarkdownDescription: "Linked BGP peer-group UUID."},
			"family":                       dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Activated lowercase address families."},
			"multi_protocol":               dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether multi-protocol support is enabled."},
			"keepalive":                    dsschema.Int64Attribute{Computed: true, MarkdownDescription: "BGP keepalive interval."},
			"holddown":                     dsschema.Int64Attribute{Computed: true, MarkdownDescription: "BGP hold time."},
			"linked_prefixlist_in":         dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Inbound linked prefix lists."},
			"linked_prefixlist_out":        dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Outbound linked prefix lists."},
			"linked_routemap_in":           dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Inbound linked route maps."},
			"linked_routemap_out":          dsschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Outbound linked route maps."},
		},
	}
}

func (d *bgpNeighborDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	d.client = client
}

func (d *bgpNeighborDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config bgpNeighborLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := bgpOwnerMu.Acquire(ctx, 1); err != nil {
		resp.Diagnostics.AddError("FRR operation canceled while waiting", frrDiagnostic(err))
		return
	}
	defer bgpOwnerMu.Release(1)
	capability, err := discoverFRR(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError("FRR compatibility check failed", frrDiagnostic(err))
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[bgpNeighborAPIResponse](ctx, d.client, strictBGPOpts(bgpNeighborReqOpts), id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading BGP neighbor", fmt.Sprintf("Could not read BGP neighbor %s: %s", id, frrDiagnostic(err)))
		return
	}
	if err := validateBGPResponse(result, capability); err != nil {
		resp.Diagnostics.AddError("Invalid FRR read-back", frrDiagnostic(err))
		return
	}
	var current BGPNeighborResourceModel
	current.fromAPI(ctx, result, id)
	var prior BGPNeighborResourceModel
	if err := normalizeBGPRead(ctx, d.client, capability, &prior, &current); err != nil {
		resp.Diagnostics.AddError("Invalid FRR relation read-back", frrDiagnostic(err))
		return
	}
	config.fromResource(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

type bgpNeighborLookupModel struct {
	RemoteASMode               types.String `tfsdk:"remote_as_mode"`
	LocalAS                    types.Int64  `tfsdk:"local_as"`
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

func (m *bgpNeighborLookupModel) fromResource(r BGPNeighborResourceModel) {
	m.RemoteASMode = r.RemoteASMode
	m.LocalAS = r.LocalAS
	m.Weight = r.Weight
	m.LocalIP = r.LocalIP
	m.LinkLocalInterface = r.LinkLocalInterface
	m.NextHopSelfAll = r.NextHopSelfAll
	m.Multihop = r.Multihop
	m.RouteReflectorClient = r.RouteReflectorClient
	m.SoftReconfigurationInbound = r.SoftReconfigurationInbound
	m.BFD = r.BFD
	m.BFDStrict = r.BFDStrict
	m.ConnectTimer = r.ConnectTimer
	m.DefaultOriginate = r.DefaultOriginate
	m.ASOverride = r.ASOverride
	m.RemovePrivateAS = r.RemovePrivateAS
	m.AllowASIn = r.AllowASIn
	m.DisableConnectedCheck = r.DisableConnectedCheck
	m.AttributeUnchanged = r.AttributeUnchanged
	m.Capabilities = r.Capabilities
	m.PeerGroup = r.PeerGroup
	m.ID = r.ID
	m.Enabled = r.Enabled
	m.Description = r.Description
	m.Address = r.Address
	m.RemoteAS = r.RemoteAS
	m.UpdateSource = r.UpdateSource
	m.NextHopSelf = r.NextHopSelf
	m.Family = r.Family
	m.MultiProtocol = r.MultiProtocol
	m.Keepalive = r.Keepalive
	m.Holddown = r.Holddown
	m.LinkedPrefixlistIn = r.LinkedPrefixlistIn
	m.LinkedPrefixlistOut = r.LinkedPrefixlistOut
	m.LinkedRoutemapIn = r.LinkedRoutemapIn
	m.LinkedRoutemapOut = r.LinkedRoutemapOut
}
