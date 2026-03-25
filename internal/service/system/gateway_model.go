// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// GatewayResourceModel is the Terraform state model for opnsense_system_gateway.
type GatewayResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Interface      types.String `tfsdk:"interface"`
	IPProtocol     types.String `tfsdk:"ip_protocol"`
	Gateway        types.String `tfsdk:"gateway"`
	DefaultGateway types.Bool   `tfsdk:"default_gateway"`
	MonitorDisable types.Bool   `tfsdk:"monitor_disable"`
	Weight         types.Int64  `tfsdk:"weight"`
	Priority       types.Int64  `tfsdk:"priority"`
}

type gatewayAPIResponse struct {
	Disabled       string               `json:"disabled"`
	Name           string               `json:"name"`
	Description    string               `json:"descr"`
	Interface      opnsense.SelectedMap `json:"interface"`
	IPProtocol     opnsense.SelectedMap `json:"ipprotocol"`
	Gateway        string               `json:"gateway"`
	DefaultGateway string               `json:"defaultgw"`
	MonitorDisable string               `json:"monitor_disable"`
	Weight         string               `json:"weight"`
	Priority       string               `json:"priority"`
}

type gatewayAPIRequest struct {
	Disabled       string `json:"disabled"`
	Name           string `json:"name"`
	Description    string `json:"descr"`
	Interface      string `json:"interface"`
	IPProtocol     string `json:"ipprotocol"`
	Gateway        string `json:"gateway"`
	DefaultGateway string `json:"defaultgw"`
	MonitorDisable string `json:"monitor_disable"`
	Weight         string `json:"weight"`
	Priority       string `json:"priority"`
}

func (m *GatewayResourceModel) toAPI(_ context.Context) *gatewayAPIRequest {
	return &gatewayAPIRequest{
		Disabled:       opnsense.BoolToString(!m.Enabled.ValueBool()),
		Name:           m.Name.ValueString(),
		Description:    m.Description.ValueString(),
		Interface:      m.Interface.ValueString(),
		IPProtocol:     m.IPProtocol.ValueString(),
		Gateway:        m.Gateway.ValueString(),
		DefaultGateway: opnsense.BoolToString(m.DefaultGateway.ValueBool()),
		MonitorDisable: opnsense.BoolToString(m.MonitorDisable.ValueBool()),
		Weight:         opnsense.Int64ToString(m.Weight.ValueInt64()),
		Priority:       opnsense.Int64ToString(m.Priority.ValueInt64()),
	}
}

func (m *GatewayResourceModel) fromAPI(_ context.Context, a *gatewayAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(!opnsense.StringToBool(a.Disabled))
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.Interface = types.StringValue(string(a.Interface))
	m.IPProtocol = types.StringValue(string(a.IPProtocol))
	m.Gateway = types.StringValue(a.Gateway)
	m.DefaultGateway = types.BoolValue(opnsense.StringToBool(a.DefaultGateway))
	m.MonitorDisable = types.BoolValue(opnsense.StringToBool(a.MonitorDisable))

	if a.Weight != "" {
		if v, err := opnsense.StringToInt64(a.Weight); err == nil {
			m.Weight = types.Int64Value(v)
		}
	}
	if a.Priority != "" {
		if v, err := opnsense.StringToInt64(a.Priority); err == nil {
			m.Priority = types.Int64Value(v)
		}
	}
}
