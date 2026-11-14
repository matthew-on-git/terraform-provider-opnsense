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

// dhcpv6SettingsLookupFields is the public GET allowlist. It is the resource
// refresh contract intersected with fields the stable DHCPv6 general model
// actually returns. Released KeaDhcpv6 general also returns socket retry,
// decline probation, and mac_sources, and getAction also returns lexpire and
// ha. Those are not lookup state. The provider resource schema copies
// dhcp_socket_type from DHCPv4, but stable 26.7.8 and 26.7.9 do not return
// that field; it is not lookup state and is not replaced with the generator
// default. There is no password or write-only secret on the general node.
// A missing required field is an error, not a synthetic model default.
var dhcpv6SettingsLookupFields = []string{
	"id",
	"enabled",
	"manual_config",
	"interfaces",
	"valid_lifetime",
	"firewall_rules",
}

var _ datasource.DataSource = &dhcpv6SettingsDataSource{}

type dhcpv6SettingsDataSource struct{ client *opnsense.Client }

type dhcpv6SettingsLookupModel struct {
	ID            types.String `tfsdk:"id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	ManualConfig  types.Bool   `tfsdk:"manual_config"`
	Interfaces    types.Set    `tfsdk:"interfaces"`
	ValidLifetime types.Int64  `tfsdk:"valid_lifetime"`
	FirewallRules types.Bool   `tfsdk:"firewall_rules"`
}

func newDHCPv6SettingsDataSource() datasource.DataSource {
	return &dhcpv6SettingsDataSource{}
}

func (d *dhcpv6SettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kea_dhcpv6_settings"
}

func (d *dhcpv6SettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads the Kea DHCPv6 general settings from the stable DHCPv6 GET. Never writes or reconfigures. The public allowlist is the resource refresh contract that the released general model actually returns: enabled, manual configuration, interfaces, valid lifetime, and firewall rules. Socket retry, decline probation, MAC sources, lease expiry, and HA nodes are not lookup state. DHCP socket type is a provider resource field copied from DHCPv4 and is absent from the stable general model, so it is not lookup state and is not replaced with a default. The released general model has no password or write-only secret. Missing required fields are not replaced with model defaults.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Synthetic singleton ID. Must be `kea-dhcpv6`.",
			},
			"enabled": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable the Kea DHCPv6 server.",
			},
			"manual_config": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Manage the Kea config file manually (disable GUI management). A blank optional value stays null.",
			},
			"interfaces": dsschema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Interfaces the DHCPv6 server listens on.",
			},
			"valid_lifetime": dsschema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Default lease valid lifetime in seconds.",
			},
			"firewall_rules": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Automatically add firewall rules for the service.",
			},
		},
	}
}

func (d *dhcpv6SettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dhcpv6SettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dhcpv6SettingsLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != dhcpv6SettingsID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid Kea DHCPv6 settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", dhcpv6SettingsID, config.ID.ValueString()),
		)
		return
	}
	result, err := opnsense.GetSingleton[dhcpv6SettingsAPIResponse](ctx, d.client, dhcpv6SettingsReqOpts)
	if err != nil {
		var missing *opnsense.NotFoundError
		if errors.As(err, &missing) {
			resp.Diagnostics.AddError(
				"Kea DHCPv6 settings not found",
				"The appliance returned an empty DHCPv6 general settings object. An empty required object is not found and is not replaced with model defaults.",
			)
			return
		}
		resp.Diagnostics.AddError("Error reading the Kea DHCPv6 general settings", fmt.Sprintf("%s", err))
		return
	}
	if field := dhcpv6SettingsLookupGap(result); field != "" {
		resp.Diagnostics.AddError(
			"Malformed Kea DHCPv6 settings response",
			fmt.Sprintf("The appliance returned DHCPv6 settings without a readable %s. Missing required fields are not replaced with model defaults.", field),
		)
		return
	}
	applyDHCPv6SettingsLookup(&config, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func dhcpv6SettingsLookupGap(result *dhcpv6SettingsAPIResponse) string {
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
	return ""
}

func applyDHCPv6SettingsLookup(model *dhcpv6SettingsLookupModel, result *dhcpv6SettingsAPIResponse) {
	model.ID = types.StringValue(dhcpv6SettingsID)
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
}
