// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// DomainOverrideResourceModel is the Terraform state model for opnsense_unbound_domain_override.
type DomainOverrideResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Domain      types.String `tfsdk:"domain"`
	Server      types.String `tfsdk:"server"`
	Description types.String `tfsdk:"description"`
}

// domainOverrideAPIResponse is the struct for unmarshaling OPNsense GET responses.
type domainOverrideAPIResponse struct {
	Enabled     string `json:"enabled"`
	Domain      string `json:"domain"`
	Server      string `json:"server"`
	Description string `json:"description"`
}

// domainOverrideAPIRequest is the struct for marshaling OPNsense POST requests.
type domainOverrideAPIRequest struct {
	Enabled     string `json:"enabled"`
	Domain      string `json:"domain"`
	Server      string `json:"server"`
	Description string `json:"description"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *DomainOverrideResourceModel) toAPI(_ context.Context) *domainOverrideAPIRequest {
	return &domainOverrideAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Domain:      m.Domain.ValueString(),
		Server:      m.Server.ValueString(),
		Description: m.Description.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *DomainOverrideResourceModel) fromAPI(_ context.Context, a *domainOverrideAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Domain = types.StringValue(a.Domain)
	m.Server = types.StringValue(a.Server)
	m.Description = types.StringValue(a.Description)
}
