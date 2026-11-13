// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/tfconv"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// dhcpv4SettingsLookupFields is the public GET allowlist. It matches the
// fields the resource refreshes and nothing else. Released KeaDhcpv4 general
// also returns socket retry, decline probation, and compatibility fields, and
// getAction also returns lexpire and ha. Those are not lookup state. There is
// no password or write-only secret. A missing required field is an error, not
// a synthetic model default.
var dhcpv4SettingsLookupFields = []string{
	"id",
	"enabled",
	"manual_config",
	"interfaces",
	"valid_lifetime",
	"firewall_rules",
	"dhcp_socket_type",
}

var _ datasource.DataSource = &dhcpv4SettingsDataSource{}

type dhcpv4SettingsDataSource struct{ client *opnsense.Client }

func newDHCPv4SettingsDataSource() datasource.DataSource {
	return &dhcpv4SettingsDataSource{}
}

func (d *dhcpv4SettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kea_dhcpv4_settings"
}

func (d *dhcpv4SettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads the Kea DHCPv4 general settings from the stable DHCPv4 GET. Never writes or reconfigures. The public allowlist matches the resource refresh contract: enabled, manual configuration, interfaces, valid lifetime, firewall rules, and DHCP socket type. Socket retry, decline probation, compatibility, lease expiry, and HA nodes are not lookup state. The released general model has no password or write-only secret. Missing required fields are not replaced with model defaults.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Synthetic singleton ID. Must be `kea-dhcpv4`.",
			},
			"enabled": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable the Kea DHCPv4 server.",
			},
			"manual_config": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Manage the Kea config file manually (disable GUI management). A blank optional value stays null.",
			},
			"interfaces": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Interfaces the DHCPv4 server listens on.",
			},
			"valid_lifetime": dsschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Default lease valid lifetime in seconds.",
			},
			"firewall_rules": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Automatically add firewall rules for the service.",
			},
			"dhcp_socket_type": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DHCP socket type: raw or udp.",
			},
		},
	}
}

func (d *dhcpv4SettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	d.client = client
}

func (d *dhcpv4SettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DHCPv4SettingsResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != dhcpv4SettingsID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid Kea DHCPv4 settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", dhcpv4SettingsID, config.ID.ValueString()),
		)
		return
	}
	result, err := opnsense.GetSingleton[dhcpv4SettingsAPIResponse](ctx, d.client, dhcpv4SettingsReqOpts)
	if err != nil {
		var missing *opnsense.NotFoundError
		if errors.As(err, &missing) {
			resp.Diagnostics.AddError(
				"Kea DHCPv4 settings not found",
				"The appliance returned an empty DHCPv4 general settings object. An empty required object is not found and is not replaced with model defaults.",
			)
			return
		}
		resp.Diagnostics.AddError("Error reading the Kea DHCPv4 general settings", fmt.Sprintf("%s", err))
		return
	}
	if field := dhcpv4SettingsLookupGap(result); field != "" {
		resp.Diagnostics.AddError(
			"Malformed Kea DHCPv4 settings response",
			fmt.Sprintf("The appliance returned DHCPv4 settings without a readable %s. Missing required fields are not replaced with model defaults.", field),
		)
		return
	}
	applyDHCPv4SettingsLookup(&config, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func dhcpv4SettingsLookupGap(result *dhcpv4SettingsAPIResponse) string {
	if result == nil {
		return "response"
	}
	switch result.Enabled {
	case "0", "1":
	default:
		return "enabled"
	}
	switch strings.TrimSpace(result.ManualConfig) {
	case "", "0", "1":
	default:
		return "manual_config"
	}
	if _, err := strconv.ParseInt(strings.TrimSpace(result.ValidLifetime), 10, 64); err != nil || strings.TrimSpace(result.ValidLifetime) == "" {
		return "valid_lifetime"
	}
	switch result.FirewallRules {
	case "0", "1":
	default:
		return "firewall_rules"
	}
	switch strings.TrimSpace(string(result.DHCPSocketType)) {
	case "raw", "udp":
	default:
		return "dhcp_socket_type"
	}
	return ""
}

func applyDHCPv4SettingsLookup(model *DHCPv4SettingsResourceModel, result *dhcpv4SettingsAPIResponse) {
	model.ID = types.StringValue(dhcpv4SettingsID)
	model.Enabled = types.BoolValue(result.Enabled == "1")
	manual := strings.TrimSpace(result.ManualConfig)
	if manual == "" {
		model.ManualConfig = types.BoolNull()
	} else {
		model.ManualConfig = types.BoolValue(manual == "1")
	}
	model.Interfaces = tfconv.SliceToSet(result.Interfaces)
	lifetime, _ := strconv.ParseInt(strings.TrimSpace(result.ValidLifetime), 10, 64)
	model.ValidLifetime = types.Int64Value(lifetime)
	model.FirewallRules = types.BoolValue(result.FirewallRules == "1")
	model.DHCPSocketType = types.StringValue(strings.TrimSpace(string(result.DHCPSocketType)))
}
