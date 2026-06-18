// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &certificateDataSource{}

type certificateDataSource struct{ client *opnsense.Client }

func newCertificateDataSource() datasource.DataSource { return &certificateDataSource{} }

func (d *certificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acme_certificate"
}

func (d *certificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads an existing ACME certificate on OPNsense by UUID.", Attributes: map[string]dsschema.Attribute{
		"id":                dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID to look up."},
		"enabled":           dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this ACME certificate is enabled."},
		"name":              dsschema.StringAttribute{Computed: true, MarkdownDescription: "Certificate name."},
		"description":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the certificate."},
		"alt_names":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Alternative names."},
		"account":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "ACME account UUID."},
		"validation_method": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Validation method UUID."},
		"key_length":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Certificate key length."},
		"auto_renewal":      dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether automatic renewal is enabled."},
		"cert_ref_id":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "HAProxy legacy certificate refid populated after successful issuance."},
		"status_code":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "ACME issuance status code reported by OPNsense."},
		"status":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable ACME issuance status reported by OPNsense."},
	}}
}

func (d *certificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *certificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CertificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[certificateAPIResponse](ctx, d.client, certificateReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading ACME certificate", fmt.Sprintf("Could not read ACME certificate %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
