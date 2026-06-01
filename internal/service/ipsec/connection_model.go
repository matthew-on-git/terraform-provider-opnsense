// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ipsecVersionOption is one entry of the IPsec connection `version` field, which
// the API returns as a positional array (index 0 = IKEv1+IKEv2, 1 = IKEv1,
// 2 = IKEv2) rather than a keyed select map. The selected index is the value.
type ipsecVersionOption struct {
	Selected int `json:"selected"`
}

// ConnectionResourceModel is the Terraform state model for opnsense_ipsec_connection.
type ConnectionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Description types.String `tfsdk:"description"`
	RemoteAddrs types.String `tfsdk:"remote_addrs"`
	Version     types.String `tfsdk:"version"`
	Proposals   types.String `tfsdk:"proposals"`
	Unique      types.String `tfsdk:"unique"`
}

// ipsecConnectionAPIResponse is the struct for unmarshaling OPNsense GET responses.
type ipsecConnectionAPIResponse struct {
	Enabled     string                   `json:"enabled"`
	Description string                   `json:"description"`
	RemoteAddrs opnsense.SelectedMapList `json:"remote_addrs"`
	Version     []ipsecVersionOption     `json:"version"`
	Proposals   opnsense.SelectedMapList `json:"proposals"`
	Unique      opnsense.SelectedMap     `json:"unique"`
}

// ipsecConnectionAPIRequest is the struct for marshaling OPNsense POST requests.
type ipsecConnectionAPIRequest struct {
	Enabled     string `json:"enabled"`
	Description string `json:"description"`
	RemoteAddrs string `json:"remote_addrs"`
	Version     string `json:"version"`
	Proposals   string `json:"proposals"`
	Unique      string `json:"unique"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *ConnectionResourceModel) toAPI(_ context.Context) *ipsecConnectionAPIRequest {
	return &ipsecConnectionAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Description: m.Description.ValueString(),
		RemoteAddrs: m.RemoteAddrs.ValueString(),
		Version:     m.Version.ValueString(),
		Proposals:   m.Proposals.ValueString(),
		Unique:      m.Unique.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *ConnectionResourceModel) fromAPI(_ context.Context, a *ipsecConnectionAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Description = types.StringValue(a.Description)
	m.RemoteAddrs = types.StringValue(strings.Join(a.RemoteAddrs, ","))
	version := ""
	for i, o := range a.Version {
		if o.Selected == 1 {
			version = strconv.Itoa(i)
			break
		}
	}
	m.Version = types.StringValue(version)
	m.Proposals = types.StringValue(strings.Join(a.Proposals, ","))
	m.Unique = types.StringValue(string(a.Unique))
}
