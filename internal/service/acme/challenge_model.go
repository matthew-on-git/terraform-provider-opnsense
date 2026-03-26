// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package acme

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// ChallengeResourceModel is the Terraform state model for opnsense_acme_challenge.
type ChallengeResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Method      types.String `tfsdk:"method"`
	DNSService  types.String `tfsdk:"dns_service"`
	DNSSleep    types.Int64  `tfsdk:"dns_sleep"`
}

type challengeAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Method      opnsense.SelectedMap `json:"method"`
	DNSService  opnsense.SelectedMap `json:"dns_service"`
	DNSSleep    string               `json:"dns_sleep"`
}

type challengeAPIRequest struct {
	Enabled     string `json:"enabled"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Method      string `json:"method"`
	DNSService  string `json:"dns_service"`
	DNSSleep    string `json:"dns_sleep"`
}

func (m *ChallengeResourceModel) toAPI(_ context.Context) *challengeAPIRequest {
	var sleepStr string
	if !m.DNSSleep.IsNull() && !m.DNSSleep.IsUnknown() {
		sleepStr = opnsense.Int64ToString(m.DNSSleep.ValueInt64())
	}

	return &challengeAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Method:      m.Method.ValueString(),
		DNSService:  m.DNSService.ValueString(),
		DNSSleep:    sleepStr,
	}
}

func (m *ChallengeResourceModel) fromAPI(_ context.Context, a *challengeAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.Method = types.StringValue(string(a.Method))
	m.DNSService = types.StringValue(string(a.DNSService))

	if a.DNSSleep != "" {
		if v, err := opnsense.StringToInt64(a.DNSSleep); err == nil {
			m.DNSSleep = types.Int64Value(v)
		}
	} else {
		m.DNSSleep = types.Int64Value(0)
	}
}
