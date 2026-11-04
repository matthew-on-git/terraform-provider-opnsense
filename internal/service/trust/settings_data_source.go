// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

type settingsDataSource struct{ client *opnsense.Client }

func newSettingsDataSource() datasource.DataSource { return &settingsDataSource{} }
func (d *settingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trust_settings"
}

func (d *settingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*opnsense.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *opnsense.Client.")
		return
	}
	d.client = c
}

func (d *settingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Reads the six audited Trust settings without mutation.", Attributes: map[string]schema.Attribute{
		"id":                        schema.StringAttribute{Computed: true, MarkdownDescription: "Fixed singleton identity trust_settings."},
		"store_intermediate_certs":  schema.BoolAttribute{Computed: true, MarkdownDescription: "Allow locally defined intermediate certificate authorities in the local trust store. Stored policy only."},
		"install_crls":              schema.BoolAttribute{Computed: true, MarkdownDescription: "Store all configured certificate revocation lists in the default trust store. Stored policy only."},
		"fetch_crls":                schema.BoolAttribute{Computed: true, MarkdownDescription: "Schedule hourly CRL downloads using distribution points from CAs in the trust store. Stored policy only."},
		"enable_legacy_sect":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Enable OpenSSL legacy providers. Stored policy only."},
		"enable_config_constraints": schema.BoolAttribute{Computed: true, MarkdownDescription: "Enable custom OpenSSL configuration constraints. Stored policy only."},
		"cipher_string":             schema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Stored selected OpenSSL CipherString choice keys; an empty set means no selected choices."},
	}}
}

func (d *settingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if err := requireSettings(d.client); err != nil {
		resp.Diagnostics.AddError("Unconfigured client", err.Error())
		return
	}
	remote, err := readSettings(ctx, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Cannot read Trust settings", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}
