// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// DHCPv6SubnetResourceModel is the Terraform state model for opnsense_kea_dhcpv6_subnet.
type DHCPv6SubnetResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Subnet      types.String `tfsdk:"subnet"`
	Interface   types.String `tfsdk:"interface"`
	Allocator   types.String `tfsdk:"allocator"`
	PDAllocator types.String `tfsdk:"pd_allocator"`
	Pools       types.String `tfsdk:"pools"`
	Description types.String `tfsdk:"description"`
}

type dhcpv6SubnetAPIResponse struct {
	Subnet      string               `json:"subnet"`
	Interface   opnsense.SelectedMap `json:"interface"`
	Allocator   opnsense.SelectedMap `json:"allocator"`
	PDAllocator opnsense.SelectedMap `json:"pd-allocator"`
	Pools       string               `json:"pools"`
	Description string               `json:"description"`
}

type dhcpv6SubnetAPIRequest struct {
	Subnet      string `json:"subnet"`
	Interface   string `json:"interface"`
	Allocator   string `json:"allocator"`
	PDAllocator string `json:"pd-allocator"`
	Pools       string `json:"pools"`
	Description string `json:"description"`
}

func (m *DHCPv6SubnetResourceModel) toAPI(_ context.Context) *dhcpv6SubnetAPIRequest {
	return &dhcpv6SubnetAPIRequest{
		Subnet:      m.Subnet.ValueString(),
		Interface:   m.Interface.ValueString(),
		Allocator:   m.Allocator.ValueString(),
		PDAllocator: m.PDAllocator.ValueString(),
		Pools:       m.Pools.ValueString(),
		Description: m.Description.ValueString(),
	}
}

func (m *DHCPv6SubnetResourceModel) fromAPI(_ context.Context, a *dhcpv6SubnetAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Subnet = types.StringValue(a.Subnet)
	m.Interface = types.StringValue(string(a.Interface))
	m.Allocator = types.StringValue(string(a.Allocator))
	m.PDAllocator = types.StringValue(string(a.PDAllocator))
	m.Pools = types.StringValue(a.Pools)
	m.Description = types.StringValue(a.Description)
}
