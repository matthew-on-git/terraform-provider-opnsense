// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// HostOverrideResourceModel is the Terraform state model for opnsense_unbound_host_override.
type HostOverrideResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Hostname    types.String `tfsdk:"hostname"`
	Domain      types.String `tfsdk:"domain"`
	RR          types.String `tfsdk:"rr"`
	Server      types.String `tfsdk:"server"`
	Description types.String `tfsdk:"description"`
}

// hostOverrideAPIResponse is the struct for unmarshaling OPNsense GET responses.
type hostOverrideAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Hostname    string               `json:"hostname"`
	Domain      string               `json:"domain"`
	RR          opnsense.SelectedMap `json:"rr"`
	Server      string               `json:"server"`
	Description string               `json:"description"`
}

// hostOverrideAPIRequest is the struct for marshaling OPNsense POST requests.
type hostOverrideAPIRequest struct {
	Enabled     string `json:"enabled"`
	Hostname    string `json:"hostname"`
	Domain      string `json:"domain"`
	RR          string `json:"rr"`
	Server      string `json:"server"`
	Description string `json:"description"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *HostOverrideResourceModel) toAPI(_ context.Context) *hostOverrideAPIRequest {
	return &hostOverrideAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Hostname:    m.Hostname.ValueString(),
		Domain:      m.Domain.ValueString(),
		RR:          m.RR.ValueString(),
		Server:      m.Server.ValueString(),
		Description: m.Description.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *HostOverrideResourceModel) fromAPI(_ context.Context, a *hostOverrideAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Hostname = types.StringValue(a.Hostname)
	m.Domain = types.StringValue(a.Domain)
	m.RR = types.StringValue(string(a.RR))
	m.Server = types.StringValue(a.Server)
	m.Description = types.StringValue(a.Description)
}
