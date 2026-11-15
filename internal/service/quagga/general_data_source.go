// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package quagga

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// generalLookupFields is the public GET allowlist. It is the resource refresh
// contract intersected with fields the stable FRR general node returns.
// Released os-frr 26.7.8 and 26.7.9 also return volatile daemons and
// manual_config from the whole-model GET. Those are not lookup state. The
// general node has no password or write-only secret. A missing required field
// is an error, not a synthetic model default. A disabled flag stays false.
var generalLookupFields = []string{
	"id",
	"enabled",
	"profile",
	"enable_carp",
	"enable_syslog",
	"enable_snmp",
	"syslog_level",
	"firewall_rules",
}

var generalProfileTokens = map[string]bool{
	"traditional": true,
	"datacenter":  true,
}

var generalSyslogTokens = map[string]bool{
	"critical":      true,
	"emergencies":   true,
	"errors":        true,
	"alerts":        true,
	"warnings":      true,
	"notifications": true,
	"informational": true,
	"debugging":     true,
}

var _ datasource.DataSource = &generalDataSource{}

type generalDataSource struct{ client *opnsense.Client }

type generalLookupModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Profile       types.String `tfsdk:"profile"`
	EnableCARP    types.Bool   `tfsdk:"enable_carp"`
	EnableSyslog  types.Bool   `tfsdk:"enable_syslog"`
	EnableSNMP    types.Bool   `tfsdk:"enable_snmp"`
	SyslogLevel   types.String `tfsdk:"syslog_level"`
	FirewallRules types.Bool   `tfsdk:"firewall_rules"`
}

func newGeneralDataSource() datasource.DataSource {
	return &generalDataSource{}
}

func (d *generalDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quagga_general"
}

func (d *generalDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads FRR general settings from the stable Quagga GET. Never writes or reconfigures. The only accepted identity is `general`. The public allowlist is the resource refresh contract that the released general node returns: enabled, profile, CARP, syslog, SNMP, syslog level, and firewall rules. Volatile daemons and manual configuration are returned by the whole-model GET and are not lookup state. The general node has no password or write-only secret. Missing required fields are not replaced with model defaults. A disabled flag stays false.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Synthetic singleton ID. Must be `general`.",
			},
			"enabled": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable the FRR routing service. A disabled service stays false and is not replaced with a default.",
			},
			"profile": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Routing profile, `traditional` or `datacenter`. A missing value is not replaced with the model default.",
			},
			"enable_carp": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable CARP failover integration. A disabled value stays false.",
			},
			"enable_syslog": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable syslog logging. A disabled value stays false and is not replaced with the model default.",
			},
			"enable_snmp": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable SNMP (AgentX). A disabled value stays false.",
			},
			"syslog_level": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Syslog verbosity. One of `critical`, `emergencies`, `errors`, `alerts`, `warnings`, `notifications`, `informational`, or `debugging`. A missing value is not replaced with the model default.",
			},
			"firewall_rules": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Automatically generate firewall rules for routing protocols. A disabled value stays false and is not replaced with the model default.",
			},
		},
	}
}

func (d *generalDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	configureQuaggaDataSource(req, resp, func(client *opnsense.Client) { d.client = client })
}

func (d *generalDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config generalLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != generalID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid FRR general settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", generalID, config.ID.ValueString()),
		)
		return
	}
	result, err := getFRRSingleton[generalAPIResponse](ctx, d.client, generalReqOpts)
	if err != nil {
		if strings.Contains(err.Error(), "missing or empty settings") {
			resp.Diagnostics.AddError(
				"FRR general settings not found",
				"The appliance returned an empty FRR general settings object. An empty required object is not found and is not replaced with model defaults.",
			)
			return
		}
		resp.Diagnostics.AddError("Error reading FRR general settings", frrDiagnostic(err))
		return
	}
	if field := generalLookupGap(result); field != "" {
		resp.Diagnostics.AddError(
			"Malformed FRR general settings response",
			fmt.Sprintf("The appliance returned FRR general settings without a readable %s. Missing required fields are not replaced with model defaults.", field),
		)
		return
	}
	applyGeneralLookup(&config, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func generalLookupGap(result *generalAPIResponse) string {
	if result == nil {
		return "response"
	}
	if !result.hasField("enabled") || !generalBoolToken(result.Enabled) {
		return "enabled"
	}
	if !result.hasField("profile") || !generalOptionToken(result.Profile, generalProfileTokens) {
		return "profile"
	}
	if !result.hasField("enablecarp") || !generalBoolToken(result.EnableCARP) {
		return "enable_carp"
	}
	if !result.hasField("enablesyslog") || !generalBoolToken(result.EnableSyslog) {
		return "enable_syslog"
	}
	if !result.hasField("enablesnmp") || !generalBoolToken(result.EnableSNMP) {
		return "enable_snmp"
	}
	if !result.hasField("sysloglevel") || !generalOptionToken(result.SyslogLevel, generalSyslogTokens) {
		return "syslog_level"
	}
	if !result.hasField("fwrules") || !generalBoolToken(result.FirewallRules) {
		return "firewall_rules"
	}
	return ""
}

func generalBoolToken(raw string) bool {
	switch strings.TrimSpace(raw) {
	case "0", "1":
		return true
	default:
		return false
	}
}

func generalOptionToken(raw frrSelectedMap, allowed map[string]bool) bool {
	trimmed := strings.TrimSpace(string(raw))
	return trimmed != "" && allowed[trimmed]
}

func applyGeneralLookup(model *generalLookupModel, result *generalAPIResponse) {
	model.ID = types.StringValue(generalID)
	model.Enabled = types.BoolValue(strings.TrimSpace(result.Enabled) == "1")
	model.Profile = types.StringValue(strings.TrimSpace(string(result.Profile)))
	model.EnableCARP = types.BoolValue(strings.TrimSpace(result.EnableCARP) == "1")
	model.EnableSyslog = types.BoolValue(strings.TrimSpace(result.EnableSyslog) == "1")
	model.EnableSNMP = types.BoolValue(strings.TrimSpace(result.EnableSNMP) == "1")
	model.SyslogLevel = types.StringValue(strings.TrimSpace(string(result.SyslogLevel)))
	model.FirewallRules = types.BoolValue(strings.TrimSpace(result.FirewallRules) == "1")
}
