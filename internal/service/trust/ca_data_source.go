// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package trust

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

type caDataSource struct{ client *opnsense.Client }

func newCADataSource() datasource.DataSource { return &caDataSource{} }
func (d *caDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trust_ca"
}

func (d *caDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{MarkdownDescription: "Reads authoritative Trust ca metadata by UUID. Private keys are sensitive and null when absent; one-time keys and creation intent cannot be recovered by lookup.", Attributes: map[string]dsschema.Attribute{
		"id":                  dsschema.StringAttribute{Required: true, MarkdownDescription: "UUID of the existing object to read; lookup never imports or mutates it."},
		"refid":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "OPNsense reference identifier used by linked resources, distinct from the UUID."},
		"description":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable description returned by the appliance."},
		"ca_ref":              dsschema.StringAttribute{Computed: true, MarkdownDescription: "Returned issuing CA reference identifier, distinct from its UUID; null when absent."},
		"common_name":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "Subject common name returned by the appliance."},
		"country":             dsschema.StringAttribute{Computed: true, MarkdownDescription: "Two-letter subject country code returned by the appliance."},
		"city":                dsschema.StringAttribute{Computed: true, MarkdownDescription: "Subject locality returned by the appliance."},
		"state":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "Subject state or province returned by the appliance."},
		"organization":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "Subject organization returned by the appliance."},
		"organizational_unit": dsschema.StringAttribute{Computed: true, MarkdownDescription: "Subject organizational unit returned by the appliance."},
		"email":               dsschema.StringAttribute{Computed: true, MarkdownDescription: "Subject email address returned by the appliance."},
		"certificate":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "certificate."},
		"private_key":         dsschema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "private key."},
		"csr":                 dsschema.StringAttribute{Computed: true, MarkdownDescription: "Returned PEM certificate signing request; null when absent."},
		"valid_from":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Actual validity start returned by the API as a numeric Unix timestamp or string; not the volatile lifetime default."},
		"valid_to":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "Actual validity end returned by the API as a numeric Unix timestamp or string."},
		"subject":             dsschema.StringAttribute{Computed: true, MarkdownDescription: "API-reconstructed subject distinguished name."},
		"purpose":             dsschema.StringAttribute{Computed: true, MarkdownDescription: "API-derived RFC 3280 EKU purpose, not generation intent or certificate profile."},
		"key_type":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "Key type observed in the API editor; a CSR-only editor may report a template default."},
		"digest":              dsschema.StringAttribute{Computed: true, MarkdownDescription: "Digest observed in the API editor; a CSR-only editor may report a template default."},
		"cert_type":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Released certificate profile: usr_cert, server_cert, combined_server_client or v3_ca. Not EKU purpose."},
		"altnames_dns":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "altnames dns."},
		"altnames_ip":         dsschema.StringAttribute{Computed: true, MarkdownDescription: "altnames ip."},
		"altnames_uri":        dsschema.StringAttribute{Computed: true, MarkdownDescription: "altnames uri."},
		"altnames_email":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "altnames email."},
		"lifetime":            dsschema.Int64Attribute{Computed: true, MarkdownDescription: "Volatile editor default in days, NOT actual certificate validity. Use valid_from and valid_to."},
		"serial":              dsschema.StringAttribute{Computed: true, MarkdownDescription: "CA signing serial counter, not necessarily certificate serial number."},
		"refcount":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "CA reference count reported by the appliance."},
		"in_use":              dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the certificate is referenced by services."},
		"is_user":             dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the subject matches an appliance user."},
		"ocsp_uri":            dsschema.StringAttribute{Computed: true, MarkdownDescription: "API-reconstructed OCSP responder URL."},
	}}
}

func (d *caDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *caDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config trustReadModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	record, err := trustRecord(ctx, d.client, "ca", config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Trust lookup failed", err.Error())
		return
	}
	config.fromRecord(record, config.ID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
