// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// HostAliasResourceModel is the Terraform state model for opnsense_unbound_host_alias.
type HostAliasResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Host        types.String `tfsdk:"host"`
	Hostname    types.String `tfsdk:"hostname"`
	Domain      types.String `tfsdk:"domain"`
	Description types.String `tfsdk:"description"`
}

// hostAliasAPIResponse is the struct for unmarshaling OPNsense GET responses.
type hostAliasAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Host        opnsense.SelectedMap `json:"host"`
	Hostname    string               `json:"hostname"`
	Domain      string               `json:"domain"`
	Description string               `json:"description"`
}

// hostAliasAPIRequest is the struct for marshaling OPNsense POST requests.
type hostAliasAPIRequest struct {
	Enabled     string `json:"enabled"`
	Host        string `json:"host"`
	Hostname    string `json:"hostname"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *HostAliasResourceModel) toAPI(_ context.Context) *hostAliasAPIRequest {
	return &hostAliasAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Host:        m.Host.ValueString(),
		Hostname:    m.Hostname.ValueString(),
		Domain:      m.Domain.ValueString(),
		Description: m.Description.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *HostAliasResourceModel) fromAPI(_ context.Context, a *hostAliasAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Host = types.StringValue(string(a.Host))
	m.Hostname = types.StringValue(a.Hostname)
	m.Domain = types.StringValue(a.Domain)
	m.Description = types.StringValue(a.Description)
}
