// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// SettingsResourceModel is the Terraform state model for opnsense_ddclient_settings.
type SettingsResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Backend   types.String `tfsdk:"backend"`
	Interval  types.Int64  `tfsdk:"interval"`
	Verbose   types.Bool   `tfsdk:"verbose"`
	AllowIPv6 types.Bool   `tfsdk:"allow_ipv6"`
}

type settingsAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Backend     opnsense.SelectedMap `json:"backend"`
	DaemonDelay string               `json:"daemon_delay"`
	Verbose     string               `json:"verbose"`
	AllowIPv6   string               `json:"allowipv6"`
}

type settingsAPIRequest struct {
	Enabled     string `json:"enabled"`
	Backend     string `json:"backend"`
	DaemonDelay string `json:"daemon_delay"`
	Verbose     string `json:"verbose"`
	AllowIPv6   string `json:"allowipv6"`
}

func (m *SettingsResourceModel) toAPI(_ context.Context) *settingsAPIRequest {
	return &settingsAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Backend:     m.Backend.ValueString(),
		DaemonDelay: strconv.FormatInt(m.Interval.ValueInt64(), 10),
		Verbose:     opnsense.BoolToString(m.Verbose.ValueBool()),
		AllowIPv6:   opnsense.BoolToString(m.AllowIPv6.ValueBool()),
	}
}

func (m *SettingsResourceModel) fromAPI(_ context.Context, api *settingsAPIResponse, id string) {
	m.ID = types.StringValue(id)
	m.Enabled = types.BoolValue(opnsense.StringToBool(api.Enabled))
	m.Backend = types.StringValue(string(api.Backend))
	interval, err := strconv.ParseInt(api.DaemonDelay, 10, 64)
	if err != nil {
		interval = 0
	}
	m.Interval = types.Int64Value(interval)
	m.Verbose = types.BoolValue(opnsense.StringToBool(api.Verbose))
	m.AllowIPv6 = types.BoolValue(opnsense.StringToBool(api.AllowIPv6))
}
