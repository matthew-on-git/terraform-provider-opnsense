// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// RouteMapResourceModel is the Terraform state model for opnsense_quagga_route_map.
type RouteMapResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
	Action      types.String `tfsdk:"action"`
	Order       types.Int64  `tfsdk:"order"`
	MatchPrefix types.Set    `tfsdk:"match_prefix"`
	Set         types.String `tfsdk:"set"`
}

type routeMapAPIResponse struct {
	Enabled     string                   `json:"enabled"`
	Description string                   `json:"description"`
	Name        string                   `json:"name"`
	Action      opnsense.SelectedMap     `json:"action"`
	Order       string                   `json:"id"`
	MatchPrefix opnsense.SelectedMapList `json:"match2"`
	Set         string                   `json:"set"`
}

type routeMapAPIRequest struct {
	Enabled     string `json:"enabled"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Action      string `json:"action"`
	Order       string `json:"id"`
	MatchPrefix string `json:"match2"`
	Set         string `json:"set"`
}

func (m *RouteMapResourceModel) toAPI(ctx context.Context) *routeMapAPIRequest {
	var matchStr string
	if !m.MatchPrefix.IsNull() && !m.MatchPrefix.IsUnknown() {
		var elems []string
		m.MatchPrefix.ElementsAs(ctx, &elems, false)
		matchStr = strings.Join(elems, ",")
	}

	return &routeMapAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Description: m.Description.ValueString(),
		Name:        m.Name.ValueString(),
		Action:      m.Action.ValueString(),
		Order:       opnsense.Int64ToString(m.Order.ValueInt64()),
		MatchPrefix: matchStr,
		Set:         m.Set.ValueString(),
	}
}

func (m *RouteMapResourceModel) fromAPI(_ context.Context, a *routeMapAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Description = types.StringValue(a.Description)
	m.Name = types.StringValue(a.Name)
	m.Action = types.StringValue(string(a.Action))
	m.Set = types.StringValue(a.Set)

	if a.Order != "" {
		if v, err := opnsense.StringToInt64(a.Order); err == nil {
			m.Order = types.Int64Value(v)
		}
	}

	if len(a.MatchPrefix) == 0 {
		m.MatchPrefix = types.SetValueMust(types.StringType, []attr.Value{})
	} else {
		vals := make([]attr.Value, len(a.MatchPrefix))
		for i, v := range a.MatchPrefix {
			vals[i] = types.StringValue(v)
		}
		m.MatchPrefix = types.SetValueMust(types.StringType, vals)
	}
}
