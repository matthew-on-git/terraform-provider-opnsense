// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

var _ datasource.DataSource = &aclDataSource{}

type aclDataSource struct{ client *opnsense.Client }

func newACLDataSource() datasource.DataSource { return &aclDataSource{} }

func (d *aclDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_acl"
}

func (d *aclDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Reads an existing HAProxy ACL on OPNsense by UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID to look up.",
			},
			"name":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "Name of the ACL rule."},
			"description":   dsschema.StringAttribute{Computed: true, MarkdownDescription: "Description of the ACL."},
			"expression":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Match expression type (e.g., 'hdr_beg', 'hdr', 'path_beg', 'path', 'ssl_fc_sni', 'ssl_sni', 'src', 'nbsrv', 'custom_acl')."},
			"negate":        dsschema.BoolAttribute{Computed: true, MarkdownDescription: "Invert the match condition. Defaults to 'false'."},
			"hdr_beg":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "HTTP Host header starts with this value."},
			"hdr_end":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "HTTP Host header ends with this value."},
			"hdr":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "HTTP Host header exact match."},
			"path_beg":      dsschema.StringAttribute{Computed: true, MarkdownDescription: "URL path starts with this value."},
			"path":          dsschema.StringAttribute{Computed: true, MarkdownDescription: "URL path exact match."},
			"ssl_sni":       dsschema.StringAttribute{Computed: true, MarkdownDescription: "SNI TLS extension matches (TCP inspection)."},
			"ssl_fc_sni":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "SNI TLS extension matches (locally deciphered)."},
			"src":           dsschema.StringAttribute{Computed: true, MarkdownDescription: "Source IP address match."},
			"nbsrv_backend": dsschema.StringAttribute{Computed: true, MarkdownDescription: "UUID of the backend to check for minimum usable servers."},
			"custom_acl":    dsschema.StringAttribute{Computed: true, MarkdownDescription: "Raw HAProxy condition (pass-through)."},
		},
	}
}

func (d *aclDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *aclDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ACLResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := config.ID.ValueString()
	result, err := opnsense.Get[aclAPIResponse](ctx, d.client, aclReqOpts, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading HAProxy ACL", fmt.Sprintf("Could not read HAProxy ACL %s: %s", id, err))
		return
	}
	config.fromAPI(ctx, result, id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
