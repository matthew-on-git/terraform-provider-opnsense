// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package kea

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/internal/tfconv"
	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// DHCPv6ReservationResourceModel is the Terraform state model for
// opnsense_kea_dhcpv6_reservation.
type DHCPv6ReservationResourceModel struct {
	ID           types.String `tfsdk:"id"`
	SubnetID     types.String `tfsdk:"subnet_id"`
	IPAddress    types.String `tfsdk:"ip_address"`
	DUID         types.String `tfsdk:"duid"`
	Hostname     types.String `tfsdk:"hostname"`
	DomainSearch types.Set    `tfsdk:"domain_search"`
	Description  types.String `tfsdk:"description"`
}

type dhcpv6ReservationAPIResponse struct {
	Subnet       opnsense.SelectedMap     `json:"subnet"`
	IPAddress    string                   `json:"ip_address"`
	DUID         string                   `json:"duid"`
	Hostname     string                   `json:"hostname"`
	DomainSearch opnsense.SelectedMapList `json:"domain_search"`
	Description  string                   `json:"description"`
}

type dhcpv6ReservationAPIRequest struct {
	Subnet       string `json:"subnet"`
	IPAddress    string `json:"ip_address"`
	DUID         string `json:"duid"`
	Hostname     string `json:"hostname"`
	DomainSearch string `json:"domain_search"`
	Description  string `json:"description"`
}

func (m *DHCPv6ReservationResourceModel) toAPI(ctx context.Context) *dhcpv6ReservationAPIRequest {
	return &dhcpv6ReservationAPIRequest{
		Subnet:       m.SubnetID.ValueString(),
		IPAddress:    m.IPAddress.ValueString(),
		DUID:         m.DUID.ValueString(),
		Hostname:     m.Hostname.ValueString(),
		DomainSearch: tfconv.SetToCSV(ctx, m.DomainSearch),
		Description:  m.Description.ValueString(),
	}
}

func (m *DHCPv6ReservationResourceModel) fromAPI(_ context.Context, a *dhcpv6ReservationAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.SubnetID = types.StringValue(string(a.Subnet))
	m.IPAddress = types.StringValue(a.IPAddress)
	m.DUID = types.StringValue(a.DUID)
	m.Hostname = types.StringValue(a.Hostname)
	m.DomainSearch = tfconv.SliceToSet(a.DomainSearch)
	m.Description = types.StringValue(a.Description)
}
