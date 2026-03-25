// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package haproxy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// HealthcheckResourceModel is the Terraform state model for opnsense_haproxy_healthcheck.
type HealthcheckResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Interval    types.String `tfsdk:"interval"`
	CheckPort   types.String `tfsdk:"check_port"`
	HTTPMethod  types.String `tfsdk:"http_method"`
	HTTPURI     types.String `tfsdk:"http_uri"`
	HTTPVersion types.String `tfsdk:"http_version"`
	ForceSSL    types.Bool   `tfsdk:"force_ssl"`
}

// healthcheckAPIResponse is the struct for unmarshaling OPNsense GET responses.
type healthcheckAPIResponse struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Type        opnsense.SelectedMap `json:"type"`
	Interval    string               `json:"interval"`
	CheckPort   string               `json:"checkport"`
	HTTPMethod  opnsense.SelectedMap `json:"http_method"`
	HTTPURI     string               `json:"http_uri"`
	HTTPVersion opnsense.SelectedMap `json:"http_version"`
	ForceSSL    string               `json:"force_ssl"`
}

// healthcheckAPIRequest is the struct for marshaling OPNsense POST requests.
type healthcheckAPIRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Interval    string `json:"interval"`
	CheckPort   string `json:"checkport"`
	HTTPMethod  string `json:"http_method"`
	HTTPURI     string `json:"http_uri"`
	HTTPVersion string `json:"http_version"`
	ForceSSL    string `json:"force_ssl"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *HealthcheckResourceModel) toAPI(_ context.Context) *healthcheckAPIRequest {
	return &healthcheckAPIRequest{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Type:        m.Type.ValueString(),
		Interval:    m.Interval.ValueString(),
		CheckPort:   m.CheckPort.ValueString(),
		HTTPMethod:  m.HTTPMethod.ValueString(),
		HTTPURI:     m.HTTPURI.ValueString(),
		HTTPVersion: m.HTTPVersion.ValueString(),
		ForceSSL:    opnsense.BoolToString(m.ForceSSL.ValueBool()),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *HealthcheckResourceModel) fromAPI(_ context.Context, a *healthcheckAPIResponse, uuid string) {
	m.ID = types.StringValue(uuid)
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.Type = types.StringValue(string(a.Type))
	m.Interval = types.StringValue(a.Interval)
	m.CheckPort = types.StringValue(a.CheckPort)
	m.HTTPMethod = types.StringValue(string(a.HTTPMethod))
	m.HTTPURI = types.StringValue(a.HTTPURI)
	m.HTTPVersion = types.StringValue(string(a.HTTPVersion))
	m.ForceSSL = types.BoolValue(opnsense.StringToBool(a.ForceSSL))
}
