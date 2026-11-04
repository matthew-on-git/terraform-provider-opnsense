// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package ipsec

import (
	"context"
	"strconv"

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
	ID                     types.String `tfsdk:"id"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	Description            types.String `tfsdk:"description"`
	RemoteAddrs            types.String `tfsdk:"remote_addrs"`
	Version                types.String `tfsdk:"version"`
	Proposals              types.String `tfsdk:"proposals"`
	Unique                 types.String `tfsdk:"unique"`
	Aggressive             types.Bool   `tfsdk:"aggressive"`
	DPDDelay               types.String `tfsdk:"dpd_delay"`
	DPDTimeout             types.String `tfsdk:"dpd_timeout"`
	IKELifetime            types.String `tfsdk:"ike_lifetime"`
	IPPools                types.String `tfsdk:"ip_pools"`
	KeyingTries            types.String `tfsdk:"keying_tries"`
	LocalAddresses         types.String `tfsdk:"local_addresses"`
	LocalPort              types.String `tfsdk:"local_port"`
	Mobike                 types.Bool   `tfsdk:"mobike"`
	ReauthenticationTime   types.String `tfsdk:"reauthentication_time"`
	RekeyTime              types.String `tfsdk:"rekey_time"`
	RemotePort             types.String `tfsdk:"remote_port"`
	SendCertificate        types.String `tfsdk:"send_certificate"`
	SendCertificateRequest types.Bool   `tfsdk:"send_certificate_request"`
	UDPEncapsulation       types.Bool   `tfsdk:"udp_encapsulation"`
}

// ipsecConnectionAPIResponse is the struct for unmarshaling OPNsense GET responses.
type ipsecConnectionAPIResponse struct {
	Enabled                string               `json:"enabled"`
	Description            string               `json:"description"`
	RemoteAddrs            ipsecSelection       `json:"remote_addrs"`
	Version                []ipsecVersionOption `json:"version"`
	Proposals              ipsecSelection       `json:"proposals"`
	Unique                 ipsecEnum            `json:"unique"`
	Aggressive             string               `json:"aggressive"`
	DPDDelay               string               `json:"dpd_delay"`
	DPDTimeout             string               `json:"dpd_timeout"`
	IKELifetime            string               `json:"over_time"`
	IPPools                ipsecSelection       `json:"pools"`
	KeyingTries            string               `json:"keyingtries"`
	LocalAddresses         ipsecSelection       `json:"local_addrs"`
	LocalPort              ipsecEnum            `json:"local_port"`
	Mobike                 string               `json:"mobike"`
	ReauthenticationTime   string               `json:"reauth_time"`
	RekeyTime              string               `json:"rekey_time"`
	RemotePort             ipsecEnum            `json:"remote_port"`
	SendCertificate        ipsecEnum            `json:"send_cert"`
	SendCertificateRequest string               `json:"send_certreq"`
	UDPEncapsulation       string               `json:"encap"`
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
	m.RemoteAddrs = stableIPsecCSV(m.RemoteAddrs, a.RemoteAddrs)
	version := ""
	for i, o := range a.Version {
		if o.Selected == 1 {
			version = strconv.Itoa(i)
			break
		}
	}
	m.Version = types.StringValue(version)
	m.Proposals = stableIPsecCSV(m.Proposals, a.Proposals)
	m.Unique = types.StringValue(string(a.Unique))
	m.Aggressive = types.BoolValue(opnsense.StringToBool(a.Aggressive))
	m.DPDDelay = types.StringValue(string(a.DPDDelay))
	m.DPDTimeout = types.StringValue(string(a.DPDTimeout))
	m.IKELifetime = types.StringValue(string(a.IKELifetime))
	m.IPPools = stableIPsecCSV(m.IPPools, a.IPPools)
	m.KeyingTries = types.StringValue(string(a.KeyingTries))
	m.LocalAddresses = stableIPsecCSV(m.LocalAddresses, a.LocalAddresses)
	m.LocalPort = types.StringValue(string(a.LocalPort))
	m.Mobike = types.BoolValue(opnsense.StringToBool(a.Mobike))
	m.ReauthenticationTime = types.StringValue(string(a.ReauthenticationTime))
	m.RekeyTime = types.StringValue(string(a.RekeyTime))
	m.RemotePort = types.StringValue(string(a.RemotePort))
	m.SendCertificate = types.StringValue(string(a.SendCertificate))
	m.SendCertificateRequest = types.BoolValue(opnsense.StringToBool(a.SendCertificateRequest))
	m.UDPEncapsulation = types.BoolValue(opnsense.StringToBool(a.UDPEncapsulation))
}

func (a *ipsecConnectionAPIResponse) UnmarshalJSON(data []byte) error {
	type plain ipsecConnectionAPIResponse
	var value plain
	if err := decodeIPsecResponse(data, &value, connectionAdvancedFields); err != nil {
		return err
	}
	*a = ipsecConnectionAPIResponse(value)
	return nil
}

// A saved object must retain a usable identity even when activation/readback fails.
func (m *ConnectionResourceModel) clearUnknownAdvanced() {
	if m.Aggressive.IsUnknown() {
		m.Aggressive = types.BoolNull()
	}
	if m.DPDDelay.IsUnknown() {
		m.DPDDelay = types.StringNull()
	}
	if m.DPDTimeout.IsUnknown() {
		m.DPDTimeout = types.StringNull()
	}
	if m.IKELifetime.IsUnknown() {
		m.IKELifetime = types.StringNull()
	}
	if m.IPPools.IsUnknown() {
		m.IPPools = types.StringNull()
	}
	if m.KeyingTries.IsUnknown() {
		m.KeyingTries = types.StringNull()
	}
	if m.LocalAddresses.IsUnknown() {
		m.LocalAddresses = types.StringNull()
	}
	if m.LocalPort.IsUnknown() {
		m.LocalPort = types.StringNull()
	}
	if m.Mobike.IsUnknown() {
		m.Mobike = types.BoolNull()
	}
	if m.ReauthenticationTime.IsUnknown() {
		m.ReauthenticationTime = types.StringNull()
	}
	if m.RekeyTime.IsUnknown() {
		m.RekeyTime = types.StringNull()
	}
	if m.RemotePort.IsUnknown() {
		m.RemotePort = types.StringNull()
	}
	if m.SendCertificate.IsUnknown() {
		m.SendCertificate = types.StringNull()
	}
	if m.SendCertificateRequest.IsUnknown() {
		m.SendCertificateRequest = types.BoolNull()
	}
	if m.UDPEncapsulation.IsUnknown() {
		m.UDPEncapsulation = types.BoolNull()
	}
}
