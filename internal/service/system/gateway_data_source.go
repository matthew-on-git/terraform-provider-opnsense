// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &gatewayDataSource{}

type gatewayDataSource struct{ client *opnsense.Client }

func newGatewayDataSource() datasource.DataSource { return &gatewayDataSource{} }

func (d *gatewayDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_gateway"
}

func (d *gatewayDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing system gateway on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"latency_low":                  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Low latency threshold in milliseconds. Empty means appliance default."},
			"latency_high":                 dsschema.StringAttribute{Computed: true, MarkdownDescription: "High latency threshold in milliseconds. Empty means appliance default."},
			"loss_low":                     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Low packet loss threshold in percent (1-99). Empty means appliance default."},
			"loss_high":                    dsschema.StringAttribute{Computed: true, MarkdownDescription: "High packet loss threshold in percent (1-100). Empty means appliance default."},
			"interval":                     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Probe interval in seconds. Empty means appliance default."},
			"time_period":                  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Time period in seconds over which monitoring results are averaged. Empty means appliance default."},
			"loss_interval":                dsschema.StringAttribute{Computed: true, MarkdownDescription: "Interval in seconds before a probe is considered lost. Empty means appliance default."},
			"data_length":                  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Probe payload length in bytes; zero is an explicit value. Empty means appliance default."},
			"monitor":                      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Monitor IP address. An empty string uses the gateway address."},
			"far_gateway":                  dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Allow a gateway outside the interface subnet."},
			"monitor_no_route":             dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Do not add a host route for the monitor address."},
			"monitor_kill_states":          dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Flush states when the gateway goes down."},
			"monitor_kill_states_priority": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Flush states when a higher-priority gateway becomes available."},
			"force_down":                   dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Treat the gateway as down regardless of monitoring."},
			"no_sync":                      dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Exclude this gateway from configuration synchronization."},
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":         dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this gateway is enabled. Defaults to 'true'."},
			"name":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "Gateway name (unique, alphanumeric + '_-', max 32 chars)."},
			"description":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description."},
			"interface":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "Interface (e.g., 'wan')."},
			"ip_protocol":     dsschema.StringAttribute{Computed: true, MarkdownDescription: "IP version: 'inet' (IPv4) or 'inet6' (IPv6)."},
			"gateway":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Gateway IP address."},
			"default_gateway": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Mark as default gateway. Defaults to 'false'."},
			"monitor_disable": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Disable gateway monitoring. Defaults to 'true'."},
			"weight":          dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Weight for gateway groups (1-5)."},
			"priority":        dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Priority (0-255). Lower = higher priority."},
		},
	}
}

func (d *gatewayDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *gatewayDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config GatewayResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[gatewayAPIResponse](ctx, d.client, gatewayReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading system gateway", fmt.Sprintf("Could not read system gateway %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
