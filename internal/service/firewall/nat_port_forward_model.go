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
	Categories      types.Set    `tfsdk:"categories"`
}

// natPortForwardAPIResponse is the struct for unmarshaling OPNsense GET responses.
// Uses dot-separated JSON keys matching the OPNsense DNat model's nested field structure.
type natPortForwardAPIResponse struct {
	Disabled      string                   `json:"disabled"`
	Interface     opnsense.SelectedMap     `json:"interface"`
	IPProtocol    opnsense.SelectedMap     `json:"ipprotocol"`
	Protocol      opnsense.SelectedMap     `json:"protocol"`
	SourceNetwork string                   `json:"source.network"`
	SourcePort    string                   `json:"source.port"`
	SourceNot     string                   `json:"source.not"`
	DestNetwork   string                   `json:"destination.network"`
	DestPort      string                   `json:"destination.port"`
	DestNot       string                   `json:"destination.not"`
	Target        string                   `json:"target"`
	LocalPort     string                   `json:"local-port"`
	Log           string                   `json:"log"`
	Description   string                   `json:"descr"`
	Categories    opnsense.SelectedMapList `json:"categories"`
}

// natPortForwardAPIRequest is the struct for marshaling OPNsense POST requests.
type natPortForwardAPIRequest struct {
	Disabled      string `json:"disabled"`
	Interface     string `json:"interface"`
	IPProtocol    string `json:"ipprotocol"`
	Protocol      string `json:"protocol"`
	SourceNetwork string `json:"source.network"`
	SourcePort    string `json:"source.port"`
	SourceNot     string `json:"source.not"`
	DestNetwork   string `json:"destination.network"`
	DestPort      string `json:"destination.port"`
	DestNot       string `json:"destination.not"`
	Target        string `json:"target"`
	LocalPort     string `json:"local-port"`
	Log           string `json:"log"`
	Description   string `json:"descr"`
	Categories    string `json:"categories"`
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

	return &natPortForwardAPIRequest{
		Disabled:      opnsense.BoolToString(!m.Enabled.ValueBool()), // Invert: enabled=true → disabled="0"
		Interface:     m.Interface.ValueString(),
		IPProtocol:    m.IPProtocol.ValueString(),
		Protocol:      m.Protocol.ValueString(),
		SourceNetwork: m.SourceNet.ValueString(),
		SourcePort:    m.SourcePort.ValueString(),
		SourceNot:     opnsense.BoolToString(m.SourceNot.ValueBool()),
		DestNetwork:   m.DestinationNet.ValueString(),
		DestPort:      m.DestinationPort.ValueString(),
		DestNot:       opnsense.BoolToString(m.DestinationNot.ValueBool()),
		Target:        m.Target.ValueString(),
		LocalPort:     m.LocalPort.ValueString(),
		Log:           opnsense.BoolToString(m.Log.ValueBool()),
		Description:   m.Description.ValueString(),
		Categories:    categoriesStr,
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
	m.SourceNet = types.StringValue(a.SourceNetwork)
	m.SourcePort = types.StringValue(a.SourcePort)
	m.SourceNot = types.BoolValue(opnsense.StringToBool(a.SourceNot))
	m.DestinationNet = types.StringValue(a.DestNetwork)
	m.DestinationPort = types.StringValue(a.DestPort)
	m.DestinationNot = types.BoolValue(opnsense.StringToBool(a.DestNot))
	m.Target = types.StringValue(a.Target)
	m.LocalPort = types.StringValue(a.LocalPort)
	m.Log = types.BoolValue(opnsense.StringToBool(a.Log))
	m.Description = types.StringValue(a.Description)

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
