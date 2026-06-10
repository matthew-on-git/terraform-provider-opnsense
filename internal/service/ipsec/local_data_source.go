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

var _ datasource.DataSource = &localDataSource{}

type localDataSource struct{ client *opnsense.Client }

func newLocalDataSource() datasource.DataSource { return &localDataSource{} }

func (d *localDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_local"
}

func (d *localDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing IPsec local identity on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"enabled":       dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this local authentication entry is enabled. Defaults to 'true'."},
			"connection_id": dsschema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the IPsec connection this entry belongs to."},
			"round":         dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Authentication round (0 = first/default)."},
			"auth":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Authentication method: 'psk', 'pubkey', 'eap-tls', 'eap-mschapv2', 'xauth-pam', or 'eap-radius'. Defaults to 'psk'."},
			"identity":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "Local IKE identity (the 'id' value sent to the peer)."},
			"eap_id":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "EAP identity, when an EAP authentication method is used."},
			"certs":         dsschema.SetAttribute{ElementType: types.StringType, Computed: true, MarkdownDescription: "UUIDs of certificates used for authentication."},
			"description":   dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the local authentication entry."},
		},
	}
}

func (d *localDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *localDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config LocalResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[localAPIResponse](ctx, d.client, localReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IPsec local identity", fmt.Sprintf("Could not read IPsec local identity %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
