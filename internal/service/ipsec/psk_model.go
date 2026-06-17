// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// PSKResourceModel is the Terraform state model for opnsense_ipsec_psk.
type PSKResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Identity       types.String `tfsdk:"identity"`
	RemoteIdentity types.String `tfsdk:"remote_identity"`
	KeyType        types.String `tfsdk:"key_type"`
	Key            types.String `tfsdk:"key"`
	Description    types.String `tfsdk:"description"`
}

// ipsecPSKAPIResponse is the struct for unmarshaling OPNsense GET responses.
type ipsecPSKAPIResponse struct {
	Identity       string               `json:"ident"`
	RemoteIdentity string               `json:"remote_ident"`
	KeyType        opnsense.SelectedMap `json:"keyType"`
	Description    string               `json:"description"`
}

// ipsecPSKAPIRequest is the struct for marshaling OPNsense POST requests.
type ipsecPSKAPIRequest struct {
	Identity       string `json:"ident"`
	RemoteIdentity string `json:"remote_ident"`
	KeyType        string `json:"keyType"`
	Key            string `json:"Key"`
	Description    string `json:"description"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *PSKResourceModel) toAPI(_ context.Context) *ipsecPSKAPIRequest {
	return &ipsecPSKAPIRequest{
		Identity:       m.Identity.ValueString(),
		RemoteIdentity: m.RemoteIdentity.ValueString(),
		KeyType:        m.KeyType.ValueString(),
		Key:            m.Key.ValueString(),
		Description:    m.Description.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
// Key is write-only and not populated from API responses.
func (m *PSKResourceModel) fromAPI(_ context.Context, a *ipsecPSKAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Identity = types.StringValue(a.Identity)
	m.RemoteIdentity = types.StringValue(a.RemoteIdentity)
	m.KeyType = types.StringValue(string(a.KeyType))
	m.Description = types.StringValue(a.Description)
}
