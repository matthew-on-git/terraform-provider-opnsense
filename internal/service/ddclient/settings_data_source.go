// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ddclient

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &settingsDataSource{}

type settingsDataSource struct{ client *opnsense.Client }

func newSettingsDataSource() datasource.DataSource { return &settingsDataSource{} }

func (d *settingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ddclient_settings"
}

func (d *settingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads Dynamic DNS daemon settings on OPNsense.", Attributes: map[string]dsschema.Attribute{
		"id":         dsschema.StringAttribute{Required: true, MarkdownDescription: "Synthetic singleton ID. Must be `ddclient-settings`."},
		"enabled":    dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the Dynamic DNS daemon is enabled."},
		"backend":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Dynamic DNS backend."},
		"interval":   dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Daemon update interval in seconds (`daemon_delay`)."},
		"verbose":    dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether verbose logging is enabled."},
		"allow_ipv6": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether IPv6 address updates are allowed."},
	}}
}

func (d *settingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *settingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SettingsResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !config.ID.IsNull() && !config.ID.IsUnknown() && config.ID.ValueString() != settingsID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid ddclient settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", settingsID, config.ID.ValueString()),
		)
		return
	}
	result, err := opnsense.GetSingleton[settingsAPIResponse](ctx, d.client, ddclientSettingsReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ddclient settings", fmt.Sprintf("%s", err))
		return
	}
	config.fromAPI(ctx, result, settingsID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
