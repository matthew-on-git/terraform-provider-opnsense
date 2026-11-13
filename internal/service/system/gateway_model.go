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
	LatencyLow                types.String `tfsdk:"latency_low"`
	LatencyHigh               types.String `tfsdk:"latency_high"`
	LossLow                   types.String `tfsdk:"loss_low"`
	LossHigh                  types.String `tfsdk:"loss_high"`
	Interval                  types.String `tfsdk:"interval"`
	TimePeriod                types.String `tfsdk:"time_period"`
	LossInterval              types.String `tfsdk:"loss_interval"`
	DataLength                types.String `tfsdk:"data_length"`
	Monitor                   types.String `tfsdk:"monitor"`
	FarGateway                types.Bool   `tfsdk:"far_gateway"`
	MonitorNoRoute            types.Bool   `tfsdk:"monitor_no_route"`
	MonitorKillStates         types.Bool   `tfsdk:"monitor_kill_states"`
	MonitorKillStatesPriority types.Bool   `tfsdk:"monitor_kill_states_priority"`
	ForceDown                 types.Bool   `tfsdk:"force_down"`
	NoSync                    types.Bool   `tfsdk:"no_sync"`
	ID                        types.String `tfsdk:"id"`
	Enabled                   types.Bool   `tfsdk:"enabled"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	Interface                 types.String `tfsdk:"interface"`
	IPProtocol                types.String `tfsdk:"ip_protocol"`
	Gateway                   types.String `tfsdk:"gateway"`
	DefaultGateway            types.Bool   `tfsdk:"default_gateway"`
	MonitorDisable            types.Bool   `tfsdk:"monitor_disable"`
	Weight                    types.Int64  `tfsdk:"weight"`
	Priority                  types.Int64  `tfsdk:"priority"`
}

type gatewayAPIResponse struct {
	LatencyLow                *string              `json:"latencylow"`
	LatencyHigh               *string              `json:"latencyhigh"`
	LossLow                   *string              `json:"losslow"`
	LossHigh                  *string              `json:"losshigh"`
	Interval                  *string              `json:"interval"`
	TimePeriod                *string              `json:"time_period"`
	LossInterval              *string              `json:"loss_interval"`
	DataLength                *string              `json:"data_length"`
	Monitor                   *string              `json:"monitor"`
	FarGateway                *gatewayAPIBool      `json:"fargw"`
	MonitorNoRoute            *gatewayAPIBool      `json:"monitor_noroute"`
	MonitorKillStates         *gatewayAPIBool      `json:"monitor_killstates"`
	MonitorKillStatesPriority *gatewayAPIBool      `json:"monitor_killstates_priority"`
	ForceDown                 *gatewayAPIBool      `json:"force_down"`
	NoSync                    *gatewayAPIBool      `json:"nosync"`
	Disabled                  string               `json:"disabled"`
	Name                      string               `json:"name"`
	Description               string               `json:"descr"`
	Interface                 opnsense.SelectedMap `json:"interface"`
	IPProtocol                opnsense.SelectedMap `json:"ipprotocol"`
	Gateway                   string               `json:"gateway"`
	DefaultGateway            string               `json:"defaultgw"`
	MonitorDisable            string               `json:"monitor_disable"`
	Weight                    string               `json:"weight"`
	Priority                  string               `json:"priority"`
}

type gatewayAPIRequest struct {
	LatencyLow                *string `json:"latencylow,omitempty"`
	LatencyHigh               *string `json:"latencyhigh,omitempty"`
	LossLow                   *string `json:"losslow,omitempty"`
	LossHigh                  *string `json:"losshigh,omitempty"`
	Interval                  *string `json:"interval,omitempty"`
	TimePeriod                *string `json:"time_period,omitempty"`
	LossInterval              *string `json:"loss_interval,omitempty"`
	DataLength                *string `json:"data_length,omitempty"`
	Monitor                   *string `json:"monitor,omitempty"`
	FarGateway                *string `json:"fargw,omitempty"`
	MonitorNoRoute            *string `json:"monitor_noroute,omitempty"`
	MonitorKillStates         *string `json:"monitor_killstates,omitempty"`
	MonitorKillStatesPriority *string `json:"monitor_killstates_priority,omitempty"`
	ForceDown                 *string `json:"force_down,omitempty"`
	NoSync                    *string `json:"nosync,omitempty"`
	Disabled                  string  `json:"disabled"`
	Name                      string  `json:"name"`
	Description               string  `json:"descr"`
	Interface                 string  `json:"interface"`
	IPProtocol                string  `json:"ipprotocol"`
	Gateway                   string  `json:"gateway"`
	DefaultGateway            string  `json:"defaultgw"`
	MonitorDisable            string  `json:"monitor_disable"`
	Weight                    string  `json:"weight"`
	Priority                  string  `json:"priority"`
}

func (m *GatewayResourceModel) toAPI(_ context.Context) *gatewayAPIRequest {
	return &gatewayAPIRequest{
		LatencyLow:                gatewayKnownString(m.LatencyLow),
		LatencyHigh:               gatewayKnownString(m.LatencyHigh),
		LossLow:                   gatewayKnownString(m.LossLow),
		LossHigh:                  gatewayKnownString(m.LossHigh),
		Interval:                  gatewayKnownString(m.Interval),
		TimePeriod:                gatewayKnownString(m.TimePeriod),
		LossInterval:              gatewayKnownString(m.LossInterval),
		DataLength:                gatewayKnownString(m.DataLength),
		Monitor:                   gatewayKnownString(m.Monitor),
		FarGateway:                gatewayKnownBool(m.FarGateway),
		MonitorNoRoute:            gatewayKnownBool(m.MonitorNoRoute),
		MonitorKillStates:         gatewayKnownBool(m.MonitorKillStates),
		MonitorKillStatesPriority: gatewayKnownBool(m.MonitorKillStatesPriority),
		ForceDown:                 gatewayKnownBool(m.ForceDown),
		NoSync:                    gatewayKnownBool(m.NoSync),
		Disabled:                  opnsense.BoolToString(!m.Enabled.ValueBool()),
		Name:                      m.Name.ValueString(),
		Description:               m.Description.ValueString(),
		Interface:                 m.Interface.ValueString(),
		IPProtocol:                m.IPProtocol.ValueString(),
		Gateway:                   m.Gateway.ValueString(),
		DefaultGateway:            opnsense.BoolToString(m.DefaultGateway.ValueBool()),
		MonitorDisable:            opnsense.BoolToString(m.MonitorDisable.ValueBool()),
		Weight:                    opnsense.Int64ToString(m.Weight.ValueInt64()),
		Priority:                  opnsense.Int64ToString(m.Priority.ValueInt64()),
	}
}

func (m *GatewayResourceModel) fromAPI(_ context.Context, a *gatewayAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.LatencyLow = gatewayReadString(a.LatencyLow)
	m.LatencyHigh = gatewayReadString(a.LatencyHigh)
	m.LossLow = gatewayReadString(a.LossLow)
	m.LossHigh = gatewayReadString(a.LossHigh)
	m.Interval = gatewayReadString(a.Interval)
	m.TimePeriod = gatewayReadString(a.TimePeriod)
	m.LossInterval = gatewayReadString(a.LossInterval)
	m.DataLength = gatewayReadString(a.DataLength)
	m.Monitor = gatewayReadString(a.Monitor)
	m.FarGateway = gatewayReadBool(a.FarGateway)
	m.MonitorNoRoute = gatewayReadBool(a.MonitorNoRoute)
	m.MonitorKillStates = gatewayReadBool(a.MonitorKillStates)
	m.MonitorKillStatesPriority = gatewayReadBool(a.MonitorKillStatesPriority)
	m.ForceDown = gatewayReadBool(a.ForceDown)
	m.NoSync = gatewayReadBool(a.NoSync)
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
