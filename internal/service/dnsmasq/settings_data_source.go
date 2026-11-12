// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dnsmasq

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// settingsLookupFields is the public settings GET allowlist. It matches the
// resource contract and nothing else. Excluded on purpose:
//   - expand_hosts: stored public boolean the resource does not own
//   - dhcp %domain: response metadata injected from the system domain
//   - hosts, domains, tags, ranges, options, boots: separate UUID lookups
//
// The released general and dhcp models have no password or write-only secret.
var settingsLookupFields = []string{
	"id",
	"enable",
	"register_dhcp",
	"register_dhcp_static",
	"dhcp_first",
	"strict_order",
	"domain_needed",
	"no_private_reverse",
	"no_resolv",
	"log_queries",
	"no_hosts",
	"strict_bind",
	"dnssec",
	"register_dhcp_domain",
	"interface",
	"port",
	"dns_port",
	"dns_forward_max",
	"cache_size",
	"local_ttl",
	"add_mac",
	"add_subnet",
	"strip_subnet",
	"dhcp",
	"no_ident",
}

var _ datasource.DataSource = &settingsDataSource{}

type settingsDataSource struct{ client *opnsense.Client }

func newSettingsDataSource() datasource.DataSource { return &settingsDataSource{} }

func (d *settingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnsmasq_settings"
}

func (d *settingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads Dnsmasq DNS/DHCP general settings from the stable settings GET. Never writes or reconfigures. Response-only %domain, stored expand_hosts, and child arrays are not lookup state.",
		Attributes: map[string]dsschema.Attribute{
			"id":                   dsschema.StringAttribute{Required: true, MarkdownDescription: "Synthetic singleton ID. Must be `dnsmasq`."},
			"enable":               dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Enable the Dnsmasq service."},
			"register_dhcp":        dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Register DHCP leases in DNS."},
			"register_dhcp_static": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Register static DHCP mappings in DNS."},
			"dhcp_first":           dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Prefer DHCP names over other sources."},
			"strict_order":         dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Query upstream servers in strict order."},
			"domain_needed":        dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Never forward queries without a domain."},
			"no_private_reverse":   dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Do not forward reverse lookups for private ranges."},
			"no_resolv":            dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Ignore /etc/resolv.conf for upstream servers."},
			"log_queries":          dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Log all DNS queries."},
			"no_hosts":             dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Do not read /etc/hosts."},
			"strict_bind":          dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Bind only to configured interfaces."},
			"dnssec":               dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Enable DNSSEC validation."},
			"register_dhcp_domain": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Domain appended to registered DHCP names."},
			"interface":            dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Interfaces Dnsmasq listens on (empty = all)."},
			"port":                 dsschema.StringAttribute{Computed: true, MarkdownDescription: "Service listen port (empty = default)."},
			"dns_port":             dsschema.StringAttribute{Computed: true, MarkdownDescription: "DNS listen port returned by settings GET. The released model marks this field volatile."},
			"dns_forward_max":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Maximum concurrent forwarded queries."},
			"cache_size":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "DNS cache size (entries)."},
			"local_ttl":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "TTL for locally-served records."},
			"add_mac":              dsschema.StringAttribute{Computed: true, MarkdownDescription: "Add requesting MAC to forwarded queries (standard, base64, text)."},
			"add_subnet":           dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Add the client subnet to forwarded queries."},
			"strip_subnet":         dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Strip the client subnet from replies."},
			"dhcp":                 dhcpDataSourceSchema(),
			"no_ident":             dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Disable the chaos TXT ident responses."},
		},
	}
}

func dhcpDataSourceSchema() dsschema.SingleNestedAttribute {
	return dsschema.SingleNestedAttribute{
		Computed:            true,
		MarkdownDescription: "Typed DHCP/RA settings from the stable settings GET. Empty integer sentinels stay null. Response-only %domain is not a field.",
		Attributes: map[string]dsschema.Attribute{
			"no_interface":     dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Interfaces excluded from DHCP and router advertisements (empty = none)."},
			"fqdn":             dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Register DHCP client names as fully qualified names."},
			"domain":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Stored DHCP domain. Empty is an empty string, not response metadata %domain."},
			"local":            dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Do not forward queries for the DHCP domain."},
			"lease_max":        dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Maximum DHCP leases. An empty API integer is null, not zero."},
			"authoritative":    dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Act as the authoritative DHCP server."},
			"default_fw_rules": dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Generate default DHCP firewall rules."},
			"reply_delay":      dsschema.Int64Attribute{Computed: true, MarkdownDescription: "DHCP reply delay in seconds. An empty API integer is null, not zero."},
			"enable_ra":        dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Enable IPv6 router advertisements."},
			"host_ping":        dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Check address availability before allocating a lease."},
			"nosync":           dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Disable configuration synchronization for DHCP settings."},
			"log_dhcp":         dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Log DHCP options and tags."},
			"log_quiet":        dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Suppress routine DHCP logging."},
		},
	}
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
	if config.ID.IsNull() || config.ID.IsUnknown() || config.ID.ValueString() != settingsID {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Invalid Dnsmasq settings ID",
			fmt.Sprintf("Expected %q for this singleton data source, got %q.", settingsID, config.ID.ValueString()),
		)
		return
	}
	result, err := opnsense.GetSingleton[settingsAPIResponse](ctx, d.client, settingsReqOpts)
	if err != nil {
		resp.Diagnostics.AddError("Error reading the Dnsmasq DNS/DHCP general settings", fmt.Sprintf("%s", err))
		return
	}
	config.fromAPI(ctx, result, settingsID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
