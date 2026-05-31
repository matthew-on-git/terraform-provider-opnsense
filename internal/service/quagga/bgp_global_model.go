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

// BGPGlobalResourceModel is the Terraform state model for opnsense_quagga_bgp_global
// (BGP global configuration — a singleton).
type BGPGlobalResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	ASNumber           types.Int64  `tfsdk:"as_number"`
	RouterID           types.String `tfsdk:"router_id"`
	Distance           types.Int64  `tfsdk:"distance"`
	GracefulRestart    types.Bool   `tfsdk:"graceful_restart"`
	NetworkImportCheck types.Bool   `tfsdk:"network_import_check"`
	EnforceFirstAS     types.Bool   `tfsdk:"enforce_first_as"`
	LogNeighborChanges types.Bool   `tfsdk:"log_neighbor_changes"`
	Networks           types.Set    `tfsdk:"networks"`
	MaximumPaths       types.Int64  `tfsdk:"maximum_paths"`
	MaximumPathsIBGP   types.Int64  `tfsdk:"maximum_paths_ibgp"`
}

type bgpGlobalAPIResponse struct {
	Enabled            string                   `json:"enabled"`
	ASNumber           string                   `json:"asnumber"`
	RouterID           string                   `json:"routerid"`
	Distance           string                   `json:"distance"`
	GracefulRestart    string                   `json:"graceful"`
	NetworkImportCheck string                   `json:"networkimportcheck"`
	EnforceFirstAS     string                   `json:"enforce_first_as"`
	LogNeighborChanges string                   `json:"logneighborchanges"`
	Networks           opnsense.SelectedMapList `json:"networks"`
	MaximumPaths       string                   `json:"maximumpaths"`
	MaximumPathsIBGP   string                   `json:"maximumpathsibgp"`
}

type bgpGlobalAPIRequest struct {
	Enabled            string `json:"enabled"`
	ASNumber           string `json:"asnumber"`
	RouterID           string `json:"routerid"`
	Distance           string `json:"distance"`
	GracefulRestart    string `json:"graceful"`
	NetworkImportCheck string `json:"networkimportcheck"`
	EnforceFirstAS     string `json:"enforce_first_as"`
	LogNeighborChanges string `json:"logneighborchanges"`
	Networks           string `json:"networks"`
	MaximumPaths       string `json:"maximumpaths"`
	MaximumPathsIBGP   string `json:"maximumpathsibgp"`
}

// setToCSV joins a Terraform string set into a comma-separated string for the API.
func setToCSV(ctx context.Context, s types.Set) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	var elems []string
	s.ElementsAs(ctx, &elems, false)
	return strings.Join(elems, ",")
}

// sliceToSet converts a string slice into a Terraform string set (never null).
func sliceToSet(items []string) types.Set {
	if len(items) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	vals := make([]attr.Value, len(items))
	for i, v := range items {
		vals[i] = types.StringValue(v)
	}
	return types.SetValueMust(types.StringType, vals)
}

func intOrEmpty(n int64) string {
	if n == 0 {
		return ""
	}
	return opnsense.Int64ToString(n)
}

func intOrZero(s string) int64 {
	if s == "" {
		return 0
	}
	if v, err := opnsense.StringToInt64(s); err == nil {
		return v
	}
	return 0
}

func (m *BGPGlobalResourceModel) toAPI(ctx context.Context) *bgpGlobalAPIRequest {
	return &bgpGlobalAPIRequest{
		Enabled:            opnsense.BoolToString(m.Enabled.ValueBool()),
		ASNumber:           opnsense.Int64ToString(m.ASNumber.ValueInt64()),
		RouterID:           m.RouterID.ValueString(),
		Distance:           intOrEmpty(m.Distance.ValueInt64()),
		GracefulRestart:    opnsense.BoolToString(m.GracefulRestart.ValueBool()),
		NetworkImportCheck: opnsense.BoolToString(m.NetworkImportCheck.ValueBool()),
		EnforceFirstAS:     opnsense.BoolToString(m.EnforceFirstAS.ValueBool()),
		LogNeighborChanges: opnsense.BoolToString(m.LogNeighborChanges.ValueBool()),
		Networks:           setToCSV(ctx, m.Networks),
		MaximumPaths:       intOrEmpty(m.MaximumPaths.ValueInt64()),
		MaximumPathsIBGP:   intOrEmpty(m.MaximumPathsIBGP.ValueInt64()),
	}
}

func (m *BGPGlobalResourceModel) fromAPI(_ context.Context, a *bgpGlobalAPIResponse, id string) {
	m.ID = types.StringValue(id)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.ASNumber = types.Int64Value(intOrZero(a.ASNumber))
	m.RouterID = types.StringValue(a.RouterID)
	m.Distance = types.Int64Value(intOrZero(a.Distance))
	m.GracefulRestart = types.BoolValue(opnsense.StringToBool(a.GracefulRestart))
	m.NetworkImportCheck = types.BoolValue(opnsense.StringToBool(a.NetworkImportCheck))
	m.EnforceFirstAS = types.BoolValue(opnsense.StringToBool(a.EnforceFirstAS))
	m.LogNeighborChanges = types.BoolValue(opnsense.StringToBool(a.LogNeighborChanges))
	m.Networks = sliceToSet(a.Networks)
	m.MaximumPaths = types.Int64Value(intOrZero(a.MaximumPaths))
	m.MaximumPathsIBGP = types.Int64Value(intOrZero(a.MaximumPathsIBGP))
}
