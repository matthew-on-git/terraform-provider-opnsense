// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ServerResourceModel is the Terraform state model for opnsense_haproxy_server.
type ServerResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Address     types.String `tfsdk:"address"`
	Port        types.Int64  `tfsdk:"port"`
	Weight      types.Int64  `tfsdk:"weight"`
	Mode        types.String `tfsdk:"mode"`
	SSL         types.Bool   `tfsdk:"ssl"`
	SSLVerify   types.Bool   `tfsdk:"ssl_verify"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

// serverAPIResponse is the struct for unmarshaling OPNsense GET responses.
// Uses SelectedMap for enum fields.
type serverAPIResponse struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Address     string               `json:"address"`
	Port        string               `json:"port"`
	Weight      string               `json:"weight"`
	Mode        opnsense.SelectedMap `json:"mode"`
	SSL         string               `json:"ssl"`
	SSLVerify   string               `json:"sslVerify"`
	Enabled     string               `json:"enabled"`
}

// serverAPIRequest is the struct for marshaling OPNsense POST requests.
// Uses plain strings since the API accepts simple values for mutations.
type serverAPIRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Port        string `json:"port"`
	Weight      string `json:"weight"`
	Mode        string `json:"mode"`
	SSL         string `json:"ssl"`
	SSLVerify   string `json:"sslVerify"`
	Enabled     string `json:"enabled"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *ServerResourceModel) toAPI(_ context.Context) *serverAPIRequest {
	var weightStr string
	if !m.Weight.IsNull() && !m.Weight.IsUnknown() {
		weightStr = opnsense.Int64ToString(m.Weight.ValueInt64())
	}

	return &serverAPIRequest{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Address:     m.Address.ValueString(),
		Port:        opnsense.Int64ToString(m.Port.ValueInt64()),
		Weight:      weightStr,
		Mode:        m.Mode.ValueString(),
		SSL:         opnsense.BoolToString(m.SSL.ValueBool()),
		SSLVerify:   opnsense.BoolToString(m.SSLVerify.ValueBool()),
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
	}
}

// fromAPI populates the Terraform model from an API response struct.
// The UUID is passed separately since it comes from the Add response, not the model.
func (m *ServerResourceModel) fromAPI(_ context.Context, a *serverAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.Address = types.StringValue(a.Address)
	m.Mode = types.StringValue(string(a.Mode))
	m.SSL = types.BoolValue(opnsense.StringToBool(a.SSL))
	m.SSLVerify = types.BoolValue(opnsense.StringToBool(a.SSLVerify))
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))

	// Port (required — always has a value from API).
	if a.Port != "" {
		portVal, err := opnsense.StringToInt64(a.Port)
		if err == nil {
			m.Port = types.Int64Value(portVal)
		}
	}

	// Weight (optional — may be empty string from API).
	if a.Weight != "" {
		weightVal, err := opnsense.StringToInt64(a.Weight)
		if err == nil {
			m.Weight = types.Int64Value(weightVal)
		} else {
			m.Weight = types.Int64Null()
		}
	} else {
		m.Weight = types.Int64Null()
	}
}
