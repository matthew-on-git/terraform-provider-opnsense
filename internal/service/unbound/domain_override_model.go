// Copyright (c) Matthew Mellor
// SPDX-License-Identifier: MPL-2.0

package unbound

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/matthew-on-git/terraform-provider-opnsense/pkg/opnsense"
)

// DomainOverrideResourceModel is the Terraform state model for opnsense_unbound_domain_override.
type DomainOverrideResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Domain      types.String `tfsdk:"domain"`
	Type        types.String `tfsdk:"type"`
	Server      types.String `tfsdk:"server"`
	ServerPort  types.Int64  `tfsdk:"server_port"`
	VerifyCN    types.String `tfsdk:"verify_cn"`
	Description types.String `tfsdk:"description"`
}

// domainOverrideAPIResponse is the struct for unmarshaling OPNsense GET responses.
type domainOverrideAPIResponse struct {
	Enabled     string               `json:"enabled"`
	Domain      string               `json:"domain"`
	Type        opnsense.SelectedMap `json:"type"`
	Server      string               `json:"server"`
	Port        string               `json:"port"`
	Verify      string               `json:"verify"`
	Description string               `json:"description"`
}

type domainOverrideLegacyAPIResponse struct {
	Enabled     string `json:"enabled"`
	Domain      string `json:"domain"`
	Server      string `json:"server"`
	Port        string `json:"port"`
	Verify      string `json:"verify"`
	Description string `json:"description"`
}

func (response *domainOverrideAPIResponse) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	domain, present := raw["domain"]
	if !present || string(domain) == "null" {
		return fmt.Errorf("missing forward domain")
	}
	var domainValue string
	if err := json.Unmarshal(domain, &domainValue); err != nil {
		return fmt.Errorf("invalid forward domain: %w", err)
	}
	type alias domainOverrideAPIResponse
	var decoded alias
	if value, ok := raw["type"]; !ok {
		return fmt.Errorf("missing forward type")
	} else if err := validateSelectedMap(value, true); err != nil {
		return fmt.Errorf("invalid forward type: %w", err)
	}
	value, ok := raw["port"]
	if !ok {
		return fmt.Errorf("missing forward port")
	}
	var port string
	if err := json.Unmarshal(value, &port); err != nil {
		return fmt.Errorf("invalid forward port: %w", err)
	}
	if port == "" {
		return fmt.Errorf("forward port is empty")
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*response = domainOverrideAPIResponse(decoded)
	return nil
}

// domainOverrideAPIRequest is the struct for marshaling OPNsense POST requests.
type domainOverrideAPIRequest struct {
	Enabled     string `json:"enabled"`
	Type        string `json:"type"`
	Domain      string `json:"domain"`
	Server      string `json:"server"`
	Port        string `json:"port"`
	Verify      string `json:"verify"`
	Description string `json:"description"`
}

// toAPI converts the Terraform model to an API request struct.
func (m *DomainOverrideResourceModel) toAPI(_ context.Context) *domainOverrideAPIRequest {
	return &domainOverrideAPIRequest{
		Enabled:     opnsense.BoolToString(m.Enabled.ValueBool()),
		Type:        domainOverrideTypeToAPI(m.Type.ValueString()),
		Domain:      m.Domain.ValueString(),
		Server:      m.Server.ValueString(),
		Port:        opnsense.Int64ToString(m.ServerPort.ValueInt64()),
		Verify:      m.VerifyCN.ValueString(),
		Description: m.Description.ValueString(),
	}
}

// fromAPI populates the Terraform model from an API response struct.
func (m *DomainOverrideResourceModel) fromAPI(_ context.Context, a *domainOverrideAPIResponse, uuid string) error {
	if uuid == "" {
		return fmt.Errorf("forward UUID is empty")
	}
	if a.Server == "" {
		return fmt.Errorf("forward server is empty")
	}
	forwardType := domainOverrideTypeFromAPI(string(a.Type))
	if forwardType != "query" && forwardType != "dot" {
		return fmt.Errorf("unexpected forward type %q", string(a.Type))
	}
	if a.Enabled != "0" && a.Enabled != "1" {
		return fmt.Errorf("invalid forward enabled value %q", a.Enabled)
	}
	port, err := opnsense.StringToInt64(a.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid forward port %q", a.Port)
	}
	m.ID = types.StringValue(uuid)
	m.Enabled = types.BoolValue(opnsense.StringToBool(a.Enabled))
	m.Domain = types.StringValue(a.Domain)
	m.Type = types.StringValue(forwardType)
	m.Server = types.StringValue(a.Server)
	m.ServerPort = types.Int64Value(port)
	m.VerifyCN = types.StringValue(a.Verify)
	m.Description = types.StringValue(a.Description)
	return nil
}

func domainOverrideTypeToAPI(value string) string {
	if value == "query" {
		return "forward"
	}
	return value
}

func domainOverrideTypeFromAPI(value string) string {
	if value == "forward" {
		return "query"
	}
	return value
}
