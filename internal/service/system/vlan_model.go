// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// VlanResourceModel is the Terraform state model for opnsense_system_vlan.
type VlanResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ParentInterface types.String `tfsdk:"parent_interface"`
	Tag             types.Int64  `tfsdk:"tag"`
	Priority        types.Int64  `tfsdk:"priority"`
	Proto           types.String `tfsdk:"proto"`
	Description     types.String `tfsdk:"description"`
	Device          types.String `tfsdk:"device"`
}

type vlanAPIResponse struct {
	ParentInterface opnsense.SelectedMap `json:"if"`
	Tag             string               `json:"tag"`
	Priority        opnsense.SelectedMap `json:"pcp"`
	Proto           opnsense.SelectedMap `json:"proto"`
	Description     string               `json:"descr"`
	Device          string               `json:"vlanif"`
}

type vlanAPIRequest struct {
	ParentInterface string `json:"if"`
	Tag             string `json:"tag"`
	Priority        string `json:"pcp"`
	Proto           string `json:"proto"`
	Description     string `json:"descr"`
	Device          string `json:"vlanif"`
}

func (m *VlanResourceModel) toAPI(_ context.Context) *vlanAPIRequest {
	return &vlanAPIRequest{
		ParentInterface: m.ParentInterface.ValueString(),
		Tag:             opnsense.Int64ToString(m.Tag.ValueInt64()),
		Priority:        opnsense.Int64ToString(m.Priority.ValueInt64()),
		Proto:           m.Proto.ValueString(),
		Description:     m.Description.ValueString(),
		Device:          m.Device.ValueString(),
	}
}

func (m *VlanResourceModel) fromAPI(_ context.Context, a *vlanAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.ParentInterface = types.StringValue(string(a.ParentInterface))
	m.Description = types.StringValue(a.Description)
	m.Device = types.StringValue(a.Device)
	m.Proto = types.StringValue(string(a.Proto))
	m.Priority = types.Int64Value(0)

	if a.Tag != "" {
		if v, err := opnsense.StringToInt64(a.Tag); err == nil {
			m.Tag = types.Int64Value(v)
		}
	}
	if pcp := string(a.Priority); pcp != "" {
		if v, err := opnsense.StringToInt64(pcp); err == nil {
			m.Priority = types.Int64Value(v)
		}
	}
}
