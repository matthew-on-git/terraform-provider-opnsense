// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &instanceDataSource{}

type instanceDataSource struct{ client *opnsense.Client }

func newInstanceDataSource() datasource.DataSource { return &instanceDataSource{} }

func (d *instanceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_openvpn_instance"
}

func (d *instanceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing OpenVPN instance on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"vpnid":              dsschema.StringAttribute{Computed: true, MarkdownDescription: "Numeric VPN instance id, auto-assigned by OPNsense."},
			"enabled":            dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the instance is enabled. Defaults to 'true'."},
			"role":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "Instance role: 'server' or 'client'."},
			"description":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description."},
			"dev_type":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Device type: 'tun', 'tap', or 'ovpn'. Defaults to 'tun'."},
			"protocol":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Protocol: 'udp', 'udp4', 'udp6', 'tcp', 'tcp4', 'tcp6'. Defaults to 'udp'."},
			"port":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "Listen/connect port."},
			"local":              dsschema.StringAttribute{Computed: true, MarkdownDescription: "Local interface address to bind."},
			"remote":             dsschema.StringAttribute{Computed: true, MarkdownDescription: "Remote server address (client role)."},
			"server":             dsschema.StringAttribute{Computed: true, MarkdownDescription: "IPv4 tunnel network (CIDR) for server role."},
			"topology":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Topology: 'subnet', 'p2p', or 'net30'. Defaults to 'subnet'."},
			"ca":                 dsschema.StringAttribute{Computed: true, MarkdownDescription: "Certificate Authority reference (UUID/refid)."},
			"cert":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "Server/client certificate reference (UUID/refid)."},
			"verify_client_cert": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Client certificate verification: 'require' or 'none'. Defaults to 'require'."},
			"tls_key":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "TLS static key reference (UUID of an opnsense_openvpn_static_key)."},
			"data_ciphers":       dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Allowed data ciphers (e.g. 'AES-256-GCM')."},
			"auth":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "Auth digest algorithm (e.g. 'SHA256')."},
			"dns_servers":        dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "DNS servers pushed to clients."},
			"push_route":         dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Networks (CIDR) pushed as routes to clients."},
			"redirect_gateway":   dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "redirect-gateway flags (e.g. 'def1', 'bypass-dhcp')."},
			"max_clients":        dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Maximum number of connected clients (0 = unset)."},
			"keepalive_interval": dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Keepalive ping interval in seconds (0 = unset)."},
			"keepalive_timeout":  dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Keepalive timeout in seconds (0 = unset)."},
			"verb":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "Log verbosity level (0-11). Defaults to the OPNsense value when unset."},
		},
	}
}

func (d *instanceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *instanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config InstanceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[instanceAPIResponse](ctx, d.client, instanceReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading OpenVPN instance", fmt.Sprintf("Could not read OpenVPN instance %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
