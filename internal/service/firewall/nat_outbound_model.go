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

// NatOutboundResourceModel is the Terraform state model for opnsense_firewall_nat_outbound.
type NatOutboundResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Sequence        types.Int64  `tfsdk:"sequence"`
	Interface       types.String `tfsdk:"interface"`
	IPProtocol      types.String `tfsdk:"ip_protocol"`
	Protocol        types.String `tfsdk:"protocol"`
	SourceNet       types.String `tfsdk:"source_net"`
	SourceNot       types.Bool   `tfsdk:"source_not"`
	SourcePort      types.String `tfsdk:"source_port"`
	DestinationNet  types.String `tfsdk:"destination_net"`
	DestinationNot  types.Bool   `tfsdk:"destination_not"`
	DestinationPort types.String `tfsdk:"destination_port"`
	Target          types.String `tfsdk:"target"`
	TargetPort      types.String `tfsdk:"target_port"`
	NoNat           types.Bool   `tfsdk:"no_nat"`
	StaticNatPort   types.Bool   `tfsdk:"static_nat_port"`
	Log             types.Bool   `tfsdk:"log"`
	Description     types.String `tfsdk:"description"`
	Categories      types.Set    `tfsdk:"categories"`
}

// natOutboundAPIResponse is the struct for unmarshaling OPNsense GET responses.
// Source NAT uses flat field names (not dot-separated like DNat).
type natOutboundAPIResponse struct {
	Enabled         string                   `json:"enabled"`
	Sequence        string                   `json:"sequence"`
	Interface       opnsense.SelectedMap     `json:"interface"`
	IPProtocol      opnsense.SelectedMap     `json:"ipprotocol"`
	Protocol        opnsense.SelectedMap     `json:"protocol"`
	SourceNet       string                   `json:"source_net"`
	SourceNot       string                   `json:"source_not"`
	SourcePort      string                   `json:"source_port"`
	DestinationNet  string                   `json:"destination_net"`
	DestinationNot  string                   `json:"destination_not"`
	DestinationPort string                   `json:"destination_port"`
	Target          string                   `json:"target"`
	TargetPort      string                   `json:"target_port"`
	NoNat           string                   `json:"nonat"`
	StaticNatPort   string                   `json:"staticnatport"`
	Log             string                   `json:"log"`
	Description     string                   `json:"description"`
	Categories      opnsense.SelectedMapList `json:"categories"`
}

// natOutboundAPIRequest is the struct for marshaling OPNsense POST requests.
type natOutboundAPIRequest struct {
	Enabled         string `json:"enabled"`
	Sequence        string `json:"sequence"`
	Interface       string `json:"interface"`
	IPProtocol      string `json:"ipprotocol"`
	Protocol        string `json:"protocol"`
	SourceNet       string `json:"source_net"`
	SourceNot       string `json:"source_not"`
	SourcePort      string `json:"source_port"`
	DestinationNet  string `json:"destination_net"`
	DestinationNot  string `json:"destination_not"`
	DestinationPort string `json:"destination_port"`
	Target          string `json:"target"`
	TargetPort      string `json:"target_port"`
	NoNat           string `json:"nonat"`
	StaticNatPort   string `json:"staticnatport"`
	Log             string `json:"log"`
	Description     string `json:"description"`
	Categories      string `json:"categories"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *NatOutboundResourceModel) toAPI(ctx context.Context) *natOutboundAPIRequest {
	var categoriesStr string
	if !m.Categories.IsNull() && !m.Categories.IsUnknown() {
		var elements []string
		m.Categories.ElementsAs(ctx, &elements, false)
		categoriesStr = strings.Join(elements, ",")
	}

	return &natOutboundAPIRequest{
		Enabled:         opnsense.BoolToString(m.Enabled.ValueBool()),
		Sequence:        opnsense.Int64ToString(m.Sequence.ValueInt64()),
		Interface:       m.Interface.ValueString(),
		IPProtocol:      m.IPProtocol.ValueString(),
		Protocol:        m.Protocol.ValueString(),
		SourceNet:       m.SourceNet.ValueString(),
		SourceNot:       opnsense.BoolToString(m.SourceNot.ValueBool()),
		SourcePort:      m.SourcePort.ValueString(),
		DestinationNet:  m.DestinationNet.ValueString(),
		DestinationNot:  opnsense.BoolToString(m.DestinationNot.ValueBool()),
		DestinationPort: m.DestinationPort.ValueString(),
		Target:          m.Target.ValueString(),
		TargetPort:      m.TargetPort.ValueString(),
		NoNat:           opnsense.BoolToString(m.NoNat.ValueBool()),
		StaticNatPort:   opnsense.BoolToString(m.StaticNatPort.ValueBool()),
		Log:             opnsense.BoolToString(m.Log.ValueBool()),
		Description:     m.Description.ValueString(),
		Categories:      categoriesStr,
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *NatOutboundResourceModel) fromAPI(_ context.Context, a *natOutboundAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Interface = types.StringValue(string(a.Interface))
	m.IPProtocol = types.StringValue(string(a.IPProtocol))
	m.Protocol = types.StringValue(string(a.Protocol))
	m.SourceNet = types.StringValue(a.SourceNet)
	m.SourceNot = types.BoolValue(opnsense.StringToBool(a.SourceNot))
	m.SourcePort = types.StringValue(a.SourcePort)
	m.DestinationNet = types.StringValue(a.DestinationNet)
	m.DestinationNot = types.BoolValue(opnsense.StringToBool(a.DestinationNot))
	m.DestinationPort = types.StringValue(a.DestinationPort)
	m.Target = types.StringValue(a.Target)
	m.TargetPort = types.StringValue(a.TargetPort)
	m.NoNat = types.BoolValue(opnsense.StringToBool(a.NoNat))
	m.StaticNatPort = types.BoolValue(opnsense.StringToBool(a.StaticNatPort))
	m.Log = types.BoolValue(opnsense.StringToBool(a.Log))
	m.Description = types.StringValue(a.Description)

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
