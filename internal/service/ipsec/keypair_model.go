// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// KeyPairResourceModel is the Terraform state model for opnsense_ipsec_key_pair.
// The public and private key PEM payloads are write-only (the API re-encodes
// them), so they are kept from configuration rather than refreshed from state.
type KeyPairResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	KeyType        types.String `tfsdk:"key_type"`
	PublicKey      types.String `tfsdk:"public_key"`
	PrivateKey     types.String `tfsdk:"private_key"`
	KeySize        types.String `tfsdk:"key_size"`
	KeyFingerprint types.String `tfsdk:"key_fingerprint"`
}

type keyPairAPIResponse struct {
	Name           string               `json:"name"`
	KeyType        opnsense.SelectedMap `json:"keyType"`
	KeySize        string               `json:"keySize"`
	KeyFingerprint string               `json:"keyFingerprint"`
}

type keyPairAPIRequest struct {
	Name       string `json:"name"`
	KeyType    string `json:"keyType"`
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

func (m *KeyPairResourceModel) toAPI(_ context.Context) *keyPairAPIRequest {
	return &keyPairAPIRequest{
		Name:       m.Name.ValueString(),
		KeyType:    m.KeyType.ValueString(),
		PublicKey:  m.PublicKey.ValueString(),
		PrivateKey: m.PrivateKey.ValueString(),
	}
}

func (m *KeyPairResourceModel) fromAPI(_ context.Context, a *keyPairAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Name = types.StringValue(a.Name)
	m.KeyType = types.StringValue(string(a.KeyType))
	m.KeySize = types.StringValue(a.KeySize)
	m.KeyFingerprint = types.StringValue(a.KeyFingerprint)
	// PublicKey and PrivateKey are write-only: preserved from configuration.
}
