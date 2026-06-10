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

var _ datasource.DataSource = &natOutboundDataSource{}

type natOutboundDataSource struct{ client *opnsense.Client }

func newNatOutboundDataSource() datasource.DataSource { return &natOutboundDataSource{} }

func (d *natOutboundDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_nat_outbound"
}

func (d *natOutboundDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing firewall outbound NAT rule on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":          dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this rule is enabled. Defaults to 'true'."},
			"sequence":         dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Rule sequence number (1-999999). Controls evaluation order."},
			"interface":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Outbound network interface (e.g., 'wan')."},
			"ip_protocol":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "IP version: 'inet' (IPv4) or 'inet6' (IPv6)."},
			"protocol":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Protocol (e.g., 'any', 'tcp', 'udp', 'TCP/UDP')."},
			"source_net":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source network ('any', CIDR, or alias)."},
			"source_not":       dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Invert source match. Defaults to 'false'."},
			"source_port":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source port or range. Empty for any."},
			"destination_net":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Destination network ('any', CIDR, or alias)."},
			"destination_not":  dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Invert destination match. Defaults to 'false'."},
			"destination_port": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Destination port or range. Empty for any."},
			"target":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Translation target IP or alias (e.g., 'wanip', interface IP, or specific address)."},
			"target_port":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Translated source port. Empty to preserve original."},
			"no_nat":           dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Suppress NAT for matching traffic. Defaults to 'false'."},
			"static_nat_port":  dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Preserve the original source port. Defaults to 'false'."},
			"log":              dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Log matching packets. Defaults to 'false'."},
			"description":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the rule."},
			"categories":       dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Set of category UUIDs assigned to this rule."},
		},
	}
}

func (d *natOutboundDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *natOutboundDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NatOutboundResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[natOutboundAPIResponse](ctx, d.client, natOutboundReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall outbound NAT rule", fmt.Sprintf("Could not read firewall outbound NAT rule %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
