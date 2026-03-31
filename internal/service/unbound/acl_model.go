// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ACLResourceModel is the Terraform state model for opnsense_unbound_acl.
type ACLResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Name        types.String `tfsdk:"name"`
	Action      types.String `tfsdk:"action"`
	Networks    types.String `tfsdk:"networks"`
	Description types.String `tfsdk:"description"`
}

// unboundACLAPIResponse is the struct for unmarshaling OPNsense GET responses.
type unboundACLAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Name        string               `json:"name"`
	Action      opnsense.SelectedMap `json:"action"`
	Networks    string               `json:"networks"`
	Description string               `json:"description"`
}

// unboundACLAPIRequest is the struct for marshaling OPNsense POST requests.
type unboundACLAPIRequest struct {
	Enabled     string `json:"enabled"`
	Name        string `json:"name"`
	Action      string `json:"action"`
	Networks    string `json:"networks"`
	Description string `json:"description"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *ACLResourceModel) toAPI(_ context.Context) *unboundACLAPIRequest {
	return &unboundACLAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Name:        m.Name.ValueString(),
		Action:      m.Action.ValueString(),
		Networks:    m.Networks.ValueString(),
		Description: m.Description.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *ACLResourceModel) fromAPI(_ context.Context, a *unboundACLAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.Action = types.StringValue(string(a.Action))
	m.Networks = types.StringValue(a.Networks)
	m.Description = types.StringValue(a.Description)
}
