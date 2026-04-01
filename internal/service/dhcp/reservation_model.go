// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package dhcp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ReservationResourceModel is the Terraform state model for opnsense_dhcpv4_reservation.
type ReservationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Subnet      types.String `tfsdk:"subnet"`
	IPAddress   types.String `tfsdk:"ip_address"`
	MACAddress  types.String `tfsdk:"mac_address"`
	Hostname    types.String `tfsdk:"hostname"`
	Description types.String `tfsdk:"description"`
}

type reservationAPIResponse struct {
	Subnet      string `json:"subnet"`
	IPAddress   string `json:"ip_address"`
	MACAddress  string `json:"hw_address"`
	Hostname    string `json:"hostname"`
	Description string `json:"description"`
}

type reservationAPIRequest struct {
	Subnet      string `json:"subnet"`
	IPAddress   string `json:"ip_address"`
	MACAddress  string `json:"hw_address"`
	Hostname    string `json:"hostname"`
	Description string `json:"description"`
}

func (m *ReservationResourceModel) toAPI(_ context.Context) *reservationAPIRequest {
	return &reservationAPIRequest{
		Subnet:      m.Subnet.ValueString(),
		IPAddress:   m.IPAddress.ValueString(),
		MACAddress:  m.MACAddress.ValueString(),
		Hostname:    m.Hostname.ValueString(),
		Description: m.Description.ValueString(),
	}
}

func (m *ReservationResourceModel) fromAPI(_ context.Context, a *reservationAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Subnet = types.StringValue(a.Subnet)
	m.IPAddress = types.StringValue(a.IPAddress)
	m.MACAddress = types.StringValue(a.MACAddress)
	m.Hostname = types.StringValue(a.Hostname)
	m.Description = types.StringValue(a.Description)
}
