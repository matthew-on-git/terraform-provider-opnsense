// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &remoteDataSource{}

type remoteDataSource struct{ client *opnsense.Client }

func newRemoteDataSource() datasource.DataSource { return &remoteDataSource{} }

func (d *remoteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_remote"
}

func (d *remoteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing IPsec remote identity on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":       dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this remote authentication entry is enabled. Defaults to 'true'."},
			"connection_id": dsschema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the IPsec connection this entry belongs to."},
			"round":         dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Authentication round (0 = first/default)."},
			"auth":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Authentication method: 'psk', 'pubkey', 'eap-tls', 'eap-mschapv2', 'xauth-pam', or 'eap-radius'. Defaults to 'psk'."},
			"identity":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Peer IKE identity (the 'id' value expected from the peer)."},
			"eap_id":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "EAP identity, when an EAP authentication method is used."},
			"groups":        dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "Local group memberships the peer must satisfy."},
			"certs":         dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "UUIDs of certificates accepted for authentication."},
			"cacerts":       dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "UUIDs of CA certificates used to validate the peer."},
			"description":   dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the remote authentication entry."},
		},
	}
}

func (d *remoteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *remoteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RemoteResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[remoteAPIResponse](ctx, d.client, remoteReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IPsec remote identity", fmt.Sprintf("Could not read IPsec remote identity %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
