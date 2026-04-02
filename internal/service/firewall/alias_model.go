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

// AliasResourceModel is the Terraform state model for opnsense_firewall_alias.
type AliasResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Content     types.Set    `tfsdk:"content"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Proto       types.String `tfsdk:"proto"`
	Categories  types.Set    `tfsdk:"categories"`
	UpdateFreq  types.String `tfsdk:"update_freq"`
}

// aliasAPIResponse is the struct for unmarshaling OPNsense GET responses.
// SelectedMap/SelectedMapList handle the OPNsense enum response format.
type aliasAPIResponse struct {
	Name        string                   `json:"name"`
	Type        opnsense.SelectedMap     `json:"type"`
	Content     opnsense.SelectedMapList `json:"content"`
	Description string                   `json:"description"`
	Enabled     string                   `json:"enabled"`
	Proto       opnsense.SelectedMap     `json:"proto"`
	Categories  opnsense.SelectedMapList `json:"categories"`
	UpdateFreq  string                   `json:"updatefreq"`
}

// aliasAPIRequest is the struct for marshaling OPNsense POST requests.
// Uses plain strings since the API accepts simple values for mutations.
type aliasAPIRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Enabled     string `json:"enabled"`
	Proto       string `json:"proto"`
	Categories  string `json:"categories"`
	UpdateFreq  string `json:"updatefreq"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *AliasResourceModel) toAPI(ctx context.Context) *aliasAPIRequest {
	var contentStr string
	if !m.Content.IsNull() && !m.Content.IsUnknown() {
		var elements []string
		m.Content.ElementsAs(ctx, &elements, false)
		contentStr = strings.Join(elements, "\n")
	}

	var categoriesStr string
	if !m.Categories.IsNull() && !m.Categories.IsUnknown() {
		var elements []string
		m.Categories.ElementsAs(ctx, &elements, false)
		categoriesStr = strings.Join(elements, ",")
	}

	return &aliasAPIRequest{
		Name:        m.Name.ValueString(),
		Type:        m.Type.ValueString(),
		Content:     contentStr,
		Description: m.Description.ValueString(),
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Proto:       m.Proto.ValueString(),
		Categories:  categoriesStr,
		UpdateFreq:  m.UpdateFreq.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
// The UUID is passed separately since it comes from the Add response, not the model.
func (m *AliasResourceModel) fromAPI(_ context.Context, a *aliasAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Name = types.StringValue(a.Name)
	m.Type = types.StringValue(string(a.Type))
	m.Description = types.StringValue(a.Description)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Proto = types.StringValue(string(a.Proto))
	m.UpdateFreq = types.StringValue(a.UpdateFreq)

	// Convert SelectedMapList content to types.Set.
	if len(a.Content) == 0 {
		m.Content = types.SetValueMust(types.StringType, []attr.Value{})
	} else {
		vals := make([]attr.Value, len(a.Content))
		for i, v := range a.Content {
			vals[i] = types.StringValue(v)
		}
		m.Content = types.SetValueMust(types.StringType, vals)
	}

	// Convert SelectedMapList to types.Set.
	if len(a.Categories) == 0 {
		m.Categories = types.SetValueMust(types.StringType, []attr.Value{})
	} else {
		catValues := make([]attr.Value, len(a.Categories))
		for i, c := range a.Categories {
			catValues[i] = types.StringValue(c)
		}
		m.Categories = types.SetValueMust(types.StringType, catValues)
	}
}
