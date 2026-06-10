// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package firewall

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &natPortForwardDataSource{}

type natPortForwardDataSource struct{ client *opnsense.Client }

func newNatPortForwardDataSource() datasource.DataSource { return &natPortForwardDataSource{} }

func (d *natPortForwardDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_nat_port_forward"
}

func (d *natPortForwardDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing firewall NAT port forward rule on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":          dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this rule is enabled. Defaults to 'true'."},
			"interface":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Network interface for incoming traffic (e.g., 'wan')."},
			"ip_protocol":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "IP version: 'inet' (IPv4), 'inet6' (IPv6), or 'inet46' (both)."},
			"protocol":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Protocol (e.g., 'tcp', 'udp', 'TCP/UDP')."},
			"source_net":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source network ('any', CIDR, or alias)."},
			"source_port":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source port or range. Empty for any."},
			"source_not":       dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Invert source match. Defaults to 'false'."},
			"destination_net":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Destination network to match (typically the WAN address, e.g., 'wanip')."},
			"destination_port": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Destination port to match (the external port)."},
			"destination_not":  dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Invert destination match. Defaults to 'false'."},
			"target":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Internal target IP address or alias to redirect to."},
			"local_port":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "Internal target port to redirect to."},
			"log":              dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Log matching packets. Defaults to 'false'."},
			"description":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the rule."},
			"categories":       dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Set of category UUIDs assigned to this rule."},
		},
	}
}

func (d *natPortForwardDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *natPortForwardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NatPortForwardResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[natPortForwardAPIResponse](ctx, d.client, natPortForwardReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall NAT port forward rule", fmt.Sprintf("Could not read firewall NAT port forward rule %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
