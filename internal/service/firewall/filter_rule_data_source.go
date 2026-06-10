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

var _ datasource.DataSource = &filterRuleDataSource{}

type filterRuleDataSource struct{ client *opnsense.Client }

var filterRuleDataSourceReqOpts = opnsense.ReqOpts{
	GetEndpoint:    "/api/firewall/filter/getRule",
	SearchEndpoint: "/api/firewall/filter/searchRule",
	Monad:          "rule",
}

func newFilterRuleDataSource() datasource.DataSource { return &filterRuleDataSource{} }

func (d *filterRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_filter_rule"
}

func (d *filterRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing firewall filter rule on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":          dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this rule is enabled. Defaults to 'true'."},
			"sequence":         dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Rule sequence number (1-999999). Controls evaluation order."},
			"action":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Rule action: 'pass', 'block', or 'reject'."},
			"quick":            dsschema.BoolAttribute{Computed: true, MarkdownDescription: "First match wins. Defaults to 'true'."},
			"interface":        dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Network interfaces this rule applies to."},
			"direction":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Traffic direction: 'in', 'out', or 'any'."},
			"ip_protocol":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "IP version: 'inet' (IPv4), 'inet6' (IPv6), or 'inet46' (both)."},
			"protocol":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Protocol (e.g., 'any', 'tcp', 'udp', 'icmp', 'TCP/UDP')."},
			"source_net":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source network ('any', interface name, CIDR, or alias)."},
			"source_port":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source port or range (e.g., '80', '80:443'). Empty for any."},
			"source_not":       dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Invert source match. Defaults to 'false'."},
			"destination_net":  dsschema.StringAttribute{Computed: true, MarkdownDescription: "Destination network ('any', interface name, CIDR, or alias)."},
			"destination_port": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Destination port or range (e.g., '443', '8080:8090'). Empty for any."},
			"destination_not":  dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Invert destination match. Defaults to 'false'."},
			"gateway":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Gateway for policy routing. Empty for default."},
			"log":              dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Log matching packets. Defaults to 'false'."},
			"description":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the rule."},
			"categories":       dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Set of category UUIDs assigned to this rule."},
		},
	}
}

func (d *filterRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *filterRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FilterRuleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[filterRuleAPIResponse](ctx, d.client, filterRuleDataSourceReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewall filter rule", fmt.Sprintf("Could not read firewall filter rule %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
