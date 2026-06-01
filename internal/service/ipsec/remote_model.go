// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// RemoteResourceModel is the Terraform state model for opnsense_ipsec_remote.
type RemoteResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Connection  types.String `tfsdk:"connection_id"`
	Round       types.Int64  `tfsdk:"round"`
	Auth        types.String `tfsdk:"auth"`
	Identity    types.String `tfsdk:"identity"`
	EAPID       types.String `tfsdk:"eap_id"`
	Groups      types.Set    `tfsdk:"groups"`
	Certs       types.Set    `tfsdk:"certs"`
	CACerts     types.Set    `tfsdk:"cacerts"`
	Description types.String `tfsdk:"description"`
}

// remoteAPIResponse is the struct for unmarshaling OPNsense GET responses.
type remoteAPIResponse struct {
	Enabled     string                   `json:"enabled"`
	Connection  opnsense.SelectedMap     `json:"connection"`
	Round       string                   `json:"round"`
	Auth        opnsense.SelectedMap     `json:"auth"`
	Identity    string                   `json:"id"`
	EAPID       string                   `json:"eap_id"`
	Groups      opnsense.SelectedMapList `json:"groups"`
	Certs       opnsense.SelectedMapList `json:"certs"`
	CACerts     opnsense.SelectedMapList `json:"cacerts"`
	Description string                   `json:"description"`
}

// remoteAPIRequest is the struct for marshaling OPNsense POST requests.
type remoteAPIRequest struct {
	Enabled     string `json:"enabled"`
	Connection  string `json:"connection"`
	Round       string `json:"round"`
	Auth        string `json:"auth"`
	Identity    string `json:"id"`
	EAPID       string `json:"eap_id"`
	Groups      string `json:"groups"`
	Certs       string `json:"certs"`
	CACerts     string `json:"cacerts"`
	Description string `json:"description"`
}

func setToCSVList(ctx context.Context, s types.Set) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	var elements []string
	s.ElementsAs(ctx, &elements, false)
	return strings.Join(elements, ",")
}

func selectedListToSet(list opnsense.SelectedMapList) types.Set {
	if len(list) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	vals := make([]attr.Value, len(list))
	for i, v := range list {
		vals[i] = types.StringValue(v)
	}
	return types.SetValueMust(types.StringType, vals)
}

// toAPI converts the Terraform model to an API request struct.
func (m *RemoteResourceModel) toAPI(ctx context.Context) *remoteAPIRequest {
	return &remoteAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Connection:  m.Connection.ValueString(),
		Round:       opnsense.Int64ToString(m.Round.ValueInt64()),
		Auth:        m.Auth.ValueString(),
		Identity:    m.Identity.ValueString(),
		EAPID:       m.EAPID.ValueString(),
		Groups:      setToCSVList(ctx, m.Groups),
		Certs:       setToCSVList(ctx, m.Certs),
		CACerts:     setToCSVList(ctx, m.CACerts),
		Description: m.Description.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *RemoteResourceModel) fromAPI(_ context.Context, a *remoteAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Connection = types.StringValue(string(a.Connection))
	m.Auth = types.StringValue(string(a.Auth))
	m.Identity = types.StringValue(a.Identity)
	m.EAPID = types.StringValue(a.EAPID)
	m.Description = types.StringValue(a.Description)

	if a.Round != "" {
		if v, err := opnsense.StringToInt64(a.Round); err == nil {
			m.Round = types.Int64Value(v)
		}
	} else {
		m.Round = types.Int64Value(0)
	}

	m.Groups = selectedListToSet(a.Groups)
	m.Certs = selectedListToSet(a.Certs)
	m.CACerts = selectedListToSet(a.CACerts)
}
