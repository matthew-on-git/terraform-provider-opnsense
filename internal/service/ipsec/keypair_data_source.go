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

// keyPairLookupFields is the public GET allowlist. It matches the fields the
// resource refreshes and nothing else. Excluded on purpose:
//   - public_key: stored PEM the resource treats as write-only because the API re-encodes it
//   - private_key: optional sensitive PEM; missing or present, never lookup state
//
// A missing secret is absent. It is not an empty synthetic value.
var keyPairLookupFields = []string{
	"id",
	"name",
	"key_type",
	"key_size",
	"key_fingerprint",
}

var _ datasource.DataSource = &keyPairDataSource{}

type keyPairDataSource struct{ client *opnsense.Client }

// keyPairLookupModel is lookup state only. It has no PEM fields, so a secret
// cannot be copied into state even when getItem returns one.
type keyPairLookupModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	KeyType        types.String `tfsdk:"key_type"`
	KeySize        types.String `tfsdk:"key_size"`
	KeyFingerprint types.String `tfsdk:"key_fingerprint"`
}

func newKeyPairDataSource() datasource.DataSource { return &keyPairDataSource{} }

func (d *keyPairDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ipsec_key_pair"
}

func (d *keyPairDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing IPsec key pair by UUID. Public metadata only: name, key type, derived size, and fingerprint. Public and private key PEM payloads are not lookup state.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the key pair.",
			},
			"name": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name for the key pair.",
			},
			"key_type": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Key type: `rsa` or `ecdsa`.",
			},
			"key_size": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Key size in bits derived by OPNsense. Empty when the appliance returns no size.",
			},
			"key_fingerprint": dsschema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Fingerprint derived by OPNsense. Empty when the appliance returns no fingerprint.",
			},
		},
	}
}

func (d *keyPairDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *keyPairDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config keyPairLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	if _, err := uuid.Parse(id); err != nil {
		resp.Diagnostics.AddError("Invalid IPsec key pair UUID", fmt.Sprintf("id must be a UUID: %s", err))
		return
	}
	result, err := opnsense.Get[keyPairAPIResponse](ctx, d.client, keyPairReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IPsec key pair", fmt.Sprintf("Could not read IPsec key pair %s: %s", id, err))
		return
	}
	if strings.TrimSpace(result.Name) == "" || strings.TrimSpace(string(result.KeyType)) == "" {
		resp.Diagnostics.AddError("Malformed IPsec key pair response", "The appliance returned a key pair without a name or key type.")
		return
	}
	config.fromLookup(result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (m *keyPairLookupModel) fromLookup(a *keyPairAPIResponse, id string) {
	m.ID = types.StringValue(id)
	m.Name = types.StringValue(a.Name)
	m.KeyType = types.StringValue(string(a.KeyType))
	m.KeySize = types.StringValue(a.KeySize)
	m.KeyFingerprint = types.StringValue(a.KeyFingerprint)
}
