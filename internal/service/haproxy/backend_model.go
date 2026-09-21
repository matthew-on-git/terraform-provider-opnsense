// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// BackendResourceModel is the Terraform state model for opnsense_haproxy_backend.
type BackendResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Mode               types.String `tfsdk:"mode"`
	Algorithm          types.String `tfsdk:"algorithm"`
	LinkedServers      types.Set    `tfsdk:"linked_servers"`
	HealthCheck        types.String `tfsdk:"health_check"`
	HealthCheckEnabled types.Bool   `tfsdk:"health_check_enabled"`
	Persistence        types.String `tfsdk:"persistence"`
	ForwardFor         types.Bool   `tfsdk:"forward_for"`
	TimeoutConnect     types.String `tfsdk:"timeout_connect"`
	TimeoutCheck       types.String `tfsdk:"timeout_check"`
	TimeoutServer      types.String `tfsdk:"timeout_server"`
	CustomOptions      types.String `tfsdk:"custom_options"`
}

// backendAPIResponse is the struct for unmarshaling OPNsense GET responses.
type backendAPIResponse struct {
	Enabled            string                   `json:"enabled"`
	Name               string                   `json:"name"`
	Description        string                   `json:"description"`
	Mode               opnsense.SelectedMap     `json:"mode"`
	Algorithm          opnsense.SelectedMap     `json:"algorithm"`
	LinkedServers      opnsense.SelectedMapList `json:"linkedServers"`
	HealthCheck        opnsense.SelectedMap     `json:"healthCheck"`
	HealthCheckEnabled string                   `json:"healthCheckEnabled"`
	Persistence        opnsense.SelectedMap     `json:"persistence"`
	ForwardFor         string                   `json:"forwardFor"`
	TimeoutConnect     string                   `json:"tuning_timeoutConnect"`
	TimeoutCheck       string                   `json:"tuning_timeoutCheck"`
	TimeoutServer      string                   `json:"tuning_timeoutServer"`
	CustomOptions      string                   `json:"customOptions"`
}

// backendAPIRequest is the struct for marshaling OPNsense POST requests.
type backendAPIRequest struct {
	Enabled            string `json:"enabled"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	Mode               string `json:"mode"`
	Algorithm          string `json:"algorithm"`
	LinkedServers      string `json:"linkedServers"`
	HealthCheck        string `json:"healthCheck"`
	HealthCheckEnabled string `json:"healthCheckEnabled"`
	Persistence        string `json:"persistence"`
	ForwardFor         string `json:"forwardFor"`
	TimeoutConnect     string `json:"tuning_timeoutConnect"`
	TimeoutCheck       string `json:"tuning_timeoutCheck"`
	TimeoutServer      string `json:"tuning_timeoutServer"`
	CustomOptions      string `json:"customOptions"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *BackendResourceModel) toAPI(ctx context.Context) *backendAPIRequest {
	var serversStr string
	if !m.LinkedServers.IsNull() && !m.LinkedServers.IsUnknown() {
		var elements []string
		m.LinkedServers.ElementsAs(ctx, &elements, false)
		serversStr = strings.Join(elements, ",")
	}

	return &backendAPIRequest{
		Enabled:            opnsense.BoolToString(m.Enabled.ValueBool()),
		Name:               m.Name.ValueString(),
		Description:        m.Description.ValueString(),
		Mode:               m.Mode.ValueString(),
		Algorithm:          m.Algorithm.ValueString(),
		LinkedServers:      serversStr,
		HealthCheck:        m.HealthCheck.ValueString(),
		HealthCheckEnabled: opnsense.BoolToString(m.HealthCheckEnabled.ValueBool()),
		Persistence:        m.Persistence.ValueString(),
		ForwardFor:         opnsense.BoolToString(m.ForwardFor.ValueBool()),
		TimeoutConnect:     m.TimeoutConnect.ValueString(),
		TimeoutCheck:       m.TimeoutCheck.ValueString(),
		TimeoutServer:      m.TimeoutServer.ValueString(),
		CustomOptions:      m.CustomOptions.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *BackendResourceModel) fromAPI(_ context.Context, a *backendAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.Mode = types.StringValue(string(a.Mode))
	m.Algorithm = types.StringValue(string(a.Algorithm))
	m.HealthCheck = types.StringValue(string(a.HealthCheck))
	m.HealthCheckEnabled = types.BoolValue(opnsense.StringToBool(a.HealthCheckEnabled))
	m.Persistence = types.StringValue(string(a.Persistence))
	m.ForwardFor = types.BoolValue(opnsense.StringToBool(a.ForwardFor))
	m.TimeoutConnect = types.StringValue(a.TimeoutConnect)
	m.TimeoutCheck = types.StringValue(a.TimeoutCheck)
	m.TimeoutServer = types.StringValue(a.TimeoutServer)
	m.CustomOptions = types.StringValue(a.CustomOptions)

	// LinkedServers — SelectedMapList → types.Set of UUID strings.
	if len(a.LinkedServers) == 0 {
		m.LinkedServers = types.SetValueMust(types.StringType, []attr.Value{})
	} else {
		vals := make([]attr.Value, len(a.LinkedServers))
		for i, v := range a.LinkedServers {
			vals[i] = types.StringValue(v)
		}
		m.LinkedServers = types.SetValueMust(types.StringType, vals)
	}
}
