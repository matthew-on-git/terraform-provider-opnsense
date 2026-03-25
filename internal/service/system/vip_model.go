// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// VipResourceModel is the Terraform state model for opnsense_system_vip.
type VipResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Interface   types.String `tfsdk:"interface"`
	Mode        types.String `tfsdk:"mode"`
	Address     types.String `tfsdk:"address"`
	SubnetBits  types.Int64  `tfsdk:"subnet_bits"`
	Description types.String `tfsdk:"description"`
	VHID        types.Int64  `tfsdk:"vhid"`
	Password    types.String `tfsdk:"password"`
	AdvBase     types.Int64  `tfsdk:"adv_base"`
	AdvSkew     types.Int64  `tfsdk:"adv_skew"`
}

type vipAPIResponse struct {
	Interface   opnsense.SelectedMap `json:"interface"`
	Mode        opnsense.SelectedMap `json:"mode"`
	Address     string               `json:"subnet"`
	SubnetBits  string               `json:"subnet_bits"`
	Description string               `json:"descr"`
	VHID        string               `json:"vhid"`
	Password    string               `json:"password"`
	AdvBase     string               `json:"advbase"`
	AdvSkew     string               `json:"advskew"`
}

type vipAPIRequest struct {
	Interface   string `json:"interface"`
	Mode        string `json:"mode"`
	Address     string `json:"subnet"`
	SubnetBits  string `json:"subnet_bits"`
	Description string `json:"descr"`
	VHID        string `json:"vhid"`
	Password    string `json:"password"`
	AdvBase     string `json:"advbase"`
	AdvSkew     string `json:"advskew"`
}

func (m *VipResourceModel) toAPI(_ context.Context) *vipAPIRequest {
	var vhidStr string
	if !m.VHID.IsNull() && !m.VHID.IsUnknown() {
		vhidStr = opnsense.Int64ToString(m.VHID.ValueInt64())
	}

	return &vipAPIRequest{
		Interface:   m.Interface.ValueString(),
		Mode:        m.Mode.ValueString(),
		Address:     m.Address.ValueString(),
		SubnetBits:  opnsense.Int64ToString(m.SubnetBits.ValueInt64()),
		Description: m.Description.ValueString(),
		VHID:        vhidStr,
		Password:    m.Password.ValueString(),
		AdvBase:     opnsense.Int64ToString(m.AdvBase.ValueInt64()),
		AdvSkew:     opnsense.Int64ToString(m.AdvSkew.ValueInt64()),
	}
}

func (m *VipResourceModel) fromAPI(_ context.Context, a *vipAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Interface = types.StringValue(string(a.Interface))
	m.Mode = types.StringValue(string(a.Mode))
	m.Address = types.StringValue(a.Address)
	m.Description = types.StringValue(a.Description)
	m.Password = types.StringValue(a.Password)

	if a.SubnetBits != "" {
		if v, err := opnsense.StringToInt64(a.SubnetBits); err == nil {
			m.SubnetBits = types.Int64Value(v)
		}
	}
	if a.VHID != "" {
		if v, err := opnsense.StringToInt64(a.VHID); err == nil {
			m.VHID = types.Int64Value(v)
		}
	} else {
		m.VHID = types.Int64Null()
	}
	if a.AdvBase != "" {
		if v, err := opnsense.StringToInt64(a.AdvBase); err == nil {
			m.AdvBase = types.Int64Value(v)
		}
	}
	if a.AdvSkew != "" {
		if v, err := opnsense.StringToInt64(a.AdvSkew); err == nil {
			m.AdvSkew = types.Int64Value(v)
		}
	}
}
