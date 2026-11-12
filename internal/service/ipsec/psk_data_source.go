// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// pskLookupFields is the public GET allowlist. It matches the fields the
// resource refreshes and nothing else. Excluded on purpose:
//   - key: sensitive pre-shared key; write-only on the resource because GET
//     is not a safe secret channel. Present or omitted, never lookup state.
//
// A missing secret is absent. It is not an empty synthetic value.
var pskLookupFields = []string{
	"id",
	"identity",
	"remote_identity",
	"key_type",
	"description",
}

var _ datasource.DataSource = &pskDataSource{}

type pskDataSource struct{ client *opnsense.Client }

// pskLookupModel is lookup state only. It has no key field, so a secret
// cannot be copied into state even when get_item returns one.
type pskLookupModel struct {
	ID             types.String `tfsdk:"id"`
	Identity       types.String `tfsdk:"identity"`
	RemoteIdentity types.String `tfsdk:"remote_identity"`
	KeyType        types.String `tfsdk:"key_type"`
	Description    types.String `tfsdk:"description"`
}

func newPSKDataSource() datasource.DataSource { return &pskDataSource{} }

func (d *pskDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_psk"
}

func (d *pskDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing IPsec pre-shared key by UUID. Public metadata only: local identity, remote identity, key type, and description. The pre-shared key value is not lookup state.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the pre-shared key.",
			},
			"identity": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Local identity for the pre-shared key.",
			},
			"remote_identity": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Remote identity for the pre-shared key. Empty when the appliance stores none.",
			},
			"key_type": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Key type: `PSK` or `EAP`.",
			},
			"description": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the pre-shared key. Empty when the appliance stores none.",
			},
		},
	}
}

func (d *pskDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *pskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config pskLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	if _, err := uuid.Parse(id); err != nil {
		resp.Diagnostics.AddError("Invalid IPsec pre-shared key UUID", fmt.Sprintf("id must be a UUID: %s", err))
		return
	}
	result, err := opnsense.Get[ipsecPSKAPIResponse](ctx, d.client, pskReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IPsec pre-shared key", fmt.Sprintf("Could not read IPsec pre-shared key %s: %s", id, err))
		return
	}
	if strings.TrimSpace(result.Identity) == "" || strings.TrimSpace(string(result.KeyType)) == "" {
		resp.Diagnostics.AddError("Malformed IPsec pre-shared key response", "The appliance returned a pre-shared key without an identity or key type.")
		return
	}
	config.fromLookup(result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (m *pskLookupModel) fromLookup(a *ipsecPSKAPIResponse, id string) {
	m.ID = types.StringValue(id)
	m.Identity = types.StringValue(a.Identity)
	m.RemoteIdentity = types.StringValue(a.RemoteIdentity)
	m.KeyType = types.StringValue(string(a.KeyType))
	m.Description = types.StringValue(a.Description)
}
