// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// AliasResourceModel is the Terraform state model for opnsense_firewall_alias.
type AliasResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Content        types.Set    `tfsdk:"content"`
	Description    types.String `tfsdk:"description"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Proto          types.String `tfsdk:"proto"`
	IPProtocol     types.Set    `tfsdk:"ip_protocol"`
	Interface      types.String `tfsdk:"interface"`
	Categories     types.Set    `tfsdk:"categories"`
	UpdateFreq     types.String `tfsdk:"update_freq"`
	PathExpression types.String `tfsdk:"path_expression"`
	Stats          types.Bool   `tfsdk:"stats"`
}

// aliasAPIResponse is the struct for unmarshaling OPNsense GET responses.
// SelectedMap/SelectedMapList handle the OPNsense enum response format.
type aliasAPIResponse struct {
	Name           string                          `json:"name"`
	Type           opnsense.SelectedMap            `json:"type"`
	Content        opnsense.NewlineSelectedMapList `json:"content"`
	Description    string                          `json:"description"`
	Enabled        string                          `json:"enabled"`
	Proto          opnsense.CSVSelectedMapList     `json:"proto"`
	Interface      opnsense.SelectedMap            `json:"interface"`
	Categories     opnsense.CSVSelectedMapList     `json:"categories"`
	UpdateFreq     string                          `json:"updatefreq"`
	PathExpression string                          `json:"path_expression"`
	Counters       string                          `json:"counters"`
}

// aliasAPIRequest is the struct for marshaling OPNsense POST requests.
// Uses plain strings since the API accepts simple values for mutations.
type aliasAPIRequest struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Content        string `json:"content"`
	Description    string `json:"description"`
	Enabled        string `json:"enabled"`
	Proto          string `json:"proto"`
	Interface      string `json:"interface"`
	Categories     string `json:"categories"`
	UpdateFreq     string `json:"updatefreq"`
	PathExpression string `json:"path_expression"`
	Counters       string `json:"counters"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *AliasResourceModel) toAPI(ctx context.Context) *aliasAPIRequest {
	var contentStr string
	if !m.Content.IsNull() && !m.Content.IsUnknown() {
		var elements []string
		m.Content.ElementsAs(ctx, &elements, false)
		sort.Strings(elements)
		contentStr = strings.Join(elements, "\n")
	}

	var categoriesStr string
	if !m.Categories.IsNull() && !m.Categories.IsUnknown() {
		var elements []string
		m.Categories.ElementsAs(ctx, &elements, false)
		sort.Strings(elements)
		categoriesStr = strings.Join(elements, ",")
	}

	var protocols []string
	if !m.IPProtocol.IsNull() && !m.IPProtocol.IsUnknown() {
		m.IPProtocol.ElementsAs(ctx, &protocols, false)
	}
	if m.IPProtocol.IsNull() && !m.Proto.IsNull() && !m.Proto.IsUnknown() && m.Proto.ValueString() != "" {
		protocols = []string{m.Proto.ValueString()}
	}
	sort.Strings(protocols)

	return &aliasAPIRequest{
		Name:           m.Name.ValueString(),
		Type:           m.Type.ValueString(),
		Content:        contentStr,
		Description:    m.Description.ValueString(),
		Enabled:        opnsense.BoolToString(m.Enabled.ValueBool()),
		Proto:          strings.Join(protocols, ","),
		Interface:      m.Interface.ValueString(),
		Categories:     categoriesStr,
		UpdateFreq:     m.UpdateFreq.ValueString(),
		PathExpression: m.PathExpression.ValueString(),
		Counters:       opnsense.BoolToString(m.Stats.ValueBool()),
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
	protocols := []string(a.Proto)
	sort.Strings(protocols)
	m.Proto = types.StringValue(firstString(protocols))
	m.IPProtocol = stringSetValue(protocols)
	m.Interface = types.StringValue(string(a.Interface))
	m.UpdateFreq = types.StringValue(a.UpdateFreq)
	m.PathExpression = types.StringValue(a.PathExpression)
	m.Stats = types.BoolValue(opnsense.StringToBool(a.Counters))

	// OPNsense represents empty alias content with a selected empty key.
	// Filter it only here: empty keys in other selectors may carry meaning.
	vals := make([]attr.Value, 0, len(a.Content))
	for _, v := range a.Content {
		if v != "" {
			vals = append(vals, types.StringValue(v))
		}
	}
	m.Content = types.SetValueMust(types.StringType, vals)

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
