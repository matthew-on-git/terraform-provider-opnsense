// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// NatPortForwardResourceModel is the Terraform state model for opnsense_firewall_nat_port_forward.
type NatPortForwardResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Sequence        types.Int64  `tfsdk:"sequence"`
	Interface       types.String `tfsdk:"interface"`
	IPProtocol      types.String `tfsdk:"ip_protocol"`
	Protocol        types.String `tfsdk:"protocol"`
	SourceNet       types.String `tfsdk:"source_net"`
	SourcePort      types.String `tfsdk:"source_port"`
	SourceNot       types.Bool   `tfsdk:"source_not"`
	DestinationNet  types.String `tfsdk:"destination_net"`
	DestinationPort types.String `tfsdk:"destination_port"`
	DestinationNot  types.Bool   `tfsdk:"destination_not"`
	Target          types.String `tfsdk:"target"`
	LocalPort       types.String `tfsdk:"local_port"`
	Log             types.Bool   `tfsdk:"log"`
	Description     types.String `tfsdk:"description"`
	Reflection      types.String `tfsdk:"reflection"`
	Categories      types.Set    `tfsdk:"categories"`
}

// natPortForwardAPIResponse is the struct for unmarshaling OPNsense GET responses.
// source/destination come back as nested objects ({"source": {"network": ...}}),
// matching the DNat model's field structure.
type natPortForwardAPIResponse struct {
	Disabled    string                   `json:"disabled"`
	Sequence    string                   `json:"sequence"`
	Interface   opnsense.SelectedMap     `json:"interface"`
	IPProtocol  opnsense.SelectedMap     `json:"ipprotocol"`
	Protocol    opnsense.SelectedMap     `json:"protocol"`
	Source      natDNatEndpoint          `json:"source"`
	Destination natDNatEndpoint          `json:"destination"`
	Target      string                   `json:"target"`
	LocalPort   string                   `json:"local-port"`
	Log         string                   `json:"log"`
	Description string                   `json:"descr"`
	Reflection  opnsense.SelectedMap     `json:"natreflection"`
	Categories  opnsense.SelectedMapList `json:"categories"`
}

// natPortForwardAPIRequest is the struct for marshaling OPNsense POST requests.
type natPortForwardAPIRequest struct {
	Disabled string `json:"disabled"`
	// Sequence is omitted when unset so OPNsense assigns the next slot
	// (max+100, append-at-end). A static default of 1 would silently slot new
	// rules to the top of the DNAT order; sending "0" is rejected (min 1).
	Sequence    *string         `json:"sequence,omitempty"`
	Interface   string          `json:"interface"`
	IPProtocol  string          `json:"ipprotocol"`
	Protocol    string          `json:"protocol"`
	Source      natDNatEndpoint `json:"source"`
	Destination natDNatEndpoint `json:"destination"`
	Target      string          `json:"target"`
	LocalPort   string          `json:"local-port"`
	Log         string          `json:"log"`
	Description string          `json:"descr"`
	Reflection  string          `json:"natreflection"`
	Categories  string          `json:"categories"`
}

// natDNatEndpoint is the nested source/destination matcher object the DNat API
// expects. network/port/not are the OPNsense field names for the nested object.
type natDNatEndpoint struct {
	Network string `json:"network"`
	Port    string `json:"port"`
	Not     string `json:"not"`
}

// toAPI converts the Terraform model to an API request struct.
// NOTE: The API uses "disabled" (inverted logic) — we invert the Terraform "enabled" value.
func (m *NatPortForwardResourceModel) toAPI(ctx context.Context) *natPortForwardAPIRequest {
	var categoriesStr string
	if !m.Categories.IsNull() && !m.Categories.IsUnknown() {
		var elements []string
		m.Categories.ElementsAs(ctx, &elements, false)
		categoriesStr = strings.Join(elements, ",")
	}

	// Omit sequence when unset so OPNsense assigns the next slot (max+100,
	// append-at-end). A hardcoded default of 1 would slot new rules to the top
	// of the DNAT order; sending "0" is rejected (min 1).
	var sequence *string
	if !m.Sequence.IsNull() && !m.Sequence.IsUnknown() {
		s := opnsense.Int64ToString(m.Sequence.ValueInt64())
		sequence = &s
	}

	return &natPortForwardAPIRequest{
		Disabled:   opnsense.BoolToString(!m.Enabled.ValueBool()), // Invert: enabled=true → disabled="0"
		Sequence:   sequence,
		Interface:  m.Interface.ValueString(),
		IPProtocol: m.IPProtocol.ValueString(),
		Protocol:   m.Protocol.ValueString(),
		Source: natDNatEndpoint{
			Network: m.SourceNet.ValueString(),
			Port:    m.SourcePort.ValueString(),
			Not:     opnsense.BoolToString(m.SourceNot.ValueBool()),
		},
		Destination: natDNatEndpoint{
			Network: m.DestinationNet.ValueString(),
			Port:    m.DestinationPort.ValueString(),
			Not:     opnsense.BoolToString(m.DestinationNot.ValueBool()),
		},
		Target:      m.Target.ValueString(),
		LocalPort:   m.LocalPort.ValueString(),
		Log:         opnsense.BoolToString(m.Log.ValueBool()),
		Description: m.Description.ValueString(),
		Reflection:  m.Reflection.ValueString(),
		Categories:  categoriesStr,
	}
}

// fromAPI populates the Terraform model from an API response struct.
// NOTE: The API uses "disabled" — we invert to Terraform "enabled".
func (m *NatPortForwardResourceModel) fromAPI(_ context.Context, a *natPortForwardAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(!opnsense.StringToBool(a.Disabled)) // Invert: disabled="0" → enabled=true
	m.Interface = types.StringValue(string(a.Interface))
	m.IPProtocol = types.StringValue(string(a.IPProtocol))
	m.Protocol = types.StringValue(string(a.Protocol))
	m.SourceNet = types.StringValue(a.Source.Network)
	m.SourcePort = types.StringValue(a.Source.Port)
	m.SourceNot = types.BoolValue(opnsense.StringToBool(a.Source.Not))
	m.DestinationNet = types.StringValue(a.Destination.Network)
	m.DestinationPort = types.StringValue(a.Destination.Port)
	m.DestinationNot = types.BoolValue(opnsense.StringToBool(a.Destination.Not))
	m.Target = types.StringValue(a.Target)
	m.LocalPort = types.StringValue(a.LocalPort)
	m.Log = types.BoolValue(opnsense.StringToBool(a.Log))
	m.Description = types.StringValue(a.Description)
	m.Reflection = types.StringValue(string(a.Reflection))

	// Sequence (required — always has a value).
	if a.Sequence != "" {
		seqVal, err := opnsense.StringToInt64(a.Sequence)
		if err == nil {
			m.Sequence = types.Int64Value(seqVal)
		}
	}

	// Categories — SelectedMapList → types.Set.
	if len(a.Categories) == 0 {
		m.Categories = types.SetValueMust(types.StringType, []attr.Value{})
	} else {
		vals := make([]attr.Value, len(a.Categories))
		for i, v := range a.Categories {
			vals[i] = types.StringValue(v)
		}
		m.Categories = types.SetValueMust(types.StringType, vals)
	}
}
