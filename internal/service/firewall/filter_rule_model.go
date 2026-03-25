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

// FilterRuleResourceModel is the Terraform state model for opnsense_firewall_filter_rule.
type FilterRuleResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Sequence        types.Int64  `tfsdk:"sequence"`
	Action          types.String `tfsdk:"action"`
	Quick           types.Bool   `tfsdk:"quick"`
	Interface       types.Set    `tfsdk:"interface"`
	Direction       types.String `tfsdk:"direction"`
	IPProtocol      types.String `tfsdk:"ip_protocol"`
	Protocol        types.String `tfsdk:"protocol"`
	SourceNet       types.String `tfsdk:"source_net"`
	SourcePort      types.String `tfsdk:"source_port"`
	SourceNot       types.Bool   `tfsdk:"source_not"`
	DestinationNet  types.String `tfsdk:"destination_net"`
	DestinationPort types.String `tfsdk:"destination_port"`
	DestinationNot  types.Bool   `tfsdk:"destination_not"`
	Gateway         types.String `tfsdk:"gateway"`
	Log             types.Bool   `tfsdk:"log"`
	Description     types.String `tfsdk:"description"`
	Categories      types.Set    `tfsdk:"categories"`
}

// filterRuleAPIResponse is the struct for unmarshaling OPNsense GET responses.
type filterRuleAPIResponse struct {
	Enabled         string                   `json:"enabled"`
	Sequence        string                   `json:"sequence"`
	Action          opnsense.SelectedMap     `json:"action"`
	Quick           string                   `json:"quick"`
	Interface       opnsense.SelectedMapList `json:"interface"`
	Direction       opnsense.SelectedMap     `json:"direction"`
	IPProtocol      opnsense.SelectedMap     `json:"ipprotocol"`
	Protocol        opnsense.SelectedMap     `json:"protocol"`
	SourceNet       string                   `json:"source_net"`
	SourcePort      string                   `json:"source_port"`
	SourceNot       string                   `json:"source_not"`
	DestinationNet  string                   `json:"destination_net"`
	DestinationPort string                   `json:"destination_port"`
	DestinationNot  string                   `json:"destination_not"`
	Gateway         opnsense.SelectedMap     `json:"gateway"`
	Log             string                   `json:"log"`
	Description     string                   `json:"description"`
	Categories      opnsense.SelectedMapList `json:"categories"`
}

// filterRuleAPIRequest is the struct for marshaling OPNsense POST requests.
type filterRuleAPIRequest struct {
	Enabled         string `json:"enabled"`
	Sequence        string `json:"sequence"`
	Action          string `json:"action"`
	Quick           string `json:"quick"`
	Interface       string `json:"interface"`
	Direction       string `json:"direction"`
	IPProtocol      string `json:"ipprotocol"`
	Protocol        string `json:"protocol"`
	SourceNet       string `json:"source_net"`
	SourcePort      string `json:"source_port"`
	SourceNot       string `json:"source_not"`
	DestinationNet  string `json:"destination_net"`
	DestinationPort string `json:"destination_port"`
	DestinationNot  string `json:"destination_not"`
	Gateway         string `json:"gateway"`
	Log             string `json:"log"`
	Description     string `json:"description"`
	Categories      string `json:"categories"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *FilterRuleResourceModel) toAPI(ctx context.Context) *filterRuleAPIRequest {
	var interfaceStr string
	if !m.Interface.IsNull() && !m.Interface.IsUnknown() {
		var elements []string
		m.Interface.ElementsAs(ctx, &elements, false)
		interfaceStr = strings.Join(elements, ",")
	}

	var categoriesStr string
	if !m.Categories.IsNull() && !m.Categories.IsUnknown() {
		var elements []string
		m.Categories.ElementsAs(ctx, &elements, false)
		categoriesStr = strings.Join(elements, ",")
	}

	return &filterRuleAPIRequest{
		Enabled:         opnsense.BoolToString(m.Enabled.ValueBool()),
		Sequence:        opnsense.Int64ToString(m.Sequence.ValueInt64()),
		Action:          m.Action.ValueString(),
		Quick:           opnsense.BoolToString(m.Quick.ValueBool()),
		Interface:       interfaceStr,
		Direction:       m.Direction.ValueString(),
		IPProtocol:      m.IPProtocol.ValueString(),
		Protocol:        m.Protocol.ValueString(),
		SourceNet:       m.SourceNet.ValueString(),
		SourcePort:      m.SourcePort.ValueString(),
		SourceNot:       opnsense.BoolToString(m.SourceNot.ValueBool()),
		DestinationNet:  m.DestinationNet.ValueString(),
		DestinationPort: m.DestinationPort.ValueString(),
		DestinationNot:  opnsense.BoolToString(m.DestinationNot.ValueBool()),
		Gateway:         m.Gateway.ValueString(),
		Log:             opnsense.BoolToString(m.Log.ValueBool()),
		Description:     m.Description.ValueString(),
		Categories:      categoriesStr,
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *FilterRuleResourceModel) fromAPI(_ context.Context, a *filterRuleAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Action = types.StringValue(string(a.Action))
	m.Quick = types.BoolValue(opnsense.StringToBool(a.Quick))
	m.Direction = types.StringValue(string(a.Direction))
	m.IPProtocol = types.StringValue(string(a.IPProtocol))
	m.Protocol = types.StringValue(string(a.Protocol))
	m.SourceNet = types.StringValue(a.SourceNet)
	m.SourcePort = types.StringValue(a.SourcePort)
	m.SourceNot = types.BoolValue(opnsense.StringToBool(a.SourceNot))
	m.DestinationNet = types.StringValue(a.DestinationNet)
	m.DestinationPort = types.StringValue(a.DestinationPort)
	m.DestinationNot = types.BoolValue(opnsense.StringToBool(a.DestinationNot))
	m.Gateway = types.StringValue(string(a.Gateway))
	m.Log = types.BoolValue(opnsense.StringToBool(a.Log))
	m.Description = types.StringValue(a.Description)

	// Sequence (required — always has a value).
	if a.Sequence != "" {
		seqVal, err := opnsense.StringToInt64(a.Sequence)
		if err == nil {
			m.Sequence = types.Int64Value(seqVal)
		}
	}

	// Interface — SelectedMapList → types.Set.
	if len(a.Interface) == 0 {
		m.Interface = types.SetValueMust(types.StringType, []attr.Value{})
	} else {
		vals := make([]attr.Value, len(a.Interface))
		for i, v := range a.Interface {
			vals[i] = types.StringValue(v)
		}
		m.Interface = types.SetValueMust(types.StringType, vals)
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
