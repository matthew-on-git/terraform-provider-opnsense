// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package openvpn

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// InstanceResourceModel is the Terraform state model for opnsense_openvpn_instance.
type InstanceResourceModel struct {
	ID                types.String `tfsdk:"id"`
	VPNID             types.String `tfsdk:"vpnid"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	Role              types.String `tfsdk:"role"`
	Description       types.String `tfsdk:"description"`
	DevType           types.String `tfsdk:"dev_type"`
	Protocol          types.String `tfsdk:"protocol"`
	Port              types.String `tfsdk:"port"`
	Local             types.String `tfsdk:"local"`
	Remote            types.String `tfsdk:"remote"`
	Server            types.String `tfsdk:"server"`
	Topology          types.String `tfsdk:"topology"`
	CA                types.String `tfsdk:"ca"`
	Cert              types.String `tfsdk:"cert"`
	VerifyClientCert  types.String `tfsdk:"verify_client_cert"`
	TLSKey            types.String `tfsdk:"tls_key"`
	DataCiphers       types.Set    `tfsdk:"data_ciphers"`
	Auth              types.String `tfsdk:"auth"`
	DNSServers        types.Set    `tfsdk:"dns_servers"`
	PushRoute         types.Set    `tfsdk:"push_route"`
	RedirectGateway   types.Set    `tfsdk:"redirect_gateway"`
	MaxClients        types.Int64  `tfsdk:"max_clients"`
	KeepaliveInterval types.Int64  `tfsdk:"keepalive_interval"`
	KeepaliveTimeout  types.Int64  `tfsdk:"keepalive_timeout"`
	Verb              types.String `tfsdk:"verb"`
}

type instanceAPIResponse struct {
	VPNID             string                   `json:"vpnid"`
	VerifyClientCert  opnsense.SelectedMap     `json:"verify_client_cert"`
	Enabled           string                   `json:"enabled"`
	Role              opnsense.SelectedMap     `json:"role"`
	Description       string                   `json:"description"`
	DevType           opnsense.SelectedMap     `json:"dev_type"`
	Protocol          opnsense.SelectedMap     `json:"proto"`
	Port              string                   `json:"port"`
	Local             string                   `json:"local"`
	Remote            opnsense.SelectedMap     `json:"remote"`
	Server            string                   `json:"server"`
	Topology          opnsense.SelectedMap     `json:"topology"`
	CA                opnsense.SelectedMap     `json:"ca"`
	Cert              opnsense.SelectedMap     `json:"cert"`
	TLSKey            opnsense.SelectedMap     `json:"tls_key"`
	DataCiphers       opnsense.SelectedMapList `json:"data-ciphers"`
	Auth              opnsense.SelectedMap     `json:"auth"`
	DNSServers        opnsense.SelectedMapList `json:"dns_servers"`
	PushRoute         opnsense.SelectedMapList `json:"push_route"`
	RedirectGateway   opnsense.SelectedMapList `json:"redirect_gateway"`
	MaxClients        string                   `json:"maxclients"`
	KeepaliveInterval string                   `json:"keepalive_interval"`
	KeepaliveTimeout  string                   `json:"keepalive_timeout"`
	Verb              opnsense.SelectedMap     `json:"verb"`
}

type instanceAPIRequest struct {
	VPNID             string `json:"vpnid"`
	VerifyClientCert  string `json:"verify_client_cert"`
	Enabled           string `json:"enabled"`
	Role              string `json:"role"`
	Description       string `json:"description"`
	DevType           string `json:"dev_type"`
	Protocol          string `json:"proto"`
	Port              string `json:"port"`
	Local             string `json:"local"`
	Remote            string `json:"remote"`
	Server            string `json:"server"`
	Topology          string `json:"topology"`
	CA                string `json:"ca"`
	Cert              string `json:"cert"`
	TLSKey            string `json:"tls_key"`
	DataCiphers       string `json:"data-ciphers"`
	Auth              string `json:"auth"`
	DNSServers        string `json:"dns_servers"`
	PushRoute         string `json:"push_route"`
	RedirectGateway   string `json:"redirect_gateway"`
	MaxClients        string `json:"maxclients"`
	KeepaliveInterval string `json:"keepalive_interval"`
	KeepaliveTimeout  string `json:"keepalive_timeout"`
	Verb              string `json:"verb,omitempty"`
}

// setToCSV joins a Terraform string set into a comma-separated string for the API.
func setToCSV(ctx context.Context, s types.Set) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	var elems []string
	s.ElementsAs(ctx, &elems, false)
	return strings.Join(elems, ",")
}

// sliceToSet converts a string slice into a Terraform string set (never null).
func sliceToSet(items []string) types.Set {
	if len(items) == 0 {
		return types.SetValueMust(types.StringType, []attr.Value{})
	}
	vals := make([]attr.Value, len(items))
	for i, v := range items {
		vals[i] = types.StringValue(v)
	}
	return types.SetValueMust(types.StringType, vals)
}

func (m *InstanceResourceModel) toAPI(ctx context.Context) *instanceAPIRequest {
	return &instanceAPIRequest{
		Enabled:           opnsense.BoolToString(m.Enabled.ValueBool()),
		Role:              m.Role.ValueString(),
		VPNID:             m.VPNID.ValueString(),
		VerifyClientCert:  m.VerifyClientCert.ValueString(),
		Description:       m.Description.ValueString(),
		DevType:           m.DevType.ValueString(),
		Protocol:          m.Protocol.ValueString(),
		Port:              m.Port.ValueString(),
		Local:             m.Local.ValueString(),
		Remote:            m.Remote.ValueString(),
		Server:            m.Server.ValueString(),
		Topology:          m.Topology.ValueString(),
		CA:                m.CA.ValueString(),
		Cert:              m.Cert.ValueString(),
		TLSKey:            m.TLSKey.ValueString(),
		DataCiphers:       setToCSV(ctx, m.DataCiphers),
		Auth:              m.Auth.ValueString(),
		DNSServers:        setToCSV(ctx, m.DNSServers),
		PushRoute:         setToCSV(ctx, m.PushRoute),
		RedirectGateway:   setToCSV(ctx, m.RedirectGateway),
		MaxClients:        int64ToStringOrEmpty(m.MaxClients.ValueInt64()),
		KeepaliveInterval: int64ToStringOrEmpty(m.KeepaliveInterval.ValueInt64()),
		KeepaliveTimeout:  int64ToStringOrEmpty(m.KeepaliveTimeout.ValueInt64()),
		Verb:              m.Verb.ValueString(),
	}
}

// int64ToStringOrEmpty returns "" for zero so optional integer fields are not
// forced to "0" on the API side.
func int64ToStringOrEmpty(n int64) string {
	if n == 0 {
		return ""
	}
	return opnsense.Int64ToString(n)
}

func (m *InstanceResourceModel) fromAPI(_ context.Context, a *instanceAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.VPNID = types.StringValue(a.VPNID)
	m.VerifyClientCert = types.StringValue(string(a.VerifyClientCert))
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Role = types.StringValue(string(a.Role))
	m.Description = types.StringValue(a.Description)
	m.DevType = types.StringValue(string(a.DevType))
	m.Protocol = types.StringValue(string(a.Protocol))
	m.Port = types.StringValue(a.Port)
	m.Local = types.StringValue(a.Local)
	m.Remote = types.StringValue(string(a.Remote))
	m.Server = types.StringValue(a.Server)
	m.Topology = types.StringValue(string(a.Topology))
	m.CA = types.StringValue(string(a.CA))
	m.Cert = types.StringValue(string(a.Cert))
	m.TLSKey = types.StringValue(string(a.TLSKey))
	m.Auth = types.StringValue(string(a.Auth))
	m.Verb = types.StringValue(string(a.Verb))

	m.DataCiphers = sliceToSet(a.DataCiphers)
	m.RedirectGateway = sliceToSet(a.RedirectGateway)
	m.DNSServers = sliceToSet(a.DNSServers)
	m.PushRoute = sliceToSet(a.PushRoute)

	m.MaxClients = types.Int64Value(stringToInt64OrZero(a.MaxClients))
	m.KeepaliveInterval = types.Int64Value(stringToInt64OrZero(a.KeepaliveInterval))
	m.KeepaliveTimeout = types.Int64Value(stringToInt64OrZero(a.KeepaliveTimeout))
}

// stringToInt64OrZero parses an OPNsense numeric string, returning 0 when empty
// or unparseable.
func stringToInt64OrZero(s string) int64 {
	if s == "" {
		return 0
	}
	if v, err := opnsense.StringToInt64(s); err == nil {
		return v
	}
	return 0
}
