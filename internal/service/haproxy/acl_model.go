// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ACLResourceModel is the Terraform state model for opnsense_haproxy_acl.
type ACLResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Expression   types.String `tfsdk:"expression"`
	Negate       types.Bool   `tfsdk:"negate"`
	HdrBeg       types.String `tfsdk:"hdr_beg"`
	HdrEnd       types.String `tfsdk:"hdr_end"`
	Hdr          types.String `tfsdk:"hdr"`
	PathBeg      types.String `tfsdk:"path_beg"`
	Path         types.String `tfsdk:"path"`
	SSLSNI       types.String `tfsdk:"ssl_sni"`
	SSLFcSNI     types.String `tfsdk:"ssl_fc_sni"`
	Src          types.String `tfsdk:"src"`
	NbsrvBackend types.String `tfsdk:"nbsrv_backend"`
	CustomACL    types.String `tfsdk:"custom_acl"`
}

// aclAPIResponse is the struct for unmarshaling OPNsense GET responses.
type aclAPIResponse struct {
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Expression   opnsense.SelectedMap `json:"expression"`
	Negate       string               `json:"negate"`
	HdrBeg       string               `json:"hdr_beg"`
	HdrEnd       string               `json:"hdr_end"`
	Hdr          string               `json:"hdr"`
	PathBeg      string               `json:"path_beg"`
	Path         string               `json:"path"`
	SSLSNI       string               `json:"ssl_sni"`
	SSLFcSNI     string               `json:"ssl_fc_sni"`
	Src          string               `json:"src"`
	NbsrvBackend opnsense.SelectedMap `json:"nbsrv_backend"`
	CustomACL    string               `json:"custom_acl"`
}

// aclAPIRequest is the struct for marshaling OPNsense POST requests.
type aclAPIRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Expression   string `json:"expression"`
	Negate       string `json:"negate"`
	HdrBeg       string `json:"hdr_beg"`
	HdrEnd       string `json:"hdr_end"`
	Hdr          string `json:"hdr"`
	PathBeg      string `json:"path_beg"`
	Path         string `json:"path"`
	SSLSNI       string `json:"ssl_sni"`
	SSLFcSNI     string `json:"ssl_fc_sni"`
	Src          string `json:"src"`
	NbsrvBackend string `json:"nbsrv_backend"`
	CustomACL    string `json:"custom_acl"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *ACLResourceModel) toAPI(_ context.Context) *aclAPIRequest {
	return &aclAPIRequest{
		Name:         m.Name.ValueString(),
		Description:  m.Description.ValueString(),
		Expression:   m.Expression.ValueString(),
		Negate:       opnsense.BoolToString(m.Negate.ValueBool()),
		HdrBeg:       m.HdrBeg.ValueString(),
		HdrEnd:       m.HdrEnd.ValueString(),
		Hdr:          m.Hdr.ValueString(),
		PathBeg:      m.PathBeg.ValueString(),
		Path:         m.Path.ValueString(),
		SSLSNI:       m.SSLSNI.ValueString(),
		SSLFcSNI:     m.SSLFcSNI.ValueString(),
		Src:          m.Src.ValueString(),
		NbsrvBackend: m.NbsrvBackend.ValueString(),
		CustomACL:    m.CustomACL.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *ACLResourceModel) fromAPI(_ context.Context, a *aclAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.Expression = types.StringValue(string(a.Expression))
	m.Negate = types.BoolValue(opnsense.StringToBool(a.Negate))
	m.HdrBeg = types.StringValue(a.HdrBeg)
	m.HdrEnd = types.StringValue(a.HdrEnd)
	m.Hdr = types.StringValue(a.Hdr)
	m.PathBeg = types.StringValue(a.PathBeg)
	m.Path = types.StringValue(a.Path)
	m.SSLSNI = types.StringValue(a.SSLSNI)
	m.SSLFcSNI = types.StringValue(a.SSLFcSNI)
	m.Src = types.StringValue(a.Src)
	m.NbsrvBackend = types.StringValue(string(a.NbsrvBackend))
	m.CustomACL = types.StringValue(a.CustomACL)
}
