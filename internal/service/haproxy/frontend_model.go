// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// FrontendResourceModel is the Terraform state model for opnsense_haproxy_frontend.
type FrontendResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Bind           types.String `tfsdk:"bind"`
	Mode           types.String `tfsdk:"mode"`
	DefaultBackend types.String `tfsdk:"default_backend"`
	SSLEnabled     types.Bool   `tfsdk:"ssl_enabled"`
	LinkedActions  types.Set    `tfsdk:"linked_actions"`
	ForwardFor     types.Bool   `tfsdk:"forward_for"`
}

// frontendAPIResponse is the struct for unmarshaling OPNsense GET responses.
type frontendAPIResponse struct {
	Enabled        string                   `json:"enabled"`
	Name           string                   `json:"name"`
	Description    string                   `json:"description"`
	Bind           string                   `json:"bind"`
	Mode           opnsense.SelectedMap     `json:"mode"`
	DefaultBackend opnsense.SelectedMap     `json:"defaultBackend"`
	SSLEnabled     string                   `json:"ssl_enabled"`
	LinkedActions  opnsense.SelectedMapList `json:"linkedActions"`
	ForwardFor     string                   `json:"forwardFor"`
}

// frontendAPIRequest is the struct for marshaling OPNsense POST requests.
type frontendAPIRequest struct {
	Enabled        string `json:"enabled"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Bind           string `json:"bind"`
	Mode           string `json:"mode"`
	DefaultBackend string `json:"defaultBackend"`
	SSLEnabled     string `json:"ssl_enabled"`
	LinkedActions  string `json:"linkedActions"`
	ForwardFor     string `json:"forwardFor"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *FrontendResourceModel) toAPI(ctx context.Context) *frontendAPIRequest {
	var actionsStr string
	if !m.LinkedActions.IsNull() && !m.LinkedActions.IsUnknown() {
		var elements []string
		m.LinkedActions.ElementsAs(ctx, &elements, false)
		actionsStr = strings.Join(elements, ",")
	}

	return &frontendAPIRequest{
		Enabled:        opnsense.BoolToString(m.Enabled.ValueBool()),
		Name:           m.Name.ValueString(),
		Description:    m.Description.ValueString(),
		Bind:           m.Bind.ValueString(),
		Mode:           m.Mode.ValueString(),
		DefaultBackend: m.DefaultBackend.ValueString(),
		SSLEnabled:     opnsense.BoolToString(m.SSLEnabled.ValueBool()),
		LinkedActions:  actionsStr,
		ForwardFor:     opnsense.BoolToString(m.ForwardFor.ValueBool()),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *FrontendResourceModel) fromAPI(_ context.Context, a *frontendAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.Bind = types.StringValue(a.Bind)
	m.Mode = types.StringValue(string(a.Mode))
	m.DefaultBackend = types.StringValue(string(a.DefaultBackend))
	m.SSLEnabled = types.BoolValue(opnsense.StringToBool(a.SSLEnabled))
	m.ForwardFor = types.BoolValue(opnsense.StringToBool(a.ForwardFor))

	// LinkedActions — SelectedMapList → types.Set of UUID strings.
	if len(a.LinkedActions) == 0 {
		m.LinkedActions = types.SetValueMust(types.StringType, []attr.Value{})
	} else {
		vals := make([]attr.Value, len(a.LinkedActions))
		for i, v := range a.LinkedActions {
			vals[i] = types.StringValue(v)
		}
		m.LinkedActions = types.SetValueMust(types.StringType, vals)
	}
}
