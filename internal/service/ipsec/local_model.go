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

// LocalResourceModel is the Terraform state model for opnsense_ipsec_local.
type LocalResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Connection  types.String `tfsdk:"connection_id"`
	Round       types.Int64  `tfsdk:"round"`
	Auth        types.String `tfsdk:"auth"`
	Identity    types.String `tfsdk:"identity"`
	EAPID       types.String `tfsdk:"eap_id"`
	Certs       types.Set    `tfsdk:"certs"`
	Description types.String `tfsdk:"description"`
}

// localAPIResponse is the struct for unmarshaling OPNsense GET responses.
type localAPIResponse struct {
	Enabled     string                   `json:"enabled"`
	Connection  opnsense.SelectedMap     `json:"connection"`
	Round       string                   `json:"round"`
	Auth        opnsense.SelectedMap     `json:"auth"`
	Identity    string                   `json:"id"`
	EAPID       string                   `json:"eap_id"`
	Certs       opnsense.SelectedMapList `json:"certs"`
	Description string                   `json:"description"`
}

// localAPIRequest is the struct for marshaling OPNsense POST requests.
type localAPIRequest struct {
	Enabled     string `json:"enabled"`
	Connection  string `json:"connection"`
	Round       string `json:"round"`
	Auth        string `json:"auth"`
	Identity    string `json:"id"`
	EAPID       string `json:"eap_id"`
	Certs       string `json:"certs"`
	Description string `json:"description"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *LocalResourceModel) toAPI(ctx context.Context) *localAPIRequest {
	var certsStr string
	if !m.Certs.IsNull() && !m.Certs.IsUnknown() {
		var elements []string
		m.Certs.ElementsAs(ctx, &elements, false)
		certsStr = strings.Join(elements, ",")
	}

	return &localAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Connection:  m.Connection.ValueString(),
		Round:       opnsense.Int64ToString(m.Round.ValueInt64()),
		Auth:        m.Auth.ValueString(),
		Identity:    m.Identity.ValueString(),
		EAPID:       m.EAPID.ValueString(),
		Certs:       certsStr,
		Description: m.Description.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *LocalResourceModel) fromAPI(_ context.Context, a *localAPIResponse, uuid string) {
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

	if len(a.Certs) == 0 {
		m.Certs = types.SetValueMust(types.StringType, []attr.Value{})
	} else {
		vals := make([]attr.Value, len(a.Certs))
		for i, v := range a.Certs {
			vals[i] = types.StringValue(v)
		}
		m.Certs = types.SetValueMust(types.StringType, vals)
	}
}
