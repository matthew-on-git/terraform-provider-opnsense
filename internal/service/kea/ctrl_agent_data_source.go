// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ctrlAgentLookupFields is the public GET allowlist. It matches the fields the
// resource refreshes and nothing else. Released KeaCtrlAgent general has only
// enabled, http_host, and http_port. There is no password or write-only secret.
// A missing required field is an error, not a synthetic model default.
var ctrlAgentLookupFields = []string{
	"id",
	"enabled",
	"http_host",
	"http_port",
}

var _ datasource.DataSource = &ctrlAgentDataSource{}

type ctrlAgentDataSource struct{ client *opnsense.Client }

func newCtrlAgentDataSource() datasource.DataSource { return &ctrlAgentDataSource{} }

func (d *ctrlAgentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kea_ctrl_agent"
}

func (d *ctrlAgentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads the Kea DHCP control agent settings from the stable control-agent GET. Never writes or reconfigures. The public allowlist is enabled, listen address, and listen port. The released general model has no password or write-only secret.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Synthetic singleton ID. Must be `kea-ctrl-agent`.",
			},
			"enabled": dsschema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Enable the Kea control agent (required for HA and the dashboard).",
			},
			"http_host": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Address the control agent listens on.",
			},
			"http_port": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Port the control agent listens on.",
			},
		},
	}
}

func (d *ctrlAgentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ctrlAgentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CtrlAgentResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != ctrlAgentID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid Kea control agent ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", ctrlAgentID, config.ID.ValueString()),
		)
		return
	}
	result, err := opnsense.GetSingleton[ctrlAgentAPIResponse](ctx, d.client, ctrlAgentReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading the Kea DHCP control agent settings", fmt.Sprintf("%s", err))
		return
	}
	if !ctrlAgentLookupComplete(result) {
		resp.Diagnostics.AddError(
			"Malformed Kea control agent response",
			"The appliance returned control-agent settings without enabled (0 or 1), a listen address, or a listen port. Missing required fields are not replaced with model defaults.",
		)
		return
	}
	config.fromAPI(ctx, result, ctrlAgentID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func ctrlAgentLookupComplete(result *ctrlAgentAPIResponse) bool {
	if result == nil {
		return false
	}
	switch result.Enabled {
	case "0", "1":
	default:
		return false
	}
	return strings.TrimSpace(result.HTTPHost) != "" && strings.TrimSpace(result.HTTPPort) != ""
}
