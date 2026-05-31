// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// GeneralResourceModel is the Terraform state model for opnsense_quagga_general
// (FRR service general settings — a singleton).
type GeneralResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Profile       types.String `tfsdk:"profile"`
	EnableCARP    types.Bool   `tfsdk:"enable_carp"`
	EnableSyslog  types.Bool   `tfsdk:"enable_syslog"`
	EnableSNMP    types.Bool   `tfsdk:"enable_snmp"`
	SyslogLevel   types.String `tfsdk:"syslog_level"`
	FirewallRules types.Bool   `tfsdk:"firewall_rules"`
}

type generalAPIResponse struct {
	Enabled       string               `json:"enabled"`
	Profile       opnsense.SelectedMap `json:"profile"`
	EnableCARP    string               `json:"enablecarp"`
	EnableSyslog  string               `json:"enablesyslog"`
	EnableSNMP    string               `json:"enablesnmp"`
	SyslogLevel   opnsense.SelectedMap `json:"sysloglevel"`
	FirewallRules string               `json:"fwrules"`
}

type generalAPIRequest struct {
	Enabled       string `json:"enabled"`
	Profile       string `json:"profile"`
	EnableCARP    string `json:"enablecarp"`
	EnableSyslog  string `json:"enablesyslog"`
	EnableSNMP    string `json:"enablesnmp"`
	SyslogLevel   string `json:"sysloglevel"`
	FirewallRules string `json:"fwrules"`
}

func (m *GeneralResourceModel) toAPI(_ context.Context) *generalAPIRequest {
	return &generalAPIRequest{
		Enabled:       opnsense.BoolToString(m.Enabled.ValueBool()),
		Profile:       m.Profile.ValueString(),
		EnableCARP:    opnsense.BoolToString(m.EnableCARP.ValueBool()),
		EnableSyslog:  opnsense.BoolToString(m.EnableSyslog.ValueBool()),
		EnableSNMP:    opnsense.BoolToString(m.EnableSNMP.ValueBool()),
		SyslogLevel:   m.SyslogLevel.ValueString(),
		FirewallRules: opnsense.BoolToString(m.FirewallRules.ValueBool()),
	}
}

func (m *GeneralResourceModel) fromAPI(_ context.Context, a *generalAPIResponse, id string) {
	m.ID = types.StringValue(id)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Profile = types.StringValue(string(a.Profile))
	m.EnableCARP = types.BoolValue(opnsense.StringToBool(a.EnableCARP))
	m.EnableSyslog = types.BoolValue(opnsense.StringToBool(a.EnableSyslog))
	m.EnableSNMP = types.BoolValue(opnsense.StringToBool(a.EnableSNMP))
	m.SyslogLevel = types.StringValue(string(a.SyslogLevel))
	m.FirewallRules = types.BoolValue(opnsense.StringToBool(a.FirewallRules))
}
